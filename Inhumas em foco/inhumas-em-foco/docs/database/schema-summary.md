# Database Schema Summary

Baseado em `migrations/001_init_postgres.sql`.

## Identidade e Acesso

- `users`: usuarios administrativos e papeis, escopados por tenant.
- `password_reset_tokens`: tokens de recuperacao de senha.
- `login_attempts`: auditoria e rate limit de login.
- `audit_logs`: trilha de auditoria administrativa.

## CMS Editorial

- `posts`: noticias/artigos.
- `post_revisions`: historico editorial.
- `categories`: categorias.
- `tags`: tags.
- `post_tags`: relacao N:N entre posts e tags.
- `slug_redirects`: redirecionamentos de slugs antigos.
- `edit_locks`: lock colaborativo de edicao.

## Midia

- `media_assets`: arquivos enviados, metadados, alt text e variantes.

## Portal e Branding Persistido

- `tenants`: entidades de portal/cliente da plataforma.
- `tenant_domains`: dominios associados a cada tenant.
- `portal_settings`: configuracoes editoriais alteraveis no painel.

Branding multiportal atual e carregado por env. `portal_settings` ja esta preparado para escopo por `tenant_id`.

## Automacao e IA

- `automation_sources`: fontes RSS, escopadas por tenant.
- `automation_runs`: execucoes de automacao, escopadas por tenant.
- `tenant_features.automation`: controla disponibilidade da automacao no painel e no worker.
- `tenant_features.media`: controla disponibilidade da biblioteca de midia no painel.
- `tenant_features.commercial`: controla disponibilidade dos modulos comerciais/locais no painel.
- `ai_usage_logs`: logs de uso de IA editorial.
- `jobs`: fila de jobs, com tenant quando processa dados de portal.
- `dead_jobs`: jobs que excederam tentativas, preservando tenant.

## Modulos Locais e Comerciais

- `neighborhoods`: bairros.
- `influencers`: influencers.
- `stores`: lojas.
- `promotions`: promocoes.
- `events`: eventos.
- `classifieds`: classificados.
- `banners`: anuncios/banners.

## Metricas

- `metrics`: eventos de visualizacao, clique e metricas simples por entidade.

## Lacuna Multi-Tenant

Ja existe base inicial:

- `tenants`
- `tenant_domains`
- `tenant_features`
- `tenant_users`
- `portal_settings.tenant_id`
- `categories.tenant_id`
- `tags.tenant_id`
- `posts.tenant_id`
- `media_assets.tenant_id`
- `banners.tenant_id`
- `neighborhoods.tenant_id`
- `stores.tenant_id`
- `promotions.tenant_id`
- `events.tenant_id`
- `classifieds.tenant_id`
- `influencers.tenant_id`
- `metrics.tenant_id`
- `users.tenant_id`
- `automation_sources.tenant_id`
- `automation_runs.tenant_id`
- `jobs.tenant_id`
- `dead_jobs.tenant_id`

Politica de jobs:

- jobs editoriais/comerciais que alteram dados de portal carregam tenant explicito: publicar post, expirar promocao, expirar banner e coletar noticias;
- jobs operacionais globais rodam sem contexto de portal: backup, vacuum, limpeza de jobs, compressao de uploads e sitemap estatico de compatibilidade.

Essa mudanca deve ser incremental e acompanhada de backfill.

Validacao PostgreSQL:

- `go test ./internal/repository -run TestMigrationFilesAreContiguous -v` valida a sequencia local das migrations;
- `POSTGRES_TEST_URL=... go test ./internal/repository -run TestPostgresMigrationsApplyWhenURLProvided -v` aplica as migrations em schema temporario de PostgreSQL.
