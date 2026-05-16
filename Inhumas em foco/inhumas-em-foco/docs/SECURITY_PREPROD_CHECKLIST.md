# Checklist de Seguranca Pre-Producao

Use antes de expor o projeto publicamente.

## Segredos e acesso

- [ ] `SESSION_SECRET` aleatorio, unico e com pelo menos 32 caracteres.
- [ ] `INITIAL_ADMIN_PASSWORD` forte no primeiro start e removido depois.
- [ ] `ADMIN_PATH_PREFIX` diferente do exemplo publico.
- [ ] Usuario do systemd sem shell e sem permissao fora de `/var/www/inhumas`.
- [ ] Arquivo `/etc/inhumas.env` com permissao `600`.

## HTTPS, proxy e cookies

- [ ] HTTPS ativo com certificado valido.
- [ ] HTTP redireciona para HTTPS.
- [ ] `APP_ENV=production`.
- [ ] `FORCE_SECURE_COOKIES=true`.
- [ ] Nginx envia `X-Forwarded-Proto https`.
- [ ] HSTS ativo depois de confirmar que HTTPS esta estavel.

## Headers

- [ ] `X-Frame-Options: SAMEORIGIN`.
- [ ] `X-Content-Type-Options: nosniff`.
- [ ] `Referrer-Policy: strict-origin-when-cross-origin`.
- [ ] `Permissions-Policy` restritivo.
- [ ] `Content-Security-Policy` aplicada com `object-src 'none'`, `base-uri 'self'` e `form-action 'self'`.
- [ ] `CSP_REPORT_ONLY=true` habilitado temporariamente em homologacao para auditar remocao futura de `unsafe-inline`.
- [ ] `unsafe-inline` ainda permitido no header aplicado por compatibilidade com templates; remover apenas depois de migrar scripts/styles inline.

## Formularios e autenticacao

- [ ] Login protegido por CSRF.
- [ ] POSTs do painel protegidos por CSRF.
- [ ] Rate limit no app e no Nginx para login.
- [ ] Teste manual de cookie seguro atras do proxy.
- [ ] Contas com papeis revisados: `admin`, `editor`, `comercial`.

## Uploads

- [ ] `UPLOAD_DIR` fora de diretorio executavel.
- [ ] Nginx serve uploads como arquivos estaticos, sem execucao.
- [ ] `MAX_UPLOAD_SIZE` alinhado com `client_max_body_size`.
- [ ] Upload invalido rejeitado.
- [ ] WebP e thumb gerados nos novos uploads.

## Operacao

- [ ] Backup diario ativo.
- [ ] Backup externo fora da VPS ativo e testado.
- [ ] `scripts/backup-check.sh` agendado apos backup diario.
- [ ] Restauracao testada.
- [ ] Alerta de disco ativo.
- [ ] `/health` monitorado.
- [ ] Logs revisados apos primeiro dia de producao.
- [ ] Smoke test pos-deploy executado.
