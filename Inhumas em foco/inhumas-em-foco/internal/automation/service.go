package automation

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"html"
	"html/template"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode"

	"inhumas-em-foco/internal/model"
	"inhumas-em-foco/internal/repository"
	"inhumas-em-foco/internal/slug"
)

const maxFeedBytes = 4 << 20

type Item struct {
	Title       string
	Link        string
	Summary     string
	PublishedAt *time.Time
}

type Fetcher func(ctx context.Context, source model.AutomationSource) ([]Item, error)

type Service struct {
	repo    *repository.Repository
	client  *http.Client
	fetcher Fetcher
	now     func() time.Time
}

func NewService(repo *repository.Repository) *Service {
	s := &Service{
		repo: repo,
		client: &http.Client{
			Timeout: 20 * time.Second,
		},
		now: time.Now,
	}
	s.fetcher = s.fetchFeed
	return s
}

func (s *Service) SetFetcher(fetcher Fetcher) {
	if fetcher != nil {
		s.fetcher = fetcher
	}
}

func (s *Service) RunAllActive(ctx context.Context) ([]model.AutomationRun, error) {
	sources, err := s.repo.AutomationSourceList(ctx, true, 200)
	if err != nil {
		return nil, err
	}
	runs := make([]model.AutomationRun, 0, len(sources))
	var firstErr error
	for _, source := range sources {
		run, err := s.RunSource(ctx, source.ID)
		if run != nil {
			runs = append(runs, *run)
		}
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return runs, firstErr
}

func (s *Service) RunSource(ctx context.Context, sourceID int64) (*model.AutomationRun, error) {
	source, err := s.repo.AutomationSourceGetByID(ctx, sourceID)
	if err != nil {
		return nil, err
	}
	if source == nil {
		return nil, fmt.Errorf("fonte de automacao nao encontrada")
	}

	startedAt := s.now()
	run := &model.AutomationRun{
		SourceID:  &source.ID,
		Status:    string(model.AutomationRunSuccess),
		StartedAt: startedAt,
	}
	if err := s.repo.AutomationRunCreate(ctx, run); err != nil {
		return nil, err
	}

	logs := []string{fmt.Sprintf("Inicio da coleta: %s", source.Name)}
	finish := func(status model.AutomationRunStatus, runErr error) (*model.AutomationRun, error) {
		now := s.now()
		run.Status = string(status)
		run.FinishedAt = &now
		run.Log = strings.Join(logs, "\n")
		if runErr != nil {
			run.Error = runErr.Error()
		}
		if err := s.repo.AutomationRunUpdate(ctx, run); err != nil {
			return run, err
		}
		if status != model.AutomationRunError {
			_ = s.repo.AutomationSourceMarkRun(ctx, source.ID, now)
		}
		return run, runErr
	}

	if !source.Active {
		err := fmt.Errorf("fonte inativa")
		logs = append(logs, err.Error())
		return finish(model.AutomationRunError, err)
	}

	items, err := s.fetcher(ctx, *source)
	if err != nil {
		logs = append(logs, "Erro ao coletar feed: "+err.Error())
		return finish(model.AutomationRunError, err)
	}
	run.ItemsFound = len(items)
	logs = append(logs, fmt.Sprintf("%d item(ns) encontrados", len(items)))

	recentPosts, err := s.repo.AutomationRecentPostTitles(ctx, 300)
	if err != nil {
		logs = append(logs, "Nao foi possivel carregar titulos recentes: "+err.Error())
		return finish(model.AutomationRunError, err)
	}

	for _, item := range items {
		item.Title = strings.TrimSpace(item.Title)
		item.Link = strings.TrimSpace(item.Link)
		if item.Title == "" {
			run.Duplicates++
			logs = append(logs, "Item ignorado sem titulo")
			continue
		}
		duplicate, reason, err := s.isDuplicate(ctx, item, recentPosts)
		if err != nil {
			logs = append(logs, "Falha na deduplicacao de "+item.Title+": "+err.Error())
			return finish(model.AutomationRunPartial, err)
		}
		if duplicate {
			run.Duplicates++
			logs = append(logs, "Duplicado ignorado ("+reason+"): "+item.Title)
			continue
		}

		post := s.postFromItem(ctx, *source, item)
		if err := s.repo.PostCreate(ctx, post); err != nil {
			logs = append(logs, "Falha ao criar rascunho de "+item.Title+": "+err.Error())
			return finish(model.AutomationRunPartial, err)
		}
		run.DraftsCreated++
		recentPosts = append([]model.Post{{ID: post.ID, Title: post.Title, SourceURL: post.SourceURL}}, recentPosts...)
		logs = append(logs, fmt.Sprintf("Rascunho #%d criado: %s", post.ID, post.Title))
	}

	logs = append(logs, fmt.Sprintf("Fim: %d rascunho(s), %d duplicado(s)", run.DraftsCreated, run.Duplicates))
	return finish(model.AutomationRunSuccess, nil)
}

func (s *Service) isDuplicate(ctx context.Context, item Item, recent []model.Post) (bool, string, error) {
	duplicate, reason, err := s.repo.AutomationPostDuplicateExact(ctx, item.Title, item.Link)
	if err != nil || duplicate {
		return duplicate, reason, err
	}
	for _, post := range recent {
		if IsSimilarTitle(item.Title, post.Title) {
			return true, "titulo similar", nil
		}
	}
	return false, "", nil
}

func (s *Service) postFromItem(ctx context.Context, source model.AutomationSource, item Item) *model.Post {
	excerpt := truncate(cleanText(item.Summary), 260)
	content := contentHTML(item, excerpt)
	categoryID := source.DefaultCategoryID
	if categoryID == nil {
		if cat, _ := s.repo.CategoryGetBySlug(ctx, "noticias"); cat != nil {
			categoryID = &cat.ID
		}
	}
	return &model.Post{
		Title:              item.Title,
		Slug:               slug.Unique(item.Title, func(candidate string) bool { return s.repo.PostSlugExists(ctx, candidate) }),
		Excerpt:            excerpt,
		Content:            content,
		MetaTitle:          truncate(item.Title, 200),
		MetaDescription:    truncate(excerpt, 300),
		SourceName:         source.Name,
		SourceURL:          item.Link,
		ReadingTimeMinutes: readingTime(content),
		CategoryID:         categoryID,
		Status:             model.StatusDraft,
		EditorialNotes:     "Rascunho criado automaticamente pela automacao de noticias. Revisao humana obrigatoria antes de publicar. Preserve e confira a fonte original.",
	}
}

func (s *Service) fetchFeed(ctx context.Context, source model.AutomationSource) ([]Item, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, source.URL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "PortalNewsBot/1.0")
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d ao baixar fonte", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxFeedBytes))
	if err != nil {
		return nil, err
	}
	return parseFeed(data)
}

func parseFeed(data []byte) ([]Item, error) {
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return nil, fmt.Errorf("feed vazio")
	}
	var rss rssFeed
	if err := xml.Unmarshal(data, &rss); err == nil && len(rss.Channel.Items) > 0 {
		items := make([]Item, 0, len(rss.Channel.Items))
		for _, item := range rss.Channel.Items {
			items = append(items, Item{
				Title:       strings.TrimSpace(item.Title),
				Link:        strings.TrimSpace(item.Link),
				Summary:     firstNonEmpty(item.Description, item.Content),
				PublishedAt: parseFeedTime(item.PubDate),
			})
		}
		return items, nil
	}
	var atom atomFeed
	if err := xml.Unmarshal(data, &atom); err == nil && len(atom.Entries) > 0 {
		items := make([]Item, 0, len(atom.Entries))
		for _, entry := range atom.Entries {
			items = append(items, Item{
				Title:       strings.TrimSpace(entry.Title),
				Link:        atomEntryLink(entry),
				Summary:     firstNonEmpty(entry.Summary, entry.Content),
				PublishedAt: parseFeedTime(firstNonEmpty(entry.Updated, entry.Published)),
			})
		}
		return items, nil
	}
	return nil, fmt.Errorf("formato RSS/Atom nao reconhecido")
}

type rssFeed struct {
	Channel struct {
		Items []rssItem `xml:"item"`
	} `xml:"channel"`
}

type rssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Content     string `xml:"encoded"`
	PubDate     string `xml:"pubDate"`
}

type atomFeed struct {
	Entries []atomEntry `xml:"entry"`
}

type atomEntry struct {
	Title     string     `xml:"title"`
	Links     []atomLink `xml:"link"`
	Summary   string     `xml:"summary"`
	Content   string     `xml:"content"`
	Updated   string     `xml:"updated"`
	Published string     `xml:"published"`
}

type atomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
}

func atomEntryLink(entry atomEntry) string {
	for _, link := range entry.Links {
		if link.Rel == "" || link.Rel == "alternate" {
			return strings.TrimSpace(link.Href)
		}
	}
	if len(entry.Links) > 0 {
		return strings.TrimSpace(entry.Links[0].Href)
	}
	return ""
}

func parseFeedTime(value string) *time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	layouts := []string{time.RFC1123Z, time.RFC1123, time.RFC3339, "Mon, 02 Jan 2006 15:04:05 -0700"}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, value); err == nil {
			return &parsed
		}
	}
	return nil
}

var tagRE = regexp.MustCompile(`<[^>]+>`)

func cleanText(value string) string {
	value = tagRE.ReplaceAllString(value, " ")
	value = html.UnescapeString(value)
	return strings.Join(strings.Fields(value), " ")
}

func contentHTML(item Item, excerpt string) string {
	var b strings.Builder
	summary := excerpt
	if summary == "" {
		summary = "Conteudo coletado automaticamente para revisao editorial."
	}
	b.WriteString("<p>")
	b.WriteString(template.HTMLEscapeString(summary))
	b.WriteString("</p>")
	if item.Link != "" {
		b.WriteString(`<p><strong>Fonte original:</strong> <a href="`)
		b.WriteString(template.HTMLEscapeString(item.Link))
		b.WriteString(`" rel="nofollow noopener" target="_blank">`)
		b.WriteString(template.HTMLEscapeString(item.Link))
		b.WriteString("</a></p>")
	}
	return b.String()
}

func readingTime(content string) int {
	words := len(strings.Fields(cleanText(content)))
	if words <= 0 {
		return 1
	}
	minutes := words / 220
	if words%220 != 0 {
		minutes++
	}
	if minutes < 1 {
		return 1
	}
	return minutes
}

func truncate(value string, max int) string {
	value = strings.TrimSpace(value)
	if max <= 0 || len([]rune(value)) <= max {
		return value
	}
	runes := []rune(value)
	return strings.TrimSpace(string(runes[:max-1])) + "..."
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func IsSimilarTitle(a, b string) bool {
	left := titleTokens(a)
	right := titleTokens(b)
	if len(left) < 4 || len(right) < 4 {
		return false
	}
	intersection := 0
	i, j := 0, 0
	for i < len(left) && j < len(right) {
		switch {
		case left[i] == right[j]:
			intersection++
			i++
			j++
		case left[i] < right[j]:
			i++
		default:
			j++
		}
	}
	union := len(left) + len(right) - intersection
	return union > 0 && float64(intersection)/float64(union) >= 0.82
}

func titleTokens(value string) []string {
	seen := map[string]bool{}
	var tokens []string
	var b strings.Builder
	flush := func() {
		token := strings.TrimSpace(b.String())
		b.Reset()
		if len([]rune(token)) < 3 || seen[token] {
			return
		}
		seen[token] = true
		tokens = append(tokens, token)
	}
	for _, r := range strings.ToLower(value) {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			b.WriteRune(r)
			continue
		}
		flush()
	}
	flush()
	sort.Strings(tokens)
	return tokens
}
