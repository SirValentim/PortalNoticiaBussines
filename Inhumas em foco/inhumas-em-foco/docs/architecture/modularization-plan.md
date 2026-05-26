# Modularization Plan

## Principio

Refatorar em ondas pequenas. A meta e reduzir acoplamento sem quebrar o portal em producao.

## Arvore Futura Desejada

```text
internal/
  core/
    auth/
    cms/
    media/
    seo/
    automation/
    commercial/
    tenant/
  portals/
    inhumasemfoco/
    lamafiamusic/
    goiasnews/
```

## Plano de Migracao

1. Documentar estado atual e regras.
2. Criar tests de smoke para rotas criticas.
3. Extrair DTOs e validators dos formularios admin.
4. Extrair services de SEO/render auxiliar.
5. Mover regras editoriais/comerciais restantes para services existentes.
6. Homologar dominio `tenant` persistente com tenants reais.
7. Separar repository por dominio apenas quando houver interfaces claras.
8. Introduzir `internal/core` como destino de modulos estabilizados.

## Quick Wins

- Teste de branding alternativo.
- Service de SEO para montar `SEOData`.
- DTOs para posts, banners e settings.
- ADR para monolito modular e multi-tenant incremental.

## Riscos

- Refatoracao ampla de diretorios quebrar imports e templates.
- Criar `core/portals` cedo demais e duplicar codigo.
- Avancar multiportal sem teste em staging com mais de um tenant real.
