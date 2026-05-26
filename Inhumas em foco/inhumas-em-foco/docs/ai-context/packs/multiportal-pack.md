# Multiportal Pack

## Estado Atual

Branding configuracional por env:

- `PORTAL_NAME`
- `SITE_URL`
- `PORTAL_CONTACT_EMAIL`
- cores, logo, favicon, SEO image, redes e textos legais

## Objetivo Futuro

Permitir varios portais no mesmo core, com isolamento de dados, dominios, usuarios, branding, modulos habilitados e limites comerciais.

## Modelo Conceitual

```text
Platform
  -> Tenant
  -> Portal
  -> Module
  -> Content
```

## Proxima Migration Recomendada

- `tenants`
- `tenant_domains`
- `tenant_features`

Depois disso, adicionar `tenant_id` gradualmente nas tabelas de conteudo.

## Cuidado

Nao adicionar `tenant_id` em todas as tabelas de uma vez sem plano de migracao, backfill e testes.
