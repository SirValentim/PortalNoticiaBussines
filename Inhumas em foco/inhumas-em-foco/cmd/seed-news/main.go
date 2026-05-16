package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"inhumas-em-foco/internal/repository"
)

func main() {
	dbPath := os.Getenv("DATABASE_URL")
	if dbPath == "" {
		dbPath = "inhumas.db"
	}
	projectRoot := os.Getenv("PROJECT_ROOT")
	if projectRoot == "" {
		wd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "erro ao resolver diretorio atual: %v\n", err)
			os.Exit(1)
		}
		projectRoot = wd
	}
	sqlPath := filepath.Join(projectRoot, "scripts", "seed-local-news-drafts.sql")
	script, err := os.ReadFile(sqlPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "erro ao ler %s: %v\n", sqlPath, err)
		os.Exit(1)
	}
	repo, err := repository.Open(os.Getenv("DB_DRIVER"), dbPath, filepath.Join(projectRoot, "migrations"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "erro ao abrir banco: %v\n", err)
		os.Exit(1)
	}
	defer repo.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if _, err := repo.DB().ExecContext(ctx, string(script)); err != nil {
		fmt.Fprintf(os.Stderr, "erro ao aplicar rascunhos: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Rascunhos locais de noticias aplicados com sucesso.")
}
