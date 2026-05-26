# Architecture Summary

## Modelo

Monolito modular em Go, com entrada HTTP unica e separacao parcial por pacotes internos.

## Fluxo Principal

```text
HTTP -> Middleware -> Handler -> Service -> Repository -> Database
```

Na pratica atual, alguns handlers ainda acessam o repository diretamente. Isso e aceitavel para o estado legado, mas novas regras de negocio devem ir para services.

## Entrada HTTP

- `cmd/web/main.go` monta config, banco, sessao, auth, storage, handler e middleware.
- `internal/handler` registra rotas publicas, administrativas e APIs.
- Templates ficam em `internal/view`.

## Jobs

- `cmd/worker` processa jobs recorrentes e pendentes.
- Jobs atuais incluem sitemap, backup, limpeza, publicacao agendada, expiracao de promocoes/banners e coleta de noticias.

## Persistencia

- `internal/repository` encapsula banco.
- SQLite e PostgreSQL sao suportados por dialeto.
- Migrations versionadas ficam em `migrations`.

## Multiportal Atual

- Branding por env em `internal/config/branding.go`.
- Middleware injeta branding no contexto.
- Templates usam `Branding` para nome, cores, logo, favicon, contatos, metadata e manifest.

## Multiportal Futuro

```text
Platform
  -> Tenant
  -> Portal
  -> Content
  -> Modules
  -> Branding
```

O proximo passo tecnico e adicionar identidade de tenant persistente e escopo de dados por tenant antes de operar multiplos portais no mesmo banco.
