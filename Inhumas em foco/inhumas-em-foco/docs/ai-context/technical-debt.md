# Technical Debt

## Alta Prioridade

- `internal/handler/handler.go` e grande demais e mistura dominios.
- Faltam DTOs explicitos para muitos formularios admin.
- Multi-tenant persistente ja existe, mas ainda precisa homologacao com tenants reais e revisao de queries criticas.
- `internal/model/model.go` centraliza muitos modelos.
- Repository central pode virar gargalo de manutencao.

## Media Prioridade

- Alguns titulos/tags de SEO ainda tem copy local generica em handlers.
- CSP ainda permite `unsafe-inline` por compatibilidade com templates/scripts.
- Redis ainda nao existe para rate limit distribuido/cache.
- PostgreSQL precisa homologacao operacional completa.

## Baixa Prioridade

- Separar templates por dominio ou modulo.
- Criar pacote `internal/core` apenas quando houver fronteiras estabilizadas.
- Docusaurus ou publicacao externa de docs.

## Quick Wins

- Criar DTOs para formularios mais sensiveis.
- Extrair SEO helpers para um service.
- Extrair renderizacao de manifest/RSS para helpers testados.
- Criar tests de smoke para paginas publicas principais com branding alternativo.
