# ADR 003: HTML Server-Side Primeiro

## Decisao

Manter templates HTML server-side com HTMX leve como frontend principal do CMS.

## Motivo

O produto editorial precisa de SEO, performance, simplicidade operacional e ergonomia de painel. O servidor renderizado atende bem ao momento atual.

## Beneficios

- SEO simples.
- Menos build frontend.
- Menor superficie de falha.
- Integra bem com Go e templates existentes.

## Riscos

- Interacoes muito complexas podem ficar verbosas.
- Scripts inline exigem cuidado com CSP.

## Mitigacao

Usar HTMX e componentes pequenos. Migrar experiencias especificas para frontend dedicado apenas quando houver necessidade real.
