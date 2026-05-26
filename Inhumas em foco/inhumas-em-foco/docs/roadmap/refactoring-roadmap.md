# Refactoring Roadmap

## Objetivo

Reduzir acoplamento sem interromper a evolucao do CMS.

## Fase 0: Protecao

- Manter `go test ./...` verde.
- Smoke tests das paginas publicas principais. Status: iniciado.
- Smoke test de branding alternativo. Status: iniciado.
- Testes de matriz RBAC admin por perfil. Status: iniciado.

## Fase 1: DTOs e Validacao

Prioridade:

- posts
- media
- settings
- banners
- users

Cada formulario deve ter DTO, normalizacao e validacao fora do handler.

## Fase 2: Services de Dominio

Extrair regras novas ou tocadas para:

- `posts`
- `commercial`
- `users`
- `automation`
- futuro `seo`
- futuro `tenant`

## Fase 3: Repository por Dominio

Separar apenas quando houver fronteira clara:

- `PostRepository`
- `MediaRepository`
- `TenantRepository`
- `MetricsRepository`

## Fase 4: Router por Modulo

Depois dos services estarem estaveis:

- `RegisterPublicRoutes`
- `RegisterAdminPostRoutes`
- `RegisterAdminCommercialRoutes`
- `RegisterAdminSystemRoutes`

## Fase 5: Core/Portals

Mover modulos estabilizados para `internal/core`.

Criar `internal/portals` apenas para customizacoes reais, evitando duplicacao artificial.

## Fase Multi-Tenant Persistente

- Tabelas `tenants` e `tenant_domains`. Status: iniciado.
- Tabela `tenant_features`. Status: iniciado.
- Tabela `tenant_users`. Status: iniciado.
- Helper `internal/tenant` para contexto/dominio. Status: iniciado.
- Resolver tenant por host em middleware. Status: iniciado.
- Escopar `portal_settings` por tenant. Status: iniciado.
- Escopar categorias por tenant. Status: iniciado.
- Escopar tags por tenant. Status: iniciado.
- Escopar posts por tenant. Status: iniciado.
- Escopar midia por tenant. Status: iniciado.
- Escopar banners por tenant. Status: iniciado.
- Escopar bairros por tenant. Status: iniciado.
- Escopar lojas por tenant. Status: iniciado.
- Escopar promocoes por tenant. Status: iniciado.
- Escopar eventos por tenant. Status: iniciado.
- Escopar classificados por tenant. Status: iniciado.
- Escopar influencers por tenant. Status: iniciado.
- Escopar metricas por tenant. Status: iniciado.
- Escopar usuarios por tenant. Status: iniciado.
- Escopar automacao editorial por tenant. Status: iniciado.
- Escopar jobs editoriais fora do HTTP por tenant. Status: iniciado.
- Definir politica de jobs globais operacionais versus jobs por tenant. Status: iniciado.
- Expor gestao de `tenant_users` no painel. Status: iniciado.
- Autocomplete de usuarios no painel de portais. Status: iniciado.
- Provisionamento inicial de portal pelo painel. Status: iniciado.
