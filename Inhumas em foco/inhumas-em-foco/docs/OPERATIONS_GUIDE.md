# Guia de Operacao Diaria

Rotina operacional para manter o Inhumas em Foco estavel depois do deploy.

## Checklist diario

- Abrir `/health` e confirmar status `ok`.
- Conferir `systemctl status inhumas-web inhumas-worker`.
- Verificar espaco em disco com `df -h`.
- Conferir se houve backup na ultima janela agendada.
- Rodar ou validar o resultado de `scripts/backup-check.sh`.
- Conferir se o monitor externo de uptime esta verde.
- Revisar erros recentes no journal.
- Conferir se posts agendados foram publicados.

Comandos uteis:

```bash
curl -f https://inhumasemfoco.com.br/health
sudo systemctl status inhumas-web --no-pager
sudo systemctl status inhumas-worker --no-pager
sudo journalctl -u inhumas-web -n 100 --no-pager
sudo journalctl -u inhumas-worker -n 100 --no-pager
df -h
/var/www/inhumas/scripts/backup-check.sh
```

## Monitoramento e alertas

- Use um monitor externo de uptime apontando para `https://inhumasemfoco.online/health` com intervalo de 1 a 5 minutos.
- Configure alerta para e-mail/WhatsApp quando `/health` falhar duas vezes seguidas.
- Execute `scripts/disk-check.sh` por cron ou systemd timer com `DISK_ALERT_THRESHOLD=80`.
- Execute `scripts/backup-check.sh` apos a rotina de backup. O padrao aceita no maximo 26 horas desde o ultimo backup de banco.
- Os logs HTTP do app saem em JSON no stdout do systemd com `request_id`, metodo, rota, status, bytes, IP e duracao.
- Para retencao de journal, rode `scripts/journal-retention.sh` diariamente com `JOURNAL_RETENTION=14d` ou valor aprovado para a VPS.

## Publicacao editorial

- Criar posts como `draft` ate a revisao final.
- Para `Politica & Bastidores`, preencher notas de apuracao e responsavel editorial.
- Usar `scheduled` quando a publicacao tiver horario definido.
- Conferir o post publicado, a imagem WebP e o selo `Patrocinado` quando aplicavel.

## Rotina comercial

- Conferir banners ativos por posicao.
- Validar datas de inicio/fim antes de publicar campanhas.
- Conferir promocoes ativas e lojas destacadas.
- Revisar metricas do painel como indicador operacional, nao como relatorio financeiro definitivo.

## Backup e restauracao

O script `scripts/backup.sh` deve rodar por cron. Copias fora da VPS sao obrigatorias para operacao real.

Exemplo de agenda diaria:

```cron
10 3 * * * BACKUP_DIR=/var/www/inhumas/backups BACKUP_SCRIPT=/var/www/inhumas/scripts/backup.sh /var/www/inhumas/scripts/backup.sh
40 3 * * * BACKUP_DIR=/var/www/inhumas/backups MAX_BACKUP_AGE_HOURS=26 /var/www/inhumas/scripts/backup-check.sh
*/30 * * * * DISK_ALERT_THRESHOLD=80 /var/www/inhumas/scripts/disk-check.sh
20 4 * * * JOURNAL_RETENTION=14d /var/www/inhumas/scripts/journal-retention.sh
```

Backup externo via rclone:

```bash
BACKUP_DIR=/var/www/inhumas/backups \
RCLONE_REMOTE='backup-provider:inhumas-em-foco' \
/var/www/inhumas/scripts/backup-offsite.sh
```

Backup externo via volume montado:

```bash
BACKUP_DIR=/var/www/inhumas/backups \
OFFSITE_BACKUP_DIR=/mnt/backup-inhumas \
/var/www/inhumas/scripts/backup-offsite.sh
```

Teste minimo de restauracao:

```bash
sqlite3 /caminho/para/backup.db "PRAGMA integrity_check;"
```

Para testar a migracao SQLite -> PostgreSQL em ambiente de homologacao:

```bash
createdb inhumas_homolog
SQLITE_DATABASE_URL=/var/www/inhumas/inhumas.db \
POSTGRES_DATABASE_URL='postgres://usuario:senha@localhost:5432/inhumas_homolog?sslmode=disable' \
MIGRATIONS_DIR=/var/www/inhumas/migrations \
/var/www/inhumas/bin/migrate-sqlite-postgres
```

O comando aplica as migrations, copia tabelas em ordem segura, ajusta sequences e compara contagens por tabela.

Para restaurar, pare os servicos, copie o backup para `DATABASE_URL`, ajuste dono/permissao e suba novamente:

```bash
sudo systemctl stop inhumas-web inhumas-worker
sudo cp backup.db /var/www/inhumas/inhumas.db
sudo chown inhumas:inhumas /var/www/inhumas/inhumas.db
sudo systemctl start inhumas-web inhumas-worker
```

## Incidentes comuns

- `health` com erro em uploads: conferir permissao de `UPLOAD_DIR`.
- Login nao segura sessao: conferir `SESSION_SECRET`, HTTPS, `FORCE_SECURE_COOKIES` e `X-Forwarded-Proto`.
- Upload falha: conferir `MAX_UPLOAD_SIZE`, `client_max_body_size` no Nginx e permissao de escrita.
- Worker sem processar: conferir journal, banco e jobs em `dead_jobs`.
- Disco enchendo: limpar backups antigos, logs e revisar crescimento de `uploads/original`.

## Atualizacao de versao

1. Rodar testes no ambiente de build.
2. Enviar novo binario e assets.
3. Parar worker.
4. Reiniciar web.
5. Reiniciar worker.
6. Rodar smoke test automatizado.
7. Conferir logs estruturados pelo `request_id` em caso de erro.

```bash
sudo systemctl stop inhumas-worker
sudo systemctl restart inhumas-web
sudo systemctl start inhumas-worker
curl -f https://inhumasemfoco.com.br/health
BASE_URL=https://inhumasemfoco.online ADMIN_PATH=/painel/1c2dhviax7 /var/www/inhumas/scripts/smoke-test.sh
```

## Rollback

Checklist rapido para emergencia:

1. Confirmar o erro em `/health`, journal ou smoke test.
2. Parar `inhumas-worker` para impedir jobs durante a reversao.
3. Restaurar o binario anterior ou o pacote de deploy anterior.
4. Se o banco foi alterado, restaurar o backup pre-deploy correspondente.
5. Conferir dono/permissao de banco, uploads e binarios.
6. Subir `inhumas-web`, validar `/health` e login admin.
7. Subir `inhumas-worker` apenas depois do painel responder.
8. Registrar no `CHECKLIST_STATUS.md` o backup usado, horario, causa e resultado.
