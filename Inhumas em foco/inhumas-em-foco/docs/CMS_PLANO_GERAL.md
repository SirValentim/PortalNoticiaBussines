# CMS Premium - Plano Geral

Documento geral do CMS do portal Inhumas em Foco.

Este arquivo consolida o cronograma do que ainda falta para transformar o MVP atual em um CMS corporativo profissional. O `CHECKLIST_STATUS.md` continua sendo o controle vivo de execucao; este documento organiza a visao macro, a ordem das fases e os criterios de pronto.

## Estado Atual

O projeto ja possui uma base operacional relevante:

- Backend Go server-side com `cmd/web` e `cmd/worker`.
- Painel admin com noticias, usuarios, banners, lojas, promocoes, bairros, influencers, metricas e jobs.
- Autenticacao com bcrypt, sessoes, CSRF, rate limit e cookies seguros.
- Upload local com validacao, WebP e thumbnails.
- SEO tecnico inicial com sitemap, robots, RSS, canonical, Open Graph e JSON-LD.
- Deploy VPS com Nginx, HTTPS, systemd, worker, backup e health check.
- RBAC inicial com papeis premium.

Mesmo assim, ainda faltam camadas importantes para um CMS corporativo: PostgreSQL real, migrations versionadas, fluxo editorial completo, biblioteca de midia, tags, auditoria ampla, configuracoes do portal, automacao de noticias, IA editorial, relatorios comerciais e refatoracao modular.

## Principios de Execucao

- Implementar em incrementos pequenos e testaveis.
- Manter o portal em producao funcionando.
- Fazer backup antes de qualquer deploy na VPS.
- Rodar `go test ./...` e build antes de publicar.
- Evitar reescrita completa quando o modulo atual puder evoluir com seguranca.
- Preferir primeiro a solidez editorial e operacional, depois automacoes avancadas.

## Cronograma Recomendado

As duracoes abaixo sao estimativas em sprints de 1 semana. Algumas fases podem ser aceleradas se o escopo for reduzido, mas a ordem recomendada evita construir automacoes sobre uma base fraca.

### Sprint 0 - Organizacao e documentacao oficial

Objetivo: consolidar os documentos vivos do projeto.

Status: concluido no escopo atual.

Entregas:

- [x] Definir `CHECKLIST_STATUS.md` como checklist vivo.
- [x] Criar documento geral do CMS.
- [ ] Marcar documentos antigos como historicos quando forem substituidos.
- [ ] Atualizar README apontando para os documentos oficiais atuais.
- [ ] Registrar a decisao de evoluir de MVP para CMS premium.

Criterio de pronto:

- Qualquer nova tarefa entra no checklist.
- O plano macro fica neste documento.
- Documentos antigos nao guiam mais a execucao diaria.

### Sprint 1 - Usuarios, permissoes e seguranca de painel

Objetivo: fechar a base corporativa de acesso.

Entregas:

- [x] Papeis premium: Super Admin, Admin, Editor, Redator, Revisor e Comercial.
- [x] Edicao completa de usuario no painel.
- [ ] Publicar edicao completa de usuario na VPS.
- [ ] Aplicar permissoes por acao nos handlers.
- [ ] Impedir Redator de publicar, excluir ou editar conteudo de terceiros.
- [ ] Permitir Revisor aprovar/reprovar sem acesso administrativo amplo.
- [ ] Restringir Comercial aos modulos comerciais.
- [ ] Criar log de auditoria para login, criacao, edicao e alteracao de senha.
- [ ] Criar politica de senha mais forte.
- [ ] Criar recuperacao segura de senha.

Criterio de pronto:

- Cada rota administrativa exige permissao clara.
- Ninguem consegue escalar privilegio por formulario.
- Acoes sensiveis aparecem na auditoria.

### Sprint 2 - Refatoracao modular da base Go

Objetivo: reduzir o risco do `handler.go` grande e preparar crescimento.

Entregas:

- [~] Separar handlers por dominio: public, admin, users, posts, commercial, system.
- [x] Separar handlers administrativos comerciais de banners e promocoes em arquivo proprio.
- [x] Separar handlers administrativos de usuarios em arquivo proprio.
- [x] Separar handlers administrativos de noticias em arquivo proprio.
- [x] Separar handlers administrativos de lojas, influencers e bairros em arquivo proprio.
- [x] Separar handlers administrativos de metricas, auditoria, jobs mortos e tracking operacional em arquivo proprio.
- [x] Criar service de recuperacao de senha em `internal/users`.
- [x] Criar service de administracao de usuarios em `internal/users`.
- [x] Criar service de regras editoriais de posts em `internal/posts`.
- [x] Criar service inicial de regras comerciais em `internal/commercial`.
- [>] Criar services para regras de negocio comerciais restantes e demais dominios.
- [ ] Manter repositories focados em persistencia.
- [ ] Padronizar validacoes e mensagens de erro.
- [ ] Criar helpers compartilhados para renderizacao de formularios admin.
- [ ] Revisar logs e substituir mensagens soltas por padrao estruturado.

Criterio de pronto:

- Novos modulos nao precisam editar um arquivo gigante.
- Regras importantes ficam em services testaveis.
- Testes atuais continuam passando.

### Sprint 3 - Categorias e tags profissionais

Objetivo: fechar taxonomia editorial.

Status: concluido no escopo atual.

Entregas:

- [x] CRUD completo de categorias no painel.
- [x] Campos de categoria: nome, slug, descricao, imagem, ordem, status.
- [x] CRUD completo de tags.
- [x] Relacao muitos-para-muitos entre posts e tags.
- [x] Filtros por categoria e tag no painel.
- [x] Paginas publicas de tag.
- [x] SEO por categoria e tag.
- [x] Slug unico com tratamento seguro de conflito em categorias.
- [x] Slug unico com tratamento seguro de conflito em tags.
- [x] Bloqueio de exclusao de categoria com noticias vinculadas.
- [x] Bloqueio de exclusao de tag com noticias vinculadas.
- [x] Auditoria de criacao, edicao e exclusao de categorias.
- [x] Auditoria de criacao, edicao e exclusao de tags.
- [x] Busca publica considera tags ativas.

Criterio de pronto:

- Editor consegue organizar todo conteudo sem acessar banco.
- Tags aparecem no post, SEO, sitemap e busca.

### Sprint 4 - Noticias e fluxo editorial premium

Objetivo: transformar o CRUD atual em fluxo editorial profissional.

Status: concluido no escopo atual.

Entregas:

- [x] Status editorial completo com transicoes controladas.
- [x] Regras: rascunho -> revisao -> aprovado/publicado.
- [x] Botao "enviar para revisao".
- [x] Botao "aprovar".
- [x] Botao "reprovar com comentario".
- [x] Historico de revisoes.
- [x] Autosave.
- [x] Preview com token seguro ou permissao.
- [x] Checklist SEO antes da publicacao.
- [x] Tempo estimado de leitura.
- [x] Campo de fonte original.
- [x] Galeria por noticia.
- [x] Fixar no topo separado de destaque da home.
- [x] Validacoes obrigatorias por status.

Criterio de pronto:

- Conteudo nao vai ao ar sem permissao adequada.
- Existe rastreabilidade de quem criou, revisou e publicou.
- O editor tem checklist claro antes de publicar.

### Sprint 5 - Biblioteca de midia

Objetivo: sair de upload por formulario para biblioteca reutilizavel.

Status: concluido no escopo atual.

Entregas:

- [x] Tela de biblioteca de midia.
- [x] Listagem paginada de imagens.
- [x] Busca por nome e alt text.
- [x] Registro de arquivo no banco.
- [x] Alt text obrigatorio/recomendado.
- [x] Organizacao por ano/mes.
- [x] Reaproveitamento de midia em noticias, lojas, banners, promocoes e influencers.
- [x] Remocao segura com checagem de uso.
- [x] Tela de detalhes da midia no proprio card.
- [x] Preparar interface para provider S3 futuro via `storage.Provider`.

Criterio de pronto:

- Editor escolhe imagem existente sem novo upload.
- Midias perigosas continuam bloqueadas.
- Remocao nao quebra conteudo publicado.

### Sprint 6 - Auditoria e configuracoes do portal

Objetivo: dar controle corporativo e rastreabilidade.

Entregas:

- [x] Tela de auditoria.
- [x] Filtros por usuario, entidade, acao e periodo.
- [x] Registro de diffs importantes.
- [x] Auditoria para usuarios, posts, banners, lojas, promocoes e configuracoes.
- [x] Painel de configuracoes gerais.
- [x] Nome do portal, logo, favicon e redes sociais.
- [x] Email, WhatsApp e contatos.
- [x] SEO global.
- [x] Configuracoes de upload.
- [x] Configuracoes de automacao.

Criterio de pronto:

- Administrador consegue entender quem mudou o que.
- Configuracoes basicas nao exigem deploy.

### Sprint 7 - Eventos, classificados e modulos locais

Objetivo: completar os produtos locais do portal.

Entregas:

- [x] CRUD completo de eventos.
- [x] Campos de evento: data, horario, local, organizador, ingresso, imagem, status e SEO.
- [x] CRUD completo de classificados.
- [x] Campos de classificados: categoria, preco, contato, validade, destaque e status.
- [x] Melhorar CRUD de lojas com SEO e status comercial.
- [x] Melhorar promocoes com regras de validade e relatorio.
- [x] Melhorar influencers com categorias/nichos e destaque.
- [x] Sitemap para todos os modulos ativos.

Criterio de pronto:

- Cada modulo local tem painel, pagina publica, SEO e status.
- Comercial consegue vender destaque sem depender de desenvolvimento.

### Sprint 8 - Anuncios e relatorios comerciais

Objetivo: profissionalizar monetizacao.

Entregas:

- [x] Relatorio por anunciante.
- [x] Relatorio por banner, posicao e periodo.
- [x] Impressao e clique por campanha.
- [x] Filtro por status comercial.
- [x] Exportacao CSV simples.
- [x] Validacao para evitar banners conflitantes por posicao/periodo.
- [x] Produto de classificados em destaque.
- [x] Produto de eventos patrocinados.
- [x] Produto de lojas patrocinadas.
- [x] Regra visual para evitar anuncios consecutivos no feed.

Criterio de pronto:

- Comercial consegue prestar contas a anunciante.
- Campanhas tem indicadores basicos confiaveis.

### Sprint 9 - Automacao de noticias

Objetivo: criar coleta automatizada sem publicacao automatica.

Entregas:

- [x] Cadastro de fontes RSS.
- [x] Cadastro de fontes oficiais.
- [x] Job de coleta.
- [x] Deduplicacao por URL.
- [x] Deduplicacao por titulo.
- [x] Similaridade basica para evitar repeticao.
- [x] Criacao automatica de rascunhos.
- [x] Fila de revisao.
- [x] Tela de execucoes.
- [x] Logs visiveis por execucao.
- [x] Botao executar agora.
- [x] Nunca publicar automaticamente.

Criterio de pronto:

- Automacao cria rascunhos rastreaveis.
- Editor humano aprova tudo antes de publicar.

### Sprint 10 - IA editorial desacoplada

Objetivo: preparar assistentes editoriais sem prender o CMS a um fornecedor.

Entregas:

- [x] Interface `EditorialAIProvider`.
- [x] Provider mock/teste.
- [x] Gerar titulo.
- [x] Gerar subtitulo.
- [x] Gerar resumo.
- [x] Gerar meta description.
- [x] Sugerir tags.
- [x] Reescrever em tom jornalistico.
- [x] Criar chamada para redes sociais.
- [x] Verificar possivel duplicidade.
- [x] Guardrails: nao inventar fatos.
- [x] Preservar fonte original.
- [x] Marcar conteudo automatizado como rascunho.
- [x] Logar uso de IA.

Criterio de pronto:

- IA auxilia, mas nao publica.
- Toda sugestao tem revisao humana.
- Camada pode trocar de fornecedor sem reescrever o CMS.

### Sprint 11 - PostgreSQL e migrations versionadas

Objetivo: evoluir a persistencia para padrao corporativo.

Entregas:

- [x] Definir estrategia SQLite -> PostgreSQL.
- [x] Adicionar driver PostgreSQL.
- [x] Configurar `DB_DRIVER` ou deteccao segura.
- [x] Criar runner de migrations.
- [x] Versionar migrations incrementais.
- [x] Portar repository para SQL compativel ou separar dialetos.
- [x] Migrar FTS5 para PostgreSQL full-text.
- [x] Ajustar worker com lock concorrente seguro.
- [x] Criar script de migracao de dados.
- [x] Testar restore e rollback.

Criterio de pronto:

- Ambiente de producao pode rodar em PostgreSQL.
- Migrations sao aplicadas de forma previsivel.
- Dados do SQLite sao migrados com validacao.

### Sprint 12 - Operacao, observabilidade e hardening

Objetivo: deixar o CMS mais confiavel para operacao continua.

Status: concluido no escopo atual.

Entregas:

- [x] Monitor externo de uptime documentado para `/health`.
- [x] Backup externo fora da VPS preparado via rclone ou volume montado.
- [x] Alertas de falha de backup adicionados no script de backup e no verificador de backup recente.
- [x] Alertas de disco configuraveis por `DISK_ALERT_THRESHOLD`.
- [x] Logs estruturados JSON com `request_id`, status, rota, IP, bytes e duracao.
- [x] Revisao CSP aplicada com diretivas mais restritivas e modo `Content-Security-Policy-Report-Only` para auditar remocao futura de `unsafe-inline`.
- [x] Politica de retencao de logs documentada e script `journal-retention.sh` criado.
- [x] Smoke test automatizado pos-deploy criado em `scripts/smoke-test.sh`.
- [x] Checklist de rollback documentado.
- [x] Rotina de publicacao, emergencia, monitoramento e backup documentada no checklist e no guia operacional.

Criterio de pronto:

- Falhas sao detectadas cedo.
- Deploy tem rollback claro.
- Backups sao restauraveis e testados.

### Sprint 13 - Refinamento premium de UX

Objetivo: elevar o painel para experiencia comercial premium.

Status: concluido no escopo atual.

Entregas:

- [x] Sidebar com grupos e melhor hierarquia.
- [x] Header admin com usuario, papel e atalhos.
- [x] Toasts de sucesso/erro.
- [x] Loading states em submits administrativos.
- [x] Estados vazios consistentes.
- [x] Confirmacoes de exclusao melhores que `confirm()` nativo.
- [x] Tabelas com filtros persistentes por tela.
- [x] Paginacao consistente preservada nos modulos que ja usam o componente atual.
- [x] Responsividade completa do painel revisada para sidebar/topbar/formularios.
- [x] Revisao visual de formularios longos com area de conteudo, topbar e acoes fixas preservadas.

Criterio de pronto:

- Painel e rapido, claro e confortavel para uso diario.
- Usuarios nao precisam entender detalhes tecnicos para operar.

### Sprint 14 - Revisao final corporativa e operacao continua

Objetivo: encerrar o ciclo de construcao do CMS com uma revisao executiva, checklist de readiness, backlog residual realista e documentacao oficial alinhada.

Status: concluido no escopo atual.

Entregas:

- [x] Documento `docs/CMS_REVISAO_FINAL.md` criado com status executivo, modulos concluidos, readiness, decisoes abertas e backlog residual.
- [x] Script `scripts/production-readiness.sh` criado para validar ambiente, services, health, home, admin protegido, backup recente, disco, Nginx e backup externo.
- [x] `.env.example` atualizado com variaveis de PostgreSQL, CSP report-only, backup externo, alertas e retencao.
- [x] `README.md` atualizado para apontar documentos oficiais, build completo e readiness pos-deploy.
- [x] `docs/DEPLOY_GUIDE.md` atualizado com PostgreSQL suportado, binarios atuais e readiness final.
- [x] `Makefile` atualizado com build completo, `build-linux`, `smoke` e `readiness`.
- [x] Backlog residual consolidado em prioridades alta, media e futura.

Criterio de pronto:

- A equipe sabe o que esta pronto, o que depende de infraestrutura externa e como validar producao.
- Deploy e manutencao possuem verificacao final repetivel.
- Documentos oficiais apontam para uma unica linha de operacao.

## Backlog Residual Curto

O CMS esta funcionalmente pronto para operacao inicial. O que resta agora e principalmente infraestrutura, governanca e evolucao:

1. Configurar backup externo real fora da VPS e testar restauracao.
2. Configurar monitor externo de uptime para `/health`.
3. Configurar SMTP transacional real para recuperacao de senha.
4. Revisar contas administrativas e remover qualquer segredo de bootstrap.
5. Homologar migracao SQLite -> PostgreSQL com base recente.
6. Criar testes de repository PostgreSQL em ambiente de CI/homologacao.
7. Remover `unsafe-inline` da CSP apos migrar scripts/styles inline.
8. Ativar Cloudflare/WAF quando DNS estiver sob gestao definitiva.

## Criterio Final de CMS Corporativo

O CMS pode ser considerado corporativo profissional quando:

- Usuarios e permissoes estao granulares e auditados.
- Noticias passam por fluxo editorial rastreavel.
- Midia, categorias e tags sao gerenciadas pelo painel.
- Automacoes criam apenas rascunhos revisaveis.
- SEO tecnico e editorial estao completos.
- Anuncios e modulos comerciais tem relatorios.
- Configuracoes principais nao dependem de deploy.
- Banco, backup, logs e deploy seguem padrao operacional seguro.
- A VPS tem monitoramento e backup externo validado.
