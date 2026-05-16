package automation

import (
	"context"
	"path/filepath"
	"testing"

	"inhumas-em-foco/internal/model"
	"inhumas-em-foco/internal/repository"
)

func TestRunSourceCreatesDraftAndSkipsDuplicates(t *testing.T) {
	repo, err := repository.New(filepath.Join(t.TempDir(), "automation.db"))
	if err != nil {
		t.Fatalf("New repo failed: %v", err)
	}
	defer repo.Close()

	ctx := context.Background()
	category, err := repo.CategoryGetBySlug(ctx, "noticias")
	if err != nil {
		t.Fatalf("CategoryGetBySlug failed: %v", err)
	}
	source := &model.AutomationSource{
		Name:              "Fonte Teste",
		SourceType:        string(model.AutomationSourceRSS),
		URL:               "https://example.com/rss.xml",
		DefaultCategoryID: &category.ID,
		Active:            true,
	}
	if err := repo.AutomationSourceCreate(ctx, source); err != nil {
		t.Fatalf("AutomationSourceCreate failed: %v", err)
	}

	svc := NewService(repo)
	svc.SetFetcher(func(ctx context.Context, source model.AutomationSource) ([]Item, error) {
		return []Item{
			{Title: "Prefeitura anuncia obra no centro", Link: "https://example.com/noticia-1", Summary: "<p>Resumo da obra no centro.</p>"},
			{Title: "Prefeitura anuncia obra no centro", Link: "https://example.com/noticia-1", Summary: "Duplicada"},
		}, nil
	})

	run, err := svc.RunSource(ctx, source.ID)
	if err != nil {
		t.Fatalf("RunSource failed: %v", err)
	}
	if run.ItemsFound != 2 || run.DraftsCreated != 1 || run.Duplicates != 1 {
		t.Fatalf("unexpected run counters: %+v", run)
	}

	queue, err := repo.AutomationDraftQueue(ctx, 10)
	if err != nil {
		t.Fatalf("AutomationDraftQueue failed: %v", err)
	}
	if len(queue) != 1 {
		t.Fatalf("draft queue length = %d, want 1", len(queue))
	}
	post, err := repo.PostGetByID(ctx, queue[0].ID)
	if err != nil {
		t.Fatalf("PostGetByID failed: %v", err)
	}
	if post.Status != model.StatusDraft || post.SourceURL != "https://example.com/noticia-1" || post.SourceName != "Fonte Teste" {
		t.Fatalf("automation draft did not preserve required fields: %+v", post)
	}

	secondRun, err := svc.RunSource(ctx, source.ID)
	if err != nil {
		t.Fatalf("second RunSource failed: %v", err)
	}
	if secondRun.DraftsCreated != 0 || secondRun.Duplicates != 2 {
		t.Fatalf("second run counters = %+v, want no drafts and two duplicates", secondRun)
	}
}

func TestIsSimilarTitle(t *testing.T) {
	if !IsSimilarTitle("Prefeitura anuncia obra no centro de Inhumas", "Prefeitura anuncia obra no centro de Inhumas hoje") {
		t.Fatal("expected similar titles to match")
	}
	if IsSimilarTitle("Festival de musica abre inscricoes", "Camara aprova projeto tributario") {
		t.Fatal("unrelated titles should not match")
	}
}
