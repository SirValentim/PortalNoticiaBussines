# PostgreSQL Validation

## Objetivo

Validar as migrations PostgreSQL em banco real sem tocar no schema publico.

## Como Rodar

Configure uma URL de PostgreSQL de teste:

```bash
POSTGRES_TEST_URL="postgres://user:pass@localhost:5432/inhumas_test?sslmode=disable" go test ./internal/repository -run TestPostgresMigrationsApplyWhenURLProvided -v
```

O teste cria um schema temporario com prefixo `codex_migration_test_`, executa as migrations `001` a `020` nesse schema e remove tudo ao final.

## Testes Locais Sempre Ativos

Mesmo sem PostgreSQL configurado, a suite valida que os arquivos de migration estao numerados em sequencia e sem versoes duplicadas:

```bash
go test ./internal/repository -run TestMigrationFilesAreContiguous -v
```
