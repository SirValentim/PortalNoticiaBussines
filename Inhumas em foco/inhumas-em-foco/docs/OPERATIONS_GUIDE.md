# Guia de Operacao Diaria

Rotina operacional para manter o Inhumas em Foco estavel depois do deploy.

## Checklist diario

- Abrir `/health` e confirmar status `ok`.
- Conferir `systemctl status inhumas-web inhumas-worker`.
- Verificar espaco em disco com `df -h`.
- Conferir se houve backup na ultima janela agendada.
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
```

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

O script `scripts/backup.sh` deve rodar por cron. Guarde copias fora da VPS quando o projeto entrar em producao real.

Teste minimo de restauracao:

```bash
sqlite3 /caminho/para/backup.db "PRAGMA integrity_check;"
```

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
6. Rodar smoke test.

```bash
sudo systemctl stop inhumas-worker
sudo systemctl restart inhumas-web
sudo systemctl start inhumas-worker
curl -f https://inhumasemfoco.com.br/health
```

