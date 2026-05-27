# NewsCore CMS - Painel Administrativo

O painel administrativo usa a marca neutra **NewsCore CMS** para posicionar o produto como uma plataforma premium multiportal.

## Posicionamento

- Nome do sistema: NewsCore CMS.
- Descricao: CMS Premium Multiportal.
- Promessa administrativa: Gestao profissional de portais de noticias.
- Primeiro tenant: Inhumas em Foco.

O nome Inhumas em Foco deve aparecer no admin apenas como portal/tenant operacional, nunca como marca principal do painel.

## Areas Principais

- Dashboard executivo.
- Portais.
- Materias.
- Categorias.
- Tags.
- Midia.
- Usuarios.
- Metricas.
- Comercial.
- Automacao.
- Auditoria.
- Configuracoes.
- Perfil.

## Regras De Produto

- Manter SSR com templates HTML Go.
- Respeitar `tenant_id` em leituras e escritas.
- Respeitar RBAC em handlers, menu e acoes destrutivas.
- Exibir IDs amigaveis no admin com o formato `#001`.
- Usar dados reais nas metricas; quando nao houver tracking suficiente, mostrar estado vazio.
- Preferir soft delete ou desativacao segura para contas com historico, posts ou auditoria.
