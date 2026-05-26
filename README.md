# Inhumas em Foco - CMS Editorial e Portal Local

Este repositorio contem a base do projeto **Inhumas em Foco**, um portal local profissional para publicacao de noticias, conteudo editorial, classificados, eventos, negocios locais, promocoes, influencers, banners comerciais e operacao administrativa completa.

O projeto foi construido com foco em producao real: painel administrativo protegido, fluxo editorial, seguranca aplicada, SEO tecnico, automacoes assistidas, observabilidade, backup, readiness operacional e evolucao preparada para PostgreSQL.

## Visao Geral

O objetivo do sistema e entregar uma plataforma editorial moderna para uma operacao local, combinando conteudo jornalistico, servicos comerciais e modulos de cidade em uma unica aplicacao.

Principais capacidades:

- Publicacao de noticias com categorias, tags, imagem destacada, galeria, SEO e status editorial.
- Painel administrativo com permissoes, auditoria, metricas, configuracoes e gestao de usuarios.
- Modulos locais para eventos, classificados, lojas, promocoes, bairros e influencers.
- Biblioteca de midia com upload seguro, validacao de MIME, thumbnails e geracao WebP.
- Automacao editorial por fontes RSS/oficiais, sempre com revisao humana antes da publicacao.
- Camada de IA editorial desacoplada para sugestoes de titulo, resumo, meta description e tags.
- Gestao comercial com banners, periodos de veiculacao, posicoes e relatorios.
- SEO tecnico com sitemap, robots, RSS, canonical, Open Graph, Twitter Cards e JSON-LD.
- Operacao com health check, metricas, scripts de backup, smoke test e readiness de producao.

## Stack Tecnica

- **Backend:** Go 1.24
- **HTTP:** `net/http` com templates HTML server-side
- **Banco:** SQLite em producao inicial, com suporte configuravel a PostgreSQL
- **Sessoes:** `gorilla/sessions`
- **Seguranca:** bcrypt, CSRF, cookies HttpOnly/SameSite, RBAC e auditoria
- **Storage:** uploads locais com originais, WebP principal e thumbnails
- **Frontend publico/admin:** HTML templates, CSS proprio e JavaScript leve
- **Prototipo visual:** app React/Vite em `Inhumas em foco/app`

## Estrutura do Repositorio

```text
.
+-- Inhumas em foco/
|   +-- inhumas-em-foco/        # Aplicacao principal em Go
|   |   +-- cmd/                # Entrypoints: web, worker, seed e migracao
|   |   +-- internal/           # Dominio, handlers, repositorios, auth e servicos
|   |   +-- migrations/         # Schema versionado para PostgreSQL
|   |   +-- static/             # CSS, imagens e assets publicos
|   |   +-- scripts/            # Backup, smoke test, readiness e manutencao
|   |   +-- deploy/             # Exemplos de Nginx e systemd
|   |   +-- docs/               # Documentacao tecnica e operacional
|   +-- app/                    # Prototipo visual React/Vite
|   +-- Mockup/                 # Referencias visuais e prompts
+-- README.md                   # Visao profissional do projeto
```

## Modulos do CMS

### Editorial

- CRUD de noticias.
- Workflow de rascunho, revisao, aprovacao, publicacao e arquivamento.
- Autosave e lock de edicao.
- Preview de posts.
- Historico de revisoes.
- Checklist editorial.
- Campos SEO por conteudo.

### Taxonomia e Midia

- Categorias e tags.
- Upload seguro de imagens.
- Biblioteca de midia reutilizavel.
- Alt text, titulo e metadados.
- Conversao para WebP e thumbnails.

### Comercial e Comunidade

- Banners comerciais com posicao, periodo e relatorio.
- Lojas locais.
- Promocoes.
- Influencers.
- Eventos.
- Classificados.
- Bairros.

### Automacao e IA

- Cadastro de fontes de automacao.
- Coleta de conteudo por fontes externas.
- Criacao de rascunhos para curadoria humana.
- Logs de execucao.
- Provedor de IA editorial desacoplado, com guardrails para evitar publicacao automatica sem revisao.

### Operacao

- Health check em `/health`.
- Metricas operacionais em `/metrics`.
- Auditoria administrativa.
- Jobs recorrentes.
- Verificacao de backups.
- Scripts de smoke test e readiness.

## Requisitos

- Go 1.24 ou superior.
- PowerShell no Windows ou shell compatvel no Linux.
- SQLite para execucao inicial.
- PostgreSQL opcional para ambiente homologado ou producao futura.
- Nginx e systemd recomendados para deploy em VPS Linux.

## Como Rodar Localmente

Entre na aplicacao principal:

```powershell
cd "Inhumas em foco\inhumas-em-foco"
```

Configure variaveis minimas para desenvolvimento:

```powershell
$env:PORT="8080"
$env:APP_ENV="development"
$env:SESSION_SECRET="12345678901234567890123456789012"
$env:INITIAL_ADMIN_PASSWORD="admin123456"
$env:DATABASE_URL="./inhumas.db"
$env:DB_DRIVER="sqlite"
$env:UPLOAD_DIR="./uploads"
$env:STATIC_DIR="./static"
$env:PROJECT_ROOT="."
$env:ADMIN_PATH_PREFIX="/painel/7x9k2m"
```

Inicie o servidor:

```powershell
go run ./cmd/web
```

Acesse:

```text
http://localhost:8080
http://localhost:8080/login
http://localhost:8080/painel/7x9k2m
```

Admin inicial em ambiente local:

```text
E-mail: admin@inhumasemfoco.com.br
Senha:  admin123456
```

> Em producao, use uma senha inicial forte, entre no painel, crie/valide as contas administrativas e remova a variavel `INITIAL_ADMIN_PASSWORD` do ambiente.

## Worker

O worker executa rotinas de manutencao e jobs operacionais.

```powershell
go run ./cmd/worker
```

## Comandos Uteis

Dentro de `Inhumas em foco/inhumas-em-foco`:

```powershell
go test ./...
go build ./cmd/web ./cmd/worker ./cmd/seed-news ./cmd/migrate-sqlite-postgres
```

Se estiver usando `make`:

```bash
make dev
make worker
make build
make smoke
make readiness
```

## Variaveis de Ambiente

As principais variaveis estao documentadas em:

- `Inhumas em foco/inhumas-em-foco/.env.example`
- `Inhumas em foco/inhumas-em-foco/docs/ENVIRONMENT.md`

Variaveis essenciais:

| Variavel | Uso |
| --- | --- |
| `PORT` | Porta HTTP do backend. |
| `APP_ENV` | Ambiente atual. Use `production` na VPS. |
| `SITE_URL` | URL publica usada em sitemap, RSS e links absolutos. |
| `DB_DRIVER` | `sqlite` ou `postgres`. |
| `DATABASE_URL` | Caminho SQLite ou URL PostgreSQL. |
| `SESSION_SECRET` | Segredo de sessao com pelo menos 32 caracteres. |
| `INITIAL_ADMIN_PASSWORD` | Senha bootstrap do primeiro admin. Remover depois do primeiro uso. |
| `ADMIN_PATH_PREFIX` | Caminho protegido do painel administrativo. |
| `UPLOAD_DIR` | Diretorio de uploads. |
| `STATIC_DIR` | Diretorio de assets servidos em `/static`. |
| `METRICS_TOKEN` | Token para proteger `/metrics`. |
| `SMTP_*` | Configuracao de e-mail transacional e recuperacao de senha. |
| `BACKUP_*` | Configuracao de backup e verificacao operacional. |

## Banco de Dados

### SQLite

SQLite e o modo conservador para operacao inicial. Basta definir:

```env
DB_DRIVER=sqlite
DATABASE_URL=/var/www/inhumas/inhumas.db
```

### PostgreSQL

O projeto tambem suporta PostgreSQL por configuracao:

```env
DB_DRIVER=postgres
DATABASE_URL=postgres://usuario:senha@localhost:5432/inhumas?sslmode=disable
MIGRATIONS_DIR=/var/www/inhumas/migrations
```

Para migrar dados de SQLite para PostgreSQL:

```bash
SQLITE_DATABASE_URL=/var/www/inhumas/inhumas.db \
POSTGRES_DATABASE_URL='postgres://usuario:senha@localhost:5432/inhumas?sslmode=disable' \
MIGRATIONS_DIR=/var/www/inhumas/migrations \
/var/www/inhumas/bin/migrate-sqlite-postgres
```

## Qualidade e Testes

Antes de qualquer deploy ou commit relevante:

```powershell
cd "Inhumas em foco\inhumas-em-foco"
go test ./...
go build ./cmd/web ./cmd/worker ./cmd/seed-news ./cmd/migrate-sqlite-postgres
```

Para validacao operacional:

```bash
BASE_URL=https://inhumasemfoco.online \
ADMIN_PATH=/painel/7x9k2m \
BACKUP_DIR=/var/backups/inhumas \
./scripts/production-readiness.sh
```

## Deploy

O deploy recomendado para VPS usa:

- Binarios Linux gerados por `make build`.
- Nginx como proxy reverso.
- systemd para `inhumas-web` e `inhumas-worker`.
- `.env` protegido no servidor.
- Backups locais e, idealmente, backup externo fora da VPS.

Arquivos de apoio:

- `Inhumas em foco/inhumas-em-foco/deploy/nginx.conf.example`
- `Inhumas em foco/inhumas-em-foco/deploy/inhumas-web.service.example`
- `Inhumas em foco/inhumas-em-foco/deploy/inhumas-worker.service.example`
- `Inhumas em foco/inhumas-em-foco/docs/DEPLOY_GUIDE.md`
- `Inhumas em foco/inhumas-em-foco/docs/OPERATIONS_GUIDE.md`

## Seguranca

O projeto inclui medidas importantes para uso em producao:

- Hash de senha com bcrypt.
- Sessoes assinadas e cookies HttpOnly/SameSite.
- Rotacao gradual de segredo de sessao via `PREVIOUS_SESSION_SECRET`.
- CSRF em formularios administrativos.
- RBAC por papeis e permissoes.
- Auditoria de acoes administrativas.
- Rate limit para login.
- Validacao de uploads e protecao contra path traversal.
- Headers de seguranca e CSP com modo report-only opcional.
- Endpoint `/metrics` protegido por token.
- Caminho administrativo configuravel por `ADMIN_PATH_PREFIX`.

Arquivos sensiveis como `.env`, bancos locais, chaves privadas, backups e dados de VPS nao devem ser versionados.

## Documentacao

Documentos principais:

- `Inhumas em foco/inhumas-em-foco/docs/PROJECT_STATUS.md` - resumo executivo do que existe, o que faz e o que falta.
- `Inhumas em foco/inhumas-em-foco/docs/CHECKLIST_STATUS.md` - checklist vivo de execucao.
- `Inhumas em foco/inhumas-em-foco/docs/CMS_PLANO_GERAL.md` - plano macro do CMS.
- `Inhumas em foco/inhumas-em-foco/docs/CMS_REVISAO_FINAL.md` - revisao executiva e readiness.
- `Inhumas em foco/inhumas-em-foco/docs/ENVIRONMENT.md` - variaveis de ambiente.
- `Inhumas em foco/inhumas-em-foco/docs/DEPLOY_GUIDE.md` - deploy em VPS.
- `Inhumas em foco/inhumas-em-foco/docs/OPERATIONS_GUIDE.md` - rotina operacional, backup e incidentes.
- `Inhumas em foco/inhumas-em-foco/docs/SECURITY_PREPROD_CHECKLIST.md` - checklist de seguranca.
- `Inhumas em foco/inhumas-em-foco/docs/ARCHITECTURE_DECISIONS.md` - decisoes tecnicas.

## Status do Projeto

O CMS esta em estado avancado e entrou em marco multiportal persistente. A base Go concentra o produto principal, enquanto o app React permanece como prototipo visual/referencia.

Ja existem tenants, dominios, features por portal, usuarios por tenant, dados escopados por `tenant_id`, middleware de resolucao por host e worker tenant-aware para jobs editoriais/comerciais.

Pendencias recomendadas para evolucao corporativa:

- Configurar SMTP real para recuperacao de senha em producao.
- Configurar backup externo fora da VPS.
- Adicionar monitor externo de uptime.
- Homologar PostgreSQL com copia real antes da troca de runtime.
- Remover `unsafe-inline` da CSP apos migrar scripts/styles inline.
- Evoluir relatorios comerciais e analytics editorial.

## Licenca

Projeto privado/proprietario. O uso, copia, distribuicao ou publicacao depende de autorizacao do proprietario do repositorio.
