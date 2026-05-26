# Coding Rules

## Regras Gerais

- Ler o codigo existente antes de alterar arquitetura.
- Fazer mudancas pequenas, testaveis e reversiveis.
- Manter handlers focados em HTTP, auth, parse de entrada e render/redirect.
- Mover regra de negocio para services quando tocar em comportamento relevante.
- Usar repository para acesso ao banco.
- Evitar dependencias circulares entre dominios.
- Validar entrada antes de chamar services.
- Nunca alterar migration antiga; criar nova migration versionada.
- Nao introduzir novo framework sem decisao registrada.

## Handlers

Handlers podem:

- validar metodo/rota;
- ler path/query/form;
- chamar services/repository existente;
- montar dados de template;
- retornar HTML, redirect, JSON ou erro.

Handlers nao devem ganhar regras novas complexas se houver como criar service.

## Services

Services devem:

- ser independentes de HTTP;
- receber tipos simples, DTOs ou models;
- concentrar invariantes de negocio;
- ser testaveis com repository fake ou banco em memoria.

## Repository

Repository deve:

- encapsular SQL;
- manter compatibilidade SQLite/PostgreSQL quando aplicavel;
- nao formatar HTML;
- nao conhecer templates ou requests.

## Templates

- Usar `Branding` para identidade do portal.
- Usar `PortalSettings` apenas para configuracoes editoriais persistidas pelo painel.
- Evitar texto hardcoded com nome/domino/contato de um portal especifico.
