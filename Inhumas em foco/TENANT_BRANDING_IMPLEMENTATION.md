# Tenant Branding Configuration — Implementation Specification

## Objetivo

Extrair todos os atributos de identidade de portal (nome, URL, branding visual, metadata SEO, contato) para um pacote `config` centralizado, injetável via variáveis de ambiente, eliminando qualquer valor hardcoded nos templates HTML e handlers Go.

Ao final desta implementação, o mesmo binário compilado deve ser capaz de servir qualquer portal apenas com um arquivo `.env` diferente, sem alteração de código-fonte.

---

## 1. Estrutura de Arquivos a Criar ou Modificar

```
inhumas-em-foco/
├── internal/
│   └── config/
│       ├── config.go          ← já existe — MODIFICAR
│       └── branding.go        ← CRIAR
├── internal/
│   └── middleware/
│       └── branding.go        ← CRIAR
├── static/
│   └── branding/              ← CRIAR diretório
│       └── .gitkeep
├── .env.example               ← MODIFICAR — adicionar variáveis de branding
└── docs/
    └── BRANDING_GUIDE.md      ← CRIAR
```

---

## 2. Pacote: `internal/config/branding.go`

### Criar o arquivo com o seguinte conteúdo exato:

```go
package config

import (
	"fmt"
	"os"
	"strings"
)

// TenantBrandingConfig encapsula todos os atributos de identidade visual
// e metadata de um portal. Todos os campos são populados exclusivamente
// via variáveis de ambiente, sem fallback para valores hardcoded de portal específico.
type TenantBrandingConfig struct {
	// Identidade do portal
	PortalName        string // PORTAL_NAME
	PortalTagline     string // PORTAL_TAGLINE
	PortalDescription string // PORTAL_DESCRIPTION
	PortalLocale      string // PORTAL_LOCALE (ex: pt_BR)
	PortalLanguage    string // PORTAL_LANGUAGE (ex: pt-BR)
	PortalCategory    string // PORTAL_CATEGORY (ex: news, music, regional)

	// URLs e domínio
	SiteURL       string // SITE_URL (ex: https://inhumasemfoco.online)
	AdminPathPrefix string // ADMIN_PATH_PREFIX

	// Identidade visual
	LogoPath       string // PORTAL_LOGO_PATH (ex: /static/branding/logo.svg)
	LogoAltText    string // PORTAL_LOGO_ALT
	FaviconPath    string // PORTAL_FAVICON_PATH
	PrimaryColor   string // PORTAL_PRIMARY_COLOR (hex, ex: #1a4a3a)
	SecondaryColor string // PORTAL_SECONDARY_COLOR (hex, ex: #f5c518)
	AccentColor    string // PORTAL_ACCENT_COLOR (hex, ex: #2d6a52)

	// SEO e metadata global
	SEOTitleSuffix    string // PORTAL_SEO_TITLE_SUFFIX (ex: " | Inhumas em Foco")
	SEODefaultImage   string // PORTAL_SEO_DEFAULT_IMAGE (URL absoluta)
	TwitterHandle     string // PORTAL_TWITTER_HANDLE (ex: @inhumasemfoco)
	FacebookPageURL   string // PORTAL_FACEBOOK_PAGE_URL
	InstagramHandle   string // PORTAL_INSTAGRAM_HANDLE

	// Contato e localização
	ContactEmail    string // PORTAL_CONTACT_EMAIL
	ContactPhone    string // PORTAL_CONTACT_PHONE
	ContactCity     string // PORTAL_CONTACT_CITY
	ContactState    string // PORTAL_CONTACT_STATE
	ContactCountry  string // PORTAL_CONTACT_COUNTRY (ex: BR)

	// Configurações editoriais
	ArticlesPerPage    int    // PORTAL_ARTICLES_PER_PAGE (default: 12)
	FeaturedTagSlug    string // PORTAL_FEATURED_TAG_SLUG (ex: destaque)
	BreakingNewsLabel  string // PORTAL_BREAKING_NEWS_LABEL (ex: Ao vivo)
	CopyrightHolder    string // PORTAL_COPYRIGHT_HOLDER
	FooterLegalText    string // PORTAL_FOOTER_LEGAL_TEXT

	// Integrações opcionais
	GoogleAnalyticsID  string // PORTAL_GA_ID (ex: G-XXXXXXXXXX)
	GoogleTagManagerID string // PORTAL_GTM_ID
	DisqusShortname    string // PORTAL_DISQUS_SHORTNAME
	RecaptchaSiteKey   string // PORTAL_RECAPTCHA_SITE_KEY
}

// LoadTenantBrandingConfig lê todas as variáveis de ambiente de branding
// e retorna um TenantBrandingConfig populado.
// Retorna erro se qualquer variável obrigatória estiver ausente ou inválida.
func LoadTenantBrandingConfig() (*TenantBrandingConfig, error) {
	cfg := &TenantBrandingConfig{}
	var missing []string

	// — Obrigatórias —
	cfg.PortalName = os.Getenv("PORTAL_NAME")
	if cfg.PortalName == "" {
		missing = append(missing, "PORTAL_NAME")
	}

	cfg.SiteURL = os.Getenv("SITE_URL")
	if cfg.SiteURL == "" {
		missing = append(missing, "SITE_URL")
	}

	cfg.ContactEmail = os.Getenv("PORTAL_CONTACT_EMAIL")
	if cfg.ContactEmail == "" {
		missing = append(missing, "PORTAL_CONTACT_EMAIL")
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("branding config: variáveis obrigatórias ausentes: %s", strings.Join(missing, ", "))
	}

	// — Com defaults —
	cfg.PortalTagline = envOrDefault("PORTAL_TAGLINE", "")
	cfg.PortalDescription = envOrDefault("PORTAL_DESCRIPTION", "")
	cfg.PortalLocale = envOrDefault("PORTAL_LOCALE", "pt_BR")
	cfg.PortalLanguage = envOrDefault("PORTAL_LANGUAGE", "pt-BR")
	cfg.PortalCategory = envOrDefault("PORTAL_CATEGORY", "news")
	cfg.AdminPathPrefix = envOrDefault("ADMIN_PATH_PREFIX", "/painel")

	cfg.LogoPath = envOrDefault("PORTAL_LOGO_PATH", "/static/branding/logo.svg")
	cfg.LogoAltText = envOrDefault("PORTAL_LOGO_ALT", cfg.PortalName)
	cfg.FaviconPath = envOrDefault("PORTAL_FAVICON_PATH", "/static/branding/favicon.ico")
	cfg.PrimaryColor = envOrDefault("PORTAL_PRIMARY_COLOR", "#1a4a3a")
	cfg.SecondaryColor = envOrDefault("PORTAL_SECONDARY_COLOR", "#f5c518")
	cfg.AccentColor = envOrDefault("PORTAL_ACCENT_COLOR", "#2d6a52")

	cfg.SEOTitleSuffix = envOrDefault("PORTAL_SEO_TITLE_SUFFIX", " | "+cfg.PortalName)
	cfg.SEODefaultImage = envOrDefault("PORTAL_SEO_DEFAULT_IMAGE", cfg.SiteURL+"/static/branding/og-default.jpg")
	cfg.TwitterHandle = envOrDefault("PORTAL_TWITTER_HANDLE", "")
	cfg.FacebookPageURL = envOrDefault("PORTAL_FACEBOOK_PAGE_URL", "")
	cfg.InstagramHandle = envOrDefault("PORTAL_INSTAGRAM_HANDLE", "")

	cfg.ContactPhone = envOrDefault("PORTAL_CONTACT_PHONE", "")
	cfg.ContactCity = envOrDefault("PORTAL_CONTACT_CITY", "")
	cfg.ContactState = envOrDefault("PORTAL_CONTACT_STATE", "")
	cfg.ContactCountry = envOrDefault("PORTAL_CONTACT_COUNTRY", "BR")

	cfg.ArticlesPerPage = envOrDefaultInt("PORTAL_ARTICLES_PER_PAGE", 12)
	cfg.FeaturedTagSlug = envOrDefault("PORTAL_FEATURED_TAG_SLUG", "destaque")
	cfg.BreakingNewsLabel = envOrDefault("PORTAL_BREAKING_NEWS_LABEL", "Ao vivo")
	cfg.CopyrightHolder = envOrDefault("PORTAL_COPYRIGHT_HOLDER", cfg.PortalName)
	cfg.FooterLegalText = envOrDefault("PORTAL_FOOTER_LEGAL_TEXT", "Todos os direitos reservados.")

	cfg.GoogleAnalyticsID = envOrDefault("PORTAL_GA_ID", "")
	cfg.GoogleTagManagerID = envOrDefault("PORTAL_GTM_ID", "")
	cfg.DisqusShortname = envOrDefault("PORTAL_DISQUS_SHORTNAME", "")
	cfg.RecaptchaSiteKey = envOrDefault("PORTAL_RECAPTCHA_SITE_KEY", "")

	return cfg, nil
}

// FullTitle retorna o título formatado para uso em <title> tags.
// Ex: "Política" → "Política | Inhumas em Foco"
func (b *TenantBrandingConfig) FullTitle(pageTitle string) string {
	if pageTitle == "" {
		return b.PortalName
	}
	return pageTitle + b.SEOTitleSuffix
}

// AbsoluteURL retorna uma URL absoluta para um path relativo.
func (b *TenantBrandingConfig) AbsoluteURL(path string) string {
	base := strings.TrimRight(b.SiteURL, "/")
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return base + path
}

// envOrDefault retorna o valor da variável de ambiente ou o default fornecido.
func envOrDefault(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

// envOrDefaultInt retorna o valor inteiro da variável de ambiente ou o default.
func envOrDefaultInt(key string, defaultValue int) int {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	var n int
	if _, err := fmt.Sscanf(v, "%d", &n); err != nil {
		return defaultValue
	}
	return n
}
```

---

## 3. Modificar: `internal/config/config.go`

### Localizar o struct `Config` principal e adicionar o campo `Branding`:

```go
// No struct Config existente, adicionar o campo:
Branding *TenantBrandingConfig
```

### Localizar a função de carregamento do Config (provavelmente `Load()` ou `New()`) e adicionar:

```go
// Após carregar os demais campos do Config, adicionar:
branding, err := LoadTenantBrandingConfig()
if err != nil {
    return nil, fmt.Errorf("falha ao carregar branding config: %w", err)
}
cfg.Branding = branding
```

---

## 4. Criar: `internal/middleware/branding.go`

Este middleware injeta o `TenantBrandingConfig` no contexto de cada requisição HTTP,
tornando-o disponível para todos os handlers sem acoplamento direto.

```go
package middleware

import (
	"context"
	"net/http"

	"github.com/SirValentim/PortalNoticiaBussines/internal/config"
)

// brandingContextKey é a chave de contexto para o TenantBrandingConfig.
// Uso de tipo privado evita colisões com outras chaves de contexto.
type brandingContextKey struct{}

// BrandingContextKey é a chave exportada para recuperação via context.Value.
var BrandingContextKey = brandingContextKey{}

// InjectBranding retorna um middleware HTTP que injeta o TenantBrandingConfig
// no contexto de cada requisição.
func InjectBranding(branding *config.TenantBrandingConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), BrandingContextKey, branding)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// BrandingFromContext recupera o TenantBrandingConfig do contexto.
// Retorna nil se não encontrado — handlers devem tratar este caso.
func BrandingFromContext(ctx context.Context) *config.TenantBrandingConfig {
	b, _ := ctx.Value(BrandingContextKey).(*config.TenantBrandingConfig)
	return b
}
```

---

## 5. Registrar o Middleware: `cmd/web/main.go`

### Localizar onde os middlewares são registrados na cadeia HTTP e adicionar:

```go
// Após inicializar cfg, registrar o middleware de branding:
// (antes do router principal ser servido)

handler = middleware.InjectBranding(cfg.Branding)(handler)
```

---

## 6. Criar `TemplateData` global com Branding

### Localizar o struct de dados de template (provavelmente em `internal/handlers/` ou similar).
### Garantir que `Branding` esteja presente em todos os renders de template:

```go
// Struct base para todos os templates — modificar ou criar se não existir:
type BaseTemplateData struct {
	Branding    *config.TenantBrandingConfig
	CurrentUser *models.User        // se aplicável
	CSRFToken   string
	FlashMsg    string
	// ... demais campos existentes
}

// Em cada handler que renderiza template, popular o Branding:
data := BaseTemplateData{
	Branding:  middleware.BrandingFromContext(r.Context()),
	CSRFToken: csrf.Token(r),
}
```

---

## 7. Atualizar Templates HTML

### Em todos os arquivos de template (`.html` ou `.gohtml`), substituir valores hardcoded:

| Substituir | Por |
|---|---|
| `Inhumas em Foco` (nome literal) | `{{.Branding.PortalName}}` |
| `inhumasemfoco.online` (domínio literal) | `{{.Branding.SiteURL}}` |
| `contato@inhumasemfoco.online` | `{{.Branding.ContactEmail}}` |
| `(62) 99999-9999` | `{{.Branding.ContactPhone}}` |
| `Inhumas, GO` | `{{.Branding.ContactCity}}, {{.Branding.ContactState}}` |
| `/static/logo.png` (logo path) | `{{.Branding.LogoPath}}` |
| `#1a4a3a` (cor hardcoded no CSS inline) | `{{.Branding.PrimaryColor}}` |
| `G-XXXXXXXXXX` (GA ID) | `{{.Branding.GoogleAnalyticsID}}` |
| `<title>Inhumas em Foco</title>` | `<title>{{.Branding.FullTitle .PageTitle}}</title>` |

### CSS Variables — no `<head>` de cada template, injetar variáveis CSS de branding:

```html
<style>
  :root {
    --portal-primary:   {{.Branding.PrimaryColor}};
    --portal-secondary: {{.Branding.SecondaryColor}};
    --portal-accent:    {{.Branding.AccentColor}};
  }
</style>
```

### Open Graph e Twitter Cards — substituir valores hardcoded:

```html
<meta property="og:site_name" content="{{.Branding.PortalName}}">
<meta property="og:locale"    content="{{.Branding.PortalLocale}}">
<meta property="og:image"     content="{{.Branding.SEODefaultImage}}">
<meta name="twitter:site"     content="{{.Branding.TwitterHandle}}">
```

### Google Analytics — condicional, apenas se configurado:

```html
{{if .Branding.GoogleAnalyticsID}}
<script async src="https://www.googletagmanager.com/gtag/js?id={{.Branding.GoogleAnalyticsID}}"></script>
<script>
  window.dataLayer = window.dataLayer || [];
  function gtag(){dataLayer.push(arguments);}
  gtag('js', new Date());
  gtag('config', '{{.Branding.GoogleAnalyticsID}}');
</script>
{{end}}
```

---

## 8. Atualizar `.env.example`

### Adicionar o bloco de variáveis de branding ao arquivo existente:

```dotenv
# ─────────────────────────────────────────
# TENANT BRANDING CONFIGURATION
# ─────────────────────────────────────────

# Identidade do portal (obrigatórias)
PORTAL_NAME=Inhumas em Foco
PORTAL_TAGLINE=O portal de notícias que conecta Inhumas
PORTAL_DESCRIPTION=Portal local de notícias, eventos, comércio e comunidade de Inhumas, Goiás.
PORTAL_LOCALE=pt_BR
PORTAL_LANGUAGE=pt-BR
PORTAL_CATEGORY=news

# URLs
SITE_URL=https://inhumasemfoco.online
ADMIN_PATH_PREFIX=/painel/7x9k2m

# Identidade visual
PORTAL_LOGO_PATH=/static/branding/logo.svg
PORTAL_LOGO_ALT=Inhumas em Foco
PORTAL_FAVICON_PATH=/static/branding/favicon.ico
PORTAL_PRIMARY_COLOR=#1a4a3a
PORTAL_SECONDARY_COLOR=#f5c518
PORTAL_ACCENT_COLOR=#2d6a52

# SEO global
PORTAL_SEO_TITLE_SUFFIX= | Inhumas em Foco
PORTAL_SEO_DEFAULT_IMAGE=https://inhumasemfoco.online/static/branding/og-default.jpg
PORTAL_TWITTER_HANDLE=@inhumasemfoco
PORTAL_FACEBOOK_PAGE_URL=
PORTAL_INSTAGRAM_HANDLE=

# Contato
PORTAL_CONTACT_EMAIL=contato@inhumasemfoco.online
PORTAL_CONTACT_PHONE=(62) 99999-9999
PORTAL_CONTACT_CITY=Inhumas
PORTAL_CONTACT_STATE=GO
PORTAL_CONTACT_COUNTRY=BR

# Editorial
PORTAL_ARTICLES_PER_PAGE=12
PORTAL_FEATURED_TAG_SLUG=destaque
PORTAL_BREAKING_NEWS_LABEL=Ao vivo
PORTAL_COPYRIGHT_HOLDER=Inhumas em Foco
PORTAL_FOOTER_LEGAL_TEXT=Todos os direitos reservados.

# Integrações (deixar vazio para desabilitar)
PORTAL_GA_ID=
PORTAL_GTM_ID=
PORTAL_DISQUS_SHORTNAME=
PORTAL_RECAPTCHA_SITE_KEY=
```

---

## 9. Assets de Branding por Portal

### Criar o diretório `static/branding/` e mover para ele:

```
static/branding/
├── logo.svg          ← logo principal (SVG preferível — escala sem perda)
├── logo-white.svg    ← variante branca para uso em fundo escuro
├── favicon.ico
├── apple-touch-icon.png   ← 180x180px
├── og-default.jpg         ← 1200x630px — imagem padrão para Open Graph
└── manifest.json          ← PWA manifest com nome e cores do portal
```

### `static/branding/manifest.json` — template:

```json
{
  "name": "{{PORTAL_NAME}}",
  "short_name": "{{PORTAL_NAME}}",
  "description": "{{PORTAL_DESCRIPTION}}",
  "start_url": "/",
  "display": "standalone",
  "background_color": "{{PORTAL_PRIMARY_COLOR}}",
  "theme_color": "{{PORTAL_PRIMARY_COLOR}}",
  "icons": [
    { "src": "/static/branding/icon-192.png", "sizes": "192x192", "type": "image/png" },
    { "src": "/static/branding/icon-512.png", "sizes": "512x512", "type": "image/png" }
  ]
}
```

> Nota: `manifest.json` deve ser servido via handler Go (não como arquivo estático) para que os valores sejam interpolados em runtime.

---

## 10. Handler para `manifest.json` dinâmico

### Criar em `internal/handlers/` (ou onde ficam os handlers públicos):

```go
// ManifestHandler serve o Web App Manifest com dados de branding injetados em runtime.
func ManifestHandler(w http.ResponseWriter, r *http.Request) {
	branding := middleware.BrandingFromContext(r.Context())
	if branding == nil {
		http.Error(w, "branding não disponível", http.StatusInternalServerError)
		return
	}

	manifest := map[string]interface{}{
		"name":             branding.PortalName,
		"short_name":       branding.PortalName,
		"description":      branding.PortalDescription,
		"start_url":        "/",
		"display":          "standalone",
		"background_color": branding.PrimaryColor,
		"theme_color":      branding.PrimaryColor,
		"icons": []map[string]string{
			{"src": "/static/branding/icon-192.png", "sizes": "192x192", "type": "image/png"},
			{"src": "/static/branding/icon-512.png", "sizes": "512x512", "type": "image/png"},
		},
	}

	w.Header().Set("Content-Type", "application/manifest+json")
	w.Header().Set("Cache-Control", "public, max-age=3600")

	if err := json.NewEncoder(w).Encode(manifest); err != nil {
		http.Error(w, "erro ao serializar manifest", http.StatusInternalServerError)
	}
}
```

### Registrar a rota no router principal:

```go
mux.HandleFunc("/manifest.json", handlers.ManifestHandler)
```

---

## 11. Testes Unitários: `internal/config/branding_test.go`

```go
package config_test

import (
	"os"
	"testing"

	"github.com/SirValentim/PortalNoticiaBussines/internal/config"
)

func TestLoadTenantBrandingConfig_Sucesso(t *testing.T) {
	os.Setenv("PORTAL_NAME", "Test Portal")
	os.Setenv("SITE_URL", "https://testportal.com")
	os.Setenv("PORTAL_CONTACT_EMAIL", "contato@testportal.com")
	defer func() {
		os.Unsetenv("PORTAL_NAME")
		os.Unsetenv("SITE_URL")
		os.Unsetenv("PORTAL_CONTACT_EMAIL")
	}()

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
	os.Unsetenv("PORTAL_NAME")
	os.Unsetenv("SITE_URL")
	os.Unsetenv("PORTAL_CONTACT_EMAIL")

	_, err := config.LoadTenantBrandingConfig()
	if err == nil {
		t.Fatal("esperava erro por variáveis obrigatórias ausentes")
	}
}

func TestFullTitle(t *testing.T) {
	os.Setenv("PORTAL_NAME", "Inhumas em Foco")
	os.Setenv("SITE_URL", "https://inhumasemfoco.online")
	os.Setenv("PORTAL_CONTACT_EMAIL", "contato@inhumasemfoco.online")
	defer func() {
		os.Unsetenv("PORTAL_NAME")
		os.Unsetenv("SITE_URL")
		os.Unsetenv("PORTAL_CONTACT_EMAIL")
	}()

	cfg, _ := config.LoadTenantBrandingConfig()

	title := cfg.FullTitle("Política")
	expected := "Política | Inhumas em Foco"
	if title != expected {
		t.Errorf("FullTitle incorreto. esperado: %q, obtido: %q", expected, title)
	}

	homeTitle := cfg.FullTitle("")
	if homeTitle != "Inhumas em Foco" {
		t.Errorf("FullTitle vazio incorreto: %q", homeTitle)
	}
}
```

---

## 12. Verificação Final

Após implementar, executar na raiz do projeto:

```bash
# Verificar compilação
go build ./...

# Executar testes
go test ./internal/config/... -v
go test ./internal/middleware/... -v

# Verificar ausência de strings hardcoded nos templates
grep -rn "Inhumas em Foco" templates/
grep -rn "inhumasemfoco.online" templates/
grep -rn "contato@inhumas" templates/
# Todos os resultados devem ser zero após a refatoração

# Smoke test com variáveis de branding do segundo portal
PORTAL_NAME="LaMafia Music" \
SITE_URL="https://lamafia.music" \
PORTAL_CONTACT_EMAIL="contato@lamafia.music" \
PORTAL_PRIMARY_COLOR="#1a0a2e" \
PORTAL_SECONDARY_COLOR="#e040fb" \
go run ./cmd/web
```

---

## 13. `.env` de Referência para LaMafia Music

Após validar o Inhumas, este é o `.env` mínimo para subir o segundo portal:

```dotenv
# Portal: LaMafia Music
PORTAL_NAME=LaMafia Music
PORTAL_TAGLINE=O portal da cena musical brasileira
PORTAL_DESCRIPTION=Releases, resenhas, perfis de artistas e a melhor cobertura da música nacional.
PORTAL_CATEGORY=music
PORTAL_LOCALE=pt_BR

SITE_URL=https://lamafia.music
ADMIN_PATH_PREFIX=/admin/studio

PORTAL_PRIMARY_COLOR=#1a0a2e
PORTAL_SECONDARY_COLOR=#e040fb
PORTAL_ACCENT_COLOR=#7c4dff

PORTAL_LOGO_PATH=/static/branding/logo.svg
PORTAL_FAVICON_PATH=/static/branding/favicon.ico
PORTAL_SEO_TITLE_SUFFIX= | LaMafia Music

PORTAL_CONTACT_EMAIL=contato@lamafia.music
PORTAL_CONTACT_CITY=Goiânia
PORTAL_CONTACT_STATE=GO

PORTAL_FEATURED_TAG_SLUG=destaque
PORTAL_BREAKING_NEWS_LABEL=Novo release
PORTAL_ARTICLES_PER_PAGE=16
PORTAL_COPYRIGHT_HOLDER=LaMafia Music

DB_DRIVER=sqlite
DATABASE_URL=/var/www/lamafia/lamafia.db
SESSION_SECRET=<gerar com: openssl rand -hex 32>
ADMIN_PATH_PREFIX=/admin/studio
```

---

## Resumo de Impacto

| Antes | Depois |
|---|---|
| Nome do portal hardcoded nos templates | `{{.Branding.PortalName}}` via contexto HTTP |
| Cores CSS fixas no código | CSS variables injetadas pelo template engine |
| Logo path fixo | Configurável por `PORTAL_LOGO_PATH` |
| Google Analytics hardcoded ou ausente | Condicional, habilitado por `PORTAL_GA_ID` |
| Novo portal = fork + busca/substituição manual | Novo portal = novo `.env` + novo `static/branding/` |
| Risco de esquecer algum valor hardcoded | Grep de validação automatizável em CI |
