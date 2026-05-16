# Variaveis de Ambiente

Referencia das variaveis suportadas pelo backend atual.

| Variavel | Obrigatoria | Padrao | Uso |
| --- | --- | --- | --- |
| `PORT` | Nao | `8080` | Porta HTTP do backend Go. |
| `APP_ENV` | Sim em producao | vazio | Define comportamento de ambiente. Use `production` na VPS. |
| `SITE_URL` | Sim | `https://inhumasemfoco.com.br` | URL publica usada em sitemap, RSS e links absolutos. |
| `DB_DRIVER` | Nao | autodetectado | `sqlite` ou `postgres`. Se vazio, URLs `postgres://`/`postgresql://` ativam PostgreSQL; caso contrario usa SQLite. |
| `DATABASE_URL` | Sim | `./inhumas.db` | Caminho SQLite ou URL PostgreSQL conforme `DB_DRIVER`. |
| `MIGRATIONS_DIR` | Sim para PostgreSQL | `./migrations` | Diretorio das migrations versionadas executadas quando `DB_DRIVER=postgres`. |
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
| `CSP_REPORT_ONLY` | Nao | `false` | Quando `true`, envia uma CSP report-only mais restritiva para auditar remocao futura de inline scripts/styles. |
| `CSP_REPORT_URI` | Nao | vazio | Endpoint que recebera violacoes CSP quando `CSP_REPORT_ONLY=true`. |
| `SMTP_HOST` | Sim para recuperacao de senha por e-mail | vazio | Servidor SMTP usado para e-mails transacionais. Se vazio, nenhum e-mail e enviado. |
| `SMTP_PORT` | Nao | `587` | Porta SMTP com STARTTLS quando suportado pelo servidor. |
| `SMTP_USERNAME` | Depende do provedor | vazio | Usuario SMTP. |
| `SMTP_PASSWORD` | Depende do provedor | vazio | Senha/token SMTP. |
| `SMTP_FROM` | Sim para SMTP | vazio | E-mail remetente, por exemplo `no-reply@inhumasemfoco.online`. |
| `SMTP_FROM_NAME` | Nao | `Inhumas em Foco` | Nome exibido como remetente. |
| `MAX_UPLOAD_SIZE` | Nao | `2097152` | Limite de upload em bytes. |
| `BCRYPT_COST` | Nao | `12` | Custo do bcrypt. Em testes pode ser menor. |
| `MAINTENANCE_MODE` | Nao | `false` | Ativa modo manutencao nas rotas publicas quando middleware estiver ligado. |
| `OFFSITE_BACKUP_DIR` | Nao | vazio | Diretorio externo/montado usado por `scripts/backup-offsite.sh` quando nao houver rclone. |
| `RCLONE_REMOTE` | Nao | vazio | Destino rclone para backup fora da VPS, por exemplo `provider:bucket/inhumas`. |
| `OFFSITE_MAX_AGE` | Nao | `48h` | Janela de arquivos recentes enviados ao destino rclone. |
| `MAX_BACKUP_AGE_HOURS` | Nao | `26` | Idade maxima aceita pelo `scripts/backup-check.sh`. |
| `DISK_ALERT_THRESHOLD` | Nao | `80` | Percentual de disco que dispara alerta em `scripts/disk-check.sh`. |
| `JOURNAL_RETENTION` | Nao | `14d` | Retencao aplicada pelo `scripts/journal-retention.sh`. |

## Exemplo SQLite de producao MVP

```env
PORT=8080
APP_ENV=production
SITE_URL=https://inhumasemfoco.com.br
DATABASE_URL=/var/www/inhumas/inhumas.db
DB_DRIVER=sqlite
MIGRATIONS_DIR=/var/www/inhumas/migrations
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
CSP_REPORT_ONLY=false
CSP_REPORT_URI=
SMTP_HOST=smtp.exemplo.com
SMTP_PORT=587
SMTP_USERNAME=no-reply@inhumasemfoco.online
SMTP_PASSWORD=troque-por-token-smtp
SMTP_FROM=no-reply@inhumasemfoco.online
SMTP_FROM_NAME=Inhumas em Foco
MAX_UPLOAD_SIZE=2097152
BCRYPT_COST=12
MAINTENANCE_MODE=false
RCLONE_REMOTE=
OFFSITE_BACKUP_DIR=
MAX_BACKUP_AGE_HOURS=26
DISK_ALERT_THRESHOLD=80
JOURNAL_RETENTION=14d
```

## PostgreSQL

O runtime suporta PostgreSQL por configuracao. Para uma instancia nova:

```env
DB_DRIVER=postgres
DATABASE_URL=postgres://usuario:senha@localhost:5432/inhumas?sslmode=disable
MIGRATIONS_DIR=/var/www/inhumas/migrations
```

Ao iniciar, o app cria `schema_migrations` e aplica `migrations/*.sql` uma vez, em ordem.

Para migrar dados do SQLite atual para PostgreSQL:

```bash
SQLITE_DATABASE_URL=/var/www/inhumas/inhumas.db \
POSTGRES_DATABASE_URL='postgres://usuario:senha@localhost:5432/inhumas?sslmode=disable' \
MIGRATIONS_DIR=/var/www/inhumas/migrations \
/var/www/inhumas/bin/migrate-sqlite-postgres
```

Depois de validar contagens, pare os servicos, altere `.env` para PostgreSQL e suba novamente.
