# Inhumas em Foco - Backend Go

Backend principal do portal local Inhumas em Foco. A aplicacao oficial usa Go, templates HTML server-side, storage local e SQLite em producao inicial, com evolucao planejada para um CMS editorial premium.

## Estado Atual

- Servidor HTTP em `cmd/web`.
- Worker de jobs em `cmd/worker`.
- Painel admin protegido por `ADMIN_PATH_PREFIX`.
- Auth com bcrypt, sessoes HttpOnly/SameSiteStrict, CSRF e RBAC editorial/comercial.
- Upload local com validacao de MIME, limite de corpo e protecao contra path traversal.
- O limite de upload tambem e aplicado durante a validacao CSRF de formularios multipart.
- Novos uploads geram original preservado, WebP principal e thumb WebP.
- RSS publico em `/rss.xml`.
- Sitemap, robots, JSON-LD, metricas operacionais e health check estao implementados.
- Migration PostgreSQL inicial existe, mas PostgreSQL fica classificado como evolucao planejada do CMS.

## Documentos Oficiais do CMS

A evolucao do CMS premium passa a usar estes documentos como fonte principal:

- `docs/CHECKLIST_STATUS.md`: checklist vivo de execucao, status e historico.
- `docs/CMS_PLANO_GERAL.md`: cronograma macro, sprints e criterios de pronto.

Os demais documentos em `docs/` ficam como referencia historica/operacional ate serem consolidados nesses dois arquivos.

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

## Worker

```powershell
go run ./cmd/worker
```

## Testes

```powershell
go test ./...
```

## Variaveis Principais

- `PORT`: porta HTTP, padrao `8080`.
- `SITE_URL`: URL publica usada no sitemap.
- `DATABASE_URL`: caminho SQLite no runtime atual.
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
- `docs/ARCHITECTURE_DECISIONS.md`: decisoes do MVP, incluindo React como prototipo visual.

## Status

O MVP operacional em Go + SQLite esta publicado e validado na VPS. A fase atual transforma essa base em um CMS corporativo premium, com foco em permissoes por acao, fluxo editorial, biblioteca de midia, automacoes, IA editorial, anuncios, modulos locais, PostgreSQL e hardening operacional.
