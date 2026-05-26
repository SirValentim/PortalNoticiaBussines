# Roadmap Summary

## Fase 1: Cerebro do Projeto

- Criar `docs/ai-context`.
- Criar ADRs iniciais.
- Criar mapa de dominio.
- Registrar dividas tecnicas e regras de engenharia.

## Fase 2: Multiportal Configuracional

- Consolidar branding por tenant.
- Criar smoke tests com segundo portal. Status: iniciado.
- Criar matriz de acesso admin por perfil. Status: iniciado.
- Garantir ausencia de hardcode em templates e handlers.
- Documentar assets e envs por portal.

## Fase 3: Multi-Tenant Persistente

- Criar tabelas `tenants` e `tenant_domains`. Status: iniciado.
- Criar tabela `tenant_features`. Status: iniciado; features `automation`, `media` e `commercial` ja aplicadas como gates operacionais.
- Criar tabela `tenant_users`. Status: iniciado.
- Definir estrategia de `tenant_id`.
- Escopar conteudo, midia, usuarios e settings.
- Criar middleware de resolucao por host. Status: iniciado.
- Escopar `portal_settings` por tenant. Status: iniciado.
- Escopar `categories` por tenant. Status: iniciado.
- Escopar `tags` por tenant. Status: iniciado.
- Escopar `posts` por tenant. Status: iniciado.
- Escopar `media_assets` por tenant. Status: iniciado.
- Escopar `banners` por tenant. Status: iniciado.
- Escopar `neighborhoods` por tenant. Status: iniciado.
- Escopar `stores` por tenant. Status: iniciado.
- Escopar `promotions` por tenant. Status: iniciado.
- Escopar `events` por tenant. Status: iniciado.
- Escopar `classifieds` por tenant. Status: iniciado.
- Escopar `influencers` por tenant. Status: iniciado.
- Escopar `metrics` por tenant. Status: iniciado.
- Escopar `users` por tenant. Status: iniciado.
- Escopar automacao editorial por tenant. Status: iniciado.
- Escopar jobs editoriais fora do HTTP por tenant. Status: iniciado.

## Fase 4: Modularizacao Real

- Extrair services por dominio.
- Criar DTOs e validators.
- Separar repository por dominio quando o ganho for claro.
- Reduzir `handler.go`.

## Fase 5: SaaS

- Billing.
- Planos e limites.
- Provisionamento de portal.
- Observabilidade por tenant.
- Backups e operacao multi-tenant.
