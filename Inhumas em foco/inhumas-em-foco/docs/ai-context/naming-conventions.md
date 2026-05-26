# Naming Conventions

## Entidades Canonicas Futuras

- Tenant
- Portal
- User
- Article
- Category
- Tag
- Media
- Banner
- Store
- Promotion
- Event
- Classified
- Influencer
- AuditLog
- Metric
- Job

## Nomes Legados Aceitos

O codigo atual usa `Post` para noticia/artigo. Ate uma migracao planejada, manter `Post` no codigo existente para evitar churn. Em documentacao e novos modulos, usar `Article` como conceito futuro e indicar equivalencia:

```text
Article = Post atual
```

## Services

- `AuthService`
- `UserService`
- `EditorialService`
- `CommercialService`
- `AutomationService`
- `MediaService`
- `TenantService`

## Repositories

Enquanto houver `Repository` central, novos metodos devem seguir prefixo do dominio:

- `PostCreate`
- `UserUpdate`
- `TenantGet`
- `MediaList`

Quando houver separacao futura:

- `ArticleRepository`
- `UserRepository`
- `TenantRepository`
- `MediaRepository`

## Rotas

Rotas publicas atuais permanecem em portugues para SEO local:

- `/noticias`
- `/noticia/{slug}`
- `/lojas`
- `/eventos`

APIs futuras devem usar versao e nomes estaveis:

- `/api/v1/articles`
- `/api/v1/media`
- `/api/v1/tenants`
