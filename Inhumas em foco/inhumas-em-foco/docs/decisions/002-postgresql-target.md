# ADR 002: PostgreSQL Como Banco Alvo

## Decisao

Usar PostgreSQL como banco principal alvo para producao madura, mantendo compatibilidade com SQLite durante desenvolvimento e operacao inicial.

## Motivo

PostgreSQL oferece confiabilidade, indices, concorrencia e recursos adequados para multi-tenant e SaaS.

## Beneficios

- Melhor escalabilidade.
- Suporte forte a constraints e indices.
- Ecossistema maduro.
- Melhor base para isolamento multi-tenant.

## Riscos

- Divergencia entre SQLite e PostgreSQL.
- Necessidade de homologacao de migrations e queries nos dois dialetos.

## Mitigacao

Testar queries criticas, manter dialeto explicito e evitar SQL que funcione apenas em um banco sem cobertura.
