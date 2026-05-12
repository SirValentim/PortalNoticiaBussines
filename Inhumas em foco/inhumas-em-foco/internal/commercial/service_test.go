package commercial

import (
	"testing"

	"inhumas-em-foco/internal/model"
)

func TestNormalizeBannerStatus(t *testing.T) {
	tests := map[string]string{
		"":           "active",
		"ACTIVE":     "active",
		" paused ":   "paused",
		"draft":      "draft",
		"expired":    "expired",
		"unexpected": "active",
	}

	for input, want := range tests {
		if got := NormalizeBannerStatus(input); got != want {
			t.Fatalf("NormalizeBannerStatus(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestServiceValidateBanner(t *testing.T) {
	svc := NewService()

	if msg := svc.ValidateBanner(&model.Banner{}); msg != "Nome da campanha e obrigatorio" {
		t.Fatalf("name validation = %q", msg)
	}
	if msg := svc.ValidateBanner(&model.Banner{Name: "Campanha"}); msg != "Cliente/anunciante e obrigatorio" {
		t.Fatalf("advertiser validation = %q", msg)
	}
	if msg := svc.ValidateBanner(&model.Banner{Name: "Campanha", AdvertiserName: "Cliente", Position: "hero", LinkURL: "https://example.com", Status: "active"}); msg != "Banner ativo precisa de imagem" {
		t.Fatalf("active image validation = %q", msg)
	}
	if msg := svc.ValidateBanner(&model.Banner{Name: "Campanha", AdvertiserName: "Cliente", Position: "hero", LinkURL: "https://example.com", Status: "draft"}); msg != "" {
		t.Fatalf("draft banner validation = %q", msg)
	}
}

func TestServiceParseDateRange(t *testing.T) {
	svc := NewService()

	if _, _, msg := svc.ParseDateRange("bad", "2026-05-10"); msg != "Data de inicio invalida" {
		t.Fatalf("start date validation = %q", msg)
	}
	if _, _, msg := svc.ParseDateRange("2026-05-11", "2026-05-10"); msg != "Data de fim deve ser igual ou posterior a data de inicio" {
		t.Fatalf("range validation = %q", msg)
	}
	start, end, msg := svc.ParseDateRange("2026-05-10", "2026-05-11")
	if msg != "" || !end.After(start) {
		t.Fatalf("valid range start=%v end=%v msg=%q", start, end, msg)
	}
}
