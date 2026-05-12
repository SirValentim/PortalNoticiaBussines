🧠 PROMPT ÚNICO — FRONTEND PRODUÇÃO
INHUMAS EM FOCO (v2.0)
🎯 0. OBJETIVO

Desenvolver o frontend completo do portal Inhumas em Foco, seguindo os mockups fornecidos, com foco em:

Mobile-first
Alta performance (VPS 1GB)
UX moderna (portal + guia comercial)
Monetização integrada
Integração total com backend Go
🧱 1. STACK (OBRIGATÓRIO)
HTML (templ)
TailwindCSS
HTMX (interatividade)
Alpine.js (somente se necessário)

🚫 PROIBIDO:

React
Vue
jQuery
SPA pesada
🎨 2. DESIGN SYSTEM (OBRIGATÓRIO)
🎯 Tokens
:root {
  --primary: #1F7A63;
  --primary-dark: #145A48;

  --accent: #F4B400;
  --accent-dark: #C49000;

  --bg: #F7F7F7;
  --card: #FFFFFF;

  --text: #1A1A1A;
  --text-light: #6B7280;

  --border: #E5E7EB;
}
📏 Spacing
xs: 4px
sm: 8px
md: 16px
lg: 24px
xl: 32px
🔲 Border Radius
sm: 6px
md: 10px
lg: 14px
🌫 Shadow
card: 0 2px 8px rgba(0,0,0,0.05)
hover: 0 4px 12px rgba(0,0,0,0.08)
🔤 Tipografia
Fonte: Inter
Títulos: 600–700
Texto: 400–500
🧩 3. COMPONENT CONTRACTS (OBRIGATÓRIO)
📰 NewsCard
Props:
- title (string, max 120)
- image_url (string)
- category (string)
- published_at (date)

Regras:
- imagem 16:9
- título máximo 2 linhas
- hover: leve zoom (desktop)
🏪 StoreCard
Props:
- name
- logo_url
- rating
- slug

Regras:
- logo circular
- nome truncado 1 linha
💸 PromotionCard
Props:
- title
- image
- price
- old_price
- discount

Regras:
- preço destaque
- badge desconto obrigatório
📢 Banner
Props:
- image
- link
- position

Regras:
- SEMPRE label "Patrocinado"
- largura total
🧷 Badge

Tipos:

categoria
patrocinado
bairro
🏠 4. HOME (CRÍTICO)
Layout obrigatório
Hero

Últimas notícias (grid)

Banner

Lojas (scroll horizontal)

Promoções

Mais notícias

Eventos

Bairros
HERO (CRÍTICO)
- imagem real
- overlay escuro
- título grande
- categoria
- botão CTA
Regras
Hero ocupa 70% largura desktop
Mobile: altura 250px
Conteúdo acima da dobra obrigatório
📰 5. PÁGINAS
/noticias
lista vertical
filtro por categoria
HTMX load more
/post/{slug}
Título
Imagem
Conteúdo
Banner inline
Relacionados
/lojas
grid responsivo
filtro
busca
/loja/{slug}
capa
logo
descrição
botão WhatsApp
promoções
/promocoes
grid visual
preço destaque
/influencers
grid
avatar + nicho
/eventos
lista com data visual
/busca
resultados mistos
/bairro/{slug}
header local
conteúdo segmentado
🔌 6. API CONTRACT (OBRIGATÓRIO)
Posts
GET /api/posts
[
  {
    "id": 1,
    "title": "...",
    "slug": "...",
    "excerpt": "...",
    "cover_image_url": "...",
    "category": "...",
    "published_at": "..."
  }
]
Stores
GET /api/stores
[
  {
    "id": 1,
    "name": "...",
    "slug": "...",
    "logo_url": "...",
    "rating": 4.5
  }
]
Promotions
GET /api/promotions
⚡ 7. HTMX (OBRIGATÓRIO)

Usar para:

paginação
filtros
busca
Exemplo
<button 
  hx-get="/noticias?page=2"
  hx-target="#list"
  hx-swap="outerHTML">
  Carregar mais
</button>
📱 8. MOBILE FIRST
1 coluna
scroll horizontal (lojas)
botões grandes
sem overflow quebrado
💰 9. MONETIZAÇÃO (EXECUTÁVEL)
- Inserir 1 banner a cada 6 conteúdos
- Nunca 2 banners seguidos
- Sempre exibir badge "Patrocinado"
- Banner ocupa 100% container
🚀 10. PERFORMANCE
WebP obrigatório
lazy loading
JS mínimo
CSS otimizado
🔐 11. SEGURANÇA
não usar HTML raw
escape automático
validar inputs
🧪 12. ESTADOS (OBRIGATÓRIO)

Cada componente deve ter:

loading (skeleton)
vazio
erro (retry)
🚫 13. ANTI-PATTERNS

PROIBIDO:

lorem ipsum
blocos vazios
imagens quebradas
layout genérico
centralizar tudo
excesso de azul (seguir paleta)
🧠 14. ORDEM DE EXECUÇÃO
1. Layout base
2. Home (mock data)
3. Componentes
4. API integração
5. Páginas
6. Mobile refine
✅ 15. DEFINITION OF DONE
todas páginas funcionando
responsivo completo
sem erros console
lighthouse ≥ 85 mobile
sem blocos vazios
consistente com design system
🎯 16. REGRA FINAL

O frontend DEVE seguir o mockup fornecido.

Diferenças precisam ser justificadas.