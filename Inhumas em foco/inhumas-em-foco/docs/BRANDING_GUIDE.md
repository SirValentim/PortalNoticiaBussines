# Guia de Branding por Tenant

O CMS carrega a identidade do portal pelo pacote `internal/config` e injeta esses dados em cada request via middleware. Isso permite rodar o mesmo binario para portais diferentes alterando apenas variaveis de ambiente e os assets em `static/branding`.

## Variaveis obrigatorias

- `PORTAL_NAME`
- `SITE_URL`
- `PORTAL_CONTACT_EMAIL`

Sem essas variaveis, `cmd/web` e `cmd/worker` encerram a inicializacao com erro.

## Assets esperados

Coloque os arquivos do portal em `static/branding/`:

- `logo.svg`
- `logo-white.svg`
- `favicon.ico`
- `apple-touch-icon.png`
- `og-default.jpg`
- `icon-192.png`
- `icon-512.png`

O `manifest.json` e servido dinamicamente em `/manifest.json`, usando as variaveis carregadas em runtime.

## Segundo portal

Para subir outro portal, configure um `.env` novo com valores como:

```dotenv
PORTAL_NAME=LaMafia Music
PORTAL_TAGLINE=O portal da cena musical brasileira
PORTAL_DESCRIPTION=Releases, resenhas, perfis de artistas e cobertura da musica nacional.
PORTAL_CATEGORY=music
PORTAL_LOCALE=pt_BR
PORTAL_LANGUAGE=pt-BR

SITE_URL=https://lamafia.music
ADMIN_PATH_PREFIX=/admin/studio

PORTAL_PRIMARY_COLOR=#1a0a2e
PORTAL_SECONDARY_COLOR=#e040fb
PORTAL_ACCENT_COLOR=#7c4dff
PORTAL_LOGO_PATH=/static/branding/logo.svg
PORTAL_FAVICON_PATH=/static/branding/favicon.ico
PORTAL_SEO_TITLE_SUFFIX= | LaMafia Music

PORTAL_CONTACT_EMAIL=contato@lamafia.music
PORTAL_CONTACT_CITY=Goiania
PORTAL_CONTACT_STATE=GO
PORTAL_COPYRIGHT_HOLDER=LaMafia Music
```

## Verificacao

Rode:

```bash
go test ./internal/config/... -v
go test ./internal/middleware/... -v
go build ./...
```
