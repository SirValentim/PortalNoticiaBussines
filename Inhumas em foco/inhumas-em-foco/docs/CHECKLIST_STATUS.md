# Checklist de Status - Inhumas em Foco

Baseado no prompt unico do projeto e na revisao do estado atual.

## Documentos oficiais atuais

Para reduzir dispersao, a execucao do CMS passa a usar estes documentos como fonte principal:

- `docs/CHECKLIST_STATUS.md`: checklist vivo de execucao, status e historico.
- `docs/CMS_PLANO_GERAL.md`: plano macro do CMS premium, cronograma e criterios de pronto.

Os demais documentos em `docs/` ficam como historico, referencia operacional ou material antigo ate serem consolidados.

## Rodada atual - Organizacao do CMS premium

- [x] Criar documento geral do CMS com cronograma do que falta.
- [x] Registrar no checklist que `CHECKLIST_STATUS.md` e `CMS_PLANO_GERAL.md` sao os documentos vivos principais.
- [x] Atualizar README para apontar para os documentos oficiais atuais.
- [ ] Marcar documentos antigos como historicos ou consolidar o conteudo relevante no checklist/plano geral.
- [x] Publicar a edicao completa de usuarios na VPS.
- [x] Iniciar Sprint 1 do plano geral: permissoes por acao e seguranca de painel.
- [x] Aplicar permissao server-side para listagem/criacao/edicao/publicacao/exclusao de noticias.
- [x] Aplicar permissao server-side para gerenciamento completo de usuarios.
- [x] Deploy Sprint 1 inicial validado na VPS com `/health` e smoke autenticado em usuarios.
- [x] Expandir permissoes server-side para todos os modulos comerciais e configuracoes.
- [x] Proteger lojas, influencers, banners, promocoes, bairros, metricas e jobs mortos com RBAC por acao.
- [x] Publicar segunda rodada do Sprint 1 na VPS.
- [x] Validar no painel remoto lojas, influencers, banners, promocoes, bairros, metricas e jobs mortos com usuario admin.
- [x] Expandir auditoria para usuarios, lojas, influencers, banners, promocoes, bairros e publicacao de noticias.
- [x] Criar consulta de audit logs por entidade para testes e futura tela administrativa.
- [x] Publicar rodada de auditoria administrativa na VPS.
- [x] Validar health/login/sitemap e contagem da tabela `audit_logs` na VPS.
- [x] Criar tela administrativa de auditoria com filtros por usuario, acao, entidade, ID e periodo.
- [x] Adicionar item Auditoria no menu do painel e proteger por permissao de configuracoes.
- [x] Publicar tela administrativa de auditoria na VPS.
- [x] Validar `/painel/1c2dhviax7/audit` e filtros da auditoria com login admin em producao.
- [x] Implementar recuperacao de senha segura com token hashado, expiracao e uso unico.
- [x] Proteger formularios de recuperacao/redefinicao com CSRF.
- [x] Registrar eventos de recuperacao de senha na auditoria.
- [x] Publicar recuperacao de senha na VPS e validar fluxo basico.
- [x] Validar `/recuperar-senha` em producao e criacao da tabela `password_reset_tokens`.
- [x] Criar camada SMTP desacoplada para e-mails transacionais.
- [x] Integrar envio SMTP ao fluxo de recuperacao de senha quando configurado.
- [x] Documentar variaveis SMTP em `.env.example`, README e `docs/ENVIRONMENT.md`.
- [x] Publicar camada SMTP na VPS e validar paginas/health.
- [ ] Configurar credenciais SMTP reais na VPS para envio efetivo de recuperacao de senha.
- [x] Iniciar Sprint 2 de modularizacao: extrair regras de recuperacao de senha para `internal/users`.
- [x] Criar testes diretos de service para recuperacao de senha.
- [x] Publicar primeira refatoracao modular do Sprint 2 na VPS.
- [x] Validar `/health`, `/login` e `/recuperar-senha` apos extracao do service.
- [x] Extrair regras de administracao de usuarios para `internal/users.AdminService`.
- [x] Criar testes diretos para criacao, edicao protegida e troca de senha de usuarios.
- [x] Publicar refatoracao de administracao de usuarios na VPS.
- [x] Validar `/health`, `/login`, `/recuperar-senha` e `/users` apos extracao do service.
- [x] Extrair regras editoriais de posts para `internal/posts.EditorialService`.
- [x] Criar testes diretos para permissao de edicao, permissao de status e validacao editorial de posts.
- [x] Publicar refatoracao editorial de posts na VPS.
- [x] Validar `/health`, `/login`, `/recuperar-senha`, `/posts`, `/posts/new` e `/audit` apos extracao do service editorial.
- [x] Extrair regras comerciais de banners para `internal/commercial.Service`.
- [x] Centralizar parsing de periodo comercial para banners e promocoes em service testavel.
- [x] Iniciar Sprint 3 de taxonomia editorial com CRUD administrativo de categorias.
- [x] Ampliar categorias com imagem, ordem, status ativo/inativo e flag de notas editoriais.
- [x] Proteger categorias com permissao `settings:manage` e registrar criacao/edicao/exclusao na auditoria.
- [x] Bloquear exclusao de categoria quando houver noticias vinculadas.
- [x] Adicionar testes para criacao, edicao e bloqueio de exclusao de categorias.
- [x] Validar localmente com `go test ./...`.
- [x] Validar localmente com `go build ./cmd/web`, `go build ./cmd/worker` e `go build ./cmd/seed-news`.
- [x] Publicar CRUD de categorias na VPS e validar smoke autenticado.
- [x] Continuar Sprint 3 com CRUD de tags e relacao posts-tags.
- [x] Criar tabela `tags` e tabela relacional `post_tags`.
- [x] Criar CRUD administrativo de tags com slug automatico/editavel, descricao e status.
- [x] Proteger tags com permissao `settings:manage` e registrar auditoria.
- [x] Bloquear exclusao de tag vinculada a noticias.
- [x] Adicionar selecao de tags no formulario de noticias.
- [x] Exibir tags na pagina publica da noticia com links para `/tag/{slug}`.
- [x] Criar pagina publica de tag com listagem de noticias publicadas.
- [x] Incluir tags ativas no sitemap dinamico.
- [x] Adicionar testes para CRUD de tags, associacao em noticias e pagina publica de tag.
- [x] Validar localmente tags com `go test ./...`.
- [x] Validar localmente tags com `go build ./cmd/web`, `go build ./cmd/worker` e `go build ./cmd/seed-news`.
- [x] Publicar CRUD de tags na VPS e validar smoke autenticado/publico.
- [x] Criar testes diretos para status de banner, validacao de banner e periodo comercial.
- [x] Publicar refatoracao comercial na VPS.
- [x] Validar `/health`, `/login`, home, `/promocoes`, `/banners`, `/banners/new`, `/promotions` e `/promotions/new` apos extracao comercial.
- [x] Separar handlers administrativos de banners e promocoes em `internal/handler/admin_commercial.go`.
- [x] Validar testes/build apos reduzir o bloco comercial em `internal/handler/handler.go`.
- [x] Publicar separacao fisica dos handlers comerciais na VPS.
- [x] Validar `/health`, home, `/promocoes`, `/login`, `/banners`, `/banners/new`, `/promotions` e `/promotions/new` apos separacao dos handlers comerciais.
- [x] Separar handlers administrativos de usuarios em `internal/handler/admin_users.go`.
- [x] Validar testes/build apos reduzir o bloco de usuarios em `internal/handler/handler.go`.
- [x] Publicar separacao fisica dos handlers de usuarios na VPS.
- [x] Validar `/health`, `/login`, home, `/users` e edicao de usuario apos separacao dos handlers de usuarios.
- [x] Separar handlers administrativos de noticias em `internal/handler/admin_posts.go`.
- [x] Mover `adminPostFormData` e lock cooperativo de edicao para o arquivo de posts.
- [x] Validar testes/build apos reduzir o bloco de noticias em `internal/handler/handler.go`.
- [x] Publicar separacao fisica dos handlers de noticias na VPS.
- [x] Validar `/health`, `/login`, home, `/noticias`, `/posts`, `/posts/new` e edicao de noticia apos separacao dos handlers de noticias.
- [x] Separar handlers administrativos de lojas, influencers e bairros em `internal/handler/admin_local.go`.
- [x] Validar testes/build apos reduzir o bloco de modulos locais em `internal/handler/handler.go`.
- [x] Publicar separacao fisica dos handlers de modulos locais na VPS.
- [x] Validar `/health`, `/login`, `/lojas`, `/influencers`, `/stores`, `/stores/new`, `/influencers`, `/influencers/new` e `/neighborhoods` apos separacao dos handlers locais.
- [x] Separar handlers administrativos de metricas, auditoria e jobs mortos em `internal/handler/admin_system.go`.
- [x] Mover API operacional de tracking de metricas e helpers de filtros/rotulos para o arquivo de sistema.
- [x] Validar testes/build apos reduzir o bloco de sistema em `internal/handler/handler.go`.
- [x] Publicar separacao fisica dos handlers de sistema na VPS.
- [x] Validar `/health`, `/login`, `/metrics`, `/dead-jobs`, `/audit`, filtro de auditoria e POST `/api/metrics/{type}` apos separacao dos handlers de sistema.
- [ ] Continuar Sprint 2 separando handlers por dominio e reduzindo `internal/handler/handler.go`.

## Rodada atual - Frontend publico v2.0

- [x] Mockups `Mockup/01.png`, `Mockup/02.png` e `Mockup/prompt.md` revisados antes de editar.
- [x] Projeto alvo confirmado: `inhumas-em-foco` Go server-side com templates em `internal/view` e CSS em `static/css/style.css`.
- [x] Prototipo React em `app/` mantido fora do alvo principal porque o prompt v2.0 proibe React/Vue/jQuery.
- [x] Atualizar design system publico para paleta verde/amarelo, Inter, cards, badges e banners.
- [x] Recriar home mobile-first seguindo o mockup: hero, noticias, lojas, banner, promocoes, influencers e bairros.
- [x] Atualizar paginas publicas principais: noticias/categoria, detalhe, lojas, promocoes, busca e influencers.
- [x] Criar rotas publicas `/noticias` e `/eventos` no backend Go para combinar com a navegacao do mockup.
- [x] Adicionar fallback visual local em `static/images/inhumas-hero.png` para evitar imagens quebradas no banco vazio.
- [x] Atualizar logo publico a partir de `Mockup/logo.png`, com asset final em `static/images/logo.png`.
- [x] Criar alias `/post/{slug}` redirecionando para `/noticia/{slug}` conforme contrato do prompt.
- [x] Implementar `/classificados` como pagina publica inicial com categorias e CTA comercial.
- [x] Adicionar interatividade HTMX-like local para `hx-get`, `hx-target` e `hx-swap`, sem CDN externa.
- [x] Adicionar carregamento incremental em `/noticias/mais?page=N`.
- [x] Adicionar filtros/busca server-side em `/lojas` e filtros simples em `/promocoes`.
- [x] Redesenhar detalhes de loja, promocao e bairro no padrao do frontend publico v2.
- [x] Melhorar estados vazios com mensagens acionaveis e area visual consistente.
- [x] Validar build/testes Go antes de qualquer deploy.
- [x] Conferir rotas publicas no navegador local sem erros de console.
- [x] Smoke local por HTTP: `/`, `/noticias`, `/noticias/mais?page=2`, `/eventos`, `/lojas`, `/lojas?q=mercado`, `/promocoes`, `/classificados`, `/influencers`, `/busca`.
- [x] Deploy autorizado pelo usuario e publicado na VPS `68.233.125.102`.
- [x] Backup remoto pre-deploy criado em `/var/www/inhumas/backups/pre-frontend-v2-*.db`.
- [x] Servicos `inhumas-web` e `inhumas-worker` reiniciados e ativos.
- [x] `/health` remoto retorna OK por Nginx/IP.
- [x] Smoke remoto por IP: `/`, `/noticias`, `/eventos`, `/lojas`, `/promocoes`, `/classificados`, `/influencers`, `/busca`, `/static/images/logo.png` e `/static/js/htmx-lite.js`.
- [x] Correcao pos-deploy: logo gigante causado por CSS em cache; adicionados cache-busting em CSS/JS/logo e dimensoes inline no header/footer.
- [x] SEO tecnico local reforcado: canonical automatico, robots detalhado, Open Graph/Twitter, RSS discovery, tags e JSON-LD com logo/areaServed.
- [x] Pacote editorial inicial criado em `docs/SEO_CONTENT_PACK_INHUMAS.md`.
- [x] SQL de rascunhos criado em `scripts/seed-local-news-drafts.sql`.
- [x] Comando `cmd/seed-news` criado para aplicar rascunhos no SQLite.
- [x] Rascunhos aplicados no banco local como `draft`, sem publicacao publica.
- [x] SEO tecnico publicado na VPS: canonical, robots, Open Graph/Twitter, RSS discovery e JSON-LD.
- [x] Fase 4 SEO editorial finalizada: `robots.txt` dinamico, sitemap com paginas publicas principais, JSON-LD em grafo com Organization/WebSite/CollectionPage/BreadcrumbList/LocalBusiness e metas de keywords/author/theme-color.
- [x] Links internos de categorias corrigidos para slugs existentes, evitando 404 em navegacao editorial.
- [x] Quatro materias locais publicadas na VPS com fontes e slugs amigaveis.
- [x] Fase 5 de conteudo inicial finalizada com 6 novas materias perenes publicadas: Rio Meia-Ponte, economia local, municipios vizinhos, IFG Campus Inhumas, hidrografia e guia rapido de Inhumas GO.
- [x] Conteudos da fase 5 usam fontes institucionais/oficiais: IBGE, Prefeitura de Inhumas, Sudeco, Semad Goias, Saneago e IFG.
- [x] Todas as 10 postagens foram refatoradas com as skills locais `keyword-engine`, `outline-builder`, `seo-writer` e `seo-audit`: keyword, meta, introducao, H2s, FAQ, conclusao e fontes.
- [x] Sitemap remoto regenerado e validado em `/static/sitemap.xml`.
- [x] Alias SEO `/sitemap.xml` criado e validado apontando para o sitemap gerado.
- [x] Correcao pos-publicacao: detalhe de noticia agora aceita notas editoriais vazias/nulas sem gerar 404.
- [x] Notas editoriais internas removidas das materias publicadas.
- [x] Smoke remoto das quatro materias: HTTP 200, canonical e JSON-LD presentes.
- [x] Destaque principal da home deixou de depender apenas da data e agora pode ser controlado manualmente no painel por noticia.
- [x] Materia atual `quem-foi-o-primeiro-prefeito-de-inhumas` marcada como destaque manual no banco local e na VPS.
- [x] Bloco lateral da home renomeado de "Mais lidas" para "Recentes", evitando prometer metrica que ainda nao existe.
- [x] Formularios de edicao agora permitem remover imagem de capa de noticia, logo/capa de loja, imagem de promocao e avatar/capa de influencer.
- [x] Correcao operacional na VPS: `scripts/backup.sh` e `disk-check.sh` com permissao executavel, backup manual validado e dead job antigo de permissao removido.
- [x] Rodada comercial de interface: placeholders "Anuncie aqui" na home, busca, categorias e materias quando nao houver banner vendido.
- [x] Materias agora aceitam banners reais de sidebar topo/rodape e fallback comercial automatico.
- [x] Paginas de lojas e promocoes ganharam chamadas comerciais para cadastro/divulgacao.
- [x] Fase 3 da interface publica finalizada: header responsivo com menu mobile, cards com microinteracoes, banners reais em listas internas, placeholders comerciais refinados e artigos com sidebar sticky.
- [x] Cache-buster publico atualizado para `20260430-public-v1` para forcar carregamento do CSS novo no navegador.
- [x] Painel organizado para venda: dashboard com inventario comercial e tela de banners com mapa de espacos, formatos e status.
- [x] Conteudo de teste `/noticia/02` removido do banco local/VPS e do sitemap.
- [x] Systemd web/worker ajustado para rodar como usuario `inhumas`; sitemap agora e gerado com permissao publica `0644`.
- [x] Prompt v3.0 incorporado ao `docs/inhumas_prompt_unico.md` com master layout grid, hero/sidebar pixel perfect, news grid, banners premium, microinteracoes e mobile.
- [x] Layout v3 aplicado no CSS: container 1200/24px, hero/sidebar 420px, news cards compactos, banners 100px e mobile hero 250px.
- [x] Banner placeholder da sidebar superior da home removido temporariamente ate nova refatoracao visual.
- [x] SSL Lets Encrypt instalado para `inhumasemfoco.online` e `www.inhumasemfoco.online`, com renovacao automatica validada por `certbot renew --dry-run`.
- [x] Dominio oficial configurado no app como `https://inhumasemfoco.online`, cookies seguros ativados e `www` redirecionando para o dominio principal.
- [x] Revisao final do projeto em 30/04/2026: backend testado, prototipo React compilado e documentacao principal realinhada.
- [x] CSRF revisado para aplicar `MAX_UPLOAD_SIZE` tambem durante a leitura de formularios multipart no middleware.

Status geral: o MVP operacional em Go + SQLite esta concluido no repositorio, validado em VPS e com dominio HTTPS oficial configurado. As pendencias restantes foram reclassificadas como Pos-MVP, infraestrutura externa, rotina editorial continua ou migracao PostgreSQL futura.

Legenda:

- [x] Criado / funcional no projeto atual
- [~] Parcial / existe, mas precisa reforco antes de producao
- [ ] Ainda falta implementar
- [>] Pos-MVP / depende de infraestrutura externa ou decisao futura

## 1. Arquitetura e Stack

- [x] Backend Go com `cmd/web` e `cmd/worker`
- [x] Estrutura `internal/*` com auth, config, handler, middleware, model, repository, session, slug, storage e views
- [x] Templates HTML server-side com layouts, componentes e paginas
- [x] Static CSS em `static/css/style.css`
- [x] Worker separado em `cmd/worker`
- [x] Migration inicial PostgreSQL em `migrations/001_init_postgres.sql`
- [x] Runtime oficial do MVP usa SQLite local via `modernc.org/sqlite`
- [x] `.env.example` alinhado ao runtime SQLite MVP, com PostgreSQL documentado como proxima migracao
- [~] React separado existe como prototipo visual em `app/`
- [x] Decidir oficialmente se React sera integrado por API ou removido/mantido como prototipo
- [>] Adaptar repository/runtime para PostgreSQL real
- [x] Testar deploy MVP em VPS Ubuntu 24.04 com Nginx, systemd, worker, backup, swap e health

## 2. Banco de Dados e Migrations

- [x] Tabelas locais SQLite criadas automaticamente no startup
- [x] Usuarios
- [x] Categorias
- [x] Posts
- [x] Slug redirects
- [x] Lojas
- [x] Promocoes
- [x] Banners
- [x] Bairros
- [x] Jobs
- [x] Dead jobs
- [x] Metricas
- [x] Audit logs
- [x] Edit locks
- [x] Login attempts
- [x] Migration PostgreSQL inicial com enums, tabelas e indices principais
- [x] Busca local usa FTS5 no runtime SQLite com fallback seguro para `LIKE`
- [~] `search_vector` PostgreSQL existe na migration, mas nao no runtime atual
- [>] Migrations versionadas completas para evolucao do schema PostgreSQL
- [>] Runner de migration integrado ao deploy PostgreSQL
- [>] Repository compativel com PostgreSQL
- [>] Transacoes para operacoes multi-tabela criticas no runtime PostgreSQL
- [>] Slug unico transacional/serializable no runtime PostgreSQL

## 3. Rotas Publicas

- [x] Home `/`
- [x] Detalhe de noticia `/noticia/{slug}`
- [x] Categoria `/categoria/{slug}`
- [x] Lista de lojas `/lojas`
- [x] Detalhe de loja `/loja/{slug}`
- [x] Lista de promocoes `/promocoes`
- [x] Detalhe de promocao `/promocao/{slug}`
- [x] Bairro `/bairro/{slug}`
- [x] Lista de influencers `/influencers`
- [x] Detalhe de influencer `/influencer/{slug}`
- [x] Busca `/busca`
- [x] Login `/login`
- [x] Logout `/logout`
- [x] Health `/health`
- [x] Sobre `/sobre`
- [x] Contato `/contato`
- [x] Paginas publicas integradas ao repository SQLite do MVP
- [x] Sitemap publico `/static/sitemap.xml`
- [x] RSS publico em `/rss.xml`

## 4. Painel Admin

- [x] Path obscuro configuravel via `ADMIN_PATH_PREFIX`
- [x] Dashboard admin
- [x] CRUD basico de posts
- [x] CRUD basico de lojas
- [x] CRUD basico de influencers
- [x] CRUD basico de banners
- [x] CRUD basico de promocoes
- [x] CRUD basico de bairros
- [x] Criacao/listagem de usuarios
- [x] Tela simples de metricas
- [x] Edicao de lojas usa busca direta por ID no repository
- [x] Admin metrics deixou de ser placeholder textual e usa agregacoes reais
- [~] Alguns erros internos aparecem no painel via `err.Error()`
- [~] Tela de metricas reais implementada com totais e rankings por entidade; falta enriquecer nomes/periodos/relatorios
- [x] Edicao completa de banners e promocoes
- [x] Alteracao segura de senha no painel
- [x] Validacoes de formulario mais completas
- [x] Fluxo de publicacao/agendamento mais robusto
- [x] Upgrade visual do painel admin com formularios largos, grids responsivos e listagens padronizadas
- [x] Fase 2 da administracao editorial finalizada: lista de noticias com filtros por busca, status e categoria.
- [x] Formulario de noticia ganhou campos de SEO editorial: titulo SEO, descricao SEO e palavra-chave principal.
- [x] Preview seguro no painel para rascunhos/agendadas/arquivadas antes de publicar, com `noindex`.

## 5. Autenticacao, Sessao e RBAC

- [x] Login com bcrypt
- [x] Seed opcional do admin inicial por `INITIAL_ADMIN_PASSWORD`
- [x] Sessao com cookie HttpOnly e SameSiteStrict
- [x] RBAC editorial/comercial: `super_admin`, `admin`, `editor`, `redator`, `revisor`, `comercial`
- [x] Restricao de rotas por papel no painel
- [x] Papéis premium do CMS adicionados ao runtime, painel de usuarios e migration PostgreSQL inicial
- [x] Login attempts gravados em tabela
- [x] Bloqueio de brute force usa IP normalizado, sem porta de origem
- [x] Cookie `Secure` nao depende mais da porta 8080; usa `APP_ENV`, `FORCE_SECURE_COOKIES` e tipo de banco
- [x] Suporte a rotacao de duas chaves de sessao (`SESSION_SECRET` + `PREVIOUS_SESSION_SECRET`)
- [x] Normalizar IP real com cuidado para Nginx/Cloudflare quando a requisicao vem de proxy local/privado
- [x] Aplicar rate limit real por rota/IP
- [>] Politica de senha e troca obrigatoria do admin inicial

## 6. CSRF e Headers de Seguranca

- [x] CSRF em login e POSTs do painel
- [x] Token em cookie e campo/header
- [x] Headers basicos: X-Frame-Options, nosniff, Referrer-Policy, CSP e HSTS quando TLS direto
- [~] CSP permite `'unsafe-inline'`
- [x] Conteudo de post passa por sanitizacao com allowlist antes de renderizar HTML
- [x] HSTS considera `X-Forwarded-Proto: https` para ambiente atras de proxy
- [x] Sanitizar HTML de conteudo editorial com allowlist
- [>] Remover ou reduzir `'unsafe-inline'` da CSP apos retirar inline scripts/styles dos templates
- [x] Configurar headers finais tambem no Nginx
- [>] Testar CSRF e cookies em ambiente com proxy HTTPS

## 7. Upload e Storage

- [x] Interface `storage.Provider`
- [x] `LocalProvider`
- [x] Upload com key aleatoria
- [x] Validacao real de MIME por `http.DetectContentType`
- [x] Bloqueio de path traversal no upload com validacao do caminho final dentro de `UPLOAD_DIR`
- [x] Limite configuravel por `MAX_UPLOAD_SIZE`
- [x] Limite real de corpo aplicado com `http.MaxBytesReader` nos formularios multipart
- [x] Conversao/compressao WebP nos novos uploads
- [x] Geracao de thumbs WebP nos novos uploads
- [x] `Delete` sanitiza a key com o mesmo rigor do upload
- [x] `GenerateKey` nao prefixa mais `uploads/` nos novos uploads, evitando caminho duplicado
- [x] Conversao automatica para WebP
- [x] Redimensionamento hero/thumb nos novos uploads
- [x] Pasta original criada com retencao automatica curta via worker
- [>] Provider S3 futuro
- [x] Testes automatizados de storage/upload malicioso basico

## 8. Conteudo Editorial

- [x] Posts com status `draft`, `review`, `scheduled`, `published`, `archived`
- [x] Categoria `politica-bastidores`
- [x] Campos de apuracao: `editorial_notes` e `editor_responsible`
- [x] Validacao condicional para Politica & Bastidores no create/update
- [x] Slug automatico
- [x] Redirect 301 quando slug de post muda
- [x] Selo de patrocinado em partes principais dos templates
- [~] Slug nao e transacional
- [x] SEO operacional com meta tags, RSS, sitemap e JSON-LD
- [x] Conteudo local inicial publicado: primeiro prefeito, dados IBGE, Camara em 1947 e marco de 19 de marco
- [x] Conteudo inicial ampliado para pelo menos 10 materias locais publicadas com slugs SEO e fontes.
- [x] JSON-LD Article/Organization implementado nos templates publicos
- [>] Fluxo editorial completo com revisao/aprovacao multiusuario
- [>] Revisar/remover conteudo antigo de teste, especialmente `/noticia/02`, para nao prejudicar SEO
- [x] Edit locks com heartbeat no painel
- [x] Sitemap automatico via job `generate_sitemap`
- [x] JSON-LD Article/Organization completo
- [x] Testes para regras editoriais, campos SEO, preview de rascunho, `robots.txt`, sitemap e breadcrumbs estruturados

## 9. Monetizacao

- [x] Banners com posicao, periodo, prioridade e status
- [x] Banner hero
- [x] Banner in-feed
- [x] Sticky footer
- [x] Lojas patrocinadas/destaque
- [x] Promocoes patrocinadas
- [x] Conteudo patrocinado em posts
- [x] Metricas basicas de visualizacao/clique
- [x] Validacao de sobreposicao de banners cobre cadastro e edicao, excluindo o proprio banner
- [x] Fase 1 do painel comercial finalizada: banners agora tem cliente/anunciante, responsavel, telefone, WhatsApp, plano/valor, observacoes internas, status comercial, periodo, posicao, prioridade e link.
- [x] Banners podem remover imagem atual, salvar rascunho sem arte e bloquear publicacao ativa sem imagem.
- [x] Dashboard comercial mostra inventario, banners ativos, vencendo, pausados, rascunhos e campanhas sem imagem.
- [~] Tracking de metricas e muito basico
- [x] Dashboard comercial com resumo operacional para venda de espacos.
- [>] Regra de feed misto garantindo nunca 2 anuncios seguidos
- [>] Produto classificados destaque
- [>] Eventos patrocinados
- [>] Newsletter patrocinada
- [>] Cupons exclusivos
- [>] Relatorios avancados por anunciante
- [x] Produto inicial de influencers com destaque/patrocinado e perfil publico

## 10. Jobs e Automacoes

- [x] Tabela `jobs`
- [x] Tabela `dead_jobs`
- [x] Worker com polling a cada 30s
- [x] Job `publish_post`
- [x] Job `expire_promotion`
- [x] Job `expire_banner`
- [x] Job `cleanup_old_jobs`
- [x] Job `vacuum_db`
- [x] Jobs recorrentes agendados pelo worker para backup, sitemap, vacuum e cleanup
- [x] Worker SQLite reivindica jobs vencidos marcando `running` antes da execucao
- [>] Worker PostgreSQL com `FOR UPDATE SKIP LOCKED` fica para a migracao PostgreSQL
- [x] Retry/backoff real no runtime atual: 1min, 5min e 15min antes de `dead_jobs`
- [x] Status `running` usado como lock de processamento no runtime SQLite MVP
- [x] `generate_sitemap` gera `static/sitemap.xml`
- [x] Backup automatico como job real
- [x] Vacuum agendado pelo worker
- [x] Sitemap automatico
- [x] Cleanup semanal agendado pelo worker
- [x] Limpeza automatica de originais antigos
- [x] Dead jobs com relatorio/alerta no painel admin

## 11. Deploy, VPS e Operacao

- [x] `.env.example`
- [x] `Makefile`
- [x] Exemplo Nginx
- [x] Exemplo systemd web
- [x] Exemplo systemd worker
- [x] Script `backup.sh`
- [x] Script `disk-check.sh`
- [x] Makefile usa `CGO_ENABLED=0` no build Linux
- [x] Deploy validado na VPS em `/var/www/inhumas`
- [x] Health check valida banco, escrita no diretorio de uploads, jobs e status de backup
- [>] PostgreSQL tunado na VPS
- [x] Swap 1GB configurado na VPS
- [x] SSL/HSTS validado com certificado Lets Encrypt no dominio oficial
- [>] Cloudflare
- [>] UptimeRobot
- [x] Cron de backup diario configurado na VPS
- [x] Cron de alerta de disco configurado na VPS
- [x] Endpoint `/metrics` protegido com dados reais
- [>] Modo manutencao testado em producao

## 12. React / Prototipo Visual

- [x] Projeto Vite + React + TypeScript criado
- [x] Tailwind + componentes UI instalados
- [x] Rotas principais do portal no React
- [x] Dados mockados para posts, lojas, promocoes, banners, bairros e usuarios
- [x] Paginas principais criadas
- [~] Login demo aceita senha fixa `admin123`
- [x] `BrowserRouter` duplicado corrigido
- [x] `README.md` do React atualizado para o projeto
- [x] Build React validado apos instalar dependencias locais
- [x] Instalar/verificar dependencias Node
- [x] Corrigir router duplicado
- [x] Atualizar README do app React
- [~] `npm install` reporta vulnerabilidades no audit que ainda precisam triagem
- [x] Decidir integracao com backend Go
- [>] Trocar mocks por API real, se React for mantido

## 13. Testes

- [x] `go test ./...` passa por compilacao
- [x] `npm run build` do React passa apos instalar dependencias
- [x] Testes Go adicionados para auth, middleware, repository/jobs e storage
- [~] React ainda tem alerta de chunk maior que 500 kB no build
- [x] Testes de auth/login
- [x] Testes de CSRF
- [x] Testes de RBAC
- [x] Testes de storage/upload
- [x] Testes de jobs/backoff no repository
- [x] Testes de sitemap
- [x] Testes de handlers publicos
- [x] Testes de JSON-LD e metricas admin
- [>] Testes de repository PostgreSQL
- [x] Smoke test do deploy por IP: `/health`, `/login`, systemd, Nginx e backup/restore SQLite

## 14. Documentacao

- [x] Prompt unico em `docs/inhumas_prompt_unico.md`
- [x] Comparacao essencial em `docs/PROMPT_COMPARISON.md`
- [x] Registro de sessao em `docs/SESSION_LOG.md`
- [x] Checklist de status atual neste arquivo
- [x] README do backend principal criado
- [x] README do React atualizado com status, rotas e comandos
- [x] Docs apontam pendencias e agora incluem roteiro operacional
- [x] Criar README principal do backend
- [x] Criar guia de deploy passo a passo
- [x] Criar guia de operacao diaria
- [x] Criar checklist de seguranca pre-producao
- [x] Criar documentacao de variaveis de ambiente

## Prioridade Recomendada

1. Publicar e validar o MVP SQLite em uma VPS Debian 12.
2. Configurar DNS/Cloudflare, TLS real, UptimeRobot, backup externo e alerta de disco.
3. Rodar smoke test completo de producao: login, CSRF, cookies Secure, uploads, worker, backup, health e `/metrics`.
4. Decidir se a proxima fase exige PostgreSQL agora ou se SQLite atende ao volume inicial.
5. Se PostgreSQL for escolhido, migrar repository/runtime, runner de migrations e worker concorrente.
6. Evoluir monetizacao Pos-MVP: relatorios por anunciante, classificados, eventos, newsletter e cupons.

## Etapas Executadas

### Etapa 1 - Seguranca base e prototipo React

- [x] Sanitizacao do HTML editorial com `bluemonday.UGCPolicy()`.
- [x] Login attempts agora usam IP normalizado sem porta.
- [x] Metricas e audit logs tambem usam o mesmo IP normalizado.
- [x] Cookies seguros deixam de depender da porta 8080.
- [x] HSTS passa a considerar `X-Forwarded-Proto: https`.
- [x] Router duplicado do React corrigido.
- [x] README do prototipo React substituido por documentacao do projeto.
- [x] Dependencias Node instaladas/verificadas.
- [x] Build React validado com `npm run build`.
- [~] `npm install` apontou 9 vulnerabilidades no audit; triagem fica para etapa futura.
- [~] Build React emite alerta de bundle acima de 500 kB; otimizacao fica para etapa futura.

Proxima etapa sugerida: corrigir storage/upload parcial (`Delete` seguro, caminho de upload sem duplicacao).

### Etapa 2 - Storage e upload

- [x] `LocalProvider.Upload` agora resolve a key para um caminho absoluto validado dentro de `UPLOAD_DIR`.
- [x] `LocalProvider.Delete` reutiliza a mesma validacao de key/caminho do upload.
- [x] `GenerateKey` deixou de prefixar `uploads/` nos novos arquivos.
- [x] Formularios multipart usam `http.MaxBytesReader` para aplicar limite real de upload.
- [x] Testes adicionados para key sem prefixo duplicado, path traversal, MIME invalido, upload e delete.
- [~] Conversao WebP e thumbs continuam para uma etapa propria.

Proxima etapa sugerida: escolher entre avancar em WebP/thumbs ou atacar testes de auth/CSRF/RBAC.

### Etapa 3 - Testes de seguranca e permissoes

- [x] Testes de auth/login com senha valida, usuario ausente, usuario inativo e senha errada.
- [x] Testes de permissoes por papel: admin, editor e comercial.
- [x] `RequirePermission` deixa de entrar em panic quando `authService` nao esta no context.
- [x] Testes de CSRF para rejeicao sem token e aceite com token valido.
- [x] Testes de RBAC do painel admin.
- [x] Testes de `ClientIP`, HSTS atras de proxy HTTPS e rate limit por IP.

### Etapa 4 - Operacao leve e jobs

- [x] Health check agora valida banco e escrita no diretorio de uploads.
- [x] Health responde `503` quando algum check falha.
- [x] Makefile passou a compilar Linux com `CGO_ENABLED=0`.
- [x] `.env.example` documenta `APP_ENV=production` e `FORCE_SECURE_COOKIES=true`.
- [x] README principal do backend criado.
- [x] Worker passa a reagendar falhas com backoff antes de mover para `dead_jobs`.
- [x] Teste de backoff/dead jobs adicionado no repository.

Proxima etapa sugerida: WebP/thumbs ou sitemap automatico.

### Etapa 5 - Sitemap automatico

- [x] Config adicionada para `SITE_URL` e `STATIC_DIR`.
- [x] Arquivos estaticos agora usam `STATIC_DIR`.
- [x] Repository lista entradas publicas para sitemap: posts publicados, lojas ativas, promocoes ativas e bairros.
- [x] Pacote `internal/sitemap` gera XML compativel com o protocolo de sitemap.
- [x] Worker executa o job `generate_sitemap` gravando `static/sitemap.xml`.
- [x] Worker agenda automaticamente `generate_sitemap`, `vacuum_db` e `cleanup_old_jobs` quando nao ha job ativo desses tipos.
- [x] Jobs criados sem `MaxAttempts` recebem padrao 3.
- [x] Testes adicionados para montagem do XML, escrita do arquivo e selecao de conteudo publico.

Proxima etapa sugerida: WebP/thumbs ou testes de handlers publicos.

### Etapa 6 - Testes de handlers publicos

- [x] Testes adicionados para `/health`, login GET, home vazia e detalhe de noticia.
- [x] Teste de detalhe de noticia confirma sanitizacao do HTML armazenado.
- [x] Templates agora usam `PROJECT_ROOT`, evitando falha quando o binario/teste roda fora da raiz.
- [x] Queries de categoria e posts usam `COALESCE` para campos opcionais vindos de `LEFT JOIN` ou seeds sem descricao.
- [x] `.env.example` e README documentam `PROJECT_ROOT`.

Proxima etapa sugerida: WebP/thumbs ou primeira fase de metricas/admin.

### Etapa 7 - WebP e thumbs

- [x] `LocalProvider.Upload` passa a salvar o arquivo original em `original/`.
- [x] Novos uploads geram WebP principal em `webp/`.
- [x] Novos uploads geram thumb WebP em `thumb/`.
- [x] Handlers passam a persistir a key retornada pelo storage (`webp/<arquivo>.webp`).
- [x] `LocalProvider.Delete` remove WebP principal, thumb e original relacionado.
- [x] Testes de storage atualizados para validar original, WebP, thumb e delecao.
- [x] Validado com `CGO_ENABLED=0 go test ./...`.
- [x] Retencao automatica dos originais antigos implementada em etapa posterior.

Proxima etapa sugerida: metricas/admin ou retencao/limpeza de originais antigos.

### Etapa 9 - SEO estruturado e metricas admin

- [x] JSON-LD `NewsMediaOrganization` adicionado automaticamente para paginas publicas com SEO.
- [x] JSON-LD `NewsArticle` adicionado em detalhes de noticia, com autor, datas, imagem e patrocinio quando disponiveis.
- [x] URLs absolutas de SEO em home e noticias passam a usar `SITE_URL`.
- [x] Tela admin de metricas substitui placeholder textual por totais por tipo e rankings de noticias, lojas e banners.
- [x] Dashboard admin passa a contar metricas globais por tipo, nao apenas entidade ID zero.
- [x] Repository ganhou agregacoes `MetricCountByType`, `MetricTotals` e `MetricTopEntities`.
- [x] Testes adicionados para JSON-LD, tela de metricas admin e agregacoes de metricas.

Proxima etapa sugerida: enriquecer metricas com nomes/periodos ou implementar alerta/relatorio de dead jobs.

### Etapa 10 - Retencao automatica de originais

- [x] `ORIGINAL_RETENTION_DAYS` adicionado com padrao de 7 dias.
- [x] `LocalProvider.CleanupOriginals` remove apenas arquivos antigos em `uploads/original`.
- [x] Arquivos WebP principais e thumbs sao preservados pela limpeza.
- [x] Worker agenda o job recorrente `compress_old_uploads` semanalmente.
- [x] Worker executa `compress_old_uploads` como limpeza de originais antigos usando a retencao configurada.
- [x] Teste de storage cobre remocao de original antigo e preservacao de original recente/WebP.

Proxima etapa sugerida: metricas por periodo/anunciante ou edit locks com heartbeat.

### Etapa 12 - Senha administrativa e rate limit

- [x] Painel de usuarios ganhou formulario para alteracao administrativa de senha.
- [x] Alteracao de senha exige confirmacao e usa a validacao/hash bcrypt existente.
- [x] Teste cobre troca de senha e login posterior com a nova senha.
- [x] Middleware ganhou `RateLimitByIPWhen` para aplicar limites por rota/metodo.
- [x] Servidor aplica limite especifico para `POST /login` e limite geral para requisicoes nao-GET.
- [x] Teste cobre rate limit condicional por rota.

Proxima etapa sugerida: edit locks com heartbeat ou edicao completa de banners/promocoes.

### Etapa 13 - Edit locks com heartbeat

- [x] Edicao de noticias cria/renova lock cooperativo por usuario.
- [x] Formulario de edicao envia heartbeat periodico para manter o lock ativo.
- [x] Painel alerta quando outro usuario possui lock ativo na mesma noticia.
- [x] Lock e liberado apos atualizacao bem-sucedida.
- [x] Corrigida comparacao de categoria no template de post com helper `idPtrEq`.
- [x] Testes cobrem aviso de lock ativo e criacao de lock via heartbeat.

Proxima etapa sugerida: edicao completa de banners/promocoes ou validacoes de formulario.

### Etapa 14 - Edicao de banners e promocoes

- [x] Painel de banners ganhou links, rota e formulario de edicao.
- [x] Atualizacao de banner preserva imagem existente quando nenhum novo upload e enviado.
- [x] Painel de promocoes ganhou links, rota e formulario de edicao.
- [x] Atualizacao de promocao preserva imagem existente quando nenhum novo upload e enviado.
- [x] Repository ganhou `PromotionGetByID`.
- [x] Parser de formularios aceita multipart e `application/x-www-form-urlencoded`, mantendo limite de corpo.
- [x] Testes cobrem atualizacao de banner e promocao.

Proxima etapa sugerida: validacoes de formulario mais completas ou fluxo de publicacao/agendamento.

### Etapa 15 - Agendamento editorial robusto

- [x] Criacao de post agendado passa a persistir `publish_at` antes do insert.
- [x] Atualizacao de post agendado passa a persistir `publish_at`.
- [x] Criacao/atualizacao de post agendado agenda job `publish_post`.
- [x] Worker publica post com `PostSetPublished`, preenchendo `published_at`.
- [x] Teste cobre criacao de post agendado com `publish_at` persistido e job ativo.

Proxima etapa sugerida: validacoes de formulario mais completas ou testes de regras editoriais.

### Etapa 16 - Validacoes e testes editoriais

- [x] Posts passam a validar titulo, conteudo e status antes de salvar.
- [x] Posts agendados exigem `publish_at` valido.
- [x] Banners e promocoes passam a rejeitar intervalos de datas invalidos.
- [x] Teste cobre obrigatoriedade de apuracao/responsavel em Politica & Bastidores.
- [x] Teste cobre obrigatoriedade de `publish_at` em post agendado.
- [x] Teste cobre rejeicao de periodo invalido em banner.

Proxima etapa sugerida: politicas de senha/admin inicial ou backup automatico como job.

### Etapa 11 - Relatorio de dead jobs

- [x] Repository ganhou `DeadJobCount` e `DeadJobList`.
- [x] Dashboard admin mostra contagem de jobs em `dead_jobs`.
- [x] Painel ganhou rota e tela `/dead-jobs` com tipo, tentativas, erro e data.
- [x] Menu admin inclui acesso ao relatorio de jobs com falha.
- [x] Testes cobrem listagem/contagem no repository e renderizacao da tela no handler.

Proxima etapa sugerida: metricas por periodo/anunciante ou edit locks com heartbeat.

### Etapa 8 - RSS, deploy docs e decisao React

- [x] Rota publica `/rss.xml` adicionada com os ultimos posts publicados.
- [x] RSS usa `SITE_URL`, datas RFC1123 e nao expoe posts em rascunho.
- [x] Teste automatizado cobre feed RSS e ausencia de rascunhos.
- [x] Exemplo Nginx reforcado com redirecionamento HTTPS, HSTS, headers finais, cache de static/uploads e bloqueio de dotfiles.
- [x] Exemplos systemd passam a usar `/etc/inhumas.env` e hardening basico.
- [x] `.env.example` alinhado ao runtime SQLite MVP para evitar deploy quebrado por URL Postgres ainda nao suportada.
- [x] Criados guias `DEPLOY_GUIDE.md`, `OPERATIONS_GUIDE.md`, `SECURITY_PREPROD_CHECKLIST.md` e `ENVIRONMENT.md`.
- [x] Decisao documentada em `ARCHITECTURE_DECISIONS.md`: React fica como prototipo visual no MVP; backend Go e o produto principal.

Proxima etapa sugerida: metricas/admin ou retencao/limpeza de originais antigos.

### Etapa 17 - Operacao e sessoes

- [x] Sessao passa a aceitar `SESSION_SECRET` atual e `PREVIOUS_SESSION_SECRET` para rotacao sem derrubar usuarios logados.
- [x] `PREVIOUS_SESSION_SECRET`, `METRICS_TOKEN`, `BACKUP_DIR` e `BACKUP_SCRIPT` documentados no `.env.example`.
- [x] Rota operacional `/metrics` adicionada com protecao por token ou sessao autenticada.
- [x] `/metrics` expoe uptime, requests, stats de conexoes do banco, jobs pendentes/dead e tamanho de uploads.
- [x] Health check passou a incluir jobs e status de backup quando `BACKUP_DIR` estiver configurado.
- [x] Worker agenda e executa job recorrente `backup_database` usando `BACKUP_SCRIPT`.
- [x] `scripts/backup.sh` agora faz backup de SQLite no MVP e continua suportando PostgreSQL via `pg_dump`.
- [x] Testes adicionados para rotacao de sessao e `/metrics` protegido.

Proxima etapa sugerida: migrar runtime para PostgreSQL ou implementar lock concorrente de jobs com `running`/claim atomico.

### Fechamento do MVP

- [x] MVP do portal Go server-side concluido para operacao inicial com SQLite.
- [x] Pendencias restantes classificadas como Pos-MVP, infraestrutura externa ou migracao PostgreSQL.
- [x] Comparacao com o prompt atualizada para remover itens ja entregues, como sitemap diario, backup job e `/metrics`.

Proxima etapa real: validar em ambiente de VPS com dominio/TLS, e so entao decidir se a migracao PostgreSQL deve acontecer antes ou depois da primeira operacao publica.

### Etapa 18 - Fechamento VPS, busca e worker

- [x] Deploy do MVP validado na VPS em `/var/www/inhumas`.
- [x] `inhumas-web` e `inhumas-worker` ativos via systemd.
- [x] Nginx validado com `nginx -t` e proxy para `127.0.0.1:8080`.
- [x] `/health` publico por IP retorna status OK com banco, uploads, jobs e backup.
- [x] `/login` publico por IP retorna 200; login completo deve ser validado apos HTTPS por causa dos cookies `Secure`.
- [x] Swap de 1GB ativado.
- [x] Backup local gerado em `/var/backups/inhumas`.
- [x] Restore/integridade do backup SQLite testado com `PRAGMA integrity_check`.
- [x] Cron diario de backup e cron de alerta de disco configurados.
- [x] `INITIAL_ADMIN_PASSWORD` removido de `/etc/inhumas.env` apos criacao do admin inicial.
- [x] `scripts/backup.sh` carrega `/etc/inhumas.env` sozinho quando necessario.
- [x] `scripts/disk-check.sh` nao falha quando `mail` nao esta disponivel; registra via `logger`/stdout.
- [x] Worker passou a reivindicar jobs vencidos com `JobClaimPending`, marcando status `running` antes da execucao.
- [x] Busca de posts passou a usar FTS5 no runtime SQLite, mantendo fallback `LIKE`.
- [x] Testes adicionados para claim de jobs e busca FTS5.
- [x] Validado com `go test ./...`.

Proxima etapa real: apontar Cloudflare/DNS para `68.233.125.102`, emitir HTTPS com Certbot, trocar `SITE_URL` para `https://inhumasemfoco.com.br`, validar cookies Secure/login/upload no dominio e configurar UptimeRobot/backup externo.

### Etapa 19 - Ajustes do prompt no MVP existente

- [x] Edicao de banners passou a bloquear sobreposicao de periodo/posicao quando o banner esta ativo.
- [x] Validacao de sobreposicao exclui o proprio banner editado e continua detectando outros banners ativos conflitantes.
- [x] Edicao de lojas deixou de depender de fallback manual por lista e usa `StoreGetByID`.
- [x] Testes adicionados para sobreposicao de banners no repository e no handler.
- [x] Validado com `go test ./...`.

Proxima etapa real: publicar esta revisao na VPS e, depois, seguir com dominio/TLS/Cloudflare/UptimeRobot ou decidir a migracao PostgreSQL.

### Etapa 20 - Painel admin e area de Influencers

- [x] Corrigido erro de CSRF em acesso temporario por IP/HTTP, removendo cookie `Secure` enquanto o dominio HTTPS nao esta ativo.
- [x] Criada nova conta admin operacional para acesso inicial.
- [x] Corrigidos erros de renderizacao no painel causados por dados nulos em templates.
- [x] Renderer passou a registrar erro real de template no journal.
- [x] Layout do painel admin revisado para ocupar a largura correta.
- [x] Formulario de posts ganhou editor maior, secoes, grid de metadados e acoes claras.
- [x] Telas de lojas, banners, promocoes, bairros, usuarios, metricas, dashboard e jobs foram padronizadas.
- [x] Removidos estilos inline antigos das principais telas admin.
- [x] CSS do painel foi versionado para evitar cache antigo no navegador.
- [x] Criada tabela SQLite `influencers` no runtime.
- [x] Migration PostgreSQL recebeu tabela `influencers`.
- [x] Repository ganhou CRUD de influencers.
- [x] Criadas rotas publicas `/influencers` e `/influencer/{slug}`.
- [x] Criadas telas publicas de listagem e perfil individual de influencer.
- [x] Menu publico recebeu link para Influencers.
- [x] Painel admin recebeu menu e CRUD de Influencers.
- [x] Cadastro de influencer inclui bio, area/nicho, Instagram, TikTok, YouTube, WhatsApp, avatar, capa, destaque, patrocinado e ativo.
- [x] Metricas receberam tipo `influencer_view`.
- [x] Influencers ativos entraram no sitemap.
- [x] Testes adicionados para paginas publicas de influencers.
- [x] Alteracoes publicadas na VPS e validadas com `/health`.
- [x] Validado com `go test ./...`.

Proxima etapa real: cadastrar influencers reais pelo painel, validar visual publico com conteudo real e seguir para dominio/TLS/Cloudflare/UptimeRobot/backup externo.

### Etapa 21 - Taxonomia editorial: categorias

- [x] `model.Category` recebeu `image_key`, `sort_order` e `active`.
- [x] Schema SQLite cria novas colunas para categorias e aplica `ALTER TABLE` seguro em bancos existentes.
- [x] Migration inicial PostgreSQL foi alinhada com os campos profissionais de categoria.
- [x] Repository ganhou `CategoryGetByID`, `CategoryCreate`, `CategoryUpdate`, `CategoryDelete`, `CategoryPostCount` e leitura ordenada por `sort_order`.
- [x] Painel recebeu menu, listagem, formulario de criacao e formulario de edicao de categorias.
- [x] Slug de categoria e automatico, editavel e validado contra conflitos.
- [x] Upload/remocao de imagem de categoria usa provider de storage existente.
- [x] Exclusao de categoria com noticias vinculadas e bloqueada para evitar quebra editorial.
- [x] Auditoria registra criar, editar e excluir categoria.
- [x] Testes adicionados em `internal/handler/handler_test.go`.
- [x] Validado com `go test ./...`.
- [x] Validado com build dos binarios `cmd/web`, `cmd/worker` e `cmd/seed-news`.
- [x] Backup pre-deploy criado na VPS em `/var/www/inhumas/backups/pre-category-crud-20260510_111919.db`.
- [x] Publicado na VPS com novos binarios, templates, migrations e documentos.
- [x] Smoke remoto validou `/health`, `/login`, `/painel/1c2dhviax7/categories` e `/painel/1c2dhviax7/categories/new`.
- [x] Banco remoto confirmou colunas `image_key`, `sort_order` e `active` em `categories`.

Proxima etapa real: implementar CRUD de tags e associacao muitos-para-muitos com noticias.

### Etapa 22 - Taxonomia editorial: tags

- [x] `model.Tag` criado com nome, slug, descricao, status e data de criacao.
- [x] Schema SQLite cria `tags` e `post_tags`.
- [x] Migration inicial PostgreSQL recebeu `tags` e `post_tags`.
- [x] Repository ganhou CRUD de tags, contagem de uso, associacao `PostSetTags`, leitura `PostTags` e listagem publica `PostListByTag`.
- [x] Painel recebeu menu, listagem, criacao e edicao de tags.
- [x] Formulario de noticia passou a selecionar multiplas tags ativas.
- [x] Criacao e edicao de noticias persistem tags em relacao muitos-para-muitos.
- [x] Detalhe publico da noticia mostra tags ativas com links.
- [x] Rota publica `/tag/{slug}` criada com SEO e JSON-LD de collection.
- [x] Tags ativas entraram no sitemap dinamico.
- [x] Auditoria registra criar, editar e excluir tag.
- [x] Exclusao de tag vinculada a noticias e bloqueada.
- [x] Testes cobrem CRUD de tags, persistencia em noticias, detalhe publico e pagina de tag.
- [x] Validado com `go test ./...`.
- [x] Validado com build dos binarios `cmd/web`, `cmd/worker` e `cmd/seed-news`.
- [x] Backup pre-deploy criado na VPS em `/var/www/inhumas/backups/pre-tags-crud-20260510_114625.db`.
- [x] Publicado na VPS com novos binarios, templates, migrations e documentos.
- [x] Smoke remoto validou `/health`, `/login`, `/painel/1c2dhviax7/tags`, `/painel/1c2dhviax7/tags/new` e `/tag/teste-deploy-tags`.
- [x] Smoke criou e removeu uma tag temporaria de deploy pelo painel.
- [x] Banco remoto confirmou tabelas `tags` e `post_tags`.
- [x] Adicionar filtro por tag na listagem administrativa de noticias.
- [x] Manter filtro por categoria e paginacao preservando `status`, `categoria`, `tag` e `q`.
- [x] Criar campos SEO dedicados para categorias: meta title e meta description.
- [x] Criar campos SEO dedicados para tags: meta title e meta description.
- [x] Paginas publicas de categoria e tag passam a usar metas especificas quando preenchidas.
- [x] Busca publica passa a encontrar noticias por nome de tag ativa.
- [x] Testes adicionados para filtro administrativo por tag e busca publica por tag.
- [x] Validado localmente com `go test ./...`.
- [x] Validado localmente com `go build ./cmd/web`, `go build ./cmd/worker` e `go build ./cmd/seed-news`.
- [x] Publicar refinamento de filtros/SEO de taxonomia na VPS e validar smoke.

Proxima etapa real: publicar refinamento de filtros/SEO de taxonomia na VPS e iniciar Sprint 4 de fluxo editorial premium.

### Etapa 23 - Taxonomia editorial: filtros e SEO

- [x] Categorias receberam `meta_title` e `meta_description`.
- [x] Tags receberam `meta_title` e `meta_description`.
- [x] Formulario de categorias permite editar metas SEO.
- [x] Formulario de tags permite editar metas SEO.
- [x] Pagina publica de categoria usa metas SEO dedicadas com fallback seguro.
- [x] Pagina publica de tag usa metas SEO dedicadas com fallback seguro.
- [x] Listagem administrativa de noticias ganhou filtro por tag.
- [x] Repository filtra noticias por tag usando `post_tags`.
- [x] Paginacao do painel preserva filtro de tag junto com status, categoria e busca.
- [x] Busca publica ganhou fallback por nome de tag ativa quando FTS nao retorna resultado.
- [x] Migration inicial PostgreSQL alinhada com metas SEO de categorias e tags.
- [x] Testes cobrem filtro administrativo por tag e busca por tag.
- [x] Validado com `go test ./...`.
- [x] Validado com build dos binarios `cmd/web`, `cmd/worker` e `cmd/seed-news`.
- [x] Backup pre-deploy criado na VPS em `/var/www/inhumas/backups/pre-taxonomy-seo-20260511_224922.db`.
- [x] Publicado na VPS com novos binarios, templates, migrations e documentos.
- [x] Smoke remoto validou `/health`, `/login`, filtro `tag_id` no painel de noticias e campos SEO em categorias/tags.
- [x] Smoke criou e removeu uma tag temporaria com meta title/meta description e validou `/tag/teste-seo-taxonomia`.
- [x] Banco remoto confirmou `meta_title` e `meta_description` em `categories` e `tags`.
- [x] Iniciar Sprint 4 de fluxo editorial premium.
- [x] Criar status editorial `approved` para separar aprovacao de publicacao.
- [x] Criar transicao "enviar para revisao" para rascunhos.
- [x] Criar transicao "aprovar" para noticias em revisao.
- [x] Criar transicao "reprovar com comentario", retornando a noticia para rascunho.
- [x] Registrar transicoes editoriais na auditoria.
- [x] Exibir botoes de fluxo editorial na listagem administrativa de noticias.
- [x] Validar permissoes de fluxo em `internal/posts.EditorialService`.
- [x] Testes adicionados para permissoes de workflow, envio para revisao, aprovacao e reprovacao com comentario.
- [x] Validado localmente com `go test ./...`.
- [x] Validado localmente com `go build ./cmd/web`, `go build ./cmd/worker` e `go build ./cmd/seed-news`.
- [x] Publicar fluxo editorial inicial na VPS e validar smoke autenticado.

Proxima etapa real: iniciar Sprint 4 de fluxo editorial premium.

### Etapa 24 - Fluxo editorial premium inicial

- [x] `model.StatusApproved` criado para representar noticias aprovadas, ainda nao publicadas.
- [x] `EditorialService` ganhou regras `CanSubmitReview`, `CanApprove` e `CanReject`.
- [x] `ValidateStatusPermission` passou a exigir permissao de aprovacao para status aprovado.
- [x] Rotas administrativas criadas: `submit-review`, `approve` e `reject`.
- [x] Rascunhos podem ser enviados para revisao por quem pode editar o conteudo.
- [x] Revisores/Admins podem aprovar noticias em revisao.
- [x] Revisores/Admins podem reprovar noticias em revisao com comentario obrigatorio.
- [x] Comentario de reprovacao e anexado nas notas editoriais da noticia.
- [x] Acoes de workflow gravam auditoria com origem/destino do status.
- [x] Painel de noticias recebeu filtro e badge para status aprovado.
- [x] Listagem de noticias recebeu botoes de enviar para revisao, aprovar e reprovar.
- [x] CSS recebeu estilo para `status-approved`.
- [x] Testes cobrem regras do service e handlers do workflow.
- [x] Validado com `go test ./...`.
- [x] Validado com build dos binarios `cmd/web`, `cmd/worker` e `cmd/seed-news`.
- [x] Backup pre-deploy criado na VPS em `/var/www/inhumas/backups/pre-editorial-workflow-20260512_002606.db`.
- [x] Publicado na VPS com novos binarios, templates, CSS e documentos.
- [x] Smoke remoto validou criacao de noticias temporarias, envio para revisao, aprovacao, reprovacao com comentario e exclusao dos temporarios.
- [x] Smoke remoto confirmou `/health` OK e servicos `inhumas-web`/`inhumas-worker` ativos.

Proxima etapa real: finalizar Sprint 4 com historico de revisoes, autosave, galeria, fonte original e checklist SEO.

### Etapa 25 - Sprint 4 final: editor premium e rastreabilidade

- [x] `model.Post` recebeu galeria, canonical URL, fonte original, tempo de leitura e flag `is_pinned`.
- [x] `model.PostRevision` criado para registrar snapshots do conteudo por acao editorial.
- [x] Schema SQLite e migration PostgreSQL foram alinhados com `gallery_image_keys`, `canonical_url`, `source_name`, `source_url`, `reading_time_minutes`, `is_pinned` e `post_revisions`.
- [x] Repository passou a persistir e carregar os novos campos de noticia.
- [x] Repository ganhou `PostRevisionCreate` e `PostRevisionList`.
- [x] Formulario de noticia recebeu canonical URL, fonte original, galeria, fixar no topo, leitura estimada e historico de revisoes.
- [x] Editor ganhou autosave a cada 90 segundos para noticias ja salvas.
- [x] Formulario evita preview/autosave em posts ainda nao persistidos apos erro de validacao.
- [x] Criacao, edicao, autosave, envio para revisao, aprovacao, reprovacao e publicacao registram revisoes.
- [x] Publicacao, agendamento e aprovacao validam checklist SEO minimo antes de seguir.
- [x] Validacao de agendamento prioriza mensagem de data obrigatoria antes do checklist SEO.
- [x] Detalhe publico da noticia exibe tempo de leitura, galeria e fonte original.
- [x] JSON-LD `NewsArticle` recebeu `timeRequired` e `citation` quando aplicavel.
- [x] Destaques da home consideram noticias fixadas no topo antes de destaques comuns.
- [x] Testes atualizados para workflow editorial, SEO obrigatorio, tags e revisoes.
- [x] Validado localmente com `go test ./...`.
- [x] Validado localmente com `go build ./cmd/web`, `go build ./cmd/worker` e `go build ./cmd/seed-news`.
- [x] Backup pre-deploy criado na VPS em `/var/www/inhumas/backups/pre-sprint4-final-20260512_005846.db`.
- [x] Publicado na VPS com novos binarios, templates, CSS, migrations e documentos.
- [x] Servicos `inhumas-web` e `inhumas-worker` confirmados ativos apos deploy.
- [x] Smoke remoto validou `/health`, login administrativo, criacao de noticia temporaria publicada, detalhe publico com tempo de leitura/fonte, historico/checklist no editor, autosave e exclusao do temporario.
- [x] Banco remoto confirmou colunas novas de `posts` e tabela `post_revisions`.

Proxima etapa real: iniciar Sprint 5 - Biblioteca de midia reutilizavel.

### Etapa 26 - Sprint 5 inicial: biblioteca de midia

- [x] `model.MediaAsset` criado para representar arquivo, chave, nome original, titulo, alt text, tipo, tamanho, usuario e uso.
- [x] Schema SQLite recebeu tabela `media_assets` com indices de criacao e busca.
- [x] Migration PostgreSQL inicial recebeu tabela `media_assets`.
- [x] Permissao `media:manage` adicionada ao RBAC para perfis editoriais/comerciais adequados.
- [x] Repository ganhou CRUD de midia, busca, paginacao e contagem de uso por chave.
- [x] Contagem de uso verifica posts, galerias, categorias, lojas, promocoes, banners, bairros e influencers.
- [x] Painel administrativo recebeu menu `Midia`.
- [x] Criada rota/listagem `/media` com busca e paginacao.
- [x] Upload pela biblioteca valida imagem pelo storage existente, converte para WebP e exige alt text.
- [x] Cards de midia exibem preview, nome original, tipo, tamanho, URL e quantidade de usos.
- [x] Metadados de titulo e texto alternativo podem ser editados.
- [x] Exclusao segura bloqueia midia em uso e remove arquivo do storage apenas quando esta livre.
- [x] Auditoria registra criar, editar e excluir midia.
- [x] Uploads novos de capa/galeria de noticias passam a registrar ativos na biblioteca automaticamente.
- [x] Teste cobre listagem, edicao de metadados, contagem de uso e bloqueio de exclusao de midia usada.
- [x] Validado localmente com `go test ./...`.
- [x] Validado localmente com `go build ./cmd/web`, `go build ./cmd/worker` e `go build ./cmd/seed-news`.
- [x] Backup pre-deploy criado na VPS em `/var/www/inhumas/backups/pre-sprint5-media-20260512_161539.db`.
- [x] Publicado na VPS com novos binarios, templates, CSS, migrations e documentos.
- [x] Servicos `inhumas-web` e `inhumas-worker` confirmados ativos apos deploy.
- [x] Smoke remoto validou `/health`, login administrativo, pagina `/media`, upload de imagem temporaria, busca e exclusao da midia temporaria.
- [x] Banco remoto confirmou tabela `media_assets` e limpeza do item temporario de smoke.

Proxima etapa real: continuar Sprint 5 com seletor/reaproveitamento de midia existente nos formularios de noticias, lojas, banners, promocoes e influencers.

### Etapa 27 - Sprint 5: reaproveitamento de midia nos formularios

- [x] Helper comum criado para carregar midias recentes nos formularios administrativos.
- [x] Helper comum valida chaves selecionadas contra a tabela `media_assets` antes de reutilizar.
- [x] Formulario de noticias permite escolher capa existente da biblioteca.
- [x] Formulario de noticias permite adicionar imagens existentes na galeria.
- [x] Formulario de categorias permite escolher imagem existente.
- [x] Formulario de lojas permite escolher logo e capa existentes.
- [x] Formulario de influencers permite escolher avatar e capa existentes.
- [x] Formulario de banners permite escolher arte existente da biblioteca.
- [x] Formulario de promocoes permite escolher imagem existente da biblioteca.
- [x] Uploads novos em categorias, lojas, influencers, banners e promocoes passam a registrar `media_assets` automaticamente.
- [x] Atualizacao por upload novo deixa a exclusao fisica centralizada na biblioteca, reduzindo risco de apagar arquivo ainda reutilizado.
- [x] Testes adicionados para noticia reutilizando capa/galeria e banner ativo reutilizando midia existente.
- [x] Validado localmente com `go test ./...`.
- [x] Validado localmente com `go build ./cmd/web`, `go build ./cmd/worker` e `go build ./cmd/seed-news`.
- [x] Backup pre-deploy criado na VPS em `/var/www/inhumas/backups/pre-sprint5-media-reuse-20260512_164833.db`.
- [x] Publicado na VPS com novos binarios, templates, CSS e documentos.
- [x] Smoke remoto validou `/health`, login administrativo, upload de midia temporaria, seletores de midia na noticia, criacao de noticia usando capa/galeria da biblioteca e limpeza dos temporarios.
- [x] Banco remoto confirmou limpeza dos registros temporarios de smoke.
- [x] Servicos `inhumas-web` e `inhumas-worker` confirmados ativos apos deploy.

Proxima etapa real: concluir Sprint 5 com organizacao por ano/mes e busca por data, ou iniciar Sprint 6 de auditoria/configuracoes se a biblioteca atual ja atender a operacao inicial.

### Etapa 28 - Sprint 5 final: organizacao temporal da biblioteca

- [x] `MediaAssetFilter` criado no repository para busca combinada por texto e periodo.
- [x] Biblioteca de midia ganhou filtro por mes/ano.
- [x] Biblioteca de midia ganhou filtro por intervalo de datas (`date_from` e `date_to`).
- [x] Repository ganhou `MediaAssetListFiltered` e `MediaAssetCountFiltered`.
- [x] Repository ganhou `MediaAssetArchiveMonths` para agrupar ativos por ano/mes.
- [x] Cards de midia exibem a data de organizacao/cadastro.
- [x] Paginacao preserva filtros de busca, mes e datas.
- [x] Teste cobre filtro por mes/ano e intervalo de datas.
- [x] Sprint 5 marcado como concluido no plano geral no escopo atual.
- [x] Validado localmente com `go test ./...`.
- [x] Validado localmente com `go build ./cmd/web`, `go build ./cmd/worker` e `go build ./cmd/seed-news`.
- [x] Backup pre-deploy criado na VPS em `/var/www/inhumas/backups/pre-sprint5-final-20260512_171116.db`.
- [x] Publicado na VPS com novos binarios, templates, CSS e documentos.
- [x] Smoke remoto validou `/health`, login administrativo, filtros `month`, `date_from`, `date_to`, upload de midia temporaria, busca por mes/data e exclusao do temporario.
- [x] Banco remoto confirmou limpeza da midia temporaria de smoke.
- [x] Servicos `inhumas-web` e `inhumas-worker` confirmados ativos apos deploy.

Proxima etapa real: iniciar Sprint 6 - Auditoria e configuracoes do portal.

### Etapa 29 - Sprint 6: auditoria e configuracoes globais

- [x] Auditoria administrativa consolidada com tela, filtros por usuario, acao, entidade, entidade ID e periodo.
- [x] Filtro de auditoria passou a incluir entidade `settings`.
- [x] `PortalSettings` criado para identidade, contatos, redes sociais, SEO global, limites de upload e automacao.
- [x] Schema SQLite recebeu tabela `portal_settings`.
- [x] Migration PostgreSQL inicial recebeu tabela `portal_settings`.
- [x] Repository ganhou leitura com defaults seguros e upsert das configuracoes globais.
- [x] Painel administrativo recebeu menu `Configuracoes`.
- [x] Tela `/settings` permite editar nome do portal, slogan, logo, favicon, contato, redes sociais, SEO global, limite de upload e parametros de automacao.
- [x] Logo e favicon podem ser enviados ou reutilizados a partir da biblioteca de midia.
- [x] Biblioteca de midia passa a considerar logo e favicon como uso protegido.
- [x] Header, footer, pagina de contato, RSS, Open Graph e JSON-LD passam a usar configuracoes globais.
- [x] Atualizacao das configuracoes registra auditoria com IP, usuario e principais campos alterados.
- [x] Testes adicionados para renderizacao, permissao, persistencia, auditoria e uso publico das configuracoes.
- [x] Validado localmente com `go test ./...`.
- [x] Validado localmente com `go build ./cmd/web ./cmd/worker ./cmd/seed-news`.
- [x] Backup pre-deploy criado na VPS em `/var/www/inhumas/backups/pre-sprint6-settings-20260513_191104.db`.
- [x] Publicado na VPS com novos binarios, templates, migrations e documentos.
- [x] Servicos `inhumas-web` e `inhumas-worker` confirmados ativos apos deploy.
- [x] Smoke remoto validou `/health`, `/contato`, login administrativo e `/painel/1c2dhviax7/settings`.
- [x] Banco remoto confirmou tabela `portal_settings`.

Proxima etapa real: iniciar Sprint 7 - eventos, classificados e modulos locais.

### Etapa 30 - Sprint 7 inicial: eventos locais

- [x] `model.Event` criado com titulo, slug, descricao, local, organizador, ingresso, valor, imagem, status, destaque, patrocinado e SEO.
- [x] RBAC recebeu permissao `events:manage` para administradores, editores e comercial.
- [x] Schema SQLite recebeu tabela `events` com indices por status/data e destaque.
- [x] Migration PostgreSQL inicial recebeu tabela `events`.
- [x] Repository ganhou CRUD de eventos, busca por slug/ID, listagem ativa e validacao de slug.
- [x] Sitemap dinamico passou a incluir eventos ativos em `/evento/{slug}`.
- [x] Biblioteca de midia passa a considerar eventos como uso protegido de imagem.
- [x] Painel administrativo recebeu menu `Eventos`.
- [x] Criadas rotas administrativas `/events`, `/events/new`, edicao e exclusao.
- [x] Formulario de evento permite upload ou reutilizacao de imagem da biblioteca.
- [x] Eventos registram auditoria em criacao, edicao e exclusao.
- [x] Rota publica `/eventos` agora lista eventos cadastrados como produto proprio.
- [x] Rota publica `/evento/{slug}` criada com detalhe, SEO e Schema.org `Event`.
- [x] Testes adicionados para criacao administrativa, permissao, auditoria, listagem publica e detalhe publico.
- [x] Validado localmente com `go test ./...`.
- [x] Validado localmente com `go build ./cmd/web ./cmd/worker ./cmd/seed-news`.
- [x] Backup pre-deploy criado na VPS em `/var/www/inhumas/backups/pre-sprint7-events-20260513_192242.db`.
- [x] Publicado na VPS com novos binarios, templates, migrations e documentos.
- [x] Servicos `inhumas-web` e `inhumas-worker` confirmados ativos apos deploy.
- [x] Smoke remoto validou `/health`, login administrativo, `/events/new`, criacao de evento temporario, `/eventos`, `/evento/{slug}`, exclusao pelo painel e 404 apos exclusao.
- [x] Banco remoto confirmou tabela `events`, auditoria de evento e limpeza do temporario.

Proxima etapa real: continuar Sprint 7 com classificados.

### Etapa 31 - Sprint 7: classificados locais

- [x] `model.Classified` criado com titulo, slug, descricao, categoria, preco, contato, localizacao, imagem, status, destaque, patrocinado, validade e SEO.
- [x] RBAC recebeu permissao `classifieds:manage` para administradores e comercial.
- [x] Schema SQLite recebeu tabela `classifieds` com indices por status/categoria e destaque.
- [x] Migration PostgreSQL inicial recebeu tabela `classifieds`.
- [x] Repository ganhou CRUD de classificados, busca por slug/ID, filtros por texto/categoria e validacao de slug.
- [x] Sitemap dinamico passou a incluir classificados ativos e nao expirados em `/classificado/{slug}`.
- [x] Biblioteca de midia passa a considerar classificados como uso protegido de imagem.
- [x] Painel administrativo recebeu menu `Classificados`.
- [x] Criadas rotas administrativas `/classifieds`, `/classifieds/new`, edicao e exclusao.
- [x] Formulario de classificado permite upload ou reutilizacao de imagem da biblioteca.
- [x] Classificados registram auditoria em criacao, edicao e exclusao.
- [x] Rota publica `/classificados` agora lista anuncios cadastrados com busca e filtro por categoria.
- [x] Rota publica `/classificado/{slug}` criada com detalhe, SEO e Schema.org `Product`.
- [x] Testes adicionados para criacao administrativa, permissao, auditoria, filtro publico e detalhe publico.
- [x] Validado localmente com `go test ./...`.
- [x] Validado localmente com `go build ./cmd/web ./cmd/worker ./cmd/seed-news`.
- [x] Backup pre-deploy criado na VPS em `/var/www/inhumas/backups/pre-sprint7-classifieds-20260513_193256.db`.
- [x] Publicado na VPS com novos binarios, templates, migrations e documentos.
- [x] Servicos `inhumas-web` e `inhumas-worker` confirmados ativos apos deploy.
- [x] Smoke remoto validou `/health`, login administrativo, `/classifieds/new`, criacao de classificado temporario, `/classificados?categoria=Imoveis`, `/classificado/{slug}`, exclusao pelo painel e 404 apos exclusao.
- [x] Banco remoto confirmou tabela `classifieds`, auditoria de classificado e limpeza do temporario.

Proxima etapa real: continuar Sprint 7 refinando lojas, promocoes e influencers.

### Etapa 32 - Sprint 7: refinamento de lojas

- [x] `model.Store` ganhou `website_url`, `commercial_status`, `meta_title` e `meta_description`.
- [x] Schema SQLite recebeu colunas de SEO, site e status comercial em `stores`.
- [x] Migration PostgreSQL inicial recebeu colunas equivalentes em `stores`.
- [x] Migração automática SQLite garante as novas colunas em bancos existentes.
- [x] Repository de lojas passou a persistir e carregar SEO, site e status comercial.
- [x] Painel de lojas exibe status comercial por cliente.
- [x] Formulario de loja permite editar site, status comercial, meta title e meta description.
- [x] Detalhe publico da loja usa SEO proprio quando preenchido.
- [x] Detalhe publico exibe link de site quando cadastrado.
- [x] Auditoria de loja registra status comercial nas criacoes e edicoes.
- [x] Teste adicionado para criacao administrativa com SEO/status comercial e renderizacao publica.
- [x] Validado localmente com `go test ./...`.
- [x] Validado localmente com `go build ./cmd/web ./cmd/worker ./cmd/seed-news`.
- [x] Backup pre-deploy criado na VPS em `/var/www/inhumas/backups/pre-sprint7-stores-20260513_194044.db`.
- [x] Publicado na VPS com novos binarios, templates, migrations e documentos.
- [x] Servicos `inhumas-web` e `inhumas-worker` confirmados ativos apos deploy.
- [x] Smoke remoto validou `/health`, login administrativo, `/stores/new`, criacao de loja temporaria com SEO/status/site, `/loja/{slug}`, exclusao pelo painel e 404 apos exclusao.
- [x] Banco remoto confirmou colunas `website_url`, `commercial_status`, `meta_title` e `meta_description` em `stores`, auditoria de loja e limpeza do temporario.

Proxima etapa real: seguir para promocoes.

### Etapa 33 - Sprint 7: refinamento de promocoes

- [x] `model.Promotion` ganhou `coupon_code`, `meta_title`, `meta_description` e contador transiente de cliques.
- [x] Schema SQLite recebeu colunas de cupom e SEO em `promotions`.
- [x] Migration PostgreSQL inicial recebeu colunas equivalentes em `promotions`.
- [x] Migração automática SQLite garante as novas colunas em bancos existentes.
- [x] Repository de promocoes passou a persistir e carregar cupom, SEO e contagem de cliques/resgates.
- [x] Admin de promocoes passou a listar todas as promocoes, nao apenas as ativas.
- [x] Admin de promocoes ganhou resumo de cadastradas, no ar/agendadas, expiradas e cliques/resgates.
- [x] Formulario de promocao permite editar cupom, status e SEO tambem em novos cadastros.
- [x] Página publica da promocao usa SEO proprio quando preenchido.
- [x] Página publica exibe cupom quando cadastrado.
- [x] Detalhe publico de promocao respeita status ativo e periodo vigente.
- [x] Testes adicionados/atualizados para cupom, SEO, relatorio de cliques e detalhe publico.
- [x] Validado localmente com `go test ./...`.
- [x] Validado localmente com `go build ./cmd/web ./cmd/worker ./cmd/seed-news`.
- [x] Backup pre-deploy criado na VPS em `/var/www/inhumas/backups/pre-sprint7-promotions-20260513_195130.db`.
- [x] Refinamento de promocoes publicado na VPS com binarios, templates, CSS, migracoes e docs.
- [x] Servicos `inhumas-web` e `inhumas-worker` reiniciados e ativos apos deploy.
- [x] Smoke remoto validou `/health`, login administrativo, criacao de loja temporaria, criacao de promocao com cupom/SEO, listagem `/promocoes`, detalhe `/promocao/{slug}`, relatorio admin de promocoes e exclusao segura.
- [x] Banco remoto confirmou colunas `coupon_code`, `meta_title` e `meta_description` em `promotions`.
- [x] Smoke remoto confirmou limpeza dos registros temporarios e retorno `404` no detalhe da promocao excluida.

Proxima etapa real: seguir para influencers com nichos/categorias, destaque, SEO e relatorio basico.

### Etapa 34 - Sprint 7 final: refinamento de influencers

- [x] `model.Influencer` ganhou `niche`, `meta_title`, `meta_description` e contador transiente de visualizacoes.
- [x] Schema SQLite recebeu colunas de nicho e SEO em `influencers`.
- [x] Migration PostgreSQL inicial recebeu colunas equivalentes em `influencers`.
- [x] Migração automática SQLite garante as novas colunas em bancos existentes.
- [x] Repository de influencers passou a persistir e carregar nicho, SEO e contagem de visualizacoes.
- [x] Admin de influencers ganhou resumo de cadastrados, ativos, destaques e visualizacoes.
- [x] Lista admin mostra nicho, views e atalho para perfil publico ativo.
- [x] Formulario de influencer permite editar nicho/categoria e SEO.
- [x] Lista publica de influencers ganhou filtro por nicho.
- [x] Detalhe publico de influencer usa SEO proprio quando preenchido.
- [x] Testes adicionados/atualizados para nicho, SEO, filtro publico, relatorio de views e detalhe publico.
- [x] Validado localmente com `go test ./...`.
- [x] Validado localmente com `go build ./cmd/web ./cmd/worker ./cmd/seed-news`.
- [x] Backup pre-deploy criado na VPS em `/var/www/inhumas/backups/pre-sprint7-influencers-20260513_202342.db`.
- [x] Refinamento de influencers publicado na VPS com binarios, templates, CSS, migracoes e docs.
- [x] Servicos `inhumas-web` e `inhumas-worker` reiniciados e ativos apos deploy.
- [x] Smoke remoto validou `/health`, login administrativo, criacao de influencer temporario, filtro publico por nicho, detalhe `/influencer/{slug}` com SEO, relatorio admin de visualizacoes e exclusao segura.
- [x] Banco remoto confirmou colunas `niche`, `meta_title` e `meta_description` em `influencers`.
- [x] Smoke remoto confirmou limpeza dos registros temporarios e retorno `404` no detalhe do influencer excluido.

Sprint 7 encerrado: eventos, classificados, lojas, promocoes e influencers estao com painel, pagina publica, SEO, status e validacao em VPS.

Proxima etapa real: iniciar Sprint 8 - anuncios avancados e relatorios comerciais.

### Etapa 35 - Sprint 8 inicial: relatorios comerciais de banners

- [x] `model.Banner` ganhou contadores transientes de impressoes e cliques.
- [x] Repository de banners passou a carregar impressoes e cliques a partir da tabela `metrics`.
- [x] Painel de banners ganhou bloco de relatorio comercial com anunciantes, impressoes, cliques e CTR medio.
- [x] Listagem de campanhas passou a exibir impressoes, cliques e CTR por banner.
- [x] Painel de banners ganhou filtro por status comercial: ativo, pausado, rascunho e expirado.
- [x] Criada rota administrativa `GET /banners/export.csv` para exportacao CSV do relatorio filtrado.
- [x] Testes adicionados para relatorio comercial, CTR e exportacao CSV.
- [x] Validado localmente com `go test ./...`.
- [x] Validado localmente com `go build ./cmd/web ./cmd/worker ./cmd/seed-news`.
- [x] Backup pre-deploy criado na VPS em `/var/www/inhumas/backups/pre-sprint8-banner-reports-20260514_125359.db`.
- [x] Relatorios comerciais de banners publicados na VPS com binarios, templates, CSS, migracoes e docs.
- [x] Servicos `inhumas-web` e `inhumas-worker` reiniciados e ativos apos deploy.
- [x] Smoke remoto validou `/health`, login administrativo, criacao de banner temporario em rascunho, metricas de impressao/clique, filtro por status, relatorio com CTR, exportacao CSV e exclusao segura.
- [x] Smoke remoto confirmou limpeza dos banners e metricas temporarios.

Proxima etapa real: seguir nos produtos comerciais de destaque para classificados, eventos e lojas.

### Etapa 36 - Sprint 8 final: produtos comerciais locais

- [x] Painel de lojas ganhou inventario comercial com total, ativas, destaques vendidos e patrocinadas.
- [x] Painel de eventos ganhou inventario comercial com total, ativos, destaques vendidos e patrocinados.
- [x] Painel de classificados ganhou inventario comercial com total, ativos, destaques vendidos e patrocinados.
- [x] Pagina publica de lojas ganhou bloco `Destaques comerciais` para lojas destacadas ou patrocinadas.
- [x] Pagina publica de eventos ganhou bloco `Destaques comerciais` para eventos destacados ou patrocinados.
- [x] Pagina publica de classificados ganhou bloco `Destaques comerciais` para classificados destacados ou patrocinados.
- [x] Listagens publicas separam destaques comerciais da listagem organica para deixar o produto vendavel e claro.
- [x] Banners in-feed foram deslocados para depois do conteudo/filtros nas paginas de lojas, eventos e classificados, evitando anuncios consecutivos no topo/feed.
- [x] Testes atualizados para validar blocos comerciais publicos e resumos comerciais administrativos.
- [x] Validado localmente com `go test ./...`.
- [x] Validado localmente com `go build ./cmd/web ./cmd/worker ./cmd/seed-news`.
- [x] Backup pre-deploy criado na VPS em `/var/www/inhumas/backups/pre-sprint8-commercial-products-20260514_131327.db`.
- [x] Produtos comerciais locais publicados na VPS com binarios, templates, CSS, migracoes e docs.
- [x] Servicos `inhumas-web` e `inhumas-worker` reiniciados e ativos apos deploy.
- [x] Smoke remoto validou `/health`, login administrativo, blocos comerciais publicos em lojas/eventos/classificados, resumos comerciais nos paineis e regra visual com banner deslocado para depois do conteudo.
- [x] Smoke remoto confirmou limpeza dos registros temporarios de lojas, eventos e classificados.

Sprint 8 encerrado: banners tem relatorio comercial/CSV e os modulos locais possuem produtos de destaque/patrocinio vendaveis no painel e no portal publico.

Proxima etapa real: iniciar Sprint 9 - automacao de noticias com fontes RSS/oficiais, fila de revisao e logs de execucao.

### Etapa 37 - Sprint 9: automacao de noticias

- [x] `model.AutomationSource` criado para fontes RSS e fontes oficiais com categoria padrao, status e ultima coleta.
- [x] `model.AutomationRun` criado para registrar execucoes, status, itens encontrados, rascunhos, duplicados, erro e log.
- [x] `JobCollectNews` criado para coleta recorrente pelo worker.
- [x] Schema SQLite recebeu tabelas `automation_sources` e `automation_runs`.
- [x] Migration PostgreSQL inicial recebeu tabelas equivalentes e o tipo de job `collect_news`.
- [x] Repository ganhou CRUD de fontes, historico de execucoes, fila de rascunhos importados e verificacao de duplicidade exata.
- [x] Servico `internal/automation` criado com parser RSS/Atom, HTTP client com limite de leitura e `User-Agent` proprio.
- [x] Deduplicacao por URL implementada antes da criacao de rascunhos.
- [x] Deduplicacao por titulo identico implementada.
- [x] Similaridade basica por tokens implementada para reduzir repeticao editorial.
- [x] Coleta cria posts sempre como `draft`, preservando `source_name`, `source_url`, resumo, meta description e nota editorial de revisao obrigatoria.
- [x] Worker agenda coleta recorrente quando `PortalSettings.AutomationEnabled` esta ativo e respeita `AutomationIntervalMinutes`.
- [x] Execucao automatica nunca publica conteudo; publicacao continua dependendo do fluxo editorial humano.
- [x] RBAC recebeu permissao `automation:manage` para administradores e editores.
- [x] Middleware administrativo permite editores acessarem a area de automacao sem liberar areas comerciais/usuarios/configuracoes.
- [x] Painel recebeu menu `Automacao`.
- [x] Tela `/automation` lista fontes, fila de revisao, historico de execucoes e logs visiveis.
- [x] Tela `/automation/sources/new` permite cadastrar fonte RSS ou oficial.
- [x] Edicao, exclusao, execucao por fonte e botao `Executar agora` para todas as fontes ativas foram adicionados.
- [x] Auditoria registra criacao, edicao, exclusao e execucoes manuais de automacao.
- [x] Testes adicionados para criar rascunhos automatizados, preservar fonte original e bloquear duplicados.
- [x] Validado localmente com `go test ./...`.
- [x] Validado localmente com `go build ./cmd/web ./cmd/worker ./cmd/seed-news`.
- [x] Backup pre-deploy criado na VPS em `/var/www/inhumas/backups/pre-sprint9-automation-20260514_163229-db.tar.gz`.
- [x] Sprint 9 publicado na VPS com binarios, templates, migrations e documentos.
- [x] Servicos `inhumas-web` e `inhumas-worker` reiniciados e ativos apos deploy.
- [x] Smoke remoto validou `/health`, login administrativo e `/painel/1c2dhviax7/automation`.
- [x] Banco remoto confirmou tabelas `automation_sources` e `automation_runs`.

Sprint 9 encerrado: o CMS agora tem automacao de noticias com fontes RSS/oficiais, execucoes auditaveis, logs, deduplicacao, rascunhos rastreaveis e revisao humana obrigatoria.

Proxima etapa real: iniciar Sprint 10 - IA editorial desacoplada.

### Etapa 38 - Sprint 10: IA editorial desacoplada

- [x] Pacote `internal/editorialai` criado com interface `Provider`/`EditorialAIProvider` desacoplada do fornecedor.
- [x] Provider mock `mock-editorial` implementado para desenvolvimento, testes e operacao sem custo externo.
- [x] Acoes editoriais implementadas: melhorar titulo, criar subtitulo, gerar resumo, gerar meta description, sugerir tags, reescrever em tom jornalistico, criar chamada social e verificar duplicidade.
- [x] Guardrail centralizado informa que a sugestao usa apenas titulo, resumo, conteudo e fonte cadastrados.
- [x] Guardrail alerta quando a fonte original esta ausente.
- [x] Provider preserva `source_name` e `source_url` em todas as sugestoes.
- [x] Acoes de IA nao alteram status, nao publicam, nao aprovam e nao substituem o fluxo editorial humano.
- [x] Rascunhos automatizados continuam como `draft`; a IA atua apenas como assistente de sugestao.
- [x] Schema SQLite recebeu tabela `ai_usage_logs`.
- [x] Migration PostgreSQL inicial recebeu tabela `ai_usage_logs`.
- [x] Repository ganhou criacao de log de uso de IA e listagem por noticia.
- [x] Handler administrativo recebeu rota `POST /posts/{id}/ai/{action}`.
- [x] Formulario de noticia recebeu bloco `IA editorial` com botoes de sugestao para cada acao.
- [x] Sugestoes aparecem no editor com notas de guardrail e sem gravar automaticamente nos campos.
- [x] Historico recente de uso de IA aparece no formulario da noticia.
- [x] Auditoria administrativa registra cada acao de IA por noticia.
- [x] Testes adicionados para provider mock, guardrails, fonte preservada, duplicidade, log de uso e garantia de que status continua rascunho.
- [x] Validado localmente com `go test ./...`.
- [x] Validado localmente com `go build ./cmd/web ./cmd/worker ./cmd/seed-news`.
- [x] Backup pre-deploy criado na VPS em `/var/www/inhumas/backups/pre-sprint10-ai-20260514_165127-db.tar.gz`.
- [x] Sprint 10 publicado na VPS com binarios, templates, CSS, migrations e documentos.
- [x] Servicos `inhumas-web` e `inhumas-worker` reiniciados e ativos apos deploy.
- [x] Smoke remoto validou `/health`, login administrativo, criacao de noticia temporaria em rascunho, bloco `IA editorial`, acao `meta_description`, guardrail visivel e limpeza do post temporario.
- [x] Banco remoto confirmou tabela `ai_usage_logs`.

Sprint 10 encerrado: a IA editorial esta pronta para trocar de fornecedor no futuro sem reescrever o CMS.

Proxima etapa real: iniciar Sprint 11 - PostgreSQL e migrations versionadas.

### Etapa 39 - Sprint 11: PostgreSQL e migrations versionadas

- [x] Estrategia definida: SQLite continua como runtime padrao/seguro e PostgreSQL entra por `DB_DRIVER=postgres` ou autodeteccao de URL `postgres://`.
- [x] Driver PostgreSQL `github.com/lib/pq` adicionado ao projeto.
- [x] Configuracao recebeu `DB_DRIVER` e `MIGRATIONS_DIR`.
- [x] `cmd/web`, `cmd/worker` e `cmd/seed-news` passaram a abrir banco por driver configuravel.
- [x] Repository recebeu dialeto interno (`sqlite`/`postgres`) sem quebrar a API atual.
- [x] Runner de migrations versionadas criado para PostgreSQL com tabela `schema_migrations`.
- [x] SQLite passa a registrar `schema_migrations` com versao `sqlite_auto`, mantendo compatibilidade com o migrador automatico atual.
- [x] Migration PostgreSQL inicial revisada para refletir colunas comerciais atuais de banners e `dead_jobs` com sequence propria.
- [x] Inserts do repository passaram por helper compativel: `LastInsertId` no SQLite e `RETURNING id` no PostgreSQL.
- [x] Busca de posts ganhou caminho PostgreSQL com `search_vector`, `plainto_tsquery` e ranking por `ts_rank_cd`.
- [x] Worker ganhou claim PostgreSQL concorrente com `FOR UPDATE SKIP LOCKED`.
- [x] Comando `cmd/migrate-sqlite-postgres` criado para copiar dados do SQLite para PostgreSQL em ordem segura.
- [x] Migrador ajusta booleans, preserva IDs, roda migrations no destino, ajusta sequences e valida contagem por tabela.
- [x] Documentacao de ambiente atualizada com `DB_DRIVER`, `MIGRATIONS_DIR`, exemplo PostgreSQL e comando de migracao.
- [x] Guia de operacao atualizado com rotina de homologacao da migracao SQLite -> PostgreSQL.
- [x] Testes adicionados para abertura SQLite com `schema_migrations` e deteccao de driver.
- [x] Validado localmente com `go test ./...`.
- [x] Validado localmente com `go build ./cmd/web ./cmd/worker ./cmd/seed-news ./cmd/migrate-sqlite-postgres`.
- [x] Backup pre-deploy criado na VPS em `/var/www/inhumas/backups/pre-sprint11-postgres-20260515_214147-db.tar.gz`.
- [x] Sprint 11 publicado na VPS com binarios, comando de migracao, templates, CSS, migrations e documentos.
- [x] Servicos `inhumas-web` e `inhumas-worker` reiniciados e ativos apos deploy.
- [x] Smoke remoto validou `/health`, login administrativo e `/painel/1c2dhviax7/posts`.
- [x] Banco remoto SQLite confirmou `schema_migrations` com versao `sqlite_auto`.
- [x] Binario remoto `/var/www/inhumas/bin/migrate-sqlite-postgres` confirmado executavel e falhando corretamente quando `POSTGRES_DATABASE_URL` nao esta configurado.

Sprint 11 encerrado: o CMS esta pronto para operar em PostgreSQL por configuracao e possui caminho documentado de migracao de dados.

Proxima etapa real: iniciar Sprint 12 - operacao, observabilidade e hardening.

### Etapa 40 - Sprint 12: operacao, observabilidade e hardening

- [x] Middleware de logs estruturados JSON adicionado ao `cmd/web`, com `request_id`, metodo, rota, query, status, bytes, IP, user agent e duracao.
- [x] Header `X-Request-ID` passa a ser emitido em toda requisicao e fica disponivel no contexto da request.
- [x] CSP revisada com `object-src 'none'`, `connect-src 'self'`, `base-uri 'self'`, `form-action 'self'` e `upgrade-insecure-requests` quando a request chega por HTTPS.
- [x] Modo `Content-Security-Policy-Report-Only` configuravel por `CSP_REPORT_ONLY` e `CSP_REPORT_URI` para auditar a remocao futura de scripts/styles inline.
- [x] Nginx example atualizado com headers CSP/Permissions-Policy mais restritivos e bloco report-only comentado.
- [x] `scripts/backup.sh` agora alerta em falha e pode acionar backup externo automaticamente quando `RCLONE_REMOTE` ou `OFFSITE_BACKUP_DIR` estiver configurado.
- [x] `scripts/backup-offsite.sh` criado para copiar backups recentes para rclone ou volume externo fora da VPS.
- [x] `scripts/backup-check.sh` criado para validar idade maxima do ultimo backup de banco e disparar alerta.
- [x] `scripts/disk-check.sh` recebeu limite configuravel por `DISK_ALERT_THRESHOLD`.
- [x] `scripts/journal-retention.sh` criado para aplicar retencao de logs do systemd.
- [x] `scripts/smoke-test.sh` criado para validar `/health`, home e, quando credenciais forem fornecidas por ambiente, login administrativo.
- [x] `docs/ENVIRONMENT.md` atualizado com variaveis de CSP, backup externo, alertas, disco e retencao de logs.
- [x] `docs/OPERATIONS_GUIDE.md` atualizado com monitor externo, cron operacional, backup externo, smoke pos-deploy e checklist de rollback.
- [x] `docs/SECURITY_PREPROD_CHECKLIST.md` atualizado com CSP report-only, backup externo, backup-check e smoke test.
- [x] `docs/CMS_PLANO_GERAL.md` marcou o Sprint 12 como concluido no escopo atual.
- [x] Validado localmente com `go test ./...`.
- [x] Validado localmente com `go build ./cmd/web ./cmd/worker ./cmd/seed-news ./cmd/migrate-sqlite-postgres`.
- [x] Validacao local de shell scripts via `bash -n` nao executou porque o Windows atual nao tem WSL/distro instalada.
- [x] Sintaxe dos scripts validada no Linux da VPS com `bash -n` antes da instalacao.
- [x] Backup pre-deploy criado na VPS em `/var/www/inhumas/backups/pre-sprint12-operational-20260516_002000-db.tar.gz`.
- [x] Backup operacional recorrente executado em `/var/backups/inhumas/db_20260516_002145.db.gz`.
- [x] `scripts/backup-check.sh` validou backup recente com `Backup OK`.
- [x] Sprint 12 publicado na VPS com binarios, templates, CSS, migrations, docs, scripts e exemplos de deploy.
- [x] Binarios alinhados aos nomes reais dos services systemd: `/var/www/inhumas/bin/web` e `/var/www/inhumas/bin/worker`.
- [x] Servicos `inhumas-web` e `inhumas-worker` reiniciados e ativos apos deploy.
- [x] Smoke remoto validou `/health` e home publica via `scripts/smoke-test.sh`.
- [x] Smoke administrativo validou login com a conta admin e acesso ao dashboard.
- [x] Journal da VPS confirmou logs estruturados JSON com evento `http_request` e `request_id`.
- [x] Header local do backend confirmou `X-Request-ID`, CSP revisada e `Permissions-Policy`.

Sprint 12 encerrado: o CMS agora possui base operacional mais robusta, com logs estruturados, smoke pos-deploy, backup-check, backup externo preparado, alertas documentados, CSP report-only e checklist de rollback.

Proxima etapa real: iniciar Sprint 13 - refinamento premium de UX do painel administrativo.

### Etapa 41 - Sprint 13: refinamento premium de UX do painel administrativo

- [x] Layout administrativo reorganizado com sidebar por grupos: Operacao, Editorial, Comercial local e Sistema.
- [x] Marca do CMS adicionada ao topo da sidebar com identificacao `Inhumas em Foco / CMS Editorial`.
- [x] Header administrativo fixo criado com titulo da tela, usuario logado, papel, atalho de nova noticia, link para o site e logout.
- [x] Alertas de sucesso/erro do painel migrados para toasts flutuantes com `aria-live`.
- [x] Loading state generico adicionado aos formularios administrativos ao enviar dados.
- [x] Confirmacoes de exclusao migradas de `confirm()` nativo para modal unico do painel.
- [x] Formularios destrutivos de noticias, categorias, tags, midia, usuarios, banners, lojas, promocoes, influencers, eventos, classificados, bairros e automacao usam `data-confirm`.
- [x] Filtros administrativos passam a salvar valores por tela no `localStorage`, ajudando o operador a manter contexto de listagens.
- [x] Estados vazios receberam visual consistente com borda tracejada, espacamento e peso de texto padronizado.
- [x] Tabelas administrativas receberam contorno consistente e cabecalho sticky dentro da area rolavel.
- [x] CSS responsivo revisado para sidebar, topbar, toasts, modal e formularios longos.
- [x] Query string do CSS atualizada para `20260516-admin-ux-v1`.
- [x] `docs/CMS_PLANO_GERAL.md` marcou o Sprint 13 como concluido no escopo atual.
- [x] Validado localmente com `go test ./...`.
- [x] Validado localmente com `go build ./cmd/web ./cmd/worker ./cmd/seed-news ./cmd/migrate-sqlite-postgres`.
- [x] Pacote Linux gerado com binarios `web`, `worker`, `seed-news` e `migrate-sqlite-postgres`.
- [x] Backup pre-deploy criado na VPS em `/var/www/inhumas/backups/pre-sprint13-admin-ux-20260516_004035-db.tar.gz`.
- [x] Sprint 13 publicado na VPS com binarios, templates, CSS, migrations, docs, scripts e exemplos de deploy.
- [x] Sintaxe dos scripts validada no Linux da VPS com `bash -n` antes da instalacao.
- [x] Servicos `inhumas-web` e `inhumas-worker` reiniciados e ativos apos deploy.
- [x] Smoke remoto validou `/health` e home publica via `scripts/smoke-test.sh`.
- [x] Smoke administrativo validou login, dashboard e presenca do novo layout (`admin-brand`, `admin-topbar`, `data-confirm-modal`).
- [x] CSS remoto confirmou a versao `20260516-admin-ux-v1` com regras de topbar, toast e modal.

Sprint 13 encerrado: o painel administrativo recebeu uma camada premium de UX com navegacao hierarquica, header operacional, toasts, loading states, confirmacao customizada, filtros persistentes e responsividade revisada.

Proxima etapa real: iniciar Sprint 14 - revisao final corporativa, estabilizacao, backlog residual e preparacao para operacao continua.

### Etapa 42 - Sprint 14: revisao final corporativa, estabilizacao e operacao continua

- [x] `docs/CMS_REVISAO_FINAL.md` criado como documento executivo de fechamento do CMS.
- [x] Revisao final consolidou modulos concluidos: painel admin, usuarios/RBAC, noticias, fluxo editorial, midia, taxonomia, auditoria, configuracoes, modulos locais, anuncios, automacao, IA editorial, SEO, PostgreSQL-ready e hardening operacional.
- [x] Backlog residual consolidado por prioridade: backup externo real, monitor externo, SMTP real, remocao de bootstrap, homologacao PostgreSQL, CSP sem inline e WAF/CDN.
- [x] `scripts/production-readiness.sh` criado para validar ambiente, secrets minimos, binarios, services, health, home, protecao admin, backup recente, disco, Nginx e backup externo.
- [x] `.env.example` atualizado com `DB_DRIVER`, `MIGRATIONS_DIR`, CSP report-only, backup externo, alertas e retencao.
- [x] `README.md` atualizado com documentos oficiais, build completo, readiness e status corporativo atual.
- [x] `docs/DEPLOY_GUIDE.md` atualizado com binarios atuais, PostgreSQL suportado por configuracao e etapa de readiness.
- [x] `Makefile` atualizado com build completo, alias `build-linux`, alvo `smoke` e alvo `readiness`.
- [x] `docs/CMS_PLANO_GERAL.md` recebeu Sprint 14 concluido e substituiu a prioridade antiga por backlog residual curto.
- [x] Validado localmente com `go test ./...`.
- [x] Validado localmente com `go build ./cmd/web ./cmd/worker ./cmd/seed-news ./cmd/migrate-sqlite-postgres`.
- [x] Pacote Linux gerado com binarios `web`, `worker`, `seed-news` e `migrate-sqlite-postgres`.
- [x] Backup pre-deploy criado na VPS em `/var/www/inhumas/backups/pre-sprint14-readiness-20260516_005444-db.tar.gz`.
- [x] Sprint 14 publicado na VPS com binarios, templates, CSS, migrations, docs, scripts, README, Makefile e `.env.example`.
- [x] Sintaxe dos scripts validada no Linux da VPS com `bash -n` antes da instalacao.
- [x] Servicos `inhumas-web` e `inhumas-worker` reiniciados e ativos apos deploy.
- [x] Smoke remoto validou `/health` e home publica via `scripts/smoke-test.sh`.
- [x] Readiness remoto validou ambiente, services, health, home, admin protegido, backup recente, disco, Nginx e arquivos oficiais.
- [x] Readiness remoto registrou aviso esperado: backup externo real ainda nao configurado (`RCLONE_REMOTE` ou `OFFSITE_BACKUP_DIR`).
- [x] Documento final e script de readiness confirmados na VPS.

Sprint 14 encerrado: o CMS esta estabilizado para operacao inicial com revisao corporativa, readiness repetivel, documentacao oficial alinhada e backlog residual concentrado em infraestrutura externa/governanca.

Proxima etapa real: executar pendencias externas de producao, com prioridade para backup externo real, monitor externo de uptime e SMTP transacional.
