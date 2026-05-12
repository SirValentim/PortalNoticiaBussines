package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"inhumas-em-foco/internal/auth"
	"inhumas-em-foco/internal/config"
	"inhumas-em-foco/internal/middleware"
	"inhumas-em-foco/internal/model"
	"inhumas-em-foco/internal/repository"
	"inhumas-em-foco/internal/session"
	"inhumas-em-foco/internal/storage"
	usersvc "inhumas-em-foco/internal/users"
)

func newTestHandler(t *testing.T) (*Handler, *repository.Repository) {
	t.Helper()
	repo, err := repository.New(":memory:")
	if err != nil {
		t.Fatalf("repository.New failed: %v", err)
	}
	cfg := &config.Config{
		Port:              "8080",
		DatabaseURL:       ":memory:",
		SessionSecret:     "12345678901234567890123456789012",
		AdminPathPrefix:   "/painel/7x9k2m",
		UploadDir:         t.TempDir(),
		StaticDir:         t.TempDir(),
		SiteURL:           "https://example.com",
		ProjectRoot:       "../..",
		MaxUploadSize:     2 * 1024 * 1024,
		DefaultBcryptCost: 4,
	}
	sessionMgr := session.NewManager(cfg.SessionSecret, false)
	authSvc := auth.NewService(repo)
	storageProvider := storage.NewLocalProvider(cfg.UploadDir, "")

	h, err := New(repo, cfg, sessionMgr, authSvc, storageProvider)
	if err != nil {
		repo.Close()
		t.Fatalf("handler.New failed: %v", err)
	}
	return h, repo
}

func TestHealthOK(t *testing.T) {
	h, repo := newTestHandler(t)
	defer repo.Close()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	h.Health(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, `"database":"ok"`) || !strings.Contains(body, `"uploads":"ok"`) {
		t.Fatalf("health body missing ok checks: %s", body)
	}
}

func TestLoginPageRendersCSRFToken(t *testing.T) {
	h, repo := newTestHandler(t)
	defer repo.Close()

	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	rec := httptest.NewRecorder()

	middleware.CSRFProtection(h.cfg.AdminPathPrefix, false)(http.HandlerFunc(h.LoginPage)).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), `name="csrf_token" value="`) {
		t.Fatalf("login page missing csrf token: %s", rec.Body.String())
	}
}

func TestHomeRendersEmptyStateWithoutError(t *testing.T) {
	h, repo := newTestHandler(t)
	defer repo.Close()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	h.Home(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "Inhumas") {
		t.Fatalf("home did not render expected shell: %s", rec.Body.String())
	}
}

func TestPostDetailSanitizesStoredHTML(t *testing.T) {
	h, repo := newTestHandler(t)
	defer repo.Close()

	now := time.Now()
	category, err := repo.CategoryGetBySlug(context.Background(), "noticias")
	if err != nil || category == nil {
		t.Fatalf("CategoryGetBySlug failed: %v", err)
	}
	post := &model.Post{
		Title:       "Post Seguro",
		Slug:        "post-seguro",
		Excerpt:     "Resumo",
		Content:     `<p>Texto</p><script>alert("x")</script><a href="javascript:alert(1)">link</a>`,
		CategoryID:  &category.ID,
		Status:      model.StatusPublished,
		PublishedAt: &now,
	}
	if err := repo.PostCreate(context.Background(), post); err != nil {
		t.Fatalf("PostCreate failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/noticia/post-seguro", nil)
	req.SetPathValue("slug", "post-seguro")
	rec := httptest.NewRecorder()

	h.PostDetail(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	body := rec.Body.String()
	if strings.Contains(body, `alert("x")`) || strings.Contains(body, "javascript:") {
		t.Fatalf("post detail rendered unsafe HTML: %s", body)
	}
	if !strings.Contains(body, "<p>Texto</p>") {
		t.Fatalf("post detail missing allowed content: %s", body)
	}
	if !strings.Contains(body, `"@type":"NewsArticle"`) || !strings.Contains(body, `"headline":"Post Seguro"`) {
		t.Fatalf("post detail missing article JSON-LD: %s", body)
	}
	if !strings.Contains(body, `"@type":"BreadcrumbList"`) {
		t.Fatalf("post detail missing breadcrumb JSON-LD: %s", body)
	}
}

func TestRSSRendersPublishedPosts(t *testing.T) {
	h, repo := newTestHandler(t)
	defer repo.Close()

	now := time.Date(2026, 4, 27, 10, 0, 0, 0, time.UTC)
	category, err := repo.CategoryGetBySlug(context.Background(), "noticias")
	if err != nil || category == nil {
		t.Fatalf("CategoryGetBySlug failed: %v", err)
	}
	published := &model.Post{
		Title:       "Noticia no Feed",
		Slug:        "noticia-no-feed",
		Excerpt:     "Resumo para leitores RSS",
		Content:     "Conteudo",
		CategoryID:  &category.ID,
		Status:      model.StatusPublished,
		PublishedAt: &now,
	}
	if err := repo.PostCreate(context.Background(), published); err != nil {
		t.Fatalf("PostCreate published failed: %v", err)
	}
	draft := &model.Post{
		Title:      "Rascunho Fora do Feed",
		Slug:       "rascunho-fora-do-feed",
		Excerpt:    "Nao deve aparecer",
		Content:    "Conteudo",
		CategoryID: &category.ID,
		Status:     model.StatusDraft,
	}
	if err := repo.PostCreate(context.Background(), draft); err != nil {
		t.Fatalf("PostCreate draft failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/rss.xml", nil)
	rec := httptest.NewRecorder()

	h.RSS(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if got := rec.Header().Get("Content-Type"); got != "application/rss+xml; charset=utf-8" {
		t.Fatalf("content-type = %q", got)
	}
	body := rec.Body.String()
	for _, want := range []string{
		"<rss",
		"<channel>",
		"<title>Inhumas em Foco</title>",
		"<title>Noticia no Feed</title>",
		"<link>https://example.com/noticia/noticia-no-feed</link>",
		"<description>Resumo para leitores RSS</description>",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("RSS missing %q: %s", want, body)
		}
	}
	if strings.Contains(body, "Rascunho Fora do Feed") {
		t.Fatalf("RSS exposed draft post: %s", body)
	}
}

func TestRobotsTxtAllowsPublicAndPointsToSitemap(t *testing.T) {
	h, repo := newTestHandler(t)
	defer repo.Close()

	req := httptest.NewRequest(http.MethodGet, "/robots.txt", nil)
	rec := httptest.NewRecorder()

	h.Robots(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	for _, want := range []string{
		"User-agent: *",
		"Allow: /",
		"Disallow: /painel/7x9k2m/",
		"Sitemap: https://example.com/sitemap.xml",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("robots.txt missing %q: %s", want, body)
		}
	}
}

func TestSitemapRootServesGeneratedSitemap(t *testing.T) {
	h, repo := newTestHandler(t)
	defer repo.Close()

	content := []byte(`<?xml version="1.0" encoding="UTF-8"?><urlset><url><loc>https://example.com/noticia/teste</loc></url></urlset>`)
	if err := os.WriteFile(filepath.Join(h.cfg.StaticDir, "sitemap.xml"), content, 0644); err != nil {
		t.Fatalf("write sitemap: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/sitemap.xml", nil)
	rec := httptest.NewRecorder()

	h.Sitemap(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("Content-Type"); !strings.Contains(got, "application/xml") {
		t.Fatalf("content-type = %q", got)
	}
	if !strings.Contains(rec.Body.String(), "https://example.com/noticia/teste") {
		t.Fatalf("sitemap body missing URL: %s", rec.Body.String())
	}
}

func TestSitemapRootBuildsFallbackWhenFileMissing(t *testing.T) {
	h, repo := newTestHandler(t)
	defer repo.Close()

	req := httptest.NewRequest(http.MethodGet, "/sitemap.xml", nil)
	rec := httptest.NewRecorder()

	h.Sitemap(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("Content-Type"); !strings.Contains(got, "application/xml") {
		t.Fatalf("content-type = %q", got)
	}
	if !strings.Contains(rec.Body.String(), "https://example.com/noticias") {
		t.Fatalf("sitemap fallback missing public URL: %s", rec.Body.String())
	}
}

func TestAdminMetricsRendersAggregatedMetrics(t *testing.T) {
	h, repo := newTestHandler(t)
	defer repo.Close()

	ctx := context.Background()
	metrics := []model.Metric{
		{MetricType: "post_view", EntityType: "post", EntityID: 1},
		{MetricType: "post_view", EntityType: "post", EntityID: 1},
		{MetricType: "banner_click", EntityType: "banner", EntityID: 3},
	}
	for i := range metrics {
		if err := repo.MetricTrack(ctx, &metrics[i]); err != nil {
			t.Fatalf("MetricTrack failed: %v", err)
		}
	}

	req := httptest.NewRequest(http.MethodGet, h.cfg.AdminPathPrefix+"/metrics", nil)
	req = req.WithContext(auth.WithUser(req.Context(), &model.User{ID: 1, Role: model.RoleAdmin}))
	rec := httptest.NewRecorder()

	h.AdminMetrics(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	body := rec.Body.String()
	for _, want := range []string{"Metricas", "Visualizacoes de noticias", "Banners com mais cliques", "#1", "#3"} {
		if !strings.Contains(body, want) {
			t.Fatalf("metrics page missing %q: %s", want, body)
		}
	}
}

func TestOperationalMetricsRequiresTokenAndReturnsRuntimeData(t *testing.T) {
	h, repo := newTestHandler(t)
	defer repo.Close()
	h.cfg.MetricsToken = "metricas-secretas"

	unauthorizedReq := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	unauthorizedRec := httptest.NewRecorder()
	h.OperationalMetrics(unauthorizedRec, unauthorizedReq)
	if unauthorizedRec.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized status = %d, want %d", unauthorizedRec.Code, http.StatusUnauthorized)
	}

	ctx := context.Background()
	if err := repo.JobCreate(ctx, &model.Job{
		Type:        model.JobGenerateSitemap,
		Payload:     "{}",
		Status:      model.JobPending,
		RunAt:       time.Now(),
		MaxAttempts: 3,
	}); err != nil {
		t.Fatalf("JobCreate failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.Header.Set("Authorization", "Bearer metricas-secretas")
	rec := httptest.NewRecorder()

	h.OperationalMetrics(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	body := rec.Body.String()
	for _, want := range []string{`"requests_total"`, `"open_connections"`, `"pending":1`, `"upload_bytes"`} {
		if !strings.Contains(body, want) {
			t.Fatalf("operational metrics missing %q: %s", want, body)
		}
	}
}

func TestAdminDeadJobsRendersReport(t *testing.T) {
	h, repo := newTestHandler(t)
	defer repo.Close()

	ctx := context.Background()
	job := &model.Job{
		Type:        model.JobCleanupOldJobs,
		Payload:     "{}",
		Status:      model.JobPending,
		RunAt:       time.Now().Add(-time.Hour),
		MaxAttempts: 1,
	}
	if err := repo.JobCreate(ctx, job); err != nil {
		t.Fatalf("JobCreate failed: %v", err)
	}
	if _, err := repo.JobRecordFailure(ctx, model.Job{ID: job.ID, Attempts: 0, MaxAttempts: 1}, "erro de teste"); err != nil {
		t.Fatalf("JobRecordFailure failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, h.cfg.AdminPathPrefix+"/dead-jobs", nil)
	req = req.WithContext(auth.WithUser(req.Context(), &model.User{ID: 1, Role: model.RoleAdmin}))
	rec := httptest.NewRecorder()

	h.AdminDeadJobs(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	body := rec.Body.String()
	for _, want := range []string{"Jobs com Falha", "cleanup_old_jobs", "erro de teste", "1 registros"} {
		if !strings.Contains(body, want) {
			t.Fatalf("dead jobs page missing %q: %s", want, body)
		}
	}
}

func TestAdminAuditLogsRendersWithFilters(t *testing.T) {
	h, repo := newTestHandler(t)
	defer repo.Close()

	ctx := context.Background()
	hash, err := h.authSvc.HashPassword("senha-forte", h.cfg.DefaultBcryptCost)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
	admin := &model.User{
		Name:         "Admin",
		Email:        "admin@example.com",
		PasswordHash: hash,
		Role:         model.RoleAdmin,
		Active:       true,
	}
	if err := repo.UserCreate(ctx, admin); err != nil {
		t.Fatalf("UserCreate failed: %v", err)
	}
	targetID := int64(77)
	if err := repo.AuditLogCreate(ctx, &model.AuditLog{
		UserID:     &admin.ID,
		Action:     "update",
		EntityType: "user",
		EntityID:   &targetID,
		Changes:    `{"email":"novo@example.com"}`,
		IPAddress:  "127.0.0.1",
		UserAgent:  "test",
	}); err != nil {
		t.Fatalf("AuditLogCreate failed: %v", err)
	}
	if err := repo.AuditLogCreate(ctx, &model.AuditLog{
		UserID:     &admin.ID,
		Action:     "delete",
		EntityType: "banner",
		EntityID:   &targetID,
		Changes:    `{"name":"Banner"}`,
		IPAddress:  "127.0.0.1",
		UserAgent:  "test",
	}); err != nil {
		t.Fatalf("AuditLogCreate second failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, h.cfg.AdminPathPrefix+"/audit?action=update&entity_type=user", nil)
	req = req.WithContext(auth.WithUser(req.Context(), admin))
	rec := httptest.NewRecorder()

	h.AdminAuditLogs(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	body := rec.Body.String()
	for _, want := range []string{"Auditoria", "admin@example.com", "novo@example.com", "1 registros"} {
		if !strings.Contains(body, want) {
			t.Fatalf("audit page missing %q: %s", want, body)
		}
	}
	if strings.Contains(body, "banner #77") {
		t.Fatalf("audit filter leaked non-matching row: %s", body)
	}
}

func TestAdminAuditLogsRequiresSettingsPermission(t *testing.T) {
	h, repo := newTestHandler(t)
	defer repo.Close()

	req := httptest.NewRequest(http.MethodGet, h.cfg.AdminPathPrefix+"/audit", nil)
	req = req.WithContext(auth.WithUser(req.Context(), &model.User{ID: 1, Role: model.RoleComercial}))
	rec := httptest.NewRecorder()

	h.AdminAuditLogs(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestPasswordResetRequestUsesGenericResponse(t *testing.T) {
	h, repo := newTestHandler(t)
	defer repo.Close()

	form := url.Values{}
	form.Set("email", "desconhecido@example.com")
	req := httptest.NewRequest(http.MethodPost, "/recuperar-senha", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	h.PasswordResetRequestPost(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "Se o e-mail estiver cadastrado") {
		t.Fatalf("missing generic reset response: %s", rec.Body.String())
	}
}

func TestPasswordResetPostUpdatesPasswordAndConsumesToken(t *testing.T) {
	h, repo := newTestHandler(t)
	defer repo.Close()

	ctx := context.Background()
	oldHash, err := h.authSvc.HashPassword("senha-antiga", h.cfg.DefaultBcryptCost)
	if err != nil {
		t.Fatalf("HashPassword old failed: %v", err)
	}
	user := &model.User{
		Name:         "Editor",
		Email:        "editor@example.com",
		PasswordHash: oldHash,
		Role:         model.RoleEditor,
		Active:       true,
	}
	if err := repo.UserCreate(ctx, user); err != nil {
		t.Fatalf("UserCreate failed: %v", err)
	}
	plainToken := "token-seguro-de-teste"
	reset := &model.PasswordResetToken{
		UserID:      user.ID,
		TokenHash:   usersvc.PasswordResetHash(plainToken),
		RequestedIP: "127.0.0.1",
		UserAgent:   "test",
		ExpiresAt:   time.Now().Add(30 * time.Minute),
	}
	if err := repo.PasswordResetCreate(ctx, reset); err != nil {
		t.Fatalf("PasswordResetCreate failed: %v", err)
	}

	form := url.Values{}
	form.Set("password", "senha-nova")
	form.Set("password_confirm", "senha-nova")
	req := httptest.NewRequest(http.MethodPost, "/redefinir-senha/"+plainToken, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetPathValue("token", plainToken)
	rec := httptest.NewRecorder()

	h.PasswordResetPost(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "Senha atualizada com sucesso") {
		t.Fatalf("missing reset success: %s", rec.Body.String())
	}
	if _, err := h.authSvc.Authenticate(ctx, "editor@example.com", "senha-nova", h.cfg.DefaultBcryptCost); err != nil {
		t.Fatalf("Authenticate with new password failed: %v", err)
	}
	if active, err := repo.PasswordResetGetActive(ctx, usersvc.PasswordResetHash(plainToken), time.Now()); err != nil || active != nil {
		t.Fatalf("token should be consumed, got token=%#v err=%v", active, err)
	}

	reuseReq := httptest.NewRequest(http.MethodPost, "/redefinir-senha/"+plainToken, strings.NewReader(form.Encode()))
	reuseReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reuseReq.SetPathValue("token", plainToken)
	reuseRec := httptest.NewRecorder()
	h.PasswordResetPost(reuseRec, reuseReq)
	if reuseRec.Code != http.StatusOK || !strings.Contains(reuseRec.Body.String(), "Link invalido ou expirado") {
		t.Fatalf("reused token should be rejected: status=%d body=%s", reuseRec.Code, reuseRec.Body.String())
	}
}

func TestAdminUserUpdatePassword(t *testing.T) {
	h, repo := newTestHandler(t)
	defer repo.Close()

	oldHash, err := h.authSvc.HashPassword("senha-antiga", h.cfg.DefaultBcryptCost)
	if err != nil {
		t.Fatalf("HashPassword old failed: %v", err)
	}
	user := &model.User{
		Name:         "Admin",
		Email:        "admin@example.com",
		PasswordHash: oldHash,
		Role:         model.RoleAdmin,
		Active:       true,
	}
	if err := repo.UserCreate(context.Background(), user); err != nil {
		t.Fatalf("UserCreate failed: %v", err)
	}

	form := url.Values{}
	form.Set("password", "senha-nova")
	form.Set("password_confirm", "senha-nova")
	req := httptest.NewRequest(http.MethodPost, h.cfg.AdminPathPrefix+"/users/1/password", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetPathValue("id", "1")
	req = req.WithContext(auth.WithUser(req.Context(), user))
	rec := httptest.NewRecorder()

	h.AdminUserUpdatePassword(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusSeeOther, rec.Body.String())
	}
	if _, err := h.authSvc.Authenticate(context.Background(), "admin@example.com", "senha-nova", h.cfg.DefaultBcryptCost); err != nil {
		t.Fatalf("Authenticate with new password failed: %v", err)
	}
}

func TestAdminUserEditAndUpdate(t *testing.T) {
	h, repo := newTestHandler(t)
	defer repo.Close()

	ctx := context.Background()
	hash, err := h.authSvc.HashPassword("senha-forte", h.cfg.DefaultBcryptCost)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
	admin := &model.User{
		Name:         "Admin",
		Email:        "admin@example.com",
		PasswordHash: hash,
		Role:         model.RoleAdmin,
		Active:       true,
	}
	if err := repo.UserCreate(ctx, admin); err != nil {
		t.Fatalf("UserCreate admin failed: %v", err)
	}
	user := &model.User{
		Name:         "Redator",
		Email:        "redator@example.com",
		PasswordHash: hash,
		Role:         model.RoleRedator,
		Active:       true,
	}
	if err := repo.UserCreate(ctx, user); err != nil {
		t.Fatalf("UserCreate target failed: %v", err)
	}

	editReq := httptest.NewRequest(http.MethodGet, h.cfg.AdminPathPrefix+"/users/2/edit", nil)
	editReq.SetPathValue("id", strconv.FormatInt(user.ID, 10))
	editReq = editReq.WithContext(auth.WithUser(editReq.Context(), admin))
	editRec := httptest.NewRecorder()
	h.AdminUserEdit(editRec, editReq)
	if editRec.Code != http.StatusOK {
		t.Fatalf("edit status = %d, want %d; body=%s", editRec.Code, http.StatusOK, editRec.Body.String())
	}
	if !strings.Contains(editRec.Body.String(), "Editar Usuario") || !strings.Contains(editRec.Body.String(), "redator@example.com") {
		t.Fatalf("edit page missing user data: %s", editRec.Body.String())
	}

	form := url.Values{}
	form.Set("name", "Revisor Atualizado")
	form.Set("email", "revisor@example.com")
	form.Set("role", string(model.RoleRevisor))
	form.Set("active", "on")
	updateReq := httptest.NewRequest(http.MethodPost, h.cfg.AdminPathPrefix+"/users/2/edit", strings.NewReader(form.Encode()))
	updateReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	updateReq.SetPathValue("id", strconv.FormatInt(user.ID, 10))
	updateReq = updateReq.WithContext(auth.WithUser(updateReq.Context(), admin))
	updateRec := httptest.NewRecorder()

	h.AdminUserUpdate(updateRec, updateReq)

	if updateRec.Code != http.StatusSeeOther {
		t.Fatalf("update status = %d, want %d; body=%s", updateRec.Code, http.StatusSeeOther, updateRec.Body.String())
	}
	updated, err := repo.UserGetByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("UserGetByID failed: %v", err)
	}
	if updated.Name != "Revisor Atualizado" || updated.Email != "revisor@example.com" || updated.Role != model.RoleRevisor || !updated.Active {
		t.Fatalf("updated user mismatch: %+v", updated)
	}
	logs, err := repo.AuditLogList(ctx, "user", user.ID, 10)
	if err != nil {
		t.Fatalf("AuditLogList failed: %v", err)
	}
	if len(logs) == 0 || logs[0].Action != "update" || !strings.Contains(logs[0].Changes, "revisor@example.com") {
		t.Fatalf("missing user update audit log: %#v", logs)
	}
}

func TestAdminPostEditShowsActiveLockWarning(t *testing.T) {
	h, repo := newTestHandler(t)
	defer repo.Close()

	ctx := context.Background()
	category, err := repo.CategoryGetBySlug(ctx, "noticias")
	if err != nil || category == nil {
		t.Fatalf("CategoryGetBySlug failed: %v", err)
	}
	authorID := int64(1)
	post := &model.Post{
		Title:      "Post em edicao",
		Slug:       "post-em-edicao",
		Content:    "conteudo",
		CategoryID: &category.ID,
		AuthorID:   &authorID,
		Status:     model.StatusDraft,
	}
	if err := repo.PostCreate(ctx, post); err != nil {
		t.Fatalf("PostCreate failed: %v", err)
	}
	if err := repo.EditLockCreate(ctx, &model.EditLock{
		EntityType: "post",
		EntityID:   post.ID,
		UserID:     99,
		LockedAt:   time.Now(),
		ExpiresAt:  time.Now().Add(time.Minute),
	}); err != nil {
		t.Fatalf("EditLockCreate failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, h.cfg.AdminPathPrefix+"/posts/1/edit", nil)
	req.SetPathValue("id", "1")
	req = req.WithContext(auth.WithUser(req.Context(), &model.User{ID: 2, Role: model.RoleEditor}))
	rec := httptest.NewRecorder()

	h.AdminPostEdit(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "editado por outro usuario") {
		t.Fatalf("missing lock warning: %s", rec.Body.String())
	}
}

func TestAdminPostLockHeartbeatCreatesLock(t *testing.T) {
	h, repo := newTestHandler(t)
	defer repo.Close()

	ctx := context.Background()
	category, err := repo.CategoryGetBySlug(ctx, "noticias")
	if err != nil || category == nil {
		t.Fatalf("CategoryGetBySlug failed: %v", err)
	}
	post := &model.Post{
		Title:      "Post heartbeat",
		Slug:       "post-heartbeat",
		Content:    "conteudo",
		CategoryID: &category.ID,
		Status:     model.StatusDraft,
	}
	if err := repo.PostCreate(ctx, post); err != nil {
		t.Fatalf("PostCreate failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, h.cfg.AdminPathPrefix+"/posts/1/lock", nil)
	req.SetPathValue("id", "1")
	req = req.WithContext(auth.WithUser(req.Context(), &model.User{ID: 7, Role: model.RoleEditor}))
	rec := httptest.NewRecorder()

	h.AdminPostLockHeartbeat(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	lock, err := repo.EditLockGet(ctx, "post", post.ID)
	if err != nil {
		t.Fatalf("EditLockGet failed: %v", err)
	}
	if lock == nil || lock.UserID != 7 || !lock.ExpiresAt.After(time.Now()) {
		t.Fatalf("unexpected lock: %#v", lock)
	}
}

func TestAdminPostCreateScheduledPersistsPublishAtAndJob(t *testing.T) {
	h, repo := newTestHandler(t)
	defer repo.Close()

	ctx := context.Background()
	category, err := repo.CategoryGetBySlug(ctx, "noticias")
	if err != nil || category == nil {
		t.Fatalf("CategoryGetBySlug failed: %v", err)
	}

	form := url.Values{}
	form.Set("title", "Post agendado")
	form.Set("category_id", strconv.FormatInt(category.ID, 10))
	form.Set("excerpt", "Resumo")
	form.Set("content", "Conteudo")
	form.Set("meta_description", "Descricao de busca para post agendado")
	form.Set("status", "scheduled")
	form.Set("publish_at", "2026-05-01T08:30")
	req := httptest.NewRequest(http.MethodPost, h.cfg.AdminPathPrefix+"/posts", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(auth.WithUser(req.Context(), &model.User{ID: 1, Role: model.RoleEditor}))
	rec := httptest.NewRecorder()

	h.AdminPostCreate(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusSeeOther, rec.Body.String())
	}
	posts, err := repo.PostListAll(ctx, 10, 0)
	if err != nil {
		t.Fatalf("PostListAll failed: %v", err)
	}
	if len(posts) != 1 || posts[0].PublishAt == nil {
		t.Fatalf("scheduled post missing publish_at: %#v", posts)
	}
	if got := posts[0].PublishAt.Format("2006-01-02T15:04"); got != "2026-05-01T08:30" {
		t.Fatalf("publish_at = %s", got)
	}
	hasJob, err := repo.JobHasActiveType(ctx, model.JobPublishPost)
	if err != nil {
		t.Fatalf("JobHasActiveType failed: %v", err)
	}
	if !hasJob {
		t.Fatal("expected publish_post job")
	}
}

func TestAdminPostCreatePersistsSEOFields(t *testing.T) {
	h, repo := newTestHandler(t)
	defer repo.Close()

	ctx := context.Background()
	category, err := repo.CategoryGetBySlug(ctx, "noticias")
	if err != nil || category == nil {
		t.Fatalf("CategoryGetBySlug failed: %v", err)
	}

	form := url.Values{}
	form.Set("title", "SEO local")
	form.Set("category_id", strconv.FormatInt(category.ID, 10))
	form.Set("excerpt", "Resumo editorial")
	form.Set("content", "Conteudo")
	form.Set("status", "draft")
	form.Set("meta_title", "SEO para Inhumas GO")
	form.Set("meta_description", "Descricao preparada para buscas locais em Inhumas GO.")
	form.Set("seo_keyword", "noticias de Inhumas")
	form.Set("canonical_url", "https://example.com/noticia/seo-local")
	form.Set("source_name", "Fonte Oficial")
	form.Set("source_url", "https://example.com/fonte")
	req := httptest.NewRequest(http.MethodPost, h.cfg.AdminPathPrefix+"/posts", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(auth.WithUser(req.Context(), &model.User{ID: 1, Role: model.RoleEditor}))
	rec := httptest.NewRecorder()

	h.AdminPostCreate(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusSeeOther, rec.Body.String())
	}
	posts, err := repo.PostListAll(ctx, 10, 0)
	if err != nil {
		t.Fatalf("PostListAll failed: %v", err)
	}
	post, err := repo.PostGetByID(ctx, posts[0].ID)
	if err != nil {
		t.Fatalf("PostGetByID failed: %v", err)
	}
	if post.MetaTitle != "SEO para Inhumas GO" || post.MetaDescription == "" || post.SEOKeyword != "noticias de Inhumas" || post.CanonicalURL == "" || post.SourceName != "Fonte Oficial" || post.ReadingTimeMinutes < 1 {
		t.Fatalf("SEO fields not persisted: %#v", post)
	}
	revisions, err := repo.PostRevisionList(ctx, post.ID, 10)
	if err != nil || len(revisions) == 0 {
		t.Fatalf("expected create revision, revisions=%#v err=%v", revisions, err)
	}
}

func TestAdminPostPreviewRendersDraftWithNoIndex(t *testing.T) {
	h, repo := newTestHandler(t)
	defer repo.Close()

	ctx := context.Background()
	category, err := repo.CategoryGetBySlug(ctx, "noticias")
	if err != nil || category == nil {
		t.Fatalf("CategoryGetBySlug failed: %v", err)
	}
	post := &model.Post{
		Title:           "Preview editorial",
		Slug:            "preview-editorial",
		Excerpt:         "Resumo",
		Content:         "<p>Conteudo em revisao</p>",
		MetaTitle:       "Titulo SEO preview",
		MetaDescription: "Descricao SEO preview",
		CategoryID:      &category.ID,
		Status:          model.StatusDraft,
	}
	if err := repo.PostCreate(ctx, post); err != nil {
		t.Fatalf("PostCreate failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, h.cfg.AdminPathPrefix+"/posts/1/preview", nil)
	req.SetPathValue("id", strconv.FormatInt(post.ID, 10))
	req = req.WithContext(auth.WithUser(req.Context(), &model.User{ID: 1, Role: model.RoleEditor}))
	rec := httptest.NewRecorder()

	h.AdminPostPreview(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	body := rec.Body.String()
	for _, want := range []string{"Preview editorial", "Conteudo em revisao", "noindex"} {
		if !strings.Contains(body, want) {
			t.Fatalf("preview missing %q: %s", want, body)
		}
	}
}

func TestInfluencerPagesRender(t *testing.T) {
	h, repo := newTestHandler(t)
	defer repo.Close()

	ctx := context.Background()
	influencer := &model.Influencer{
		Name:      "Criadora Local",
		Slug:      "criadora-local",
		Bio:       "Conteudo sobre a cidade",
		CityArea:  "Gastronomia",
		Instagram: "https://instagram.com/criadora",
		Active:    true,
	}
	if err := repo.InfluencerCreate(ctx, influencer); err != nil {
		t.Fatalf("InfluencerCreate failed: %v", err)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/influencers", nil)
	listRec := httptest.NewRecorder()
	h.InfluencerList(listRec, listReq)
	if listRec.Code != http.StatusOK || !strings.Contains(listRec.Body.String(), "Criadora Local") {
		t.Fatalf("list status/body = %d %s", listRec.Code, listRec.Body.String())
	}

	detailReq := httptest.NewRequest(http.MethodGet, "/influencer/criadora-local", nil)
	detailReq.SetPathValue("slug", "criadora-local")
	detailRec := httptest.NewRecorder()
	h.InfluencerDetail(detailRec, detailReq)
	if detailRec.Code != http.StatusOK || !strings.Contains(detailRec.Body.String(), "https://instagram.com/criadora") {
		t.Fatalf("detail status/body = %d %s", detailRec.Code, detailRec.Body.String())
	}
}

func TestAdminPostNewRendersWithNilPost(t *testing.T) {
	h, repo := newTestHandler(t)
	defer repo.Close()

	req := httptest.NewRequest(http.MethodGet, h.cfg.AdminPathPrefix+"/posts/new", nil)
	req = req.WithContext(auth.WithUser(req.Context(), &model.User{ID: 1, Role: model.RoleEditor}))
	rec := httptest.NewRecorder()

	h.AdminPostNew(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	body := rec.Body.String()
	for _, want := range []string{"Nova Noticia", `name="cover_image"`, `name="publish_at"`} {
		if !strings.Contains(body, want) {
			t.Fatalf("new post form missing %q: %s", want, body)
		}
	}
}

func TestAdminPostCreateRequiresEditorialNotesForPolitics(t *testing.T) {
	h, repo := newTestHandler(t)
	defer repo.Close()

	ctx := context.Background()
	category, err := repo.CategoryGetBySlug(ctx, "politica-bastidores")
	if err != nil || category == nil {
		t.Fatalf("CategoryGetBySlug failed: %v", err)
	}

	form := url.Values{}
	form.Set("title", "Materia politica")
	form.Set("category_id", strconv.FormatInt(category.ID, 10))
	form.Set("content", "Conteudo")
	form.Set("status", "draft")
	req := httptest.NewRequest(http.MethodPost, h.cfg.AdminPathPrefix+"/posts", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(auth.WithUser(req.Context(), &model.User{ID: 1, Role: model.RoleEditor}))
	rec := httptest.NewRecorder()

	h.AdminPostCreate(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "Apuracao e responsavel editorial") {
		t.Fatalf("missing editorial validation error: %s", rec.Body.String())
	}
}

func TestAdminPostCreateScheduledRequiresPublishAt(t *testing.T) {
	h, repo := newTestHandler(t)
	defer repo.Close()

	ctx := context.Background()
	category, err := repo.CategoryGetBySlug(ctx, "noticias")
	if err != nil || category == nil {
		t.Fatalf("CategoryGetBySlug failed: %v", err)
	}

	form := url.Values{}
	form.Set("title", "Post sem data")
	form.Set("category_id", strconv.FormatInt(category.ID, 10))
	form.Set("content", "Conteudo")
	form.Set("status", "scheduled")
	req := httptest.NewRequest(http.MethodPost, h.cfg.AdminPathPrefix+"/posts", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(auth.WithUser(req.Context(), &model.User{ID: 1, Role: model.RoleEditor}))
	rec := httptest.NewRecorder()

	h.AdminPostCreate(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "data de publicacao valida") {
		t.Fatalf("missing publish_at validation error: %s", rec.Body.String())
	}
}

func TestAdminPostCreateBlocksRedatorPublish(t *testing.T) {
	h, repo := newTestHandler(t)
	defer repo.Close()

	ctx := context.Background()
	category, err := repo.CategoryGetBySlug(ctx, "noticias")
	if err != nil || category == nil {
		t.Fatalf("CategoryGetBySlug failed: %v", err)
	}

	form := url.Values{}
	form.Set("title", "Tentativa de publicacao")
	form.Set("category_id", strconv.FormatInt(category.ID, 10))
	form.Set("content", "Conteudo")
	form.Set("status", string(model.StatusPublished))
	req := httptest.NewRequest(http.MethodPost, h.cfg.AdminPathPrefix+"/posts", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(auth.WithUser(req.Context(), &model.User{ID: 5, Role: model.RoleRedator}))
	rec := httptest.NewRecorder()

	h.AdminPostCreate(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "nao publicar/agendar") {
		t.Fatalf("missing permission validation error: %s", rec.Body.String())
	}
	posts, err := repo.PostListAll(ctx, 10, 0)
	if err != nil {
		t.Fatalf("PostListAll failed: %v", err)
	}
	if len(posts) != 0 {
		t.Fatalf("redator publish attempt created post: %#v", posts)
	}
}

func TestAdminPostEditorialWorkflow(t *testing.T) {
	h, repo := newTestHandler(t)
	defer repo.Close()

	ctx := context.Background()
	authorID := int64(10)
	post := &model.Post{
		Title:           "Pauta de bairro",
		Slug:            "pauta-de-bairro",
		Excerpt:         "Resumo",
		Content:         "<p>Texto</p>",
		MetaDescription: "Descricao para aprovacao editorial",
		AuthorID:        &authorID,
		Status:          model.StatusDraft,
	}
	if err := repo.PostCreate(ctx, post); err != nil {
		t.Fatalf("PostCreate failed: %v", err)
	}

	submitReq := httptest.NewRequest(http.MethodPost, h.cfg.AdminPathPrefix+"/posts/"+strconv.FormatInt(post.ID, 10)+"/submit-review", nil)
	submitReq.SetPathValue("id", strconv.FormatInt(post.ID, 10))
	submitReq = submitReq.WithContext(auth.WithUser(submitReq.Context(), &model.User{ID: authorID, Name: "Redator", Role: model.RoleRedator}))
	submitRec := httptest.NewRecorder()
	h.AdminPostSubmitReview(submitRec, submitReq)
	if submitRec.Code != http.StatusSeeOther {
		t.Fatalf("submit status = %d, want %d; body=%s", submitRec.Code, http.StatusSeeOther, submitRec.Body.String())
	}
	reviewPost, _ := repo.PostGetByID(ctx, post.ID)
	if reviewPost.Status != model.StatusReview {
		t.Fatalf("status after submit = %s, want review", reviewPost.Status)
	}

	approveReq := httptest.NewRequest(http.MethodPost, h.cfg.AdminPathPrefix+"/posts/"+strconv.FormatInt(post.ID, 10)+"/approve", nil)
	approveReq.SetPathValue("id", strconv.FormatInt(post.ID, 10))
	approveReq = approveReq.WithContext(auth.WithUser(approveReq.Context(), &model.User{ID: 20, Name: "Revisor", Role: model.RoleRevisor}))
	approveRec := httptest.NewRecorder()
	h.AdminPostApprove(approveRec, approveReq)
	if approveRec.Code != http.StatusSeeOther {
		t.Fatalf("approve status = %d, want %d; body=%s", approveRec.Code, http.StatusSeeOther, approveRec.Body.String())
	}
	approvedPost, _ := repo.PostGetByID(ctx, post.ID)
	if approvedPost.Status != model.StatusApproved {
		t.Fatalf("status after approve = %s, want approved", approvedPost.Status)
	}

	logs, err := repo.AuditLogList(ctx, "post", post.ID, 10)
	if err != nil {
		t.Fatalf("AuditLogList failed: %v", err)
	}
	if len(logs) < 2 {
		t.Fatalf("expected workflow audit logs, got %d", len(logs))
	}
}

func TestAdminPostRejectReturnsToDraftWithComment(t *testing.T) {
	h, repo := newTestHandler(t)
	defer repo.Close()

	ctx := context.Background()
	post := &model.Post{
		Title:          "Pauta para revisar",
		Slug:           "pauta-para-revisar",
		Excerpt:        "Resumo",
		Content:        "<p>Texto</p>",
		Status:         model.StatusReview,
		EditorialNotes: "Nota inicial",
	}
	if err := repo.PostCreate(ctx, post); err != nil {
		t.Fatalf("PostCreate failed: %v", err)
	}

	form := url.Values{}
	form.Set("comment", "Faltou confirmar a fonte")
	req := httptest.NewRequest(http.MethodPost, h.cfg.AdminPathPrefix+"/posts/"+strconv.FormatInt(post.ID, 10)+"/reject", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetPathValue("id", strconv.FormatInt(post.ID, 10))
	req = req.WithContext(auth.WithUser(req.Context(), &model.User{ID: 20, Name: "Revisor", Role: model.RoleRevisor}))
	rec := httptest.NewRecorder()
	h.AdminPostReject(rec, req)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("reject status = %d, want %d; body=%s", rec.Code, http.StatusSeeOther, rec.Body.String())
	}
	rejected, _ := repo.PostGetByID(ctx, post.ID)
	if rejected.Status != model.StatusDraft {
		t.Fatalf("status after reject = %s, want draft", rejected.Status)
	}
	if !strings.Contains(rejected.EditorialNotes, "Faltou confirmar a fonte") || !strings.Contains(rejected.EditorialNotes, "Nota inicial") {
		t.Fatalf("rejection comment not appended: %q", rejected.EditorialNotes)
	}
}

func TestAdminCategoryCreateAndUpdate(t *testing.T) {
	h, repo := newTestHandler(t)
	defer repo.Close()

	admin := &model.User{ID: 1, Role: model.RoleAdmin}
	createForm := url.Values{}
	createForm.Set("name", "Cultura Local")
	createForm.Set("description", "Agenda cultural e cobertura comunitaria")
	createForm.Set("meta_title", "Cultura de Inhumas")
	createForm.Set("meta_description", "Noticias culturais de Inhumas e comunidade local")
	createForm.Set("sort_order", "12")
	createForm.Set("active", "on")
	createForm.Set("requires_editorial_notes", "on")
	createReq := httptest.NewRequest(http.MethodPost, h.cfg.AdminPathPrefix+"/categories", strings.NewReader(createForm.Encode()))
	createReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	createReq = createReq.WithContext(auth.WithUser(createReq.Context(), admin))
	createRec := httptest.NewRecorder()

	h.AdminCategoryCreate(createRec, createReq)

	if createRec.Code != http.StatusSeeOther {
		t.Fatalf("create status = %d, want %d; body=%s", createRec.Code, http.StatusSeeOther, createRec.Body.String())
	}
	category, err := repo.CategoryGetBySlug(context.Background(), "cultura-local")
	if err != nil || category == nil {
		t.Fatalf("CategoryGetBySlug failed: %v", err)
	}
	if !category.Active || !category.RequiresEditorialNotes || category.SortOrder != 12 || category.MetaTitle != "Cultura de Inhumas" {
		t.Fatalf("category fields not persisted: %+v", category)
	}

	updateForm := url.Values{}
	updateForm.Set("name", "Cultura e Comunidade")
	updateForm.Set("slug", "cultura-comunidade")
	updateForm.Set("description", "Nova descricao")
	updateForm.Set("meta_title", "Cultura e Comunidade em Inhumas")
	updateForm.Set("meta_description", "Cobertura cultural e comunitaria de Inhumas")
	updateForm.Set("sort_order", "4")
	updateReq := httptest.NewRequest(http.MethodPost, h.cfg.AdminPathPrefix+"/categories/"+strconv.FormatInt(category.ID, 10), strings.NewReader(updateForm.Encode()))
	updateReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	updateReq.SetPathValue("id", strconv.FormatInt(category.ID, 10))
	updateReq = updateReq.WithContext(auth.WithUser(updateReq.Context(), admin))
	updateRec := httptest.NewRecorder()

	h.AdminCategoryUpdate(updateRec, updateReq)

	if updateRec.Code != http.StatusSeeOther {
		t.Fatalf("update status = %d, want %d; body=%s", updateRec.Code, http.StatusSeeOther, updateRec.Body.String())
	}
	updated, err := repo.CategoryGetBySlug(context.Background(), "cultura-comunidade")
	if err != nil || updated == nil {
		t.Fatalf("updated category not found: %v", err)
	}
	if updated.Name != "Cultura e Comunidade" || updated.Active || updated.RequiresEditorialNotes || updated.SortOrder != 4 || updated.MetaDescription == "" {
		t.Fatalf("category update not persisted: %+v", updated)
	}
}

func TestAdminCategoryDeleteBlocksWhenPostsExist(t *testing.T) {
	h, repo := newTestHandler(t)
	defer repo.Close()

	ctx := context.Background()
	category, err := repo.CategoryGetBySlug(ctx, "noticias")
	if err != nil || category == nil {
		t.Fatalf("CategoryGetBySlug failed: %v", err)
	}
	post := &model.Post{
		Title:      "Materia vinculada",
		Slug:       "materia-vinculada",
		Excerpt:    "Resumo",
		Content:    "<p>Texto</p>",
		CategoryID: &category.ID,
		Status:     model.StatusDraft,
	}
	if err := repo.PostCreate(ctx, post); err != nil {
		t.Fatalf("PostCreate failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, h.cfg.AdminPathPrefix+"/categories/"+strconv.FormatInt(category.ID, 10)+"/delete", nil)
	req.SetPathValue("id", strconv.FormatInt(category.ID, 10))
	req = req.WithContext(auth.WithUser(req.Context(), &model.User{ID: 1, Role: model.RoleAdmin}))
	rec := httptest.NewRecorder()

	h.AdminCategoryDelete(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "Nao e possivel excluir categoria") {
		t.Fatalf("delete block message missing: %s", rec.Body.String())
	}
	stillThere, err := repo.CategoryGetByID(ctx, category.ID)
	if err != nil || stillThere == nil {
		t.Fatalf("category should not be deleted: %v", err)
	}
}

func TestAdminTagCreateAndUpdate(t *testing.T) {
	h, repo := newTestHandler(t)
	defer repo.Close()

	admin := &model.User{ID: 1, Role: model.RoleAdmin}
	createForm := url.Values{}
	createForm.Set("name", "Saude")
	createForm.Set("description", "Pautas de saude local")
	createForm.Set("meta_title", "Saude em Inhumas")
	createForm.Set("meta_description", "Noticias de saude em Inhumas")
	createForm.Set("active", "on")
	createReq := httptest.NewRequest(http.MethodPost, h.cfg.AdminPathPrefix+"/tags", strings.NewReader(createForm.Encode()))
	createReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	createReq = createReq.WithContext(auth.WithUser(createReq.Context(), admin))
	createRec := httptest.NewRecorder()

	h.AdminTagCreate(createRec, createReq)

	if createRec.Code != http.StatusSeeOther {
		t.Fatalf("create status = %d, want %d; body=%s", createRec.Code, http.StatusSeeOther, createRec.Body.String())
	}
	tag, err := repo.TagGetBySlug(context.Background(), "saude")
	if err != nil || tag == nil {
		t.Fatalf("TagGetBySlug failed: %v", err)
	}
	if !tag.Active || tag.Description != "Pautas de saude local" || tag.MetaTitle != "Saude em Inhumas" {
		t.Fatalf("tag fields not persisted: %+v", tag)
	}

	updateForm := url.Values{}
	updateForm.Set("name", "Saude Publica")
	updateForm.Set("slug", "saude-publica")
	updateForm.Set("description", "Nova descricao")
	updateForm.Set("meta_title", "Saude publica em Inhumas")
	updateForm.Set("meta_description", "Cobertura de saude publica em Inhumas")
	updateReq := httptest.NewRequest(http.MethodPost, h.cfg.AdminPathPrefix+"/tags/"+strconv.FormatInt(tag.ID, 10), strings.NewReader(updateForm.Encode()))
	updateReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	updateReq.SetPathValue("id", strconv.FormatInt(tag.ID, 10))
	updateReq = updateReq.WithContext(auth.WithUser(updateReq.Context(), admin))
	updateRec := httptest.NewRecorder()

	h.AdminTagUpdate(updateRec, updateReq)

	if updateRec.Code != http.StatusSeeOther {
		t.Fatalf("update status = %d, want %d; body=%s", updateRec.Code, http.StatusSeeOther, updateRec.Body.String())
	}
	updated, err := repo.TagGetBySlug(context.Background(), "saude-publica")
	if err != nil || updated == nil {
		t.Fatalf("updated tag not found: %v", err)
	}
	if updated.Name != "Saude Publica" || updated.Active || updated.Description != "Nova descricao" || updated.MetaDescription == "" {
		t.Fatalf("tag update not persisted: %+v", updated)
	}
}

func TestAdminPostCreatePersistsTagsAndPublicTagPage(t *testing.T) {
	h, repo := newTestHandler(t)
	defer repo.Close()

	ctx := context.Background()
	category, err := repo.CategoryGetBySlug(ctx, "noticias")
	if err != nil || category == nil {
		t.Fatalf("CategoryGetBySlug failed: %v", err)
	}
	tag := &model.Tag{Name: "Economia", Slug: "economia", Description: "Economia local", Active: true}
	if err := repo.TagCreate(ctx, tag); err != nil {
		t.Fatalf("TagCreate failed: %v", err)
	}

	form := url.Values{}
	form.Set("title", "Economia local em alta")
	form.Set("category_id", strconv.FormatInt(category.ID, 10))
	form.Set("content", "Conteudo da noticia")
	form.Set("meta_description", "Cobertura economica local com contexto para leitores de Inhumas")
	form.Set("status", string(model.StatusPublished))
	form.Add("tag_ids", strconv.FormatInt(tag.ID, 10))
	req := httptest.NewRequest(http.MethodPost, h.cfg.AdminPathPrefix+"/posts", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(auth.WithUser(req.Context(), &model.User{ID: 1, Role: model.RoleEditor}))
	rec := httptest.NewRecorder()

	h.AdminPostCreate(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusSeeOther, rec.Body.String())
	}
	post, err := repo.PostGetBySlug(ctx, "economia-local-em-alta")
	if err != nil || post == nil {
		t.Fatalf("PostGetBySlug failed: %v", err)
	}
	if len(post.Tags) != 1 || post.Tags[0].Slug != "economia" {
		t.Fatalf("post tags not persisted: %+v", post.Tags)
	}

	detailReq := httptest.NewRequest(http.MethodGet, "/noticia/economia-local-em-alta", nil)
	detailReq.SetPathValue("slug", "economia-local-em-alta")
	detailRec := httptest.NewRecorder()
	h.PostDetail(detailRec, detailReq)
	if detailRec.Code != http.StatusOK || !strings.Contains(detailRec.Body.String(), "/tag/economia") {
		t.Fatalf("post detail missing tag link: status=%d body=%s", detailRec.Code, detailRec.Body.String())
	}

	tagReq := httptest.NewRequest(http.MethodGet, "/tag/economia", nil)
	tagReq.SetPathValue("slug", "economia")
	tagRec := httptest.NewRecorder()
	h.TagPosts(tagRec, tagReq)
	if tagRec.Code != http.StatusOK || !strings.Contains(tagRec.Body.String(), "Economia local em alta") {
		t.Fatalf("tag page missing post: status=%d body=%s", tagRec.Code, tagRec.Body.String())
	}
}

func TestAdminMediaLibraryListsUpdatesAndBlocksUsedDelete(t *testing.T) {
	h, repo := newTestHandler(t)
	defer repo.Close()

	ctx := context.Background()
	asset := &model.MediaAsset{
		Key:          "webp/used.webp",
		OriginalName: "used.png",
		Title:        "Foto usada",
		AltText:      "Imagem em uso",
		ContentType:  "image/webp",
		SizeBytes:    2048,
	}
	if err := repo.MediaAssetCreate(ctx, asset); err != nil {
		t.Fatalf("MediaAssetCreate failed: %v", err)
	}
	post := &model.Post{
		Title:         "Post com midia",
		Slug:          "post-com-midia",
		Content:       "Conteudo",
		CoverImageKey: asset.Key,
		Status:        model.StatusDraft,
	}
	if err := repo.PostCreate(ctx, post); err != nil {
		t.Fatalf("PostCreate failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, h.cfg.AdminPathPrefix+"/media", nil)
	req = req.WithContext(auth.WithUser(req.Context(), &model.User{ID: 1, Role: model.RoleEditor}))
	rec := httptest.NewRecorder()
	h.AdminMedia(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "Foto usada") || !strings.Contains(rec.Body.String(), "1 uso(s)") || !strings.Contains(rec.Body.String(), "disabled") {
		t.Fatalf("media page missing expected asset state: %s", rec.Body.String())
	}

	updateForm := url.Values{}
	updateForm.Set("title", "Foto atualizada")
	updateForm.Set("alt_text", "Alt atualizado")
	updateReq := httptest.NewRequest(http.MethodPost, h.cfg.AdminPathPrefix+"/media/1", strings.NewReader(updateForm.Encode()))
	updateReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	updateReq.SetPathValue("id", strconv.FormatInt(asset.ID, 10))
	updateReq = updateReq.WithContext(auth.WithUser(updateReq.Context(), &model.User{ID: 1, Role: model.RoleEditor}))
	updateRec := httptest.NewRecorder()
	h.AdminMediaUpdate(updateRec, updateReq)
	if updateRec.Code != http.StatusSeeOther {
		t.Fatalf("update status = %d, want %d; body=%s", updateRec.Code, http.StatusSeeOther, updateRec.Body.String())
	}
	updated, _ := repo.MediaAssetGetByID(ctx, asset.ID)
	if updated == nil || updated.Title != "Foto atualizada" || updated.AltText != "Alt atualizado" {
		t.Fatalf("media update not persisted: %+v", updated)
	}

	deleteReq := httptest.NewRequest(http.MethodPost, h.cfg.AdminPathPrefix+"/media/1/delete", nil)
	deleteReq.SetPathValue("id", strconv.FormatInt(asset.ID, 10))
	deleteReq = deleteReq.WithContext(auth.WithUser(deleteReq.Context(), &model.User{ID: 1, Role: model.RoleEditor}))
	deleteRec := httptest.NewRecorder()
	h.AdminMediaDelete(deleteRec, deleteReq)
	if deleteRec.Code != http.StatusOK || !strings.Contains(deleteRec.Body.String(), "midia em uso") {
		t.Fatalf("used delete should be blocked: status=%d body=%s", deleteRec.Code, deleteRec.Body.String())
	}
	stillExists, _ := repo.MediaAssetGetByID(ctx, asset.ID)
	if stillExists == nil {
		t.Fatal("used media should not be deleted")
	}

	unused := &model.MediaAsset{Key: "webp/free.webp", OriginalName: "free.png", Title: "Livre", AltText: "Livre", ContentType: "image/webp"}
	if err := repo.MediaAssetCreate(ctx, unused); err != nil {
		t.Fatalf("MediaAssetCreate unused failed: %v", err)
	}
	freeDeleteReq := httptest.NewRequest(http.MethodPost, h.cfg.AdminPathPrefix+"/media/2/delete", nil)
	freeDeleteReq.SetPathValue("id", strconv.FormatInt(unused.ID, 10))
	freeDeleteReq = freeDeleteReq.WithContext(auth.WithUser(freeDeleteReq.Context(), &model.User{ID: 1, Role: model.RoleEditor}))
	freeDeleteRec := httptest.NewRecorder()
	h.AdminMediaDelete(freeDeleteRec, freeDeleteReq)
	if freeDeleteRec.Code != http.StatusSeeOther {
		t.Fatalf("unused delete status = %d, want %d; body=%s", freeDeleteRec.Code, http.StatusSeeOther, freeDeleteRec.Body.String())
	}
	deleted, _ := repo.MediaAssetGetByID(ctx, unused.ID)
	if deleted != nil {
		t.Fatal("unused media should be deleted")
	}
}

func TestAdminMediaFiltersByDateAndMonth(t *testing.T) {
	h, repo := newTestHandler(t)
	defer repo.Close()

	ctx := context.Background()
	jan := &model.MediaAsset{Key: "webp/janeiro.webp", OriginalName: "janeiro.png", Title: "Imagem Janeiro", AltText: "Janeiro", ContentType: "image/webp"}
	may := &model.MediaAsset{Key: "webp/maio.webp", OriginalName: "maio.png", Title: "Imagem Maio", AltText: "Maio", ContentType: "image/webp"}
	if err := repo.MediaAssetCreate(ctx, jan); err != nil {
		t.Fatalf("MediaAssetCreate jan failed: %v", err)
	}
	if err := repo.MediaAssetCreate(ctx, may); err != nil {
		t.Fatalf("MediaAssetCreate may failed: %v", err)
	}
	if _, err := repo.DB().ExecContext(ctx, `UPDATE media_assets SET created_at = $1, updated_at = $1 WHERE id = $2`, time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC), jan.ID); err != nil {
		t.Fatalf("update jan date failed: %v", err)
	}
	if _, err := repo.DB().ExecContext(ctx, `UPDATE media_assets SET created_at = $1, updated_at = $1 WHERE id = $2`, time.Date(2026, 5, 12, 10, 0, 0, 0, time.UTC), may.ID); err != nil {
		t.Fatalf("update may date failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, h.cfg.AdminPathPrefix+"/media?month=2026-05", nil)
	req = req.WithContext(auth.WithUser(req.Context(), &model.User{ID: 1, Role: model.RoleEditor}))
	rec := httptest.NewRecorder()
	h.AdminMedia(rec, req)
	body := rec.Body.String()
	if rec.Code != http.StatusOK || !strings.Contains(body, "Imagem Maio") || strings.Contains(body, "Imagem Janeiro") || !strings.Contains(body, "Maio 2026") {
		t.Fatalf("month filter failed: status=%d body=%s", rec.Code, body)
	}

	dateReq := httptest.NewRequest(http.MethodGet, h.cfg.AdminPathPrefix+"/media?date_from=2026-01-01&date_to=2026-01-31", nil)
	dateReq = dateReq.WithContext(auth.WithUser(dateReq.Context(), &model.User{ID: 1, Role: model.RoleEditor}))
	dateRec := httptest.NewRecorder()
	h.AdminMedia(dateRec, dateReq)
	dateBody := dateRec.Body.String()
	if dateRec.Code != http.StatusOK || !strings.Contains(dateBody, "Imagem Janeiro") || strings.Contains(dateBody, "Imagem Maio") {
		t.Fatalf("date filter failed: status=%d body=%s", dateRec.Code, dateBody)
	}
}

func TestAdminPostCreateReusesMediaLibraryAssets(t *testing.T) {
	h, repo := newTestHandler(t)
	defer repo.Close()

	ctx := context.Background()
	category, err := repo.CategoryGetBySlug(ctx, "noticias")
	if err != nil || category == nil {
		t.Fatalf("CategoryGetBySlug failed: %v", err)
	}
	cover := &model.MediaAsset{Key: "webp/cover.webp", OriginalName: "cover.png", Title: "Capa", AltText: "Capa", ContentType: "image/webp"}
	gallery := &model.MediaAsset{Key: "webp/gallery.webp", OriginalName: "gallery.png", Title: "Galeria", AltText: "Galeria", ContentType: "image/webp"}
	if err := repo.MediaAssetCreate(ctx, cover); err != nil {
		t.Fatalf("MediaAssetCreate cover failed: %v", err)
	}
	if err := repo.MediaAssetCreate(ctx, gallery); err != nil {
		t.Fatalf("MediaAssetCreate gallery failed: %v", err)
	}

	form := url.Values{}
	form.Set("title", "Noticia com midia reaproveitada")
	form.Set("category_id", strconv.FormatInt(category.ID, 10))
	form.Set("content", "Conteudo da noticia")
	form.Set("status", string(model.StatusDraft))
	form.Set("cover_image_media_key", cover.Key)
	form.Add("gallery_media_keys", gallery.Key)
	req := httptest.NewRequest(http.MethodPost, h.cfg.AdminPathPrefix+"/posts", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(auth.WithUser(req.Context(), &model.User{ID: 1, Role: model.RoleEditor}))
	rec := httptest.NewRecorder()

	h.AdminPostCreate(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusSeeOther, rec.Body.String())
	}
	post, err := repo.PostGetBySlug(ctx, "noticia-com-midia-reaproveitada")
	if err != nil || post == nil {
		t.Fatalf("PostGetBySlug failed: %v", err)
	}
	if post.CoverImageKey != cover.Key || len(post.GalleryImageKeys) != 1 || post.GalleryImageKeys[0] != gallery.Key {
		t.Fatalf("media keys not reused: cover=%q gallery=%v", post.CoverImageKey, post.GalleryImageKeys)
	}
}

func TestAdminBannerCreateReusesMediaLibraryAsset(t *testing.T) {
	h, repo := newTestHandler(t)
	defer repo.Close()

	ctx := context.Background()
	asset := &model.MediaAsset{Key: "webp/banner-library.webp", OriginalName: "banner.png", Title: "Banner Biblioteca", AltText: "Banner", ContentType: "image/webp"}
	if err := repo.MediaAssetCreate(ctx, asset); err != nil {
		t.Fatalf("MediaAssetCreate failed: %v", err)
	}

	form := url.Values{}
	form.Set("name", "Campanha biblioteca")
	form.Set("advertiser_name", "Cliente")
	form.Set("position", "hero")
	form.Set("link_url", "https://example.com")
	form.Set("status", "active")
	form.Set("start_date", "2026-05-01")
	form.Set("end_date", "2026-05-31")
	form.Set("image_media_key", asset.Key)
	req := httptest.NewRequest(http.MethodPost, h.cfg.AdminPathPrefix+"/banners", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(auth.WithUser(req.Context(), &model.User{ID: 1, Role: model.RoleComercial}))
	rec := httptest.NewRecorder()

	h.AdminBannerCreate(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusSeeOther, rec.Body.String())
	}
	banners, err := repo.BannerList(ctx)
	if err != nil {
		t.Fatalf("BannerList failed: %v", err)
	}
	if len(banners) != 1 || banners[0].ImageKey != asset.Key {
		t.Fatalf("banner did not reuse media asset: %+v", banners)
	}
}

func TestAdminPostsFiltersByTag(t *testing.T) {
	h, repo := newTestHandler(t)
	defer repo.Close()

	ctx := context.Background()
	category, err := repo.CategoryGetBySlug(ctx, "noticias")
	if err != nil || category == nil {
		t.Fatalf("CategoryGetBySlug failed: %v", err)
	}
	tag := &model.Tag{Name: "Mobilidade", Slug: "mobilidade", Active: true}
	if err := repo.TagCreate(ctx, tag); err != nil {
		t.Fatalf("TagCreate failed: %v", err)
	}
	tagOther := &model.Tag{Name: "Educacao", Slug: "educacao", Active: true}
	if err := repo.TagCreate(ctx, tagOther); err != nil {
		t.Fatalf("TagCreate other failed: %v", err)
	}
	postTagged := &model.Post{
		Title:      "Obra de mobilidade",
		Slug:       "obra-de-mobilidade",
		Excerpt:    "Resumo",
		Content:    "<p>Texto</p>",
		CategoryID: &category.ID,
		Status:     model.StatusDraft,
	}
	if err := repo.PostCreate(ctx, postTagged); err != nil {
		t.Fatalf("PostCreate tagged failed: %v", err)
	}
	if err := repo.PostSetTags(ctx, postTagged.ID, []int64{tag.ID}); err != nil {
		t.Fatalf("PostSetTags tagged failed: %v", err)
	}
	postOther := &model.Post{
		Title:      "Matricula escolar",
		Slug:       "matricula-escolar",
		Excerpt:    "Resumo",
		Content:    "<p>Texto</p>",
		CategoryID: &category.ID,
		Status:     model.StatusDraft,
	}
	if err := repo.PostCreate(ctx, postOther); err != nil {
		t.Fatalf("PostCreate other failed: %v", err)
	}
	if err := repo.PostSetTags(ctx, postOther.ID, []int64{tagOther.ID}); err != nil {
		t.Fatalf("PostSetTags other failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, h.cfg.AdminPathPrefix+"/posts?tag_id="+strconv.FormatInt(tag.ID, 10), nil)
	req = req.WithContext(auth.WithUser(req.Context(), &model.User{ID: 1, Role: model.RoleAdmin}))
	rec := httptest.NewRecorder()

	h.AdminPosts(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Obra de mobilidade") || strings.Contains(body, "Matricula escolar") {
		t.Fatalf("tag filter did not isolate posts: %s", body)
	}
}

func TestPostSearchFindsPostsByTagName(t *testing.T) {
	_, repo := newTestHandler(t)
	defer repo.Close()

	ctx := context.Background()
	category, err := repo.CategoryGetBySlug(ctx, "noticias")
	if err != nil || category == nil {
		t.Fatalf("CategoryGetBySlug failed: %v", err)
	}
	tag := &model.Tag{Name: "Mobilidade Urbana", Slug: "mobilidade-urbana", Active: true}
	if err := repo.TagCreate(ctx, tag); err != nil {
		t.Fatalf("TagCreate failed: %v", err)
	}
	now := time.Now()
	post := &model.Post{
		Title:       "Obra importante",
		Slug:        "obra-importante",
		Excerpt:     "Resumo local",
		Content:     "<p>Texto sem o termo pesquisado.</p>",
		CategoryID:  &category.ID,
		Status:      model.StatusPublished,
		PublishedAt: &now,
	}
	if err := repo.PostCreate(ctx, post); err != nil {
		t.Fatalf("PostCreate failed: %v", err)
	}
	if err := repo.PostSetTags(ctx, post.ID, []int64{tag.ID}); err != nil {
		t.Fatalf("PostSetTags failed: %v", err)
	}

	results, err := repo.PostSearch(ctx, "Mobilidade Urbana", 10)
	if err != nil {
		t.Fatalf("PostSearch failed: %v", err)
	}
	if len(results) != 1 || results[0].Slug != "obra-importante" {
		t.Fatalf("search by tag returned unexpected results: %+v", results)
	}
}

func TestAdminBannerUpdate(t *testing.T) {
	h, repo := newTestHandler(t)
	defer repo.Close()

	ctx := context.Background()
	banner := &model.Banner{
		Name:      "Banner antigo",
		Position:  "hero",
		ImageKey:  "webp/banner.webp",
		LinkURL:   "https://example.com/old",
		StartDate: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC),
		Active:    true,
		Priority:  1,
	}
	if err := repo.BannerCreate(ctx, banner); err != nil {
		t.Fatalf("BannerCreate failed: %v", err)
	}

	form := url.Values{}
	form.Set("name", "Banner novo")
	form.Set("position", "in_feed")
	form.Set("link_url", "https://example.com/new")
	form.Set("start_date", "2026-05-01")
	form.Set("end_date", "2026-05-31")
	form.Set("priority", "5")
	form.Set("active", "on")
	req := httptest.NewRequest(http.MethodPost, h.cfg.AdminPathPrefix+"/banners/1", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetPathValue("id", "1")
	req = req.WithContext(auth.WithUser(req.Context(), &model.User{ID: 1, Role: model.RoleComercial}))
	rec := httptest.NewRecorder()

	h.AdminBannerUpdate(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusSeeOther, rec.Body.String())
	}
	updated, err := repo.BannerGetByID(ctx, banner.ID)
	if err != nil {
		t.Fatalf("BannerGetByID failed: %v", err)
	}
	if updated.Name != "Banner novo" || updated.Position != "in_feed" || updated.Priority != 5 || updated.ImageKey != "webp/banner.webp" {
		t.Fatalf("unexpected banner update: %#v", updated)
	}
	logs, err := repo.AuditLogList(ctx, "banner", banner.ID, 10)
	if err != nil {
		t.Fatalf("AuditLogList failed: %v", err)
	}
	if len(logs) == 0 || logs[0].Action != "update" || !strings.Contains(logs[0].Changes, "in_feed") {
		t.Fatalf("missing banner update audit log: %#v", logs)
	}
}

func TestAdminBannerUpdateRejectsInvalidDateRange(t *testing.T) {
	h, repo := newTestHandler(t)
	defer repo.Close()

	ctx := context.Background()
	banner := &model.Banner{
		Name:      "Banner",
		Position:  "hero",
		ImageKey:  "webp/banner.webp",
		LinkURL:   "https://example.com",
		StartDate: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC),
		Active:    true,
	}
	if err := repo.BannerCreate(ctx, banner); err != nil {
		t.Fatalf("BannerCreate failed: %v", err)
	}

	form := url.Values{}
	form.Set("name", "Banner")
	form.Set("position", "hero")
	form.Set("link_url", "https://example.com")
	form.Set("start_date", "2026-05-31")
	form.Set("end_date", "2026-05-01")
	req := httptest.NewRequest(http.MethodPost, h.cfg.AdminPathPrefix+"/banners/1", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetPathValue("id", "1")
	req = req.WithContext(auth.WithUser(req.Context(), &model.User{ID: 1, Role: model.RoleComercial}))
	rec := httptest.NewRecorder()

	h.AdminBannerUpdate(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "Data de fim") {
		t.Fatalf("missing date validation error: %s", rec.Body.String())
	}
}

func TestAdminBannerUpdateRejectsOverlap(t *testing.T) {
	h, repo := newTestHandler(t)
	defer repo.Close()

	ctx := context.Background()
	current := &model.Banner{
		Name:      "Hero atual",
		Position:  "hero",
		ImageKey:  "webp/banner.webp",
		LinkURL:   "https://example.com/current",
		StartDate: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC),
		Active:    true,
	}
	if err := repo.BannerCreate(ctx, current); err != nil {
		t.Fatalf("BannerCreate current failed: %v", err)
	}
	other := &model.Banner{
		Name:      "Hero ocupado",
		Position:  "hero",
		ImageKey:  "webp/banner2.webp",
		LinkURL:   "https://example.com/other",
		StartDate: time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, 5, 20, 0, 0, 0, 0, time.UTC),
		Active:    true,
	}
	if err := repo.BannerCreate(ctx, other); err != nil {
		t.Fatalf("BannerCreate other failed: %v", err)
	}

	form := url.Values{}
	form.Set("name", "Hero atual")
	form.Set("position", "hero")
	form.Set("link_url", "https://example.com/current")
	form.Set("start_date", "2026-05-01")
	form.Set("end_date", "2026-05-15")
	form.Set("active", "on")
	req := httptest.NewRequest(http.MethodPost, h.cfg.AdminPathPrefix+"/banners/1", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetPathValue("id", strconv.FormatInt(current.ID, 10))
	req = req.WithContext(auth.WithUser(req.Context(), &model.User{ID: 1, Role: model.RoleComercial}))
	rec := httptest.NewRecorder()

	h.AdminBannerUpdate(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "Ja existe banner ativo") {
		t.Fatalf("missing overlap validation error: %s", rec.Body.String())
	}
}

func TestAdminPromoUpdate(t *testing.T) {
	h, repo := newTestHandler(t)
	defer repo.Close()

	ctx := context.Background()
	store := &model.Store{Slug: "loja", Name: "Loja", Active: true}
	if err := repo.StoreCreate(ctx, store); err != nil {
		t.Fatalf("StoreCreate failed: %v", err)
	}
	promo := &model.Promotion{
		StoreID:      store.ID,
		Title:        "Promo antiga",
		Slug:         "promo-antiga",
		Description:  "Descricao antiga",
		PriceDisplay: "R$ 10",
		ImageKey:     "webp/promo.webp",
		StartDate:    time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		EndDate:      time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC),
		Status:       "active",
	}
	if err := repo.PromotionCreate(ctx, promo); err != nil {
		t.Fatalf("PromotionCreate failed: %v", err)
	}

	form := url.Values{}
	form.Set("store_id", strconv.FormatInt(store.ID, 10))
	form.Set("title", "Promo nova")
	form.Set("description", "Descricao nova")
	form.Set("price_display", "R$ 20")
	form.Set("start_date", "2026-05-01")
	form.Set("end_date", "2026-05-31")
	form.Set("status", "draft")
	form.Set("is_sponsored", "on")
	req := httptest.NewRequest(http.MethodPost, h.cfg.AdminPathPrefix+"/promotions/1", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetPathValue("id", "1")
	req = req.WithContext(auth.WithUser(req.Context(), &model.User{ID: 1, Role: model.RoleComercial}))
	rec := httptest.NewRecorder()

	h.AdminPromoUpdate(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusSeeOther, rec.Body.String())
	}
	updated, err := repo.PromotionGetByID(ctx, promo.ID)
	if err != nil {
		t.Fatalf("PromotionGetByID failed: %v", err)
	}
	if updated.Title != "Promo nova" || updated.Status != "draft" || !updated.IsSponsored || updated.ImageKey != "webp/promo.webp" {
		t.Fatalf("unexpected promo update: %#v", updated)
	}
	logs, err := repo.AuditLogList(ctx, "promotion", promo.ID, 10)
	if err != nil {
		t.Fatalf("AuditLogList failed: %v", err)
	}
	if len(logs) == 0 || logs[0].Action != "update" || !strings.Contains(logs[0].Changes, "Promo nova") {
		t.Fatalf("missing promotion update audit log: %#v", logs)
	}
}
