# ADR 004: Multi-Tenant Incremental

## Decisao

Evoluir multiportal em etapas: primeiro branding/configuracao por tenant, depois tenants persistentes, depois isolamento completo de dados.

## Motivo

O sistema ja esta em operacao como portal unico. Uma migracao multi-tenant ampla de uma vez aumentaria risco de regressao.

## Beneficios

- Permite validar segundo portal rapidamente.
- Reduz risco operacional.
- Mantem o core atual funcionando.
- Cria caminho claro para SaaS.

## Riscos

- Durante a transicao, parte do sistema ainda sera single-tenant.
- Exige disciplina para nao adicionar novos hardcodes.

## Mitigacao

Criar testes com branding alternativo, documentar pontos single-tenant e planejar migrations de `tenant_id` por dominio.
