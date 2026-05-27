# Guia Mestre — CODEX como Cérebro do Projeto

## Objetivo

Transformar o projeto em uma plataforma organizada, reutilizável e escalável para múltiplos portais e produtos.

Exemplos:

- Inhumas em Foco
- LaMafiaMusic
- GoiasNews
- futuros portais
- futuros SaaS

A ideia é criar:

- um núcleo reutilizável
- uma arquitetura padrão
- documentação viva
- contexto persistente para IA
- workflow AI-first
- um “cérebro do projeto” para o Codex

---

# Filosofia do Sistema

O projeto NÃO é mais um site.

Ele passa a ser:

> Uma plataforma editorial/CMS modular baseada em Golang.

Cada novo portal será apenas:

- tema
- branding
- conteúdo
- configurações
- módulos habilitados

A arquitetura principal será compartilhada.

---

# Objetivo Final

Ter:

```txt
Core CMS Platform
    ↓
Inhumas em Foco
LaMafiaMusic
GoiasNews
Outro Portal
```

---

# Stack Recomendada

## Backend

- Golang
- PostgreSQL
- Redis
- HTMX ou React
- Tailwind

## Infra

- Docker
- Docker Compose
- Traefik ou Nginx

## Documentação

- Markdown
- Obsidian
- Mermaid
- Docusaurus futuramente

## IA

- Codex
- ChatGPT
- Cursor opcional

---

# Estrutura Ideal do Projeto

```txt
/project-root
    /cmd
    /internal
    /pkg
    /web
    /configs
    /deployments
    /scripts
    /docs
    /tests
    /storage
```

---

# Estrutura do Internal

```txt
/internal
    /auth
    /billing
    /cms
    /media
    /seo
    /analytics
    /tenant
    /users
    /ai
    /notifications
```

Cada módulo:

```txt
/module
    handler.go
    service.go
    repository.go
    model.go
    dto.go
    routes.go
    validator.go
```

---

# Estrutura do Cérebro do Projeto

```txt
/docs
    /ai-context
    /architecture
    /backend
    /database
    /product
    /roadmap
    /decisions
    /deprecated
```

---

# A Parte Mais Importante

# O Cérebro do Projeto

O Codex precisa entender:

- o produto
- a arquitetura
- as regras
- o padrão de código
- o estado atual
- prioridades
- convenções
- visão futura

Sem isso:

- a IA alucina
- duplica código
- cria arquitetura conflitante
- quebra padrões

---

# PASSO 1 — Criar o AI Context

Crie:

```txt
/docs/ai-context
```

---

# PASSO 2 — Criar os Arquivos Mestres

## project-overview.md

```md
# Projeto

Plataforma CMS Editorial baseada em Golang.

## Objetivo

Criar uma plataforma reutilizável para múltiplos portais:

- Inhumas em Foco
- LaMafiaMusic
- GoiasNews

## Stack

- Golang
- PostgreSQL
- Redis
- HTMX
- Tailwind

## Arquitetura

- Monolito modular
- Multi-tenant preparado
- API REST
- CMS Premium
- Painel administrativo

## Padrões

- Sem lógica no handler
- Service Layer obrigatória
- Repository Pattern
- DTO obrigatório
- Middleware apenas para cross-cutting

## Objetivo Atual

Estruturar o Core CMS reutilizável.
```

---

# PASSO 3 — Criar ENGINEERING.md

```md
# Engineering Rules

## Regras Gerais

- Nunca acessar banco diretamente no handler
- Services devem ser independentes de HTTP
- Toda entrada deve validar DTO
- Evitar lógica duplicada
- Priorizar composição ao invés de herança
- Todo módulo deve ser isolado
- Nenhum módulo deve depender diretamente de outro

## Estrutura

- Handler
- Service
- Repository
- DTO
- Model

## Banco

- Toda migration deve ser versionada
- Nunca alterar migration antiga

## Naming

- User
- Tenant
- Organization
- Article
- Media

Nunca misturar nomenclaturas.
```

---

# PASSO 4 — Criar Current State

Arquivo:

```txt
/docs/ai-context/current-state.md
```

Exemplo:

```md
# Estado Atual

## Já Implementado

- Sistema de notícias
- CMS básico
- Upload de mídia
- SEO básico
- Painel administrativo

## Em Desenvolvimento

- Billing
- Multi-tenant
- ACL
- Sistema premium

## Problemas atuais

- Estrutura antiga misturada
- Código legado
- Docs duplicadas
- Falta modularização
```

---

# PASSO 5 — Criar Architecture Summary

Arquivo:

```txt
/docs/ai-context/architecture-summary.md
```

Exemplo:

```md
# Arquitetura

## Modelo

Monolito modular.

## Fluxo

HTTP -> Handler -> Service -> Repository -> Database

## Frontend

HTMX + Tailwind.

## Banco

PostgreSQL.

## Cache

Redis.

## Objetivo futuro

Separar módulos críticos em microserviços apenas se necessário.
```

---

# PASSO 6 — Criar Naming Conventions

Arquivo:

```txt
/docs/ai-context/naming-conventions.md
```

Exemplo:

```md
# Naming Conventions

## Entidades

Article
Media
User
Tenant
Organization
Subscription
Invoice

## APIs

/api/v1/articles
/api/v1/users

## Services

ArticleService
UserService
BillingService

## Repository

ArticleRepository
UserRepository
```

---

# PASSO 7 — Criar ADRs

```txt
/docs/decisions
```

Exemplo:

```txt
001-monolith-first.md
002-use-postgresql.md
003-use-htmx.md
```

---

# Exemplo de ADR

```md
# ADR 001

## Decisão

Usar monolito modular.

## Motivo

Equipe pequena.

## Benefícios

- Deploy simples
- Menor custo
- Menor complexidade

## Riscos

- Crescimento futuro
```

---

# PASSO 8 — Criar Roadmap

```txt
/docs/roadmap
```

Exemplo:

```md
# Q2 2026

## Alta Prioridade

- Billing
- ACL
- Multi-tenant

## Média

- Newsletter
- Analytics avançado

## Baixa

- Mobile app
```

---

# PASSO 9 — Criar Packs para IA

```txt
/docs/ai-context/packs
```

Exemplos:

```txt
billing-pack.md
cms-pack.md
media-pack.md
seo-pack.md
```

Cada pack deve conter:

- arquitetura daquele domínio
- tabelas
- endpoints
- regras
- problemas
- roadmap

---

# PASSO 10 — Criar Mapa do Domínio

Arquivo:

```txt
/docs/architecture/domain-map.md
```

Exemplo:

```txt
Tenant
  ↓
Portal
  ↓
Users
  ↓
Articles
  ↓
Media
  ↓
SEO
  ↓
Analytics
  ↓
Subscriptions
```

---

# PASSO 11 — Organizar o Monorepo Mental

Separar:

## CORE

Código compartilhado.

## PORTAL

Customização específica.

---

# Estrutura Ideal

```txt
/internal/core
/internal/portals
```

---

# Core

Tudo reutilizável:

```txt
/internal/core
    /auth
    /billing
    /cms
    /media
    /seo
```

---

# Portais

```txt
/internal/portals
    /inhumasemfoco
    /lamafiamusic
    /goiasnews
```

Cada portal terá:

- config
- tema
- branding
- regras locais

---

# PASSO 12 — Sistema Multi-Tenant Futuro

O Codex deve entender:

- um core
- vários tenants
- isolamento lógico
- branding separado

---

# Modelo Conceitual

```txt
Platform
    ↓
Tenant
    ↓
Portal
    ↓
Content
```

---

# PASSO 13 — Criar Contexto Permanente para o Codex

Arquivo:

```txt
.codex-context.md
```

Esse arquivo é CRÍTICO.

---

# Exemplo

```md
# Projeto

Plataforma CMS modular em Golang.

## Objetivo

Criar um core reutilizável para múltiplos portais.

## Arquitetura

- Monolito modular
- Repository Pattern
- Service Layer
- DTO

## Regras

- Sem lógica no handler
- Services independentes de HTTP
- Código reutilizável
- Multi-tenant ready

## Estrutura

/internal/core
/internal/portals
/docs

## Prioridade Atual

Organizar arquitetura e remover legado.
```

---

# PROMPT MESTRE PARA O CODEX

# Objetivo

Usar esse prompt para fazer o Codex estruturar o cérebro do projeto.

---

# PROMPT

```txt
Você é um arquiteto senior de software especializado em:

- Golang
- arquitetura SaaS
- CMS enterprise
- monolito modular
- sistemas multi-tenant
- engenharia AI-first

Sua missão é transformar este projeto em uma plataforma reutilizável.

Objetivos:

1. Organizar a arquitetura do projeto
2. Separar core reutilizável de portais específicos
3. Estruturar documentação viva
4. Criar contexto persistente para IA
5. Melhorar escalabilidade
6. Reduzir código legado
7. Criar padrão reutilizável para múltiplos portais

Portais atuais:

- Inhumas em Foco
- LaMafiaMusic
- GoiasNews

Stack:

- Golang
- PostgreSQL
- Redis
- HTMX
- Tailwind

Regras obrigatórias:

- Sem lógica no handler
- Service Layer obrigatória
- Repository Pattern
- DTO obrigatório
- Código modular
- Multi-tenant ready
- Reutilização máxima

Sua tarefa inicial:

1. Mapear estrutura atual
2. Identificar legado
3. Propor modularização
4. Criar organização de docs
5. Criar AI Context
6. Criar roadmap técnico
7. Criar ADRs
8. Criar estrutura core/portal
9. Criar naming conventions
10. Criar estratégia escalável

Sempre:

- explicar decisões
- evitar complexidade desnecessária
- priorizar simplicidade
- pensar em longo prazo
- gerar código limpo
- gerar documentação clara
- manter consistência arquitetural
```

---

# PROMPT PARA ORGANIZAR O PROJETO EXISTENTE

```txt
Analise a estrutura atual do projeto e:

1. Identifique módulos
2. Identifique acoplamentos ruins
3. Identifique duplicações
4. Identifique legado
5. Sugira modularização
6. Sugira separação core/portal
7. Sugira melhorias arquiteturais
8. Crie roadmap de refatoração

Gere:

- árvore ideal de diretórios
- plano de migração
- lista de prioridades
- riscos
- quick wins
```

---

# PROMPT PARA GERAR DOCUMENTAÇÃO AUTOMÁTICA

```txt
Com base no código atual:

1. Gere documentação técnica
2. Gere arquitetura do módulo
3. Gere fluxos
4. Gere tabelas
5. Gere endpoints
6. Gere regras de negócio
7. Gere dependências
8. Gere riscos técnicos

Salvar em:

/docs
```

---

# PROMPT PARA O CODEX CRIAR O PRÓPRIO CÉREBRO

```txt
Crie um sistema de contexto persistente para IA.

Objetivo:

Permitir que futuras IAs entendam rapidamente:

- arquitetura
- regras
- domínio
- módulos
- objetivos
- roadmap
- padrões
- estado atual

Crie:

/docs/ai-context

Arquivos:

- project-overview.md
- current-state.md
- coding-rules.md
- architecture-summary.md
- naming-conventions.md
- roadmap-summary.md
- module-map.md
- technical-debt.md

O conteúdo deve ser:

- claro
- técnico
- objetivo
- reutilizável
- otimizado para IA
```

---

# PROMPT PARA MULTI-PROJETOS

```txt
Estruture o sistema para suportar múltiplos portais.

Separar:

- Core compartilhado
- Branding
- Configurações
- Features específicas
- Temas
- Assets

Objetivo:

Permitir criar novos portais rapidamente usando o mesmo núcleo.
```

---

# Estrutura Final Esperada

```txt
/internal
    /core
    /portals

/docs
    /ai-context
    /architecture
    /backend
    /roadmap
    /decisions
```

---

# Resultado Esperado

Ao final:

- projeto organizado
- IA entendendo o sistema
- documentação viva
- arquitetura consistente
- escalabilidade
- reutilização
- onboarding fácil
- menor caos estrutural
- base pronta para crescer

---

# Visão Final

Você não está criando apenas um portal.

Está criando:

> Uma plataforma editorial inteligente reutilizável baseada em Golang.

Esse é o mindset correto para o próximo estágio do projeto.
