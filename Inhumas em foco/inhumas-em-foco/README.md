# NewsCore CMS - Backend Go SSR

Backend principal do NewsCore CMS, uma plataforma premium multiportal para gestao profissional de portais de noticias. A aplicacao oficial usa Go, templates HTML server-side, storage local e SQLite no runtime inicial, com base multiportal persistente e suporte configuravel a PostgreSQL.

O Inhumas em Foco continua existindo como primeiro portal/tenant cadastrado. Ele nao e a marca principal do sistema administrativo.

## Estado Atual

- Servidor HTTP em `cmd/web`.
- Worker de jobs em `cmd/worker`.
- Servidor MCP em `cmd/mcp` para integracao editorial controlada com clientes compativeis.
- Base multiportal persistente com tenants, dominios, features, usuarios por portal e dados por `tenant_id`.
- Painel admin NewsCore CMS protegido por `ADMIN_PATH_PREFIX`.
- Auth com bcrypt, sessoes HttpOnly/SameSiteStrict, CSRF e RBAC editorial/comercial.
- Upload local com validacao de MIME, limite de corpo e protecao contra path traversal.
- O limite de upload tambem e aplicado durante a validacao CSRF de formularios multipart.
- Novos uploads geram original preservado, WebP principal e thumb WebP.
- RSS publico em `/rss.xml`.
- Sitemap, robots, JSON-LD, metricas operacionais e health check estao implementados.
- PostgreSQL e suportado por configuracao, com migrations versionadas e comando de migracao SQLite -> PostgreSQL; a producao atual segue em SQLite.

## Documentos Oficiais do CMS

A evolucao do CMS premium passa a usar estes documentos como fonte principal:

- `.codex-context.md`: contexto persistente rapido para IA e engenharia.
- `docs/PROJECT_STATUS.md`: resumo executivo do que existe, o que o projeto faz e o que falta.
- `docs/ai-context/`: cerebro vivo do projeto, arquitetura atual, regras, mapa de modulos, roadmap e dividas tecnicas.
- `docs/CHECKLIST_STATUS.md`: checklist vivo de execucao, status e historico.
- `docs/CMS_PLANO_GERAL.md`: cronograma macro, sprints e criterios de pronto.
- `docs/CMS_REVISAO_FINAL.md`: revisao executiva final, readiness operacional e backlog residual.

Os demais documentos em `docs/` ficam como referencia operacional ou historica. Quando houver conflito, prevalecem o checklist vivo, o plano geral e a revisao final.

## Rodar Localmente

```powershell
$env:SESSION_SECRET="12345678901234567890123456789012"
$env:INITIAL_ADMIN_PASSWORD="admin123456"
$env:DATABASE_URL="./inhumas.db"
$env:UPLOAD_DIR="./uploads"
go run ./cmd/web
```

Abrir:

```text
http://localhost:8080/login
```

Admin local inicial:

```text
admin@inhumasemfoco.com.br
admin123456
```

O usuario inicial criado por `INITIAL_ADMIN_PASSWORD` nasce como `super_admin`, para permitir a gestao da plataforma multiportal. Usuarios `admin` gerenciam o portal atual, mas nao a lista global de portais.

## Worker

```powershell
go run ./cmd/worker
```

## Testes

```powershell
go test ./...
```

## Build e readiness

```powershell
go build ./cmd/web ./cmd/worker ./cmd/seed-news ./cmd/migrate-sqlite-postgres ./cmd/mcp
```

Para compilar o servidor MCP:

```powershell
go build -o bin/inhumas-mcp.exe ./cmd/mcp
```

Na VPS, depois de deploy ou manutencao:

```bash
BASE_URL=https://inhumasemfoco.online \
ADMIN_PATH=/painel/1c2dhviax7 \
BACKUP_DIR=/var/backups/inhumas \
/var/www/inhumas/scripts/production-readiness.sh
```

## Variaveis Principais

- `PORT`: porta HTTP, padrao `8080`.
- `SITE_URL`: URL publica usada no sitemap.
- `DATABASE_URL`: caminho SQLite no runtime atual.
- `DB_DRIVER`: `sqlite`, `postgres` ou vazio para autodeteccao.
- `MIGRATIONS_DIR`: diretorio de migrations versionadas.
- `SESSION_SECRET`: segredo com pelo menos 32 caracteres.
- `PREVIOUS_SESSION_SECRET`: segredo anterior para rotacao gradual de sessoes.
- `INITIAL_ADMIN_PASSWORD`: cria admin inicial se ainda nao houver usuarios.
- `ADMIN_PATH_PREFIX`: caminho obscuro do painel.
- `UPLOAD_DIR`: pasta de uploads.
- `STATIC_DIR`: pasta servida em `/static` e destino do `sitemap.xml`.
- `PROJECT_ROOT`: raiz do projeto usada para localizar templates HTML.
- `METRICS_TOKEN`: token para acesso ao endpoint operacional `/metrics`.
- `BACKUP_DIR`: pasta onde backups recentes sao verificados pelo health.
- `BACKUP_SCRIPT`: script executado pelo job recorrente de backup.
- `SMTP_HOST`, `SMTP_PORT`, `SMTP_USERNAME`, `SMTP_PASSWORD`, `SMTP_FROM`, `SMTP_FROM_NAME`: envio de e-mails transacionais, incluindo recuperacao de senha.
- `MAX_UPLOAD_SIZE`: limite maximo de upload em bytes.
- `BCRYPT_COST`: custo bcrypt.
- `ORIGINAL_RETENTION_DAYS`: dias de retencao dos originais de upload.
- `APP_ENV`: use `production` em producao.
- `FORCE_SECURE_COOKIES`: use `true` para forcar cookies `Secure`.
- `MAINTENANCE_MODE`: ativa modo manutencao nas rotas publicas.
- `CSP_REPORT_ONLY`, `CSP_REPORT_URI`: auditoria de CSP mais restritiva.
- `RCLONE_REMOTE`, `OFFSITE_BACKUP_DIR`: destino de backup externo.
- `MAX_BACKUP_AGE_HOURS`, `DISK_ALERT_THRESHOLD`, `JOURNAL_RETENTION`: operacao e alertas.

## Health Check

```text
GET /health
```

Retorna status geral e checks de banco, uploads, jobs e backup configurado. Se algum check critico falhar, responde `503`.

## Metricas Operacionais

```text
GET /metrics
Authorization: Bearer <METRICS_TOKEN>
```

Retorna uptime, total de requests, conexoes do banco, jobs pendentes/dead e tamanho atual da pasta de uploads.

## Documentacao Operacional

- `docs/DEPLOY_GUIDE.md`: deploy passo a passo na VPS.
- `docs/OPERATIONS_GUIDE.md`: rotina diaria, backup, restauracao e incidentes.
- `docs/SECURITY_PREPROD_CHECKLIST.md`: checklist antes de expor em producao.
- `docs/ENVIRONMENT.md`: referencia completa de variaveis de ambiente.
- `docs/MCP_INTEGRATION.md`: uso do servidor MCP, build e exemplos de configuracao.
- `docs/CMS_REVISAO_FINAL.md`: status executivo, readiness e backlog residual.
- `docs/ARCHITECTURE_DECISIONS.md`: decisoes do MVP, incluindo React como prototipo visual.

## Status

O CMS Go esta em marco multiportal persistente: tenants, dominios, features, usuarios por portal, dados por `tenant_id` e jobs tenant-aware ja existem no codigo. O backlog residual principal agora e homologacao real: PostgreSQL/staging, dominios reais, backup externo, monitor externo, SMTP real e revisao final de isolamento com tenants reais.
