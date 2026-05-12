# Registro Essencial da Sessao

## O que foi feito

- Aplicado o prompt no backend Go principal em `inhumas-em-foco`.
- Corrigido build do Go.
- Trocado SQLite com CGO por `modernc.org/sqlite`, para rodar no Windows sem GCC.
- Corrigido login/sessao do admin.
- Protegido painel `/painel/7x9k2m` por autenticacao.
- Adicionado RBAC basico:
  - `admin`: tudo.
  - `editor`: posts.
  - `comercial`: lojas, banners e promocoes.
- Adicionado CSRF em login e formularios admin.
- Reforcados headers de seguranca.
- Reforcado upload local com validacao real de MIME e bloqueio de path traversal.
- Corrigido render dos templates Go.
- Criados arquivos de producao:
  - `.env.example`
  - `migrations/001_init_postgres.sql`
  - `scripts/backup.sh`
  - `scripts/disk-check.sh`
  - `deploy/nginx.conf.example`
  - `deploy/inhumas-web.service.example`
  - `deploy/inhumas-worker.service.example`
- Atualizado `Makefile`.

## Como testar localmente

```powershell
cd "C:\Users\Valentim\Documents\New project\Inhumas em foco\inhumas-em-foco"
$env:SESSION_SECRET="12345678901234567890123456789012"
$env:INITIAL_ADMIN_PASSWORD="admin123456"
$env:PORT="8080"
go run ./cmd/web
```

Abrir:

```text
http://localhost:8080/login
```

Login local:

```text
admin@inhumasemfoco.com.br
admin123456
```

Painel:

```text
http://localhost:8080/painel/7x9k2m
```

## Validado

```powershell
go test ./...
```

Resultado: todos os pacotes Go passaram.

Fluxo validado:

```text
GET /health -> 200
GET /login -> 200 com csrf_token
POST /login -> redireciona para /painel/7x9k2m
GET /painel/7x9k2m autenticado -> 200
```

## Pendencias

- Migrar runtime de SQLite local para PostgreSQL.
- Implementar worker com `FOR UPDATE SKIP LOCKED`.
- Validar deploy real na VPS com Nginx, systemd, TLS, backup e monitoramento.
- Completar metricas/admin comercial.
- Implementar retencao automatica dos originais antigos.
- Trocar mocks do React por API real somente se o prototipo for promovido a produto.

## Atualizacao posterior

- RSS publico `/rss.xml` implementado e testado.
- Exemplo Nginx reforcado com HTTPS, HSTS, headers finais, cache e bloqueio de dotfiles.
- Exemplos systemd ajustados para `/etc/inhumas.env` e hardening basico.
- `.env.example` alinhado ao SQLite MVP, evitando apontar para PostgreSQL antes do runtime suportar Postgres.
- Criados guias de deploy, operacao diaria, variaveis de ambiente e checklist pre-producao.
- Decisao React documentada: permanece como prototipo visual no MVP.

## Atualizacao 2026-04-28

- Deploy do MVP validado na VPS em `/var/www/inhumas`.
- `inhumas-web` e `inhumas-worker` ativos via systemd.
- Nginx configurado como proxy para `127.0.0.1:8080`.
- Swap de 1GB ativado.
- Backup local em `/var/backups/inhumas` configurado e restore/integridade SQLite testado.
- Cron diario de backup e cron de alerta de disco configurados.
- Corrigido `scripts/backup.sh` para carregar `/etc/inhumas.env` quando necessario.
- Corrigido `scripts/disk-check.sh` para nao falhar quando `mail` nao estiver disponivel.
- Criada nova conta admin para operacao inicial por IP/HTTP, com credenciais salvas em `dados do VPS/inhumas-admin2-credentials.txt`.
- Ajuste temporario na VPS para acesso por HTTP/IP sem cookie `Secure`; deve voltar para `APP_ENV=production` e `FORCE_SECURE_COOKIES=true` apos HTTPS.
- Corrigidos erros de renderizacao no painel admin causados por campos nulos ou chaves ausentes em templates.
- Renderer passou a registrar erro real de template no journal.
- Upgrade visual amplo do painel admin:
  - layout admin mais largo;
  - formularios em grids responsivos;
  - tela de posts com editor maior;
  - telas de lojas, banners, promocoes, bairros, usuarios, metricas, dashboard e jobs padronizadas;
  - remocao de estilos inline antigos;
  - status exibidos como etiquetas visuais;
  - CSS versionado para evitar cache antigo no navegador.
- Worker passou a reivindicar jobs vencidos com status `running` antes da execucao.
- Busca de posts passou a usar FTS5 no runtime SQLite, com fallback `LIKE`.
- Modulo de influencers implementado:
  - tabela `influencers` no SQLite runtime;
  - migration PostgreSQL alinhada;
  - repository com create/update/get/list/delete;
  - rotas publicas `/influencers` e `/influencer/{slug}`;
  - telas publicas de listagem e perfil;
  - menu publico com link para Influencers;
  - painel admin com listagem, criacao, edicao e exclusao de influencers;
  - campos de bio, area/nicho, Instagram, TikTok, YouTube, WhatsApp, avatar, capa, destaque, patrocinado e ativo;
  - metrica `influencer_view`;
  - influencers ativos incluidos no sitemap.
- Testes adicionados para busca FTS5, claim de jobs e paginas publicas de influencers.
- Validado com `go test ./...`.

## Pendencias apos 2026-04-28

- Apontar Cloudflare/DNS para `68.233.125.102`.
- Emitir HTTPS com Certbot.
- Voltar cookies seguros em producao: `APP_ENV=production` e `FORCE_SECURE_COOKIES=true`.
- Validar login, painel, upload e cookies Secure no dominio HTTPS.
- Configurar UptimeRobot.
- Configurar backup externo.
- Decidir se PostgreSQL sera migracao obrigatoria antes da operacao publica ampliada.
