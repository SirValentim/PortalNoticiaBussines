# Multi-Tenant Migration Plan

## Objetivo

Evoluir de portal unico com branding por env para plataforma multi-tenant persistente, sem quebrar a operacao atual.

## Principio

Adicionar isolamento por tenant em ondas pequenas, com backfill e testes por dominio.

## Onda 1: Tabelas Base

Criar nova migration com:

- `tenants`: feito em `002_tenants.sql`.
- `tenant_domains`: feito em `002_tenants.sql`.
- `tenant_features`: feito em `019_tenant_features.sql`.

Campos sugeridos para `tenants`:

- `id`
- `name`
- `slug`
- `status`
- `primary_domain`
- `created_at`
- `updated_at`

Campos sugeridos para `tenant_domains`:

- `id`
- `tenant_id`
- `domain`
- `is_primary`
- `created_at`

Campos implementados para `tenant_features`:

- `id`
- `tenant_id`
- `feature`
- `enabled`
- `limit_value`
- `created_at`
- `updated_at`

## Onda 2: Resolver Tenant por Request

- Criar `internal/tenant`. Status: iniciado com helpers de contexto e normalizacao de dominio.
- Resolver tenant por host/header em middleware. Status: iniciado com fallback para `default`.
- Manter fallback para tenant default enquanto a base ainda for single-tenant.

## Onda 3: Escopar Configuracoes

- Adicionar `tenant_id` em `portal_settings`. Status: iniciado em `003_portal_settings_tenant.sql`.
- Backfill do tenant default. Status: iniciado.
- Atualizar repository de settings. Status: iniciado.

## Onda 4: Escopar Conteudo

Adicionar `tenant_id` gradualmente em:

- `categories`: iniciado em `004_categories_tenant.sql`.
- `tags`: iniciado em `005_tags_tenant.sql`.
- `posts`: iniciado em `006_posts_tenant.sql`.
- `media_assets`: iniciado em `007_media_assets_tenant.sql`.
- `banners`: iniciado em `008_banners_tenant.sql`.
- `neighborhoods`: iniciado em `009_neighborhoods_tenant.sql`.
- `stores`: iniciado em `010_stores_tenant.sql`.
- `promotions`: iniciado em `011_promotions_tenant.sql`.
- `events`: iniciado em `012_events_tenant.sql`.
- `classifieds`: iniciado em `013_classifieds_tenant.sql`.
- `influencers`: iniciado em `014_influencers_tenant.sql`.

## Onda 5: Usuarios e Permissoes por Tenant

Metricas: iniciado em `015_metrics_tenant.sql`.
Usuarios: iniciado em `016_users_tenant.sql` com a opcao simples `users.tenant_id` e e-mail unico por tenant.
Automacao editorial: iniciado em `017_automation_tenant.sql` para `automation_sources` e `automation_runs`.
Usuarios multiportal: iniciado em `020_tenant_users.sql` como tabela de vinculo usuario-tenant.
Permissao de plataforma: `/tenants` usa `tenants:manage`, reservada a `super_admin`; `admin` nao gerencia tenants da plataforma.

Definir se usuario pode pertencer a multiplos tenants:

- opcao simples: `users.tenant_id`. Status: iniciado.
- opcao SaaS: tabela `tenant_users`. Status: iniciada no repository e usada por login/RBAC para papel efetivo por tenant.

Para SaaS real, preferir `tenant_users`; a gestao inicial de vinculos multiportal ja existe no painel de portais com entrada por email ou ID de usuario.

## Onda 5.1: Provisionamento Inicial de Portal

- Criacao pelo painel gera tenant e provisiona dominio principal quando informado.
- `portal_settings` inicial recebe nome, SEO basico e email de contato derivado do dominio.
- Features padrao sao criadas: `automation`, `media` e `commercial`.
- `automation` ja e feature operacional: painel bloqueia acoes quando desabilitada e o worker nao agenda nem executa `collect_news` para esse tenant.
- `media` ja e feature operacional: painel bloqueia biblioteca/upload/edicao/remocao de midia quando desabilitada.
- `commercial` ja e feature operacional: painel bloqueia banners, lojas, promocoes, eventos, classificados, influencers e bairros quando desabilitada.

## Onda 6: Jobs Recorrentes Fora do HTTP

- Adicionar tenant explicito no payload ou na tabela `jobs` quando o job processar dados de portal. Status: iniciado em `018_jobs_tenant.sql` com `jobs.tenant_id` e `dead_jobs.tenant_id`.
- Executar `collect_news` por tenant com automacao habilitada, em vez de apenas pelo fallback default. Status: iniciado no worker.
- Respeitar `tenant_features.automation` antes de agendar ou executar `collect_news`. Status: iniciado no worker e coberto por teste no painel.
- Manter jobs operacionais globais, como backup e vacuum, sem tenant quando fizer sentido. Status: politica explicitada no worker com testes.

## Riscos

- Vazamento de conteudo entre tenants por query sem filtro.
- Backfill incompleto.
- Rotas admin acessando dados sem escopo.
- Jobs processando registros de outro tenant.

## Testes Minimos

- Tenant A nao lista conteudo do Tenant B.
- Admin de Tenant A nao acessa settings do Tenant B.
- Sitemap e RSS respeitam tenant atual.
- Branding por dominio funciona sem alterar binario.
- Pagina publica por host usa settings do tenant resolvido e fallback default sem vazamento. Status: coberto por teste local.
