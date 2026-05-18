package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"inhumas-em-foco/internal/config"
	"inhumas-em-foco/internal/model"
	"inhumas-em-foco/internal/repository"
	"inhumas-em-foco/internal/slug"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type mcpApp struct {
	repo   *repository.Repository
	db     *sql.DB
	driver string
}

type articleDTO struct {
	ID          int64      `json:"id"`
	Title       string     `json:"title"`
	Slug        string     `json:"slug"`
	Excerpt     string     `json:"excerpt"`
	Content     string     `json:"content,omitempty"`
	Category    string     `json:"category"`
	Tags        string     `json:"tags"`
	Status      string     `json:"status"`
	AuthorName  string     `json:"author_name"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type categoryDTO struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

func main() {
	cfg := config.Load()
	repo, err := repository.Open(cfg.DBDriver, cfg.DatabaseURL, cfg.MigrationsDir)
	if err != nil {
		log.Fatalf("mcp: abrir repository: %v", err)
	}
	defer repo.DB().Close()

	app := &mcpApp{
		repo:   repo,
		db:     repo.DB(),
		driver: repo.Driver(),
	}

	s := server.NewMCPServer(
		"inhumas-em-foco-mcp",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	s.AddTool(mcp.NewTool("list_articles",
		mcp.WithDescription("Lista posts do portal. Pode filtrar por status: draft, review, approved, scheduled, published ou archived."),
		mcp.WithString("status", mcp.Description("Filtro opcional de status.")),
		mcp.WithNumber("limit", mcp.Description("Maximo de resultados. Padrao 10, maximo 50.")),
	), app.handleListArticles)

	s.AddTool(mcp.NewTool("get_article",
		mcp.WithDescription("Retorna o conteudo completo de um post pelo ID."),
		mcp.WithNumber("id", mcp.Required(), mcp.Description("ID do post.")),
	), app.handleGetArticle)

	s.AddTool(mcp.NewTool("create_draft",
		mcp.WithDescription("Cria um rascunho editorial. Nunca publica automaticamente; o conteudo fica como draft para revisao humana."),
		mcp.WithString("title", mcp.Required(), mcp.Description("Titulo do post.")),
		mcp.WithString("excerpt", mcp.Description("Resumo ou subtitulo.")),
		mcp.WithString("content", mcp.Description("Corpo completo em HTML ou texto.")),
		mcp.WithString("tags", mcp.Description("Tags separadas por virgula.")),
		mcp.WithNumber("category_id", mcp.Description("ID da categoria.")),
	), app.handleCreateDraft)

	s.AddTool(mcp.NewTool("update_article_status",
		mcp.WithDescription("Atualiza o status editorial de um post existente."),
		mcp.WithNumber("id", mcp.Required(), mcp.Description("ID do post.")),
		mcp.WithString("status", mcp.Required(), mcp.Description("Novo status: draft, review, approved, scheduled, published ou archived.")),
	), app.handleUpdateStatus)

	s.AddTool(mcp.NewTool("list_categories",
		mcp.WithDescription("Lista categorias ativas disponiveis no portal."),
	), app.handleListCategories)

	s.AddTool(mcp.NewTool("search_articles",
		mcp.WithDescription("Pesquisa posts por palavra-chave em titulo, resumo ou conteudo."),
		mcp.WithString("query", mcp.Required(), mcp.Description("Texto para pesquisar.")),
	), app.handleSearchArticles)

	s.AddTool(mcp.NewTool("get_portal_stats",
		mcp.WithDescription("Retorna metricas editoriais gerais do portal."),
	), app.handleGetStats)

	log.Printf("mcp: iniciado via stdio usando driver=%s database=%s", repo.Driver(), cfg.DatabaseURL)
	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("mcp: server erro: %v", err)
	}
}

func (a *mcpApp) handleListArticles(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	status := strings.TrimSpace(req.GetString("status", ""))
	limit := req.GetInt("limit", 10)
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	if status != "" && !validStatus(status) {
		return errResult("Status invalido. Use: draft, review, approved, scheduled, published ou archived"), nil
	}

	args := []any{}
	where := ""
	if status != "" {
		args = append(args, status)
		where = " WHERE p.status = " + a.placeholder(len(args))
	}
	args = append(args, limit)

	query := fmt.Sprintf(`
		SELECT p.id, p.title, p.slug, COALESCE(p.excerpt, ''), p.status,
		       COALESCE(c.name, ''), COALESCE(u.name, ''), %s,
		       p.published_at, p.created_at, p.updated_at
		FROM posts p
		LEFT JOIN categories c ON c.id = p.category_id
		LEFT JOIN users u ON u.id = p.author_id
		%s
		ORDER BY p.created_at DESC
		LIMIT %s`, a.tagsExpr("p.id"), where, a.placeholder(len(args)))

	rows, err := a.db.QueryContext(ctx, query, args...)
	if err != nil {
		return errResult("Erro ao consultar posts: " + err.Error()), nil
	}
	defer rows.Close()

	articles, err := scanArticles(rows, false)
	if err != nil {
		return errResult("Erro ao ler posts: " + err.Error()), nil
	}
	return textResultJSON(articles), nil
}

func (a *mcpApp) handleGetArticle(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireInt("id")
	if err != nil || id <= 0 {
		return errResult("Parametro 'id' invalido ou ausente"), nil
	}

	query := fmt.Sprintf(`
		SELECT p.id, p.title, p.slug, COALESCE(p.excerpt, ''), COALESCE(p.content, ''),
		       p.status, COALESCE(c.name, ''), COALESCE(u.name, ''), %s,
		       p.published_at, p.created_at, p.updated_at
		FROM posts p
		LEFT JOIN categories c ON c.id = p.category_id
		LEFT JOIN users u ON u.id = p.author_id
		WHERE p.id = %s`, a.tagsExpr("p.id"), a.placeholder(1))

	var article articleDTO
	var publishedAt sql.NullTime
	err = a.db.QueryRowContext(ctx, query, id).Scan(
		&article.ID,
		&article.Title,
		&article.Slug,
		&article.Excerpt,
		&article.Content,
		&article.Status,
		&article.Category,
		&article.AuthorName,
		&article.Tags,
		&publishedAt,
		&article.CreatedAt,
		&article.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return errResult(fmt.Sprintf("Post nao encontrado: id=%d", id)), nil
	}
	if err != nil {
		return errResult("Erro ao buscar post: " + err.Error()), nil
	}
	if publishedAt.Valid {
		article.PublishedAt = &publishedAt.Time
	}
	return textResultJSON(article), nil
}

func (a *mcpApp) handleCreateDraft(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	title := strings.TrimSpace(req.GetString("title", ""))
	excerpt := strings.TrimSpace(req.GetString("excerpt", ""))
	content := strings.TrimSpace(req.GetString("content", ""))
	tags := strings.TrimSpace(req.GetString("tags", ""))
	categoryID := req.GetInt("category_id", 0)
	if title == "" {
		return errResult("Campo 'title' e obrigatorio"), nil
	}
	if content == "" {
		content = excerpt
	}

	tx, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		return errResult("Erro ao iniciar transacao: " + err.Error()), nil
	}
	defer tx.Rollback()

	postSlug := slug.Unique(title, func(candidate string) bool {
		var count int
		_ = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM posts WHERE slug = `+a.placeholder(1), candidate).Scan(&count)
		return count > 0
	})

	var categoryArg any
	if categoryID > 0 {
		categoryArg = categoryID
	}

	postID, err := a.insertPost(ctx, tx, title, postSlug, excerpt, content, categoryArg)
	if err != nil {
		return errResult("Erro ao criar rascunho: " + err.Error()), nil
	}
	if err := a.attachTags(ctx, tx, postID, tags); err != nil {
		return errResult("Erro ao associar tags: " + err.Error()), nil
	}
	if err := tx.Commit(); err != nil {
		return errResult("Erro ao confirmar rascunho: " + err.Error()), nil
	}

	return textResultJSON(map[string]any{
		"id":      postID,
		"slug":    postSlug,
		"status":  string(model.StatusDraft),
		"message": "Rascunho criado com sucesso",
	}), nil
}

func (a *mcpApp) handleUpdateStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireInt("id")
	if err != nil || id <= 0 {
		return errResult("Parametro 'id' invalido"), nil
	}
	status := strings.TrimSpace(req.GetString("status", ""))
	if !validStatus(status) {
		return errResult("Status invalido. Use: draft, review, approved, scheduled, published ou archived"), nil
	}

	setPublished := ""
	if status == string(model.StatusPublished) {
		setPublished = ", published_at = CURRENT_TIMESTAMP"
	}
	query := fmt.Sprintf(`UPDATE posts SET status = %s, updated_at = CURRENT_TIMESTAMP%s WHERE id = %s`,
		a.placeholder(1), setPublished, a.placeholder(2))
	res, err := a.db.ExecContext(ctx, query, status, id)
	if err != nil {
		return errResult("Erro ao atualizar status: " + err.Error()), nil
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return errResult(fmt.Sprintf("Post nao encontrado: id=%d", id)), nil
	}
	return textResultJSON(map[string]any{"id": id, "status": status, "message": "Status atualizado"}), nil
}

func (a *mcpApp) handleListCategories(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	rows, err := a.db.QueryContext(ctx, `SELECT id, name, slug FROM categories WHERE active = true ORDER BY sort_order, name`)
	if err != nil {
		return errResult("Erro ao listar categorias: " + err.Error()), nil
	}
	defer rows.Close()

	var categories []categoryDTO
	for rows.Next() {
		var c categoryDTO
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug); err != nil {
			return errResult("Erro ao ler categoria: " + err.Error()), nil
		}
		categories = append(categories, c)
	}
	return textResultJSON(categories), nil
}

func (a *mcpApp) handleSearchArticles(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	queryText := strings.TrimSpace(req.GetString("query", ""))
	if queryText == "" {
		return errResult("Parametro 'query' e obrigatorio"), nil
	}
	pattern := "%" + strings.ToLower(queryText) + "%"
	args := []any{pattern, pattern, pattern}

	query := fmt.Sprintf(`
		SELECT p.id, p.title, p.slug, COALESCE(p.excerpt, ''), p.status,
		       COALESCE(c.name, ''), COALESCE(u.name, ''), %s,
		       p.published_at, p.created_at, p.updated_at
		FROM posts p
		LEFT JOIN categories c ON c.id = p.category_id
		LEFT JOIN users u ON u.id = p.author_id
		WHERE lower(p.title) LIKE %s OR lower(COALESCE(p.excerpt, '')) LIKE %s OR lower(COALESCE(p.content, '')) LIKE %s
		ORDER BY p.created_at DESC
		LIMIT 20`, a.tagsExpr("p.id"), a.placeholder(1), a.placeholder(2), a.placeholder(3))

	rows, err := a.db.QueryContext(ctx, query, args...)
	if err != nil {
		return errResult("Erro na busca: " + err.Error()), nil
	}
	defer rows.Close()

	articles, err := scanArticles(rows, false)
	if err != nil {
		return errResult("Erro ao ler resultados: " + err.Error()), nil
	}
	return textResultJSON(articles), nil
}

func (a *mcpApp) handleGetStats(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	type stats struct {
		TotalArticles     int `json:"total_articles"`
		Published         int `json:"published"`
		Drafts            int `json:"drafts"`
		InReview          int `json:"in_review"`
		Scheduled         int `json:"scheduled"`
		Archived          int `json:"archived"`
		TotalCategories   int `json:"total_categories"`
		PublishedThisWeek int `json:"published_this_week"`
	}

	var s stats
	_ = a.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM posts`).Scan(&s.TotalArticles)
	_ = a.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM posts WHERE status = 'published'`).Scan(&s.Published)
	_ = a.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM posts WHERE status = 'draft'`).Scan(&s.Drafts)
	_ = a.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM posts WHERE status = 'review'`).Scan(&s.InReview)
	_ = a.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM posts WHERE status = 'scheduled'`).Scan(&s.Scheduled)
	_ = a.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM posts WHERE status = 'archived'`).Scan(&s.Archived)
	_ = a.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM categories WHERE active = true`).Scan(&s.TotalCategories)
	_ = a.db.QueryRowContext(ctx, a.publishedThisWeekQuery()).Scan(&s.PublishedThisWeek)

	return textResultJSON(s), nil
}

func (a *mcpApp) insertPost(ctx context.Context, tx *sql.Tx, title, postSlug, excerpt, content string, categoryID any) (int64, error) {
	query := fmt.Sprintf(`
		INSERT INTO posts (title, slug, excerpt, content, category_id, status, created_at, updated_at)
		VALUES (%s, %s, %s, %s, %s, 'draft', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		a.placeholder(1), a.placeholder(2), a.placeholder(3), a.placeholder(4), a.placeholder(5))

	if a.driver == "postgres" {
		var id int64
		err := tx.QueryRowContext(ctx, query+" RETURNING id", title, postSlug, excerpt, content, categoryID).Scan(&id)
		return id, err
	}
	res, err := tx.ExecContext(ctx, query, title, postSlug, excerpt, content, categoryID)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (a *mcpApp) attachTags(ctx context.Context, tx *sql.Tx, postID int64, rawTags string) error {
	for _, tagName := range splitTags(rawTags) {
		tagSlug := slug.Generate(tagName)
		tagID, err := a.ensureTag(ctx, tx, tagName, tagSlug)
		if err != nil {
			return err
		}
		query := fmt.Sprintf(`INSERT INTO post_tags (post_id, tag_id) VALUES (%s, %s)`, a.placeholder(1), a.placeholder(2))
		if a.driver == "postgres" {
			query = `INSERT INTO post_tags (post_id, tag_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
		} else {
			query = `INSERT OR IGNORE INTO post_tags (post_id, tag_id) VALUES (?, ?)`
		}
		if _, err := tx.ExecContext(ctx, query, postID, tagID); err != nil {
			return err
		}
	}
	return nil
}

func (a *mcpApp) ensureTag(ctx context.Context, tx *sql.Tx, name, tagSlug string) (int64, error) {
	var id int64
	err := tx.QueryRowContext(ctx, `SELECT id FROM tags WHERE slug = `+a.placeholder(1), tagSlug).Scan(&id)
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}

	insert := fmt.Sprintf(`INSERT INTO tags (slug, name, active) VALUES (%s, %s, true)`, a.placeholder(1), a.placeholder(2))
	if a.driver == "postgres" {
		err = tx.QueryRowContext(ctx, insert+" RETURNING id", tagSlug, name).Scan(&id)
		return id, err
	}
	res, err := tx.ExecContext(ctx, insert, tagSlug, name)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (a *mcpApp) placeholder(position int) string {
	if a.driver == "postgres" {
		return fmt.Sprintf("$%d", position)
	}
	return "?"
}

func (a *mcpApp) tagsExpr(postIDExpr string) string {
	if a.driver == "postgres" {
		return fmt.Sprintf(`COALESCE((SELECT string_agg(t.name, ', ' ORDER BY t.name) FROM post_tags pt JOIN tags t ON t.id = pt.tag_id WHERE pt.post_id = %s), '')`, postIDExpr)
	}
	return fmt.Sprintf(`COALESCE((SELECT group_concat(t.name, ', ') FROM post_tags pt JOIN tags t ON t.id = pt.tag_id WHERE pt.post_id = %s), '')`, postIDExpr)
}

func (a *mcpApp) publishedThisWeekQuery() string {
	if a.driver == "postgres" {
		return `SELECT COUNT(*) FROM posts WHERE status = 'published' AND published_at >= NOW() - INTERVAL '7 days'`
	}
	return `SELECT COUNT(*) FROM posts WHERE status = 'published' AND published_at >= datetime('now', '-7 days')`
}

func scanArticles(rows *sql.Rows, includeContent bool) ([]articleDTO, error) {
	var articles []articleDTO
	for rows.Next() {
		var a articleDTO
		var publishedAt sql.NullTime
		var err error
		if includeContent {
			err = rows.Scan(&a.ID, &a.Title, &a.Slug, &a.Excerpt, &a.Content, &a.Status, &a.Category, &a.AuthorName, &a.Tags, &publishedAt, &a.CreatedAt, &a.UpdatedAt)
		} else {
			err = rows.Scan(&a.ID, &a.Title, &a.Slug, &a.Excerpt, &a.Status, &a.Category, &a.AuthorName, &a.Tags, &publishedAt, &a.CreatedAt, &a.UpdatedAt)
		}
		if err != nil {
			return nil, err
		}
		if publishedAt.Valid {
			a.PublishedAt = &publishedAt.Time
		}
		articles = append(articles, a)
	}
	return articles, rows.Err()
}

func splitTags(raw string) []string {
	seen := map[string]bool{}
	var tags []string
	for _, item := range strings.Split(raw, ",") {
		name := strings.TrimSpace(item)
		if name == "" {
			continue
		}
		key := strings.ToLower(name)
		if seen[key] {
			continue
		}
		seen[key] = true
		tags = append(tags, name)
	}
	return tags
}

func validStatus(status string) bool {
	switch model.PostStatus(status) {
	case model.StatusDraft, model.StatusReview, model.StatusApproved, model.StatusScheduled, model.StatusPublished, model.StatusArchived:
		return true
	default:
		return false
	}
}

func textResultJSON(v any) *mcp.CallToolResult {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return errResult("Erro ao serializar resposta: " + err.Error())
	}
	return mcp.NewToolResultText(string(b))
}

func errResult(msg string) *mcp.CallToolResult {
	return mcp.NewToolResultError(msg)
}

func init() {
	log.SetOutput(os.Stderr)
}
