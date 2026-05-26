# Project Status - Marco Multiportal

Data: 2026-05-26

Este documento resume, de forma direta, o que o projeto e hoje, o que ele faz de verdade e o que ainda falta para a proxima etapa.

## O Que O Projeto E

O Inhumas em Foco deixou de ser apenas um portal local isolado. O projeto agora e uma plataforma CMS editorial em Go, com base multiportal persistente, capaz de operar diferentes portais no mesmo binario, separando conteudo, configuracoes, usuarios, features e jobs por tenant.

O primeiro portal continua sendo o Inhumas em Foco, mas a arquitetura ja prepara a plataforma para outros portais locais, editoriais ou comerciais, como LaMafiaMusic, GoiasNews e futuros produtos SaaS editoriais.

## O Que O Projeto Faz Hoje

- Serve o portal publico com noticias, categorias, tags, busca, eventos, classificados, lojas, promocoes, bairros, influencers e paginas institucionais.
- Oferece painel administrativo protegido por login, sessoes seguras, CSRF, RBAC e caminho admin configuravel.
- Gerencia conteudo editorial com posts, status, preview, categorias, tags, midia, SEO e publicacao.
- Gerencia modulos comerciais e locais: banners, lojas, promocoes, eventos, classificados, influencers e bairros.
- Processa uploads com validacao, protecao de path, WebP e thumbnails.
- Gera sitemap, robots, RSS, manifest, Open Graph, Twitter Cards e JSON-LD.
- Executa worker de jobs recorrentes, com politica clara entre jobs globais e jobs por tenant.
- Mantem automacao editorial por fontes, criando rascunhos para revisao humana.
- Expoe servidor MCP para integracao editorial controlada.
- Entrega health check, metricas protegidas, scripts de backup, smoke test e readiness operacional.
- Suporta SQLite no runtime inicial e PostgreSQL por driver/migrations.
- Resolve tenant por dominio, com fallback para o tenant `default`.
- Escopa por `tenant_id` os principais dominios de dados: settings, categorias, tags, posts, midia, banners, bairros, lojas, promocoes, eventos, classificados, influencers, metricas, usuarios, automacao e jobs.
- Permite gestao de portais no painel por usuario `super_admin`, incluindo dominios, features e vinculos de usuarios por portal.
- Usa features por tenant para controlar automacao, midia e modulos comerciais/locais.

## O Que Foi Consolidado Neste Marco

- Base persistente de tenants e dominios.
- Middleware de tenant por host.
- Contexto de tenant no request.
- Painel admin de portais.
- Provisionamento inicial de portal.
- Features operacionais por tenant.
- RBAC efetivo por vinculo `tenant_users`.
- Escopo de dados por `tenant_id` nos principais repositories.
- Worker tenant-aware para jobs editoriais/comerciais.
- Migrations PostgreSQL versionadas de `002` a `020`.
- Testes de isolamento, features, migrations, middleware, RBAC e worker.
- Documentacao AI-first em `docs/ai-context`, arquitetura, banco, backend, decisoes e roadmap.

## O Que Falta Implementar

### Alta Prioridade

- Homologar as migrations em PostgreSQL real usando `POSTGRES_TEST_URL`.
- Rodar smoke test completo em VPS/staging com dominios reais.
- Configurar SMTP real para recuperacao de senha e e-mails transacionais.
- Configurar backup externo fora da VPS.
- Configurar monitor externo de uptime e alerta.
- Revisar queries criticas em ambiente com dois ou mais tenants reais.

### Media Prioridade

- Extrair DTOs e validadores dos formularios admin mais sensiveis: posts, settings, banners, media e users.
- Extrair service de SEO para reduzir responsabilidade dos handlers.
- Separar gradualmente repositories por dominio quando as interfaces ficarem estaveis.
- Remover `unsafe-inline` da CSP depois de migrar scripts/styles inline.
- Melhorar relatorios comerciais e analytics editorial.
- Criar rotinas de exportacao/importacao por tenant.

### Baixa Prioridade

- Criar tema visual separado por portal.
- Publicar documentacao externa com Docusaurus ou equivalente.
- Avaliar Redis para cache, filas leves ou rate limit distribuido quando houver necessidade real.
- Avaliar billing somente depois da operacao multiportal estar homologada.

## O Que Nao Fazer Agora

- Reescrever o projeto inteiro para `internal/core` em uma unica mudanca.
- Trocar server-side rendering por SPA sem justificativa operacional.
- Publicar conteudo automatizado sem aprovacao humana.
- Alterar migrations antigas em vez de criar novas migrations versionadas.
- Criar billing antes de validar tenants reais em producao/staging.

## Como Validar O Marco

```powershell
cd "Inhumas em foco\inhumas-em-foco"
go test ./...
go build ./cmd/web ./cmd/worker ./cmd/seed-news ./cmd/migrate-sqlite-postgres
```

Validacao PostgreSQL opcional:

```powershell
$env:POSTGRES_TEST_URL="postgres://usuario:senha@localhost:5432/inhumas_test?sslmode=disable"
go test ./internal/repository -run TestPostgresMigrationsApplyWhenURLProvided -v
```

## Proximo Marco Recomendado

O proximo marco deve ser "homologacao multiportal real": subir dois tenants com dominios/hosts distintos, conteudo separado, features diferentes e usuario com papeis diferentes por portal. Depois disso, a plataforma passa de base tecnica multiportal para operacao multiportal comprovada.
