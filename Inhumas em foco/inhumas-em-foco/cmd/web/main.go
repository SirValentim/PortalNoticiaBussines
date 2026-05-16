package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"inhumas-em-foco/internal/auth"
	"inhumas-em-foco/internal/config"
	"inhumas-em-foco/internal/handler"
	"inhumas-em-foco/internal/middleware"
	"inhumas-em-foco/internal/model"
	"inhumas-em-foco/internal/repository"
	"inhumas-em-foco/internal/session"
	"inhumas-em-foco/internal/storage"
)

func main() {
	cfg := config.Load()

	if !cfg.IsValidSessionSecret() {
		log.Fatal("SESSION_SECRET deve ter pelo menos 32 caracteres")
	}
	if !cfg.IsValidPreviousSessionSecret() {
		log.Fatal("PREVIOUS_SESSION_SECRET deve estar vazio ou ter pelo menos 32 caracteres")
	}

	// Setup database
	repo, err := repository.Open(cfg.DBDriver, cfg.DatabaseURL, cfg.MigrationsDir)
	if err != nil {
		log.Fatal("Erro ao conectar ao banco: ", err)
	}
	defer repo.Close()

	// Setup session manager
	secure := !cfg.IsLocal()
	sessionMgr := session.NewManagerWithPrevious(cfg.SessionSecret, cfg.PreviousSessionSecret, secure)

	// Setup auth service
	authSvc := auth.NewService(repo)

	// Setup storage
	storageProvider := storage.NewLocalProvider(cfg.UploadDir, "")

	// Setup handler
	h, err := handler.New(repo, cfg, sessionMgr, authSvc, storageProvider)
	if err != nil {
		log.Fatal("Erro ao inicializar handler: ", err)
	}

	// Setup routes
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	// Chain middleware
	var root http.Handler = mux
	root = middleware.Recovery(root)
	root = middleware.SecurityHeadersWithConfig(cfg)(root)
	root = middleware.RequestTimeout(30 * time.Second)(root)
	root = middleware.RateLimitByIPWhen(10, time.Minute, func(r *http.Request) bool {
		return r.URL.Path == "/login" && r.Method == http.MethodPost
	})(root)
	root = middleware.RateLimitByIPWhen(120, time.Minute, func(r *http.Request) bool {
		return r.Method != http.MethodGet && r.Method != http.MethodHead && r.Method != http.MethodOptions
	})(root)
	root = middleware.MaintenanceMode(cfg)(root)
	root = middleware.RequireAdmin(cfg.AdminPathPrefix)(root)
	root = middleware.AuthMiddleware(sessionMgr, repo, authSvc)(root)
	root = middleware.CSRFProtection(cfg.AdminPathPrefix, secure, cfg.MaxUploadSize)(root)
	root = middleware.MetricsMiddleware(repo)(root)
	root = middleware.StructuredLogger(root)

	// Create server
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      root,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Seed default admin user if none exists
	go seedDefaultAdmin(repo, authSvc, cfg.DefaultBcryptCost)

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		fmt.Println("\nShutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		server.Shutdown(ctx)
	}()

	fmt.Printf("Server starting on http://localhost:%s\n", cfg.Port)
	fmt.Printf("Admin panel: http://localhost:%s%s\n", cfg.Port, cfg.AdminPathPrefix)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("Server error: ", err)
	}
}

func seedDefaultAdmin(repo *repository.Repository, authSvc *auth.Service, cost int) {
	time.Sleep(1 * time.Second)
	ctx := context.Background()

	// Check if any user exists
	users, err := repo.UserList(ctx)
	if err != nil || len(users) > 0 {
		return
	}

	initialPassword := os.Getenv("INITIAL_ADMIN_PASSWORD")
	if initialPassword == "" {
		log.Println("INITIAL_ADMIN_PASSWORD nao definido; admin padrao nao sera criado automaticamente")
		return
	}

	hash, err := authSvc.HashPassword(initialPassword, cost)
	if err != nil {
		log.Println("Erro ao criar senha padrao:", err)
		return
	}

	admin := &model.User{
		Name:         "Administrador",
		Email:        "admin@inhumasemfoco.com.br",
		PasswordHash: hash,
		Role:         "admin",
		Active:       true,
	}

	if err := repo.UserCreate(ctx, admin); err != nil {
		log.Println("Erro ao criar admin padrao:", err)
		return
	}

	fmt.Println("Default admin created: admin@inhumasemfoco.com.br")
	fmt.Println("IMPORTANTE: Altere a senha padrao apos o primeiro acesso!")
}
