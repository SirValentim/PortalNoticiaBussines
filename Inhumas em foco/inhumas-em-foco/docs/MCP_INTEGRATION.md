# Integracao MCP

O projeto inclui um servidor MCP em `cmd/mcp` para expor operacoes editoriais controladas a clientes compativeis, como Claude Desktop ou inspetores MCP.

O servidor roda via stdio, usa o mesmo banco configurado pelo CMS e nunca publica conteudo automaticamente ao criar um novo post. A ferramenta `create_draft` sempre gera rascunho com status `draft`, mantendo revisao humana obrigatoria.

## Ferramentas disponiveis

- `list_articles`: lista posts, com filtro opcional por status.
- `get_article`: retorna o conteudo completo de um post pelo ID.
- `create_draft`: cria rascunho editorial.
- `update_article_status`: altera status editorial de um post.
- `list_categories`: lista categorias ativas.
- `search_articles`: pesquisa posts por titulo, resumo ou conteudo.
- `get_portal_stats`: retorna estatisticas editoriais gerais.

## Build local

```powershell
cd "Inhumas em foco\inhumas-em-foco"
go build -o bin/inhumas-mcp.exe ./cmd/mcp
```

## Build Linux para VPS

```bash
cd "Inhumas em foco/inhumas-em-foco"
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/inhumas-mcp ./cmd/mcp
```

Tambem existem targets no Makefile:

```bash
make mcp-build-linux
make mcp-build-windows
make mcp-dev
```

## Exemplo Claude Desktop no Windows

Arquivo:

```text
%APPDATA%\Claude\claude_desktop_config.json
```

Exemplo:

```json
{
  "mcpServers": {
    "inhumas-em-foco": {
      "command": "C:\\Users\\Valentim\\Documents\\New project\\Inhumas em foco\\inhumas-em-foco\\bin\\inhumas-mcp.exe",
      "env": {
        "DATABASE_URL": "C:\\Users\\Valentim\\Documents\\New project\\Inhumas em foco\\inhumas-em-foco\\inhumas.db",
        "DB_DRIVER": "sqlite",
        "PROJECT_ROOT": "C:\\Users\\Valentim\\Documents\\New project\\Inhumas em foco\\inhumas-em-foco",
        "SESSION_SECRET": "change-this-to-at-least-32-random-bytes"
      }
    }
  }
}
```

## Exemplo em producao

```json
{
  "mcpServers": {
    "inhumas-em-foco": {
      "command": "/var/www/inhumas/bin/inhumas-mcp",
      "env": {
        "DATABASE_URL": "/var/www/inhumas/inhumas.db",
        "DB_DRIVER": "sqlite",
        "PROJECT_ROOT": "/var/www/inhumas",
        "MIGRATIONS_DIR": "/var/www/inhumas/migrations",
        "SESSION_SECRET": "change-this-to-at-least-32-random-bytes"
      }
    }
  }
}
```

## Inspecao

Com o inspetor MCP instalado:

```bash
npx @modelcontextprotocol/inspector ./bin/inhumas-mcp
```

Ou pelo Makefile:

```bash
make mcp-inspect
```

## Cuidados

- Nao use o MCP com credenciais de producao em maquinas compartilhadas.
- O binario deve rodar com acesso apenas ao banco e aos arquivos necessarios.
- A publicacao final de conteudo deve continuar sendo feita por usuario autorizado no painel.
- Em producao, prefira caminhos absolutos nas variaveis `DATABASE_URL`, `PROJECT_ROOT` e `MIGRATIONS_DIR`.
