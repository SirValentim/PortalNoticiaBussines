package editorialai

import (
	"context"
	"strings"
	"testing"

	"inhumas-em-foco/internal/model"
)

func TestMockProviderGeneratesSafeMetaAndKeepsSource(t *testing.T) {
	provider := NewMockProvider()
	result, err := provider.Suggest(context.Background(), Request{
		Task: TaskMetaDescription,
		Post: model.Post{
			Title:      "Prefeitura anuncia obra",
			Excerpt:    "Prefeitura anunciou obra no centro com impacto no transito local.",
			SourceName: "Prefeitura de Inhumas",
			SourceURL:  "https://example.com/fonte",
			Status:     model.StatusDraft,
		},
	})
	if err != nil {
		t.Fatalf("Suggest failed: %v", err)
	}
	if result.MetaDescription == "" || len([]rune(result.MetaDescription)) > 160 {
		t.Fatalf("invalid meta description: %q", result.MetaDescription)
	}
	if result.SourceName != "Prefeitura de Inhumas" || result.SourceURL == "" {
		t.Fatalf("source was not preserved: %+v", result)
	}
	if !strings.Contains(result.Notes, "revise fatos") {
		t.Fatalf("guardrail note missing: %q", result.Notes)
	}
}

func TestMockProviderChecksDuplicate(t *testing.T) {
	provider := NewMockProvider()
	result, err := provider.Suggest(context.Background(), Request{
		Task: TaskCheckDuplicate,
		Post: model.Post{ID: 2, Title: "Prefeitura anuncia obra no centro de Inhumas", Status: model.StatusDraft},
		RecentPosts: []model.Post{
			{ID: 1, Title: "Prefeitura anuncia obra no centro de Inhumas hoje"},
		},
	})
	if err != nil {
		t.Fatalf("Suggest failed: %v", err)
	}
	if !strings.Contains(result.DuplicateRisk, "Possivel duplicidade") {
		t.Fatalf("unexpected duplicate result: %q", result.DuplicateRisk)
	}
}

func TestTaskValidate(t *testing.T) {
	if err := Task("unknown").Validate(); err != ErrInvalidTask {
		t.Fatalf("Validate error = %v, want ErrInvalidTask", err)
	}
}
