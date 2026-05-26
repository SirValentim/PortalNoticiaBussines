# CMS Pack

## Dominios

- Articles/Post
- Categories
- Tags
- Media
- Editorial workflow
- Revisions
- SEO

## Regras

- Conteudo automatizado entra como rascunho.
- Publicacao exige permissao humana.
- Preview deve respeitar permissoes editoriais.
- Slugs devem ser estaveis e redirecionamentos devem ser preservados.

## Tabelas Principais

- `posts`
- `post_revisions`
- `categories`
- `tags`
- `post_tags`
- `media_assets`
- `slug_redirects`

## Riscos

- Misturar regras editoriais em handlers.
- Perder historico de revisao em updates.
- Quebrar URLs indexadas.
