# Decisoes de Arquitetura

## ADR-001 - React fica como prototipo visual separado

Status: aceito para o MVP atual.

Contexto:

- O backend Go ja entrega rotas publicas, painel admin, templates server-side, auth, CSRF, RBAC, uploads, jobs e sitemap.
- O app React em `../app` e util como referencia visual e exploracao de UX, mas ainda usa mocks e login demonstrativo.
- Integrar React agora exigiria uma API publica/admin completa, duplicaria fluxos ja funcionais e aumentaria a superficie de seguranca antes do deploy inicial.

Decisao:

- O produto principal do MVP continua sendo o backend Go com templates server-side.
- O React permanece como prototipo visual documentado, sem ser caminho de producao.
- Novas funcionalidades funcionais devem entrar primeiro no backend Go.

Criterios para reabrir:

- Necessidade clara de experiencia SPA.
- API versionada pronta para conteudo publico e painel.
- Plano de auth/CSRF/CORS para frontend separado.
- Tempo reservado para substituir mocks por dados reais e cobrir build/deploy do frontend.

Impacto no checklist:

- "Decidir oficialmente se React sera integrado por API ou removido/mantido como prototipo" passa para concluido.
- "Decidir integracao com backend Go" passa para concluido como "nao integrar no MVP".
- "Trocar mocks por API real" permanece pendente apenas se a decisao futura for promover React a produto.

