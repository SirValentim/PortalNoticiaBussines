# ADR 001: Monolito Modular Primeiro

## Decisao

Manter a plataforma como monolito modular em Go.

## Motivo

O produto ainda esta em fase de consolidacao e a equipe se beneficia de deploy simples, menor custo operacional e menor complexidade de observabilidade.

## Beneficios

- Deploy mais simples.
- Menos infraestrutura.
- Refatoracao mais direta.
- Compartilhamento facil entre modulos.

## Riscos

- Arquivos grandes podem concentrar responsabilidades.
- Modulos podem ficar acoplados se nao houver disciplina.

## Mitigacao

Criar services por dominio, DTOs e documentacao viva antes de qualquer divisao em microservicos.
