# Domain Map

```mermaid
flowchart TD
    Platform["Platform"]
    Tenant["Tenant"]
    Portal["Portal"]
    Branding["Branding"]
    Users["Users and RBAC"]
    CMS["CMS Content"]
    Media["Media"]
    SEO["SEO and Discovery"]
    Automation["Automation"]
    Commercial["Commercial Modules"]
    Analytics["Analytics and Metrics"]
    Jobs["Jobs and Operations"]

    Platform --> Tenant
    Tenant --> Portal
    Portal --> Branding
    Portal --> Users
    Portal --> CMS
    CMS --> Media
    CMS --> SEO
    CMS --> Automation
    Portal --> Commercial
    Portal --> Analytics
    Platform --> Jobs
```

## Dominios Atuais

- Tenant/Branding: tenants, dominios, features, settings persistidos e fallback por env.
- Users/RBAC: usuarios, roles, permissoes e auditoria.
- CMS: posts, categorias, tags, revisoes, status editorial.
- Media: upload, validacao, thumbnails e storage local.
- SEO: sitemap, robots, RSS, JSON-LD, OG/Twitter.
- Automation: RSS sources, runs e jobs.
- Commercial: lojas, promocoes, eventos, classificados, influencers, bairros e banners.
- Operations: health, metrics, backups, jobs e logs.
