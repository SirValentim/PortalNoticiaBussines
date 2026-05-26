# Routes Inventory

Inventario das rotas registradas por `internal/handler/handler.go`.

## Publicas

| Metodo | Rota | Responsabilidade |
|---|---|---|
| GET | `/` | Home |
| GET | `/noticias` | Lista de noticias |
| GET | `/noticias/mais` | Parcial HTMX de noticias |
| GET | `/noticia/{slug}` | Detalhe de noticia |
| GET | `/post/{slug}` | Alias legado para noticia |
| GET | `/categoria/{slug}` | Noticias por categoria |
| GET | `/tag/{slug}` | Noticias por tag |
| GET | `/eventos` | Lista de eventos |
| GET | `/evento/{slug}` | Detalhe de evento |
| GET | `/lojas` | Lista de lojas |
| GET | `/loja/{slug}` | Detalhe de loja |
| GET | `/influencers` | Lista de influencers |
| GET | `/influencer/{slug}` | Detalhe de influencer |
| GET | `/promocoes` | Lista de promocoes |
| GET | `/promocao/{slug}` | Detalhe de promocao |
| GET | `/classificados` | Lista de classificados |
| GET | `/classificado/{slug}` | Detalhe de classificado |
| GET | `/bairro/{slug}` | Pagina de bairro |
| GET | `/busca` | Busca publica |
| GET | `/sobre` | Pagina institucional |
| GET | `/contato` | Pagina de contato |

## Autenticacao

| Metodo | Rota | Responsabilidade |
|---|---|---|
| GET | `/login` | Tela de login |
| POST | `/login` | Login |
| GET | `/logout` | Logout |
| GET | `/recuperar-senha` | Solicitar recuperacao |
| POST | `/recuperar-senha` | Enviar solicitacao |
| GET | `/redefinir-senha/{token}` | Form de nova senha |
| POST | `/redefinir-senha/{token}` | Salvar nova senha |

## SEO e Operacao

| Metodo | Rota | Responsabilidade |
|---|---|---|
| GET | `/health` | Health check |
| GET | `/metrics` | Metricas operacionais protegidas por token |
| GET | `/robots.txt` | Robots |
| GET | `/sitemap.xml` | Sitemap |
| GET | `/rss.xml` | Feed RSS |
| GET | `/manifest.json` | Manifest dinamico por branding |
| POST | `/api/metrics/{type}` | Tracking de metrica |

## Admin

Todas as rotas admin usam `ADMIN_PATH_PREFIX`.

| Area | Rotas principais |
|---|---|
| Dashboard | `/`, `/metrics`, `/dead-jobs`, `/audit`, `/settings` |
| Posts | `/posts`, `/posts/new`, `/posts/{id}/edit`, `/posts/{id}/preview`, acoes de autosave, lock, review, approve, reject, publish e IA |
| Automacao | `/automation`, sources CRUD, run source, run all |
| Taxonomia | `/categories`, `/tags` |
| Midia | `/media` |
| Comercial | `/banners`, `/stores`, `/promotions` |
| Local | `/events`, `/classifieds`, `/influencers`, `/neighborhoods` |
| Usuarios | `/users`, edit e troca de senha |
| Portais | `/tenants`, criacao/edicao, dominios, features e vinculos de usuarios |

## Observacoes

- Rotas publicas em portugues devem ser preservadas por SEO.
- APIs futuras devem usar `/api/v1`.
- Antes de extrair routers por modulo, criar testes de smoke para os grupos acima.
- Rotas de portais exigem permissao `tenants:manage` e sao reservadas ao papel `super_admin`.
