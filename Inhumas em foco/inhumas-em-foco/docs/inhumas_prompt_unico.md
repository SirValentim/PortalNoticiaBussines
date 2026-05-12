# PROMPT ÚNICO — INHUMAS EM FOCO
## Portal local full-stack Go 1.24 + Monetização integrada | VPS 1GB RAM

---

## 0. STATUS FINAL DO MVP — 30/04/2026

Este prompt continua sendo a visão completa do produto. Para fechamento do projeto atual, a leitura oficial é:

- Produto principal: `inhumas-em-foco`, backend Go server-side com templates HTML, CSS próprio e interatividade leve local.
- Runtime do MVP: SQLite em produção inicial, com migration PostgreSQL mantida para evolução futura.
- Protótipo React: pasta `app/`, mantida como referência visual histórica, fora do deploy oficial.
- Infra publicada: VPS com Nginx, systemd web/worker, swap, backup, domínio `https://inhumasemfoco.online` e TLS Lets Encrypt.
- MVP concluído: portal público, painel admin, auth/RBAC, CSRF, uploads WebP, SEO técnico, sitemap, RSS, JSON-LD, jobs recorrentes, backup, health, métricas e área comercial inicial.
- Pós-MVP: PostgreSQL runtime, S3, UptimeRobot, backup externo, relatórios comerciais avançados, Cloudflare completo e novos produtos comerciais.

Regra de execução para agents: antes de implementar qualquer mudança nova, diferencie o que é requisito do MVP fechado do que é evolução Pós-MVP. Não reabrir PostgreSQL, React/API real ou S3 como bloqueio do MVP salvo pedido explícito.

---

## 1. STACK E ARQUITETURA

```
VPS 1GB RAM (Debian 12)
├── Nginx (proxy + estáticos + rate limit + WAF)
├── App Go (HTTP server, porta 8080)
├── App Go (Worker, modo -worker, fila PostgreSQL)
└── PostgreSQL 15/16 (tunado)
```

### 1.1 PostgreSQL Tuning
```ini
shared_buffers = 64MB
effective_cache_size = 256MB
work_mem = 4MB
max_connections = 20
random_page_cost = 1.1
```

### 1.2 Swap Obrigatório
```bash
fallocate -l 1G /swapfile && chmod 600 /swapfile && mkswap /swapfile && swapon /swapfile
```

### 1.3 Systemd Limits
```ini
MemoryMax=200M
CPUQuota=80%
Restart=always
```

---

## 2. SEGURANÇA TOTAL

| Ameaça | Mitigação |
|--------|-----------|
| SQL Injection | Queries parametrizadas APENAS. NUNCA concatenação de string em SQL. Sanitizar slugs com `[a-z0-9-]` |
| XSS | `templ` escapa por padrão. NUNCA `raw/html` sem `bluemonday`. CSP header |
| CSRF | `gorilla/csrf` em TODOS os forms do painel. Token em meta tag + header HTMX |
| Rate Limit | Nginx: 10r/s geral, 5r/m login, 2r/m upload |
| Admin exposto | Path obscuro via `ADMIN_PATH_PREFIX` (ex: `/painel/7x9k2m`). Bloquear IP estrangeiro |
| Brute force | Tabela `login_attempts`: bloqueia IP após 5 falhas por 30min |
| Upload malicioso | Max 5MB, validar MIME real via `http.DetectContentType`, magic bytes, renomear para UUID |
| Sessão | HttpOnly, Secure, SameSiteStrict. `SESSION_SECRET` ≥32 bytes. Suportar 2 chaves (atual+anterior) |
| DDoS | Cloudflare gratuito (DNS + Under Attack). Esconder IP real da VPS |

### Headers Nginx obrigatórios
```nginx
add_header X-Frame-Options "SAMEORIGIN" always;
add_header X-Content-Type-Options "nosniff" always;
add_header X-XSS-Protection "1; mode=block" always;
add_header Referrer-Policy "strict-origin-when-cross-origin" always;
add_header Content-Security-Policy "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self';" always;
add_header Strict-Transport-Security "max-age=63072000; includeSubDomains; preload" always;
server_tokens off;
```

---

## 3. SCHEMA DO BANCO (Completo)

```sql
-- ENUMS
CREATE TYPE job_status AS ENUM ('pending','running','completed','failed');
CREATE TYPE job_type AS ENUM ('publish_post','expire_promotion','expire_banner','backup_database','vacuum_db','generate_sitemap','cleanup_old_jobs','compress_old_uploads');
CREATE TYPE user_role AS ENUM ('admin','editor','comercial');
CREATE TYPE post_status AS ENUM ('draft','scheduled','published','archived');

-- USUÁRIOS
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    email VARCHAR(200) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role user_role DEFAULT 'editor',
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW()
);

-- CATEGORIAS
CREATE TABLE categories (
    id SERIAL PRIMARY KEY,
    slug VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    requires_editorial_notes BOOLEAN DEFAULT false
);
INSERT INTO categories (slug, name, requires_editorial_notes) VALUES
('noticias','Notícias',false),
('politica-bastidores','Política & Bastidores',true),
('influencers','Influencers da Cidade',false),
('eventos','Eventos',false);

-- POSTS
CREATE TABLE posts (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(300) NOT NULL,
    slug VARCHAR(300) UNIQUE NOT NULL,
    excerpt TEXT,
    content TEXT NOT NULL,
    cover_image_key VARCHAR(500),
    category_id INT REFERENCES categories(id),
    author_id BIGINT REFERENCES users(id),
    status post_status DEFAULT 'draft',
    is_sponsored BOOLEAN DEFAULT false,
    editorial_notes TEXT,
    editor_responsible VARCHAR(200),
    published_at TIMESTAMP,
    publish_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    search_vector tsvector
);
CREATE INDEX idx_posts_status ON posts(status, published_at DESC);
CREATE INDEX idx_posts_category ON posts(category_id, status);
CREATE INDEX idx_posts_search ON posts USING GIN(search_vector);

-- TRIGGER tsvector PT-BR
CREATE OR REPLACE FUNCTION posts_search_update() RETURNS trigger AS $$
BEGIN
    NEW.search_vector :=
        setweight(to_tsvector('portuguese', COALESCE(NEW.title,'')), 'A') ||
        setweight(to_tsvector('portuguese', COALESCE(NEW.excerpt,'')), 'B') ||
        setweight(to_tsvector('portuguese', COALESCE(NEW.content,'')), 'C');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
CREATE TRIGGER posts_search_trigger BEFORE INSERT OR UPDATE ON posts
    FOR EACH ROW EXECUTE FUNCTION posts_search_update();

-- SLUG REDIRECTS (SEO)
CREATE TABLE slug_redirects (
    old_slug VARCHAR(300) PRIMARY KEY,
    new_slug VARCHAR(300) NOT NULL,
    entity_type VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- LOJAS
CREATE TABLE stores (
    id BIGSERIAL PRIMARY KEY,
    slug VARCHAR(200) UNIQUE NOT NULL,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    category VARCHAR(100),
    address TEXT,
    phone VARCHAR(50),
    whatsapp VARCHAR(50),
    logo_key VARCHAR(500),
    cover_image_key VARCHAR(500),
    is_sponsored BOOLEAN DEFAULT false,
    is_featured BOOLEAN DEFAULT false,
    neighborhood_id INT,
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW()
);

-- PROMOÇÕES
CREATE TABLE promotions (
    id BIGSERIAL PRIMARY KEY,
    store_id BIGINT REFERENCES stores(id),
    title VARCHAR(300) NOT NULL,
    slug VARCHAR(300) UNIQUE NOT NULL,
    description TEXT,
    price_display VARCHAR(100),
    image_key VARCHAR(500),
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    status VARCHAR(50) DEFAULT 'active',
    is_sponsored BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW()
);
CREATE INDEX idx_promotions_active ON promotions(store_id, status, end_date)
    WHERE status = 'active';

-- BANNERS
CREATE TABLE banners (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    position VARCHAR(50) NOT NULL, -- hero, sidebar_top, sidebar_bottom, in_feed, sticky_footer
    image_key VARCHAR(500) NOT NULL,
    link_url TEXT NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    active BOOLEAN DEFAULT true,
    priority INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW()
);
CREATE INDEX idx_banners_position ON banners(position, active, start_date, end_date);

-- BAIRROS
CREATE TABLE neighborhoods (
    id SERIAL PRIMARY KEY,
    slug VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    meta_title VARCHAR(60),
    meta_description VARCHAR(160),
    cover_image_key VARCHAR(500),
    created_at TIMESTAMP DEFAULT NOW()
);

-- JOBS (fila persistente)
CREATE TABLE jobs (
    id BIGSERIAL PRIMARY KEY,
    type job_type NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}',
    status job_status DEFAULT 'pending',
    run_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    attempts INT DEFAULT 0,
    max_attempts INT DEFAULT 3,
    error TEXT,
    processed_at TIMESTAMP WITH TIME ZONE
);
CREATE INDEX idx_jobs_pending ON jobs(run_at) WHERE status = 'pending';

-- DEAD JOBS (falhas após retry)
CREATE TABLE dead_jobs (LIKE jobs INCLUDING ALL);

-- MÉTRICAS (monetização)
CREATE TABLE metrics (
    id BIGSERIAL PRIMARY KEY,
    metric_type VARCHAR(50) NOT NULL, -- banner_impression, banner_click, store_view, promo_click, coupon_used, store_whatsapp_click
    entity_type VARCHAR(50) NOT NULL,
    entity_id BIGINT NOT NULL,
    user_id BIGINT,
    ip_address INET,
    user_agent TEXT,
    referrer TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);
CREATE INDEX idx_metrics_entity ON metrics(entity_type, entity_id, metric_type);
CREATE INDEX idx_metrics_date ON metrics(created_at);

-- AUDIT LOG
CREATE TABLE audit_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT,
    action VARCHAR(50) NOT NULL,
    entity_type VARCHAR(50),
    entity_id BIGINT,
    changes JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- EDIT LOCKS
CREATE TABLE edit_locks (
    entity_type VARCHAR(50) NOT NULL,
    entity_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    locked_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL,
    PRIMARY KEY (entity_type, entity_id)
);

-- LOGIN ATTEMPTS (segurança)
CREATE TABLE login_attempts (
    id BIGSERIAL PRIMARY KEY,
    ip_address INET NOT NULL,
    email VARCHAR(200),
    success BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW()
);
CREATE INDEX idx_login_ip ON login_attempts(ip_address, created_at DESC);
```

---

## 4. AUTOMAÇÕES E JOBS

### Worker Go (mesma binary, flag `-worker`)
- Polling a cada 30s via `SELECT ... FOR UPDATE SKIP LOCKED`
- Retry com backoff: 1min, 5min, 15min
- Após 3 falhas: mover para `dead_jobs`

### Jobs Obrigatórios
| Job | Gatilho | Ação |
|-----|---------|------|
| `publish_post` | `run_at = post.publish_at` | `UPDATE posts SET status='published'` |
| `expire_promotion` | `run_at = promo.end_date` | `UPDATE promotions SET status='expired'` |
| `expire_banner` | `run_at = banner.end_date` | `UPDATE banners SET active=false` |
| `backup_database` | Diário 03:00 | `pg_dump + gzip` |
| `vacuum_db` | Domingo 04:00 | `VACUUM ANALYZE` |
| `generate_sitemap` | Diário 02:00 | Gera `/static/sitemap.xml` |
| `cleanup_old_jobs` | Semanal | `DELETE jobs completed > 30 dias` |
| `compress_old_uploads` | Mensal | Comprimir imagens > 6 meses |

### Agendamento no código
```go
func (s *JobService) Schedule(ctx context.Context, jobType string, payload map[string]any, runAt time.Time) error {
    payloadJSON, _ := json.Marshal(payload)
    _, err := s.db.ExecContext(ctx,
        `INSERT INTO jobs (type, payload, run_at) VALUES ($1, $2, $3)`,
        jobType, payloadJSON, runAt)
    return err
}
```

---

## 5. BACKUP AUTOMÁTICO

### Script `/opt/inhumas/backup.sh`
```bash
#!/bin/bash
set -e
BACKUP_DIR="/var/backups/inhumas"
DATE=$(date +%Y%m%d_%H%M%S)
mkdir -p $BACKUP_DIR
pg_dump -U inhumas -d inhumasdb | gzip > $BACKUP_DIR/db_$DATE.sql.gz
tar -czf $BACKUP_DIR/uploads_$DATE.tar.gz /var/www/inhumas/uploads/
tar -czf $BACKUP_DIR/config_$DATE.tar.gz /var/www/inhumas/.env /etc/nginx/sites-available/ /etc/systemd/system/inhumas.service
find $BACKUP_DIR -name "*.gz" -mtime +30 -delete
echo "Backup: $DATE"
```

### Cron
```cron
0 3 * * * root /opt/inhumas/backup.sh >> /var/log/inhumas/backup.log 2>&1
```

### Alerta de disco `/opt/inhumas/disk-check.sh`
```bash
USAGE=$(df / | tail -1 | awk '{print $5}' | sed 's/%//')
if [ $USAGE -gt 80 ]; then
    echo "Disco em ${USAGE}%" | mail -s "Inhumas - Disco Cheio" admin@inhumasemfoco.com.br
fi
```

---

## 6. MONETIZAÇÃO + HOME (Feed Misto)

### Princípio: Editorial e Comercial alternados. Nunca 2 anúncios seguidos.

### Produtos e Preços
| Produto | Preço | Posição na Home |
|---------|-------|-----------------|
| Banner Hero | R$ 300-400/mês | Topo, full-width |
| Vitrine de Lojas | R$ 80-150/loja/mês | Carousel horizontal |
| Promoção do Dia | R$ 30-50/promo | Bloco com countdown HTMX |
| Banner In-Feed | R$ 150-250/mês | Entre blocos editoriais |
| Conteúdo Patrocinado | R$ 150-300/matéria | Mesmo layout de notícia, selo âmbar |
| Classificados Destaque | R$ 10-20/classificado | Bloco com fotos |
| Eventos Patrocinados | R$ 200-500/evento | Bloco eventos |
| Guia por Bairro | R$ 50/loja/mês | Bloco "Comércio do seu bairro" |
| Newsletter Patrocinada | R$ 100-200/envio | Topo do email diário |
| Cupons Exclusivos | R$ 50/mês + comissão | Página da loja |

### Regras de Exibição
- Máximo 1 banner hero (só se houver pagante)
- 1 banner in-feed a cada 3 blocos editoriais
- Vitrine: 4-6 lojas em carousel
- Sticky footer: 1, mobile only, fechável
- Selo "Patrocinado" obrigatório em TODOS os templates comerciais

### Estrutura da Home (model)
```go
type HomeData struct {
    Headline       *Post
    LatestNews     []Post
    Bastidores     []Post
    HeroBanner     *Banner
    InFeedBanner   *Banner
    StickyBanner   *Banner
    FeaturedStores []Store
    FeaturedPromos []Promotion
    SponsoredPosts []Post
    Events         []Event
    Classifieds    []Classified
    Neighborhood   *NeighborhoodHighlight
    LatestJobs     []JobListing
    Utilities      Utilities
}
```

### Query de banners ativos
```sql
SELECT * FROM banners
WHERE active = true
  AND position = $1
  AND CURRENT_DATE BETWEEN start_date AND end_date
ORDER BY priority DESC, created_at DESC
LIMIT 1;
```

### Query de promoções ativas
```sql
SELECT p.*, s.name as store_name, s.slug as store_slug
FROM promotions p
JOIN stores s ON s.id = p.store_id
WHERE p.status = 'active'
  AND CURRENT_DATE BETWEEN p.start_date AND p.end_date
ORDER BY p.created_at DESC
LIMIT 3;
```

### Métricas (tracking assíncrono)
```go
func MetricsMiddleware(metricType string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            go func() {
                h.metricsRepo.Track(r.Context(), metricType, entityType, entityID, r)
            }()
            next.ServeHTTP(w, r)
        })
    }
}
```

---

## 7. SEO AUTOMÁTICO

### Componente Head (templ)
Parâmetros: Title, Description, Image, URL, Type, NoIndex, PublishedAt, ModifiedAt, Author, Tags.
Gera automaticamente: `<title>`, `<meta description>`, Open Graph, Twitter Card, canonical, JSON-LD (Article/Organization), meta robots.

### Sitemap
Job diário gera `/static/sitemap.xml` com: posts publicados, páginas estáticas, lojas, promoções ativas, páginas de bairro.

### Busca (tsvector PT-BR)
```sql
SELECT id, title, slug, excerpt, cover_image_key,
       ts_rank_cd(search_vector, plainto_tsquery('portuguese', $1), 32) as rank
FROM posts
WHERE search_vector @@ plainto_tsquery('portuguese', $1)
  AND status = 'published'
ORDER BY rank DESC, published_at DESC
LIMIT $2;
```

### Páginas de Bairro
Rota `/bairro/{slug}`. SEO: `"Notícias e comércio do bairro {Nome} em Inhumas, GO"`. Listar posts, lojas, eventos do bairro.

### Slugs Únicos
Transação SERIALIZABLE. Se colisão, sufixo incremental (-2, -3). Tabela `slug_redirects` para 301 quando slug mudar.

---

## 8. RBAC E REGRAS DE NEGÓCIO

### Papéis
- `admin`: tudo
- `editor`: posts (criar, publicar, editar qualquer, arquivar)
- `comercial`: lojas (criar, editar própria), banners, promoções

### Permissões via middleware Chi
```go
func RequirePermission(perm auth.Permission) func(http.Handler) http.Handler
```

### Lock de Edição
Heartbeat a cada 2min via HTMX. Expira em 10min. Se outro usuário tentar: aviso "Bloqueado por {nome} até {hora}".

### Validação Condicional: Política & Bastidores
```go
if post.CategorySlug == "politica-bastidores" {
    if post.EditorialNotes == "" || post.EditorResponsible == "" {
        return error("campo de apuração e responsável editorial obrigatórios")
    }
}
```

### Banners — Sem Sobreposição
```sql
SELECT COUNT(*) FROM banners
WHERE position = $1 AND active = true
  AND (start_date, end_date) OVERLAPS ($2, $3);
```
Se > 0, erro: "Já existe banner ativo na posição X nesse período".

### Conteúdo Patrocinado
Campo `is_sponsored` em posts, stores, promotions. Componente `SponsoredBadge` em TODOS os templates: home, listagens, busca, páginas internas, RSS.

---

## 9. STORAGE ABSTRAÍDO (S3-Ready)

```go
type Provider interface {
    Upload(ctx context.Context, key string, r io.Reader, contentType string) (*FileInfo, error)
    Delete(ctx context.Context, key string) error
    URL(ctx context.Context, key string) string
}
```

- MVP: `LocalProvider` (`/var/www/inhumas/uploads/`)
- Futuro: `S3Provider` (mesma interface, zero refatoração)
- Regra: armazenar apenas `key` no banco, NUNCA path absoluto. NUNCA `os.ReadFile` direto em handlers.

### Compressão de Mídia
- Max upload: 2MB (Nginx `client_max_body_size 2M`)
- Conversão automática para WebP
- Redimensionamento: Hero 1200x630, Thumbnail 400x300, Banner conforme posição
- Lib: `github.com/disintegration/imaging`

---

## 10. MANUTENÇÃO E DEPLOY

### Makefile
```makefile
templ:
	templ generate
dev: templ
	go run ./cmd/web
worker:
	go run ./cmd/worker
build: templ
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/web ./cmd/web
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/worker ./cmd/worker
deploy: build
	ssh vps "sudo systemctl stop inhumas-web inhumas-worker"
	rsync -avz --delete bin/ vps:/var/www/inhumas/bin/
	ssh vps "sudo systemctl start inhumas-web inhumas-worker"
status:
	ssh vps "sudo systemctl status inhumas-web inhumas-worker"
logs:
	ssh vps "sudo journalctl -u inhumas-web -u inhumas-worker -f"
backup:
	ssh vps "sudo /opt/inhumas/backup.sh"
migrate-up:
	migrate -path migrations -database "$(DATABASE_URL)" up
migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down
health:
	curl -s https://inhumasemfoco.com.br/health | jq .
```

### Health Check
Endpoint `/health` retorna: status, timestamp, version, checks (database, disk, backup).

### Modo Manutenção
Env var `MAINTENANCE_MODE=true`. Middleware retorna página estática para rotas públicas. Painel admin continua acessível.

### Monitoramento
- UptimeRobot (gratuito): monitorar `/health` a cada 5min
- Métricas internas leves em `/metrics` (protegido): uptime, requests, db_connections, jobs_pending, disk_usage

---

## 11. PLANO DE CONTINGÊNCIA

| Incidente | Detecção | Resposta Imediata | Recuperação |
|-----------|----------|-------------------|-------------|
| Site fora do ar | UptimeRobot | `systemctl status` + restart | Restaurar backup mais recente |
| DDoS/Ataque | Cloudflare + logs | Ativar "Under Attack" no CF | Bloquear IPs, analisar logs |
| Banco corrompido | Health check | Parar app, não escrever | Restaurar pg_dump + WAL |
| Disco cheio | Script monitoramento | Limpar logs, comprimir uploads | Expandir disco ou migrar storage |
| Vazamento de dados | Alerta anômalo | Isolar endpoint, rotacionar secrets | Notificar afetados (LGPD) |
| Processo judicial | Notificação | Preservar logs e backups | Publicar direito de resposta/errata |
| Anunciante cancela 40% receita | Relatório financeiro | Ativar fundo de reserva | Campanha de captação emergencial |
| Burnout operador | Autoavaliação | Reduzir frequência de posts | Automatizar mais, contratar freelancer |

---

## 12. ESTRUTURA DE DIRETÓRIOS

```
inhumas-em-foco/
├── cmd/
│   ├── web/main.go
│   └── worker/main.go
├── internal/
│   ├── auth/          # RBAC
│   ├── config/        # Env vars
│   ├── handler/       # HTTP handlers
│   ├── middleware/    # Auth, CSRF, rate limit, maintenance, metrics
│   ├── model/         # Structs
│   ├── repository/    # SQL queries
│   ├── service/       # Regras de negócio
│   ├── session/       # gorilla/sessions
│   ├── slug/          # Gerador único
│   ├── storage/       # Interface + LocalProvider
│   ├── view/
│   │   ├── components/  # SEOHead, SponsoredBadge, Countdown, FormError
│   │   ├── layouts/
│   │   └── pages/       # Home, Post, Loja, etc.
│   └── worker/        # Processamento de jobs
├── migrations/
├── scripts/           # backup.sh, disk-check.sh
├── static/            # CSS, JS, sitemap.xml
├── uploads/
│   ├── original/      # preservado 7 dias
│   ├── webp/
│   └── thumb/
├── .env.example
├── Makefile
├── go.mod
└── README.md
```

---

## 13. CHECKLIST DE IMPLEMENTAÇÃO

### Fase 1 — Fundação (Semana 1)
- [ ] Migrations completas (todas as tabelas acima)
- [ ] PostgreSQL tunado + swap configurado
- [ ] `storage.Provider` + `LocalProvider`
- [ ] `slug.Generator` transacional + `slug_redirects`
- [ ] `session` (gorilla/sessions, Secure, HttpOnly, SameSiteStrict)
- [ ] `auth.RBAC` (roles, permissions, middleware)
- [ ] Middleware: auth, CSRF, rate limit, maintenance mode, metrics
- [ ] Componente `Head` SEO parametrizável
- [ ] Componente `SponsoredBadge` reutilizável

### Fase 2 — Automações (Semana 2)
- [ ] Tabela `jobs` + worker completo
- [ ] Agendamento de posts (`publish_post`)
- [ ] Expiração automática (`expire_promotion`, `expire_banner`)
- [ ] Backup automático (script + cron)
- [ ] Sitemap automático
- [ ] Vacuum e manutenção automática
- [ ] Alerta de disco > 80%

### Fase 3 — Conteúdo + Comercial (Semana 3)
- [ ] CRUD posts com validação condicional (Política & Bastidores)
- [ ] CRUD lojas, promoções, banners
- [ ] Validação de sobreposição de banners
- [ ] Busca tsvector PT-BR
- [ ] Páginas de bairro
- [ ] Upload de mídia com compressão WebP
- [ ] Tabela `metrics` + tracking assíncrono

### Fase 4 — Home Monetizada + Segurança (Semana 4)
- [ ] Home com feed misto (editorial + comercial)
- [ ] Vitrine de lojas (carousel)
- [ ] Promoção do dia com countdown HTMX
- [ ] Banners posicionados (hero, in-feed, sticky)
- [ ] Conteúdo patrocinado com selo
- [ ] Classificados destaque
- [ ] Nginx: WAF, rate limit, headers de segurança, SSL HSTS
- [ ] Cloudflare ativo
- [ ] Admin path obscuro
- [ ] Health checks + UptimeRobot
- [ ] Deploy com Makefile

---

## 14. REGRAS PARA AGENTS

- NUNCA exponha detalhes de erro do banco para o cliente (log interno apenas)
- SEMPRE use `context.Context` com timeout em operações de banco e HTTP
- NUNCA armazene senhas em plain text. Use bcrypt cost ≥ 12
- SEMPRE valide e sanitize inputs antes de processamento
- NUNCA commite `.env` ou chaves secretas
- SEMPRE use transactions para operações multi-tabela
- NUNCA use `SELECT *` em queries de produção
- SEMPRE armazene apenas `key` no banco (nunca path absoluto de arquivo)
- NUNCA use `os.ReadFile` direto em handlers (sempre via `storage.Provider`)
- SEMPRE inclua selo `SponsoredBadge` em conteúdo comercial
- NUNCA concatene strings em queries SQL (parametrizadas apenas)

---

## 19. MASTER LAYOUT GRID (BASE DO SITE)

Layout global desktop:

- Container: `max-width: 1200px`
- `margin: 0 auto`
- Padding horizontal: `24px`

Secoes:

- Espacamento entre secoes: `32px`
- Titulo de secao:
  - `font-size: 20px`
  - `font-weight: 600`
  - `margin-bottom: 16px`

Regra: todas as secoes devem alinhar exatamente no mesmo eixo vertical.

---

## 20. HERO + SIDEBAR (PIXEL PERFECT)

Grid:

- `display: grid`
- `grid-template-columns: 2fr 1fr`
- `gap: 24px`

Hero:

- Altura: `420px`
- `border-radius: 12px`
- `overflow: hidden`
- `position: relative`

Imagem:

- `width: 100%`
- `height: 100%`
- `object-fit: cover`

Overlay:

- `background: linear-gradient(to top, rgba(0,0,0,0.7), rgba(0,0,0,0.2))`

Conteudo:

- `position: absolute`
- `bottom: 24px`
- `left: 24px`
- `right: 24px`

Titulo:

- `font-size: 34px`
- `line-height: 1.2`
- `font-weight: 700`
- `color: white`
- `max-width: 80%`

Descricao:

- `font-size: 14px`
- `color: rgba(255,255,255,0.9)`
- `margin-top: 8px`

Botao:

- `margin-top: 12px`
- `padding: 10px 18px`
- `border-radius: 8px`

Sidebar:

- `background: white`
- `border-radius: 12px`
- `padding: 16px`
- `height: 420px`
- `display: flex`
- `flex-direction: column`
- `gap: 12px`

Item:

- `border-bottom: 1px solid var(--border)`
- `padding-bottom: 8px`

Categoria:

- `font-size: 10px`
- `color: var(--primary)`
- `font-weight: 600`

Titulo:

- `font-size: 13px`
- `font-weight: 500`

---

## 21. NEWS GRID (CORRETO)

Grid:

- Desktop: 4 colunas
- `gap: 16px`

Card:

- `background: white`
- `border-radius: 10px`
- `overflow: hidden`

Imagem:

- `width: 100%`
- `height: 120px`
- `object-fit: cover`

Conteudo:

- `padding: 10px`

Categoria:

- `font-size: 10px`
- `margin-bottom: 4px`

Titulo:

- `font-size: 14px`
- `font-weight: 600`
- `line-height: 1.3`
- `max-height: 2.6em`
- `overflow: hidden`

---

## 22. BANNER (AJUSTE PREMIUM)

Banner:

- `height: 100px`
- `border-radius: 12px`
- `padding: 16px 20px`

Layout:

- `display: flex`
- `justify-content: space-between`
- `align-items: center`

Texto:

- `font-size: 16px`
- `font-weight: 600`

Botao:

- `background: var(--accent)`
- `color: black`
- `padding: 10px 16px`
- `border-radius: 8px`

Label:

- `position: absolute`
- `top: 8px`
- `left: 8px`
- `font-size: 10px`

---

## 23. VISUAL DENSITY CONTROL

Objetivo: a tela deve mostrar mais conteudo sem parecer poluida.

Regras:

- Reduzir alturas excessivas
- Evitar espacos grandes vazios
- Cards mais compactos
- Manter consistencia

Se parecer espacado demais: reduzir padding em 20%.

---

## 24. MICRO INTERACOES (NIVEL PRODUTO)

Hover:

- Cards: `transform: translateY(-2px)` com `transition: 0.2s`
- Imagens: `scale: 1.03` no hover
- Botoes: `opacity: 0.9` no hover

---

## 25. MOBILE (CORRECAO REAL)

Hero:

- Altura: `250px`
- Titulo: `20px`

Grid:

- 1 coluna

Sidebar:

- Vira lista abaixo do hero

News:

- 1 coluna
- Imagem: `100% width`

Banner:

- Stack vertical

---

## 26. UI POLISH CHECKLIST (OBRIGATORIO)

Antes de finalizar:

- Tudo alinhado verticalmente?
- Espacamentos consistentes?
- Nenhum elemento sobrando?
- Visual bate com mockup?

Se nao: ajustar automaticamente.
