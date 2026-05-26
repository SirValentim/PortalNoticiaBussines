# Project Overview

## Produto

CMS editorial modular em Golang para operacao de portais locais, editoriais e comerciais. O primeiro tenant e o Inhumas em Foco; a arquitetura deve permitir novos portais com o mesmo binario, alterando configuracoes, branding, assets e conteudo.

## Portais Alvo

- Inhumas em Foco
- LaMafiaMusic
- GoiasNews
- futuros portais locais ou SaaS editoriais

## Stack Atual

- Go
- Templates HTML server-side
- HTMX leve
- SQLite em runtime inicial
- PostgreSQL suportado por driver/migrations
- Storage local com processamento de imagem
- MCP para automacoes editoriais controladas

## Stack Planejada

- PostgreSQL como banco principal de producao madura
- Redis para cache, filas leves ou rate limit distribuido quando necessario
- Nginx ou Traefik na borda
- Docker Compose para ambientes reproduziveis

## Direcao Arquitetural

- Monolito modular primeiro.
- Multi-tenant preparado por configuracao e branding.
- Separacao gradual entre core reutilizavel e customizacoes de portal.
- Services independentes de HTTP para regras de negocio.
- Repository pattern para acesso a dados.

## Objetivo Atual

Criar documentacao viva e contexto persistente para futuras iteracoes de IA e engenharia, reduzindo risco de duplicacao, mudancas conflitantes e refatoracoes grandes demais.
