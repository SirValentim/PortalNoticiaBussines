# Comparacao Essencial com o Prompt

## Rodada final - Fechamento do MVP em 30/04/2026

Decisao final: o produto principal e o projeto Go server-side em `inhumas-em-foco`, rodando com SQLite no MVP operacional. O projeto React em `app/` permanece apenas como prototipo visual historico, sem entrar no deploy oficial.

Validacao desta revisao:

- `go test ./...` passou no backend Go.
- `npm run build` passou no prototipo React, mantendo apenas o aviso conhecido de chunk acima de 500 kB.
- CSRF foi revisado para aplicar `MAX_UPLOAD_SIZE` tambem durante a leitura multipart feita pelo middleware.
- Documentacao final realinhada para separar MVP concluido de itens Pos-MVP.

## Rodada anterior - Prompt Frontend Producao v2.0

Fonte revisada: `Mockup/prompt.md`, `Mockup/01.png` e `Mockup/02.png`.

Decisao desta rodada: aplicar o mockup no projeto Go atual (`inhumas-em-foco`) antes de qualquer envio para VPS. O projeto React em `app/` fica apenas como prototipo historico, pois o novo prompt exige HTML server-side, Tailwind/HTMX quando necessario e proibe React/Vue/jQuery.

Itens em execucao:

- Design system publico com verde `#1F7A63`, amarelo `#F4B400`, fundo `#F7F7F7`, cards brancos, radius moderado e sombra leve.
- Home com hero editorial, ultimas noticias, lojas em destaque, banner patrocinado, promocoes, influencers e bairros.
- Componentes visuais equivalentes a NewsCard, StoreCard, PromotionCard, Banner e Badge.
- Paginas publicas principais alinhadas ao mockup e aos dados reais do backend.
- Validacao local obrigatoria antes de pensar em VPS.

Resultado local desta rodada:

- Header/footer publicos refeitos para a navegacao do mockup.
- Home publica recebeu hero editorial, cards de noticias, lojas, banner, promocoes, influencers e bairros.
- Listas de noticias, lojas, promocoes, busca e influencers foram redesenhadas com os componentes visuais do prompt.
- Detalhe de noticia ganhou layout editorial com sidebar de relacionadas.
- Backend recebeu `/noticias` e `/eventos` para evitar links quebrados na navegacao.
- Fallback visual local criado em `static/images/inhumas-hero.png`.
- Novo logo de `Mockup/logo.png` aplicado no header/footer publico via `static/images/logo.png`.
- Alias `/post/{slug}` criado para cumprir o contrato do prompt sem quebrar a rota real `/noticia/{slug}`.
- Pagina `/classificados` adicionada com categorias iniciais e chamada comercial.
- Filtros server-side adicionados em lojas e promocoes.
- Load more de noticias adicionado em `/noticias/mais?page=N` com atributos `hx-*` e script local minimo.
- Detalhes de loja, promocao e bairro redesenhados para o padrao visual do mockup.
- Estados vazios foram reforcados com mensagens claras e sem blocos genericos.
- `go test ./...` passou e as rotas publicas principais foram abertas no navegador local sem erros de console.
- Deploy publicado na VPS `68.233.125.102`, com backup SQLite pre-deploy, servicos systemd ativos e smoke remoto por IP validado.
- Rodada SEO publicada na VPS: canonical, robots, Open Graph/Twitter, RSS discovery e JSON-LD reforcados.
- Quatro materias locais publicadas com fontes: primeiro prefeito de Inhumas, dados oficiais do IBGE, Camara Municipal em 1947 e importancia de 19 de marco.
- Sitemap remoto regenerado e validado em `/static/sitemap.xml`.
- Correcao pos-publicacao aplicada: campos editoriais vazios/nulos nao quebram mais o detalhe da noticia.
- Destaque principal da home agora e manual por noticia marcada no painel, com fallback para a noticia publicada mais recente.
- Edicoes de conteudo ganharam controles para remover imagens ja enviadas em noticias, lojas, promocoes e influencers.
- Interface comercial organizada: placeholders "Anuncie aqui" entram quando nao ha banner vendido, banners reais ocupam os mesmos pontos quando cadastrados.
- Painel iniciou a evolucao para ferramenta comercial com inventario de espacos, especificacoes de formato e status por posicao.
- Limpeza SEO aplicada: `/noticia/02` removida do banco e do sitemap.

## Feito

- Estrutura Go com `cmd/web`, `cmd/worker`, `internal/*`, templates, static e storage.
- Rotas publicas: home, noticia, categoria, loja, promocao, bairro, influencers, busca, login, contato e sobre.
- Painel admin com posts, lojas, influencers, banners, promocoes, bairros, usuarios e metricas basicas.
- Area de Influencers da Cidade implementada com cadastro admin, listagem publica e perfil individual.
- Edicao de banners e promocoes pelo painel.
- Edicao de lojas usa busca direta por ID no repository.
- Fluxo de publicacao/agendamento com `publish_at` persistido e job de publicacao.
- Validacoes de formulario para posts, agendamentos, banners e promocoes.
- Validacao de banners sem sobreposicao no cadastro e na edicao, excluindo o proprio banner editado.
- Testes automatizados para regras editoriais de Politica & Bastidores.
- Auth com bcrypt, sessao HttpOnly/SameSiteStrict e login funcional.
- Rotacao de sessao com chave atual e chave anterior.
- Admin path obscuro via `ADMIN_PATH_PREFIX`.
- RBAC basico por papel.
- Alteracao segura de senha no painel.
- CSRF em login e formularios admin.
- Rate limit por IP aplicado a login e a requisicoes de escrita.
- Headers de seguranca no app.
- Schema local equivalente em SQLite.
- Migration PostgreSQL inicial criada.
- Worker com claim atomico no SQLite, status `running`, retry/backoff, publicacao de post e expiracao de promocao/banner.
- Busca em runtime SQLite usando FTS5 com fallback seguro para `LIKE`.
- Sitemap automatico diario pelo job `generate_sitemap`.
- Worker agenda backup automatico diario via job `backup_database`.
- Relatorio de dead jobs no painel admin.
- Edit locks com heartbeat no painel de noticias.
- RSS publico em `/rss.xml`.
- JSON-LD Article/Organization nas paginas publicas.
- Tela admin de metricas com totais e rankings baseados em dados reais.
- Upgrade visual do painel admin, com formularios largos, grids responsivos, tabelas padronizadas e CSS versionado.
- Storage abstraido com provider local.
- Upload com validacao de MIME, key aleatoria, WebP principal, thumb e retencao automatica de originais.
- Slug unico e redirects.
- Selo `Patrocinado` em templates principais.
- Scripts/exemplos de VPS: Nginx, systemd, backup, disk-check e `.env.example`.
- `/metrics` operacional protegido com uptime, requests, conexoes de banco, jobs e tamanho de uploads.
- Deploy validado na VPS em `/var/www/inhumas` com systemd web/worker, Nginx, swap 1GB, cron de backup/disco, `/health`, `/login` e restore de backup SQLite.
- `go test ./...` passando.
- MVP Go server-side com SQLite pronto para primeira validacao operacional em VPS.

## Parcial

- PostgreSQL: migration existe, mas runtime ainda usa SQLite local.
- Busca: runtime SQLite usa FTS5; `tsvector` fica apenas para a futura migracao PostgreSQL.
- Jobs: runtime SQLite tem claim atomico com `running`, backoff, backup, sitemap e relatorio de falhas; `FOR UPDATE SKIP LOCKED` fica para PostgreSQL.
- Metricas: tracking e painel com agregacoes existem, incluindo `influencer_view`; falta relatorio comercial mais rico por periodo/anunciante.
- SEO: meta tags, sitemap, RSS e JSON-LD existem; dominio oficial e HTTPS ja foram configurados. Ainda faltam Search Console e rotina editorial continua.
- Upload: fluxo local esta bem coberto; falta apenas provider S3 futuro se necessario.
- Health: verifica banco, uploads, jobs e status de backup configurado; detalhes de disco ficam em `/metrics`.
- Rate limit: aplicado no app e reforcado no exemplo Nginx; ajustes finos ficam para producao real.
- Deploy: fluxo MVP validado na VPS; dominio `https://inhumasemfoco.online` e TLS Lets Encrypt foram configurados.

## Pos-MVP / Infraestrutura Externa

- Completar configuracao Cloudflare, caso seja usado como camada de DNS/protecao.
- Configurar UptimeRobot e copia externa de backups.
- Migrar repository/runtime para PostgreSQL se o volume ou a operacao exigirem.
- Criar migrations versionadas completas e runner de migration para a fase PostgreSQL.
- Implementar worker PostgreSQL com `FOR UPDATE SKIP LOCKED` se a proxima fase trocar o banco.
- Reforcar validacao de banners sem sobreposicao com garantia transacional se houver migracao PostgreSQL/concorrencia alta.
- Implementar provider S3 se a operacao exigir storage externo.
- Adicionar testes automatizados para PostgreSQL e smoke test de deploy.
- Decidir sobre trocar mocks do React por API real caso o prototipo vire produto.

## Conclusao

O prompt completo descreve um produto final mais amplo do que o MVP atual. O repositorio agora cobre o MVP operacional principal: portal publico, painel admin, auth/RBAC, CSRF com limite real de multipart, uploads WebP, busca FTS5 no runtime SQLite, jobs com claim `running`, area de influencers, sitemap, RSS, metricas, backup interno, health, dominio HTTPS e documentacao. O que resta nao e bloqueio de codigo do MVP; e proxima fase de monitoramento externo, backup externo, escala, PostgreSQL e produtos comerciais adicionais.
