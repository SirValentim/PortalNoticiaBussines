# Guia de Deploy - Inhumas em Foco

Este guia descreve o caminho recomendado para publicar o backend Go em uma VPS Debian 12. O runtime atual do projeto ainda usa SQLite; PostgreSQL ja tem migration inicial, mas nao esta ligado ao codigo principal.

## 1. Preparar servidor

```bash
sudo apt update
sudo apt install -y nginx certbot python3-certbot-nginx sqlite3 curl ca-certificates
sudo useradd --system --home /var/www/inhumas --shell /usr/sbin/nologin inhumas
sudo mkdir -p /var/www/inhumas/{uploads,static,backups,releases}
sudo chown -R inhumas:inhumas /var/www/inhumas
```

Crie swap de 1 GB se a VPS for pequena:

```bash
sudo fallocate -l 1G /swapfile
sudo chmod 600 /swapfile
sudo mkswap /swapfile
sudo swapon /swapfile
echo '/swapfile none swap sw 0 0' | sudo tee -a /etc/fstab
```

## 2. Build

No ambiente de build:

```bash
make build-linux
```

Envie os binarios `bin/web` e `bin/worker`, alem de `internal/view`, `static`, `migrations`, `scripts`, `deploy` e `.env.example` para a VPS. O diretorio final esperado pelo exemplo e:

```text
/var/www/inhumas
```

## 3. Configurar ambiente

Copie `.env.example` para `/etc/inhumas.env` e ajuste valores sensiveis:

```bash
sudo cp /var/www/inhumas/.env.example /etc/inhumas.env
sudo chmod 600 /etc/inhumas.env
sudo nano /etc/inhumas.env
```

Valores obrigatorios antes do primeiro start:

- `SESSION_SECRET`: string aleatoria com pelo menos 32 caracteres.
- `INITIAL_ADMIN_PASSWORD`: senha inicial forte; depois remova ou troque.
- `ADMIN_PATH_PREFIX`: caminho obscuro do painel.
- `DATABASE_URL`: para o runtime atual, use caminho SQLite, exemplo `/var/www/inhumas/inhumas.db`.
- `UPLOAD_DIR`: `/var/www/inhumas/uploads`.
- `STATIC_DIR`: `/var/www/inhumas/static`.
- `PROJECT_ROOT`: `/var/www/inhumas`.
- `SITE_URL`: URL publica final com HTTPS.
- `APP_ENV=production`.
- `FORCE_SECURE_COOKIES=true`.

## 4. Systemd

```bash
sudo cp /var/www/inhumas/deploy/inhumas-web.service.example /etc/systemd/system/inhumas-web.service
sudo cp /var/www/inhumas/deploy/inhumas-worker.service.example /etc/systemd/system/inhumas-worker.service
sudo systemctl daemon-reload
sudo systemctl enable --now inhumas-web
sudo systemctl enable --now inhumas-worker
```

Verifique:

```bash
sudo systemctl status inhumas-web --no-pager
sudo systemctl status inhumas-worker --no-pager
curl -f http://127.0.0.1:8080/health
```

## 5. Nginx e TLS

Copie o exemplo e ajuste dominios/caminhos:

```bash
sudo cp /var/www/inhumas/deploy/nginx.conf.example /etc/nginx/sites-available/inhumas
sudo ln -s /etc/nginx/sites-available/inhumas /etc/nginx/sites-enabled/inhumas
sudo nginx -t
```

Emita certificados:

```bash
sudo certbot --nginx -d inhumasemfoco.com.br -d www.inhumasemfoco.com.br
sudo nginx -t
sudo systemctl reload nginx
```

Valide por fora:

```bash
curl -I https://inhumasemfoco.com.br/health
curl -I https://inhumasemfoco.com.br/rss.xml
curl -I https://inhumasemfoco.com.br/static/sitemap.xml
```

## 6. Backup e alertas

Configure backup diario do SQLite e alerta de disco:

```bash
sudo crontab -e
```

Exemplo:

```cron
10 2 * * * /var/www/inhumas/scripts/backup.sh >> /var/log/inhumas-backup.log 2>&1
*/30 * * * * /var/www/inhumas/scripts/disk-check.sh >> /var/log/inhumas-disk.log 2>&1
```

## 7. Smoke test

Antes de considerar o deploy pronto:

- `/health` retorna HTTP 200.
- `/login` abre com cookie seguro via HTTPS.
- Login inicial redireciona para o painel.
- Upload pequeno de imagem gera WebP e thumb.
- Criar post publicado aparece na home, em `/rss.xml` e no proximo sitemap.
- Worker esta ativo e sem erros repetidos no journal.
- Backup gera arquivo restauravel.
