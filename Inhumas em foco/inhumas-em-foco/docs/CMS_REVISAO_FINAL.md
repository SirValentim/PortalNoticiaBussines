# Revisao Final Corporativa do CMS

Este documento resume o estado corporativo do CMS Inhumas em Foco apos a conclusao dos sprints 0 a 14. O checklist vivo continua sendo `docs/CHECKLIST_STATUS.md`; este arquivo serve como leitura executiva para operacao, manutencao e proximas decisoes.

## Status executivo

O CMS esta pronto para operacao inicial de um portal local profissional em producao, com publicacao editorial, painel administrativo, midia, taxonomia, automacao assistida, IA desacoplada, anuncios, modulos locais, auditoria, configuracoes, SEO tecnico e rotina operacional.

O runtime de producao segue em SQLite por decisao conservadora, mas o projeto ja possui suporte configuravel a PostgreSQL, migrations versionadas e comando de migracao validado. A troca para PostgreSQL deve acontecer em uma janela propria, com homologacao e backup restauravel.

## Modulos concluidos

- Painel administrativo responsivo com sidebar hierarquica, topbar, toasts, loading states e confirmacao customizada.
- Usuarios, papeis e RBAC para administracao, editorial, revisao e comercial.
- Noticias com fluxo editorial, revisoes, autosave, preview, SEO, fonte original, galeria e checklist.
- Categorias, tags e biblioteca de midia reutilizavel.
- Auditoria e configuracoes do portal.
- Eventos, classificados, lojas, promocoes, influencers e bairros.
- Anuncios com relatorios, validacao de conflito e produtos comerciais locais.
- Automacao RSS/oficial com rascunhos e revisao humana obrigatoria.
- Camada de IA editorial desacoplada com guardrails.
- SEO tecnico: sitemap, robots, RSS, canonical, OG/Twitter Cards e JSON-LD.
- Operacao: health, metrics, logs JSON, smoke test, backups, backup-check, alertas e rollback.

## Readiness operacional

Antes de considerar uma janela de publicacao ou manutencao encerrada:

```bash
BASE_URL=https://inhumasemfoco.online \
ADMIN_PATH=/painel/1c2dhviax7 \
BACKUP_DIR=/var/backups/inhumas \
/var/www/inhumas/scripts/production-readiness.sh
```

O script verifica ambiente, services, health, home publica, protecao do admin, backup recente, disco, Nginx e destino de backup externo quando configurado.

## Decisoes abertas

- Configurar backup externo real fora da VPS usando `RCLONE_REMOTE` ou volume externo.
- Configurar monitor externo de uptime, por exemplo UptimeRobot, apontando para `/health`.
- Configurar SMTP real para recuperacao de senha em producao.
- Planejar migracao SQLite -> PostgreSQL em janela propria.
- Remover `unsafe-inline` da CSP depois de migrar scripts/styles inline para arquivos estaticos.
- Opcional: ativar Cloudflare/WAF quando DNS estiver sob gestao definitiva.

## Backlog residual corporativo

Prioridade alta:

- Backup externo real e teste documentado de restauracao.
- Monitor externo com alerta para indisponibilidade.
- SMTP transacional real.
- Revisao das contas administrativas e remocao de credenciais bootstrap.

Prioridade media:

- Homologacao PostgreSQL com copia recente da base.
- Testes especificos de repository PostgreSQL em CI.
- CSP sem `unsafe-inline`.
- Relatorios comerciais avancados por periodo/anunciante.

Prioridade futura:

- S3 ou storage externo para midia.
- WAF/CDN completo.
- Otimizacoes finas de bundle/prototipo React, se o prototipo voltar a ser usado.
- Analytics editorial mais profundo.

## Rotina recomendada

- Diario: `/health`, services, backup-check, disco e journal.
- Semanal: revisar auditoria, jobs com falha, usuarios ativos, posts agendados e relatorios comerciais.
- Mensal: testar restore, revisar secrets, atualizar sistema operacional e validar certificados TLS.
- Antes de deploy: `go test ./...`, build Linux, backup pre-deploy, smoke e readiness.
