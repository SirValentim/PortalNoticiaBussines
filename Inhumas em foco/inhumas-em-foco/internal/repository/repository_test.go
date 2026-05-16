package repository

import (
	"context"
	"testing"
	"time"

	"inhumas-em-foco/internal/model"
)

func TestUserUpdate(t *testing.T) {
	repo, err := New(":memory:")
	if err != nil {
		t.Fatalf("New repo failed: %v", err)
	}
	defer repo.Close()

	ctx := context.Background()
	user := &model.User{
		Name:         "Usuario",
		Email:        "usuario@example.com",
		PasswordHash: "hash",
		Role:         model.RoleRedator,
		Active:       true,
	}
	if err := repo.UserCreate(ctx, user); err != nil {
		t.Fatalf("UserCreate failed: %v", err)
	}

	user.Name = "Usuario Editado"
	user.Email = "editado@example.com"
	user.Role = model.RoleEditor
	user.Active = false
	if err := repo.UserUpdate(ctx, user); err != nil {
		t.Fatalf("UserUpdate failed: %v", err)
	}

	updated, err := repo.UserGetByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("UserGetByID failed: %v", err)
	}
	if updated.Name != user.Name || updated.Email != user.Email || updated.Role != user.Role || updated.Active {
		t.Fatalf("updated user mismatch: %+v", updated)
	}
	if updated.PasswordHash != user.PasswordHash {
		t.Fatalf("password hash changed: %q", updated.PasswordHash)
	}
}

func TestOpenSQLiteMarksSchemaMigration(t *testing.T) {
	repo, err := Open("sqlite", ":memory:", "")
	if err != nil {
		t.Fatalf("Open sqlite failed: %v", err)
	}
	defer repo.Close()
	if repo.Driver() != "sqlite" {
		t.Fatalf("driver = %q, want sqlite", repo.Driver())
	}
	var count int
	if err := repo.DB().QueryRow(`SELECT COUNT(*) FROM schema_migrations WHERE version = 'sqlite_auto'`).Scan(&count); err != nil {
		t.Fatalf("schema migration lookup failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("schema migration count = %d, want 1", count)
	}
}

func TestNormalizeDriver(t *testing.T) {
	cases := []struct {
		driver string
		url    string
		want   string
	}{
		{"", "postgres://user:pass@localhost/db", "postgres"},
		{"pg", "./inhumas.db", "postgres"},
		{"sqlite3", "postgres://ignored", "sqlite"},
		{"", "./inhumas.db", "sqlite"},
	}
	for _, tc := range cases {
		if got := normalizeDriver(tc.driver, tc.url); got != tc.want {
			t.Fatalf("normalizeDriver(%q,%q) = %q, want %q", tc.driver, tc.url, got, tc.want)
		}
	}
}

func TestJobRecordFailureBackoffAndDeadJobs(t *testing.T) {
	repo, err := New(":memory:")
	if err != nil {
		t.Fatalf("New repo failed: %v", err)
	}
	defer repo.Close()

	ctx := context.Background()
	job := &model.Job{
		Type:        model.JobGenerateSitemap,
		Payload:     "{}",
		Status:      model.JobPending,
		RunAt:       time.Now().Add(-time.Minute),
		MaxAttempts: 3,
	}
	if err := repo.JobCreate(ctx, job); err != nil {
		t.Fatalf("JobCreate failed: %v", err)
	}

	moved, err := repo.JobRecordFailure(ctx, model.Job{ID: job.ID, Attempts: 0, MaxAttempts: 3}, "primeira falha")
	if err != nil {
		t.Fatalf("JobRecordFailure attempt 1 failed: %v", err)
	}
	if moved {
		t.Fatal("job should not move to dead_jobs on first failure")
	}

	stored := readJob(t, repo, job.ID)
	if stored.Status != model.JobPending {
		t.Fatalf("status after first failure = %q, want %q", stored.Status, model.JobPending)
	}
	if stored.Attempts != 1 {
		t.Fatalf("attempts after first failure = %d, want 1", stored.Attempts)
	}
	if !stored.RunAt.After(time.Now()) {
		t.Fatalf("run_at should be scheduled in the future, got %v", stored.RunAt)
	}

	moved, err = repo.JobRecordFailure(ctx, model.Job{ID: job.ID, Attempts: 2, MaxAttempts: 3}, "falha final")
	if err != nil {
		t.Fatalf("JobRecordFailure final failed: %v", err)
	}
	if !moved {
		t.Fatal("job should move to dead_jobs on max attempts")
	}

	var jobsCount int
	if err := repo.DB().QueryRowContext(ctx, `SELECT COUNT(*) FROM jobs WHERE id = $1`, job.ID).Scan(&jobsCount); err != nil {
		t.Fatalf("count jobs failed: %v", err)
	}
	if jobsCount != 0 {
		t.Fatalf("jobs count = %d, want 0", jobsCount)
	}

	var deadCount int
	if err := repo.DB().QueryRowContext(ctx, `SELECT COUNT(*) FROM dead_jobs WHERE id = $1`, job.ID).Scan(&deadCount); err != nil {
		t.Fatalf("count dead_jobs failed: %v", err)
	}
	if deadCount != 1 {
		t.Fatalf("dead_jobs count = %d, want 1", deadCount)
	}
}

func TestSitemapEntriesReturnsPublicContent(t *testing.T) {
	repo, err := New(":memory:")
	if err != nil {
		t.Fatalf("New repo failed: %v", err)
	}
	defer repo.Close()

	ctx := context.Background()
	now := time.Now()
	_, err = repo.DB().ExecContext(ctx, `
		INSERT INTO posts (title, slug, content, status, published_at, updated_at)
		VALUES
			('Publicado', 'post-publicado', 'conteudo', 'published', $1, $1),
			('Rascunho', 'post-rascunho', 'conteudo', 'draft', NULL, $1);
		INSERT INTO stores (slug, name, active, created_at)
		VALUES
			('loja-ativa', 'Loja Ativa', true, $1),
			('loja-inativa', 'Loja Inativa', false, $1);
		INSERT INTO neighborhoods (slug, name, created_at)
		VALUES ('centro', 'Centro', $1);
	`, now)
	if err != nil {
		t.Fatalf("seed failed: %v", err)
	}

	entries, err := repo.SitemapEntries(ctx)
	if err != nil {
		t.Fatalf("SitemapEntries failed: %v", err)
	}

	paths := map[string]bool{}
	for _, entry := range entries {
		paths[entry.Path] = true
	}
	for _, want := range []string{"/noticia/post-publicado", "/loja/loja-ativa", "/bairro/centro"} {
		if !paths[want] {
			t.Fatalf("missing sitemap path %q in %#v", want, paths)
		}
	}
	for _, unwanted := range []string{"/noticia/post-rascunho", "/loja/loja-inativa"} {
		if paths[unwanted] {
			t.Fatalf("unexpected private sitemap path %q", unwanted)
		}
	}
}

func TestJobCreateDefaultsMaxAttempts(t *testing.T) {
	repo, err := New(":memory:")
	if err != nil {
		t.Fatalf("New repo failed: %v", err)
	}
	defer repo.Close()

	ctx := context.Background()
	job := &model.Job{
		Type:    model.JobGenerateSitemap,
		Payload: "{}",
		Status:  model.JobPending,
		RunAt:   time.Now(),
	}
	if err := repo.JobCreate(ctx, job); err != nil {
		t.Fatalf("JobCreate failed: %v", err)
	}
	if job.MaxAttempts != 3 {
		t.Fatalf("MaxAttempts = %d, want 3", job.MaxAttempts)
	}

	exists, err := repo.JobHasActiveType(ctx, model.JobGenerateSitemap)
	if err != nil {
		t.Fatalf("JobHasActiveType failed: %v", err)
	}
	if !exists {
		t.Fatal("expected active job type to exist")
	}
}

func TestJobClaimPendingMarksJobsRunning(t *testing.T) {
	repo, err := New(":memory:")
	if err != nil {
		t.Fatalf("New repo failed: %v", err)
	}
	defer repo.Close()

	ctx := context.Background()
	due := &model.Job{
		Type:        model.JobGenerateSitemap,
		Payload:     "{}",
		Status:      model.JobPending,
		RunAt:       time.Now().Add(-time.Minute),
		MaxAttempts: 3,
	}
	future := &model.Job{
		Type:        model.JobBackupDatabase,
		Payload:     "{}",
		Status:      model.JobPending,
		RunAt:       time.Now().Add(time.Hour),
		MaxAttempts: 3,
	}
	if err := repo.JobCreate(ctx, due); err != nil {
		t.Fatalf("JobCreate due failed: %v", err)
	}
	if err := repo.JobCreate(ctx, future); err != nil {
		t.Fatalf("JobCreate future failed: %v", err)
	}

	claimed, err := repo.JobClaimPending(ctx, 10)
	if err != nil {
		t.Fatalf("JobClaimPending failed: %v", err)
	}
	if len(claimed) != 1 {
		t.Fatalf("claimed jobs = %d, want 1", len(claimed))
	}
	if claimed[0].ID != due.ID || claimed[0].Status != model.JobRunning {
		t.Fatalf("claimed job = %#v, want due job running", claimed[0])
	}

	secondClaim, err := repo.JobClaimPending(ctx, 10)
	if err != nil {
		t.Fatalf("second JobClaimPending failed: %v", err)
	}
	if len(secondClaim) != 0 {
		t.Fatalf("second claim returned %d jobs, want 0", len(secondClaim))
	}
}

func TestPostSearchUsesFTSContent(t *testing.T) {
	repo, err := New(":memory:")
	if err != nil {
		t.Fatalf("New repo failed: %v", err)
	}
	defer repo.Close()

	ctx := context.Background()
	publishedAt := time.Now()
	post := &model.Post{
		Title:       "Feira de tecnologia em Inhumas",
		Slug:        "feira-tecnologia-inhumas",
		Excerpt:     "Agenda local",
		Content:     "Startups e inovacao no centro da cidade.",
		Status:      model.StatusPublished,
		PublishedAt: &publishedAt,
	}
	if err := repo.PostCreate(ctx, post); err != nil {
		t.Fatalf("PostCreate failed: %v", err)
	}

	results, err := repo.PostSearch(ctx, "startups inovacao", 10)
	if err != nil {
		t.Fatalf("PostSearch failed: %v", err)
	}
	if len(results) != 1 || results[0].ID != post.ID {
		t.Fatalf("search results = %#v, want created post", results)
	}
}

func TestBannerOverlapExcludesCurrentBanner(t *testing.T) {
	repo, err := New(":memory:")
	if err != nil {
		t.Fatalf("New repo failed: %v", err)
	}
	defer repo.Close()

	ctx := context.Background()
	first := &model.Banner{
		Name:      "Hero abril",
		Position:  "hero",
		ImageKey:  "webp/hero.webp",
		LinkURL:   "https://example.com",
		StartDate: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC),
		Active:    true,
	}
	if err := repo.BannerCreate(ctx, first); err != nil {
		t.Fatalf("BannerCreate first failed: %v", err)
	}
	second := &model.Banner{
		Name:      "Hero maio",
		Position:  "hero",
		ImageKey:  "webp/hero-maio.webp",
		LinkURL:   "https://example.com",
		StartDate: time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC),
		Active:    true,
	}
	if err := repo.BannerCreate(ctx, second); err != nil {
		t.Fatalf("BannerCreate second failed: %v", err)
	}

	count, err := repo.BannerCountActiveInPeriodExcluding(ctx, "hero", first.StartDate, first.EndDate, first.ID)
	if err != nil {
		t.Fatalf("BannerCountActiveInPeriodExcluding failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("overlap count excluding first = %d, want 1", count)
	}

	count, err = repo.BannerCountActiveInPeriodExcluding(ctx, "hero", first.StartDate, first.EndDate, second.ID)
	if err != nil {
		t.Fatalf("BannerCountActiveInPeriodExcluding failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("overlap count excluding second = %d, want 1", count)
	}

	count, err = repo.BannerCountActiveInPeriodExcluding(ctx, "hero", first.StartDate, first.EndDate, 0)
	if err != nil {
		t.Fatalf("BannerCountActiveInPeriodExcluding failed: %v", err)
	}
	if count != 2 {
		t.Fatalf("overlap count without exclusion = %d, want 2", count)
	}
}

func TestMetricTotalsAndTopEntities(t *testing.T) {
	repo, err := New(":memory:")
	if err != nil {
		t.Fatalf("New repo failed: %v", err)
	}
	defer repo.Close()

	ctx := context.Background()
	metrics := []model.Metric{
		{MetricType: "post_view", EntityType: "post", EntityID: 10},
		{MetricType: "post_view", EntityType: "post", EntityID: 10},
		{MetricType: "post_view", EntityType: "post", EntityID: 11},
		{MetricType: "store_view", EntityType: "store", EntityID: 5},
	}
	for i := range metrics {
		if err := repo.MetricTrack(ctx, &metrics[i]); err != nil {
			t.Fatalf("MetricTrack failed: %v", err)
		}
	}

	count, err := repo.MetricCountByType(ctx, "post_view")
	if err != nil {
		t.Fatalf("MetricCountByType failed: %v", err)
	}
	if count != 3 {
		t.Fatalf("post_view count = %d, want 3", count)
	}

	totals, err := repo.MetricTotals(ctx, 10)
	if err != nil {
		t.Fatalf("MetricTotals failed: %v", err)
	}
	if len(totals) != 2 || totals[0].MetricType != "post_view" || totals[0].Total != 3 {
		t.Fatalf("unexpected totals: %#v", totals)
	}

	top, err := repo.MetricTopEntities(ctx, "post_view", 10)
	if err != nil {
		t.Fatalf("MetricTopEntities failed: %v", err)
	}
	if len(top) != 2 || top[0].EntityID != 10 || top[0].Total != 2 {
		t.Fatalf("unexpected top entities: %#v", top)
	}
}

func TestDeadJobListAndCount(t *testing.T) {
	repo, err := New(":memory:")
	if err != nil {
		t.Fatalf("New repo failed: %v", err)
	}
	defer repo.Close()

	ctx := context.Background()
	job := &model.Job{
		Type:        model.JobGenerateSitemap,
		Payload:     "{}",
		Status:      model.JobPending,
		RunAt:       time.Now().Add(-time.Hour),
		MaxAttempts: 1,
	}
	if err := repo.JobCreate(ctx, job); err != nil {
		t.Fatalf("JobCreate failed: %v", err)
	}
	moved, err := repo.JobRecordFailure(ctx, model.Job{ID: job.ID, Attempts: 0, MaxAttempts: 1}, "falha permanente")
	if err != nil {
		t.Fatalf("JobRecordFailure failed: %v", err)
	}
	if !moved {
		t.Fatal("expected job to move to dead_jobs")
	}

	count, err := repo.DeadJobCount(ctx)
	if err != nil {
		t.Fatalf("DeadJobCount failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("dead job count = %d, want 1", count)
	}

	jobs, err := repo.DeadJobList(ctx, 10)
	if err != nil {
		t.Fatalf("DeadJobList failed: %v", err)
	}
	if len(jobs) != 1 || jobs[0].Type != model.JobGenerateSitemap || jobs[0].Error != "falha permanente" {
		t.Fatalf("unexpected dead jobs: %#v", jobs)
	}
}

func readJob(t *testing.T, repo *Repository, id int64) model.Job {
	t.Helper()
	var job model.Job
	err := repo.DB().QueryRowContext(context.Background(), `
		SELECT id, type, payload, status, run_at, created_at, attempts, max_attempts, error, processed_at
		FROM jobs WHERE id = $1`, id).
		Scan(&job.ID, &job.Type, &job.Payload, &job.Status, &job.RunAt, &job.CreatedAt, &job.Attempts, &job.MaxAttempts, &job.Error, &job.ProcessedAt)
	if err != nil {
		t.Fatalf("read job failed: %v", err)
	}
	return job
}
