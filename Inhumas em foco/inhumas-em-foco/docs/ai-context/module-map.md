# Module Map

## `cmd/web`

Inicializa aplicacao HTTP, banco, sessoes, auth, storage, handler e middleware.

## `cmd/worker`

Processa jobs recorrentes e pendentes.

## `cmd/mcp`

Exponibiliza ferramentas editoriais controladas para clientes MCP.

## `internal/config`

Carrega configuracao de ambiente e tenant branding.

## `internal/middleware`

Cross-cutting HTTP: seguranca, CSRF, auth, RBAC admin, logs, metricas, timeout, manutencao e branding context.

## `internal/handler`

Rotas publicas, admin, renderizacao, SEO, RSS, sitemap, manifest e APIs simples.

## `internal/repository`

Persistencia SQL, migrations runtime e consultas por dominio.

## `internal/model`

Modelos compartilhados atuais.

## `internal/auth`, `internal/users`, `internal/session`

Autenticacao, autorizacao, usuarios admin, recuperacao de senha e sessoes.

## `internal/storage`

Upload local, validacao, geracao de WebP/thumbs e limpeza de originais.

## `internal/automation`

Coleta RSS e criacao de rascunhos editoriais.

## `internal/editorialai`

Provider de IA editorial, atualmente mockado.

## `internal/commercial`

Regras comerciais de banners/promocoes/status.

## `internal/posts`

Regras editoriais de posts, status e permissoes de edicao.

## `internal/view`

Layouts, componentes e paginas HTML.
