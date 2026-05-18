package config_test

import (
	"strings"
	"testing"

	"inhumas-em-foco/internal/config"
)

func TestLoadTenantBrandingConfig_Sucesso(t *testing.T) {
	t.Setenv("PORTAL_NAME", "Test Portal")
	t.Setenv("SITE_URL", "https://testportal.com")
	t.Setenv("PORTAL_CONTACT_EMAIL", "contato@testportal.com")

	cfg, err := config.LoadTenantBrandingConfig()
	if err != nil {
		t.Fatalf("esperava sucesso, obteve erro: %v", err)
	}
	if cfg.PortalName != "Test Portal" {
		t.Errorf("PortalName incorreto: %s", cfg.PortalName)
	}
	if cfg.PortalLocale != "pt_BR" {
		t.Errorf("PortalLocale default incorreto: %s", cfg.PortalLocale)
	}
	if cfg.ArticlesPerPage != 12 {
		t.Errorf("ArticlesPerPage default incorreto: %d", cfg.ArticlesPerPage)
	}
}

func TestLoadTenantBrandingConfig_VariaveisObrigatoriasAusentes(t *testing.T) {
	t.Setenv("PORTAL_NAME", "")
	t.Setenv("SITE_URL", "")
	t.Setenv("PORTAL_CONTACT_EMAIL", "")

	_, err := config.LoadTenantBrandingConfig()
	if err == nil {
		t.Fatal("esperava erro por variaveis obrigatorias ausentes")
	}
	if !strings.Contains(err.Error(), "PORTAL_NAME") || !strings.Contains(err.Error(), "SITE_URL") || !strings.Contains(err.Error(), "PORTAL_CONTACT_EMAIL") {
		t.Fatalf("erro nao lista variaveis ausentes: %v", err)
	}
}

func TestFullTitle(t *testing.T) {
	t.Setenv("PORTAL_NAME", "Inhumas em Foco")
	t.Setenv("SITE_URL", "https://inhumasemfoco.online")
	t.Setenv("PORTAL_CONTACT_EMAIL", "contato@inhumasemfoco.online")

	cfg, err := config.LoadTenantBrandingConfig()
	if err != nil {
		t.Fatalf("LoadTenantBrandingConfig: %v", err)
	}

	title := cfg.FullTitle("Politica")
	expected := "Politica | Inhumas em Foco"
	if title != expected {
		t.Errorf("FullTitle incorreto. esperado: %q, obtido: %q", expected, title)
	}

	homeTitle := cfg.FullTitle("")
	if homeTitle != "Inhumas em Foco" {
		t.Errorf("FullTitle vazio incorreto: %q", homeTitle)
	}
}
