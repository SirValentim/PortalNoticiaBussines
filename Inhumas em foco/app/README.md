# Inhumas em Foco - Protótipo React

Protótipo visual do portal Inhumas em Foco criado com Vite, React, TypeScript, Tailwind CSS e componentes UI.

Este app ainda usa dados mockados em `src/data/mock.ts`. O backend principal do projeto fica em `../inhumas-em-foco` e hoje é a implementação mais completa para produção.

## Comandos

```bash
npm install
npm run dev
npm run build
npm run preview
```

## Rotas Principais

- `/`
- `/noticia/:slug`
- `/categoria/:slug`
- `/lojas`
- `/loja/:slug`
- `/promocoes`
- `/promocao/:slug`
- `/bairro/:slug`
- `/busca`
- `/sobre`
- `/contato`
- `/login`
- `/painel/7x9k2m`

## Estado Atual

- Estrutura visual e navegação criadas.
- Dados ainda são locais/mockados.
- Login é apenas demonstrativo.
- A integração com o backend Go ainda precisa ser decidida.
