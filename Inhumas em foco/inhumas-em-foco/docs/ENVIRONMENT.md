# Variaveis de Ambiente

Referencia das variaveis suportadas pelo backend atual.

| Variavel | Obrigatoria | Padrao | Uso |
| --- | --- | --- | --- |
| `PORT` | Nao | `8080` | Porta HTTP do backend Go. |
| `APP_ENV` | Sim em producao | vazio | Define comportamento de ambiente. Use `production` na VPS. |
| `SITE_URL` | Sim | `https://inhumasemfoco.com.br` | URL publica usada em sitemap, RSS e links absolutos. |
| `DATABASE_URL` | Sim | `./inhumas.db` | Caminho SQLite no runtime atual. A URL Postgres ainda e pendencia de runtime. |
| `SESSION_SECRET` | Sim | valor inseguro de dev | Segredo de sessao; minimo de 32 caracteres. |
| `PREVIOUS_SESSION_SECRET` | Nao | vazio | Segredo anterior aceito temporariamente para rotacao de sessoes. |
| `FORCE_SECURE_COOKIES` | Sim em producao | `false` | Forca cookies `Secure` mesmo quando o app esta atras do proxy. |
| `INITIAL_ADMIN_PASSWORD` | Sim no primeiro start | vazio | Cria o admin inicial se ainda nao houver usuarios. Remova depois. |
| `ADMIN_PATH_PREFIX` | Sim | `/painel/7x9k2m` | Prefixo obscuro do painel admin. |
| `UPLOAD_DIR` | Sim | `./uploads` | Diretorio de uploads locais. |
| `STATIC_DIR` | Sim | `./static` | Diretorio servido em `/static` e destino do sitemap. |
| `PROJECT_ROOT` | Sim em deploy | `.` | Raiz usada para localizar templates e assets do projeto. |
| `METRICS_TOKEN` | Sim se `/metrics` ficar publico | vazio | Token Bearer ou `X-Metrics-Token` aceito pelo endpoint operacional `/metrics`. |
| `BACKUP_DIR` | Sim em producao | vazio | Diretorio onde o health procura backups recentes e o worker orienta o script de backup. |
| `BACKUP_SCRIPT` | Sim em producao | vazio | Caminho do script executado pelo job recorrente `backup_database`. |
| `SMTP_HOST` | Sim para recuperacao de senha por e-mail | vazio | Servidor SMTP usado para e-mails transacionais. Se vazio, nenhum e-mail e enviado. |
| `SMTP_PORT` | Nao | `587` | Porta SMTP com STARTTLS quando suportado pelo servidor. |
| `SMTP_USERNAME` | Depende do provedor | vazio | Usuario SMTP. |
| `SMTP_PASSWORD` | Depende do provedor | vazio | Senha/token SMTP. |
| `SMTP_FROM` | Sim para SMTP | vazio | E-mail remetente, por exemplo `no-reply@inhumasemfoco.online`. |
| `SMTP_FROM_NAME` | Nao | `Inhumas em Foco` | Nome exibido como remetente. |
| `MAX_UPLOAD_SIZE` | Nao | `2097152` | Limite de upload em bytes. |
| `BCRYPT_COST` | Nao | `12` | Custo do bcrypt. Em testes pode ser menor. |
| `MAINTENANCE_MODE` | Nao | `false` | Ativa modo manutencao nas rotas publicas quando middleware estiver ligado. |

## Exemplo SQLite de producao MVP

```env
PORT=8080
APP_ENV=production
SITE_URL=https://inhumasemfoco.com.br
DATABASE_URL=/var/www/inhumas/inhumas.db
SESSION_SECRET=troque-por-um-segredo-longo-e-aleatorio
PREVIOUS_SESSION_SECRET=
FORCE_SECURE_COOKIES=true
INITIAL_ADMIN_PASSWORD=troque-no-primeiro-start
ADMIN_PATH_PREFIX=/painel/caminho-unico
UPLOAD_DIR=/var/www/inhumas/uploads
STATIC_DIR=/var/www/inhumas/static
PROJECT_ROOT=/var/www/inhumas
METRICS_TOKEN=troque-por-token-longo-e-aleatorio
BACKUP_DIR=/var/backups/inhumas
BACKUP_SCRIPT=/var/www/inhumas/scripts/backup.sh
SMTP_HOST=smtp.exemplo.com
SMTP_PORT=587
SMTP_USERNAME=no-reply@inhumasemfoco.online
SMTP_PASSWORD=troque-por-token-smtp
SMTP_FROM=no-reply@inhumasemfoco.online
SMTP_FROM_NAME=Inhumas em Foco
MAX_UPLOAD_SIZE=2097152
BCRYPT_COST=12
MAINTENANCE_MODE=false
```

## Observacao sobre PostgreSQL

O prompt completo mira PostgreSQL, e a migration inicial ja existe em `migrations/001_init_postgres.sql`. O codigo Go atual ainda instancia o repository SQLite. Ate a migracao de runtime ser feita, configure `DATABASE_URL` como caminho de arquivo SQLite.
