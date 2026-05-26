# Current State

## Ja Implementado

- Servidor web em `cmd/web`.
- Worker de jobs recorrentes em `cmd/worker`.
- Servidor MCP em `cmd/mcp`.
- Autenticacao com bcrypt, sessoes seguras, CSRF e RBAC.
- Painel administrativo com dashboard, usuarios, auditoria, settings e metricas.
- CRUD editorial de posts, categorias, tags, midia e preview.
- Modulos locais/comerciais: lojas, promocoes, eventos, classificados, influencers, bairros e banners.
- Automacao editorial por fontes RSS com criacao de rascunhos e logs.
- IA editorial mockada para sugestoes de texto.
- Sitemap, robots, RSS, JSON-LD, Open Graph e Twitter Cards.
- Storage local com validacao de upload, WebP e thumbnails.
- Branding multiportal por variaveis de ambiente via `TenantBrandingConfig`.
- Smoke tests de paginas publicas e branding alternativo.
- Testes de matriz RBAC admin por perfil e secao.
- Base persistente inicial de tenants: migration `002_tenants.sql`, repository e helpers `internal/tenant`.
- Base de features por tenant: migration `019_tenant_features.sql`, model e repository.
- Feature `automation` por tenant ja controla painel e worker: quando desabilitada, bloqueia CRUD/execucao manual da automacao e impede agendamento/execucao de `collect_news`.
- Features `media` e `commercial` tambem sao gates operacionais no admin: bloqueiam biblioteca de midia e modulos comerciais/locais quando desabilitadas.
- Base de usuarios multiportal: migration `020_tenant_users.sql`, model e repository; login/RBAC usam o papel efetivo de `tenant_users` quando existe vinculo no tenant atual.
- Painel admin de portais em `/tenants` para criar/editar tenants, dominios, features e vinculos `tenant_users` com sugestoes por email/nome.
- Gestao de portais usa permissao dedicada `tenants:manage`, disponivel para `super_admin`; `admin` continua restrito a administracao do portal atual.
- Criacao de portal faz provisionamento inicial: dominio principal, `portal_settings` basico e features padrao.
- Middleware de resolucao de tenant por dominio com fallback para `default`.
- Smoke multiportal por host cobre pagina publica usando settings do tenant correto e fallback sem vazamento entre tenants.
- SQLite `:memory:` em testes limitado a uma conexao para evitar bancos isolados por pool.
- `portal_settings` escopado por `tenant_id` no repository e na migration `003_portal_settings_tenant.sql`.
- `categories` escopado por `tenant_id` no repository e na migration `004_categories_tenant.sql`.
- `tags` escopado por `tenant_id` no repository e na migration `005_tags_tenant.sql`.
- `posts` escopado por `tenant_id` no repository e na migration `006_posts_tenant.sql`.
- `media_assets` escopado por `tenant_id` no repository e na migration `007_media_assets_tenant.sql`.
- `banners` escopado por `tenant_id` no repository e na migration `008_banners_tenant.sql`.
- `neighborhoods` escopado por `tenant_id` no repository e na migration `009_neighborhoods_tenant.sql`.
- `stores` escopado por `tenant_id` no repository e na migration `010_stores_tenant.sql`.
- `promotions` escopado por `tenant_id` no repository e na migration `011_promotions_tenant.sql`.
- `events` escopado por `tenant_id` no repository e na migration `012_events_tenant.sql`.
- `classifieds` escopado por `tenant_id` no repository e na migration `013_classifieds_tenant.sql`.
- `influencers` escopado por `tenant_id` no repository e na migration `014_influencers_tenant.sql`.
- `metrics` escopado por `tenant_id` no repository e na migration `015_metrics_tenant.sql`.
- `users` escopado por `tenant_id` no repository e na migration `016_users_tenant.sql`.
- `automation_sources` e `automation_runs` escopados por `tenant_id` no repository e na migration `017_automation_tenant.sql`.
- `jobs` e `dead_jobs` preservam `tenant_id` no repository, no worker e na migration `018_jobs_tenant.sql`.
- Homologacao PostgreSQL preparada com teste opcional `POSTGRES_TEST_URL` e validacao estatica de sequencia das migrations.

## Em Desenvolvimento

- Homologacao multiportal real com dois ou mais tenants em staging/producao.
- Separacao gradual de handlers grandes em services e DTOs.
- Rodar homologacao PostgreSQL em ambiente real de banco quando `POSTGRES_TEST_URL` estiver configurado.
- Melhor isolamento entre core reutilizavel e regras especificas de portal.
- Organizacao documental AI-first.

## Problemas Atuais

- `internal/handler/handler.go` concentra muitas rotas, renderizacao, SEO e regras de apresentacao.
- `internal/model/model.go` centraliza muitos modelos de dominios diferentes.
- `internal/repository` concentra acesso a tabelas de varios dominios.
- Algumas regras ainda vivem em handlers por historico do MVP.
- O banco tem tabelas base de tenant, settings, conteudo editorial, midia, modulos comerciais/locais, metricas, usuarios, automacao editorial e jobs por tenant; o risco restante e validar isolamento com dados reais em staging.
- Jobs fora do HTTP tem politica explicita: jobs editoriais/comerciais carregam tenant; jobs operacionais globais rodam sem contexto de portal.
- Existem documentos antigos em `docs/` que podem conflitar se nao houver uma fonte viva clara.

## Regra de Evolucao

Nao mover tudo para `internal/core` de uma vez. Primeiro documentar, testar e criar interfaces pequenas; depois migrar modulo por modulo.
