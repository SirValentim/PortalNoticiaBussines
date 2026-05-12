PROMPT REFINADO — NÍVEL ABSURDO (v3.0)

Adicione isso abaixo do seu prompt atual:

🎯 19. MASTER LAYOUT GRID (BASE DO SITE)
Layout global desktop:

- container: max-width 1200px
- margin: 0 auto
- padding horizontal: 24px

SEÇÕES:
- espaçamento entre seções: 32px
- título de seção:
  - font-size: 20px
  - font-weight: 600
  - margin-bottom: 16px

REGRA:
→ Todas as seções devem alinhar exatamente no mesmo eixo vertical
🔥 20. HERO + SIDEBAR (PIXEL PERFECT)
GRID:

- display: grid
- grid-template-columns: 2fr 1fr
- gap: 24px

HERO:

- altura: 420px
- border-radius: 12px
- overflow: hidden
- position: relative

IMAGEM:
- width: 100%
- height: 100%
- object-fit: cover

OVERLAY:
- background: linear-gradient(
  to top,
  rgba(0,0,0,0.7),
  rgba(0,0,0,0.2)
)

CONTEÚDO:
- position: absolute
- bottom: 24px
- left: 24px
- right: 24px

TÍTULO:
- font-size: 34px
- line-height: 1.2
- font-weight: 700
- color: white
- max-width: 80%

DESCRIÇÃO:
- font-size: 14px
- color: rgba(255,255,255,0.9)
- margin-top: 8px

BOTÃO:
- margin-top: 12px
- padding: 10px 18px
- border-radius: 8px
🧱 SIDEBAR (CORREÇÃO VISUAL FORTE)
SIDEBAR:

- background: white
- border-radius: 12px
- padding: 16px
- height: 420px (igual ao hero)
- display: flex
- flex-direction: column
- gap: 12px

ITEM:

- border-bottom: 1px solid var(--border)
- padding-bottom: 8px

CATEGORIA:
- font-size: 10px
- color: var(--primary)
- font-weight: 600

TÍTULO:
- font-size: 13px
- font-weight: 500
📰 21. NEWS GRID (AGORA CORRETO)
GRID:

- desktop: 4 colunas
- gap: 16px

CARD:

- background: white
- border-radius: 10px
- overflow: hidden

IMAGEM:
- width: 100%
- height: 120px
- object-fit: cover

CONTEÚDO:
- padding: 10px

CATEGORIA:
- font-size: 10px
- margin-bottom: 4px

TÍTULO:
- font-size: 14px
- font-weight: 600
- line-height: 1.3
- max-height: 2.6em
- overflow: hidden
📢 22. BANNER (AJUSTE PREMIUM)
BANNER:

- height: 100px
- border-radius: 12px
- padding: 16px 20px

LAYOUT:
- display: flex
- justify-content: space-between
- align-items: center

TEXTO:
- font-size: 16px
- font-weight: 600

BOTÃO:
- background: var(--accent)
- color: black
- padding: 10px 16px
- border-radius: 8px

LABEL:
- position: absolute
- top: 8px
- left: 8px
- font-size: 10px
🧠 23. VISUAL DENSITY CONTROL
OBJETIVO:

A tela deve mostrar MAIS conteúdo sem parecer poluída.

REGRAS:

- reduzir alturas excessivas
- evitar espaços grandes vazios
- cards mais compactos
- manter consistência

Se parecer “espaçado demais”:
→ reduzir padding em 20%
🎯 24. MICRO INTERAÇÕES (NÍVEL PRODUTO)
HOVER:

Cards:
- transform: translateY(-2px)
- transition: 0.2s

Imagens:
- scale: 1.03 no hover

Botões:
- opacity: 0.9 no hover
📱 25. MOBILE (CORREÇÃO REAL)
HERO:
- altura: 250px
- título: 20px

GRID:
- 1 coluna

SIDEBAR:
- vira lista abaixo do hero

NEWS:
- 1 coluna
- imagem: 100% width

BANNER:
- stack vertical
🧪 26. UI POLISH CHECKLIST (OBRIGATÓRIO)
ANTES DE FINALIZAR:

✔ Tudo alinhado verticalmente?
✔ Espaçamentos consistentes?
✔ Nenhum elemento “sobrando”?
✔ Visual bate com mockup?

SE NÃO:
→ ajustar automaticamente