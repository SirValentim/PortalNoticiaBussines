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

- [ ] Tela de auditoria.
- [ ] Filtros por usuario, entidade, acao e periodo.
- [ ] Registro de diffs importantes.
- [ ] Auditoria para usuarios, posts, banners, lojas, promocoes e configuracoes.
- [ ] Painel de configuracoes gerais.
- [ ] Nome do portal, logo, favicon e redes sociais.
- [ ] Email, WhatsApp e contatos.
- [ ] SEO global.
- [ ] Configuracoes de upload.
- [ ] Configuracoes de automacao.

Criterio de pronto:

- Administrador consegue entender quem mudou o que.
- Configuracoes basicas nao exigem deploy.

### Sprint 7 - Eventos, classificados e modulos locais

Objetivo: completar os produtos locais do portal.

Entregas:

- [ ] CRUD completo de eventos.
- [ ] Campos de evento: data, horario, local, organizador, ingresso, imagem, status e SEO.
- [ ] CRUD completo de classificados.
- [ ] Campos de classificados: categoria, preco, contato, validade, destaque e status.
- [ ] Melhorar CRUD de lojas com SEO e status comercial.
- [ ] Melhorar promocoes com regras de validade e relatorio.
- [ ] Melhorar influencers com categorias/nichos e destaque.
- [ ] Sitemap para todos os modulos ativos.

Criterio de pronto:

- Cada modulo local tem painel, pagina publica, SEO e status.
- Comercial consegue vender destaque sem depender de desenvolvimento.

### Sprint 8 - Anuncios e relatorios comerciais

Objetivo: profissionalizar monetizacao.

Entregas:

- [ ] Relatorio por anunciante.
- [ ] Relatorio por banner, posicao e periodo.
- [ ] Impressao e clique por campanha.
- [ ] Filtro por status comercial.
- [ ] Exportacao CSV simples.
- [ ] Validacao para evitar banners conflitantes por posicao/periodo.
- [ ] Produto de classificados em destaque.
- [ ] Produto de eventos patrocinados.
- [ ] Produto de lojas patrocinadas.
- [ ] Regra visual para evitar anuncios consecutivos no feed.

Criterio de pronto:

- Comercial consegue prestar contas a anunciante.
- Campanhas tem indicadores basicos confiaveis.

### Sprint 9 - Automacao de noticias

Objetivo: criar coleta automatizada sem publicacao automatica.

Entregas:

- [ ] Cadastro de fontes RSS.
- [ ] Cadastro de fontes oficiais.
- [ ] Job de coleta.
- [ ] Deduplicacao por URL.
- [ ] Deduplicacao por titulo.
- [ ] Similaridade basica para evitar repeticao.
- [ ] Criacao automatica de rascunhos.
- [ ] Fila de revisao.
- [ ] Tela de execucoes.
- [ ] Logs visiveis por execucao.
- [ ] Botao executar agora.
- [ ] Nunca publicar automaticamente.

Criterio de pronto:

- Automacao cria rascunhos rastreaveis.
- Editor humano aprova tudo antes de publicar.

### Sprint 10 - IA editorial desacoplada

Objetivo: preparar assistentes editoriais sem prender o CMS a um fornecedor.

Entregas:

- [ ] Interface `EditorialAIProvider`.
- [ ] Provider mock/teste.
- [ ] Gerar titulo.
- [ ] Gerar subtitulo.
- [ ] Gerar resumo.
- [ ] Gerar meta description.
- [ ] Sugerir tags.
- [ ] Reescrever em tom jornalistico.
- [ ] Criar chamada para redes sociais.
- [ ] Verificar possivel duplicidade.
- [ ] Guardrails: nao inventar fatos.
- [ ] Preservar fonte original.
- [ ] Marcar conteudo automatizado como rascunho.
- [ ] Logar uso de IA.

Criterio de pronto:

- IA auxilia, mas nao publica.
- Toda sugestao tem revisao humana.
- Camada pode trocar de fornecedor sem reescrever o CMS.

### Sprint 11 - PostgreSQL e migrations versionadas

Objetivo: evoluir a persistencia para padrao corporativo.

Entregas:

- [ ] Definir estrategia SQLite -> PostgreSQL.
- [ ] Adicionar driver PostgreSQL.
- [ ] Configurar `DB_DRIVER` ou deteccao segura.
- [ ] Criar runner de migrations.
- [ ] Versionar migrations incrementais.
- [ ] Portar repository para SQL compativel ou separar dialetos.
- [ ] Migrar FTS5 para PostgreSQL full-text.
- [ ] Ajustar worker com lock concorrente seguro.
- [ ] Criar script de migracao de dados.
- [ ] Testar restore e rollback.

Criterio de pronto:

- Ambiente de producao pode rodar em PostgreSQL.
- Migrations sao aplicadas de forma previsivel.
- Dados do SQLite sao migrados com validacao.

### Sprint 12 - Operacao, observabilidade e hardening

Objetivo: deixar o CMS mais confiavel para operacao continua.

Entregas:

- [ ] Monitor externo de uptime.
- [ ] Backup externo fora da VPS.
- [ ] Alertas de falha de backup.
- [ ] Alertas de disco.
- [ ] Logs estruturados.
- [ ] Revisao CSP para reduzir `unsafe-inline`.
- [ ] Politica de retencao de logs.
- [ ] Smoke test automatizado pos-deploy.
- [ ] Checklist de rollback.
- [ ] Documentar rotina de publicacao e emergencia no checklist.

Criterio de pronto:

- Falhas sao detectadas cedo.
- Deploy tem rollback claro.
- Backups sao restauraveis e testados.

### Sprint 13 - Refinamento premium de UX

Objetivo: elevar o painel para experiencia comercial premium.

Entregas:

- [ ] Sidebar com grupos e melhor hierarquia.
- [ ] Header admin com usuario, papel e atalhos.
- [ ] Toasts de sucesso/erro.
- [ ] Loading states.
- [ ] Estados vazios consistentes.
- [ ] Confirmacoes de exclusao melhores que `confirm()` nativo.
- [ ] Tabelas com filtros persistentes.
- [ ] Paginacao consistente em todos os modulos.
- [ ] Responsividade completa do painel.
- [ ] Revisao visual de formularios longos.

Criterio de pronto:

- Painel e rapido, claro e confortavel para uso diario.
- Usuarios nao precisam entender detalhes tecnicos para operar.

## Ordem de Prioridade Curta

Para as proximas execucoes, a ordem mais segura e:

1. Publicar edicao completa de usuarios na VPS.
2. Aplicar permissoes por acao nos handlers.
3. Criar CRUD de categorias e tags.
4. Criar biblioteca de midia.
5. Refatorar handlers grandes em modulos.
6. Implementar fluxo editorial de revisao/aprovacao.
7. Criar auditoria e configuracoes.
8. Implementar automacao RSS.
9. Implementar IA editorial.
10. Migrar para PostgreSQL.

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
