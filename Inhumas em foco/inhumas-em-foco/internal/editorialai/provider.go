package editorialai

import (
	"context"
	"errors"
	"html"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"inhumas-em-foco/internal/model"
)

type Task string

const (
	TaskImproveTitle        Task = "improve_title"
	TaskGenerateSubtitle    Task = "generate_subtitle"
	TaskGenerateSummary     Task = "generate_summary"
	TaskMetaDescription     Task = "meta_description"
	TaskSuggestTags         Task = "suggest_tags"
	TaskRewriteJournalistic Task = "rewrite_journalistic"
	TaskSocialCall          Task = "social_call"
	TaskCheckDuplicate      Task = "check_duplicate"
)

var ErrInvalidTask = errors.New("acao de IA invalida")

type Provider interface {
	Name() string
	Suggest(ctx context.Context, req Request) (Suggestion, error)
}

type Request struct {
	Task        Task
	Post        model.Post
	Tags        []model.Tag
	RecentPosts []model.Post
}

type Suggestion struct {
	Title           string   `json:"title,omitempty"`
	Subtitle        string   `json:"subtitle,omitempty"`
	Summary         string   `json:"summary,omitempty"`
	MetaDescription string   `json:"meta_description,omitempty"`
	Tags            []string `json:"tags,omitempty"`
	RewrittenText   string   `json:"rewritten_text,omitempty"`
	SocialCall      string   `json:"social_call,omitempty"`
	DuplicateRisk   string   `json:"duplicate_risk,omitempty"`
	Notes           string   `json:"notes,omitempty"`
	SourceName      string   `json:"source_name,omitempty"`
	SourceURL       string   `json:"source_url,omitempty"`
}

type MockProvider struct{}

func NewMockProvider() *MockProvider {
	return &MockProvider{}
}

func (p *MockProvider) Name() string {
	return "mock-editorial"
}

func (p *MockProvider) Suggest(ctx context.Context, req Request) (Suggestion, error) {
	if err := req.Task.Validate(); err != nil {
		return Suggestion{}, err
	}
	post := req.Post
	body := cleanText(firstNonEmpty(post.Content, post.Excerpt))
	summary := truncate(firstNonEmpty(post.Excerpt, body), 240)
	base := Suggestion{
		Notes:      guardrailNote(post),
		SourceName: post.SourceName,
		SourceURL:  post.SourceURL,
	}

	switch req.Task {
	case TaskImproveTitle:
		base.Title = improveTitle(post.Title)
	case TaskGenerateSubtitle:
		base.Subtitle = subtitleFrom(post.Title, summary)
	case TaskGenerateSummary:
		base.Summary = truncate(summary, 320)
	case TaskMetaDescription:
		base.MetaDescription = truncate(metaDescriptionFrom(post.Title, summary), 160)
	case TaskSuggestTags:
		base.Tags = suggestTags(post, req.Tags)
	case TaskRewriteJournalistic:
		base.RewrittenText = rewriteJournalistic(post.Title, body)
	case TaskSocialCall:
		base.SocialCall = socialCall(post.Title)
	case TaskCheckDuplicate:
		base.DuplicateRisk = duplicateRisk(post, req.RecentPosts)
	}
	return base, nil
}

func (t Task) Validate() error {
	switch t {
	case TaskImproveTitle, TaskGenerateSubtitle, TaskGenerateSummary, TaskMetaDescription, TaskSuggestTags, TaskRewriteJournalistic, TaskSocialCall, TaskCheckDuplicate:
		return nil
	default:
		return ErrInvalidTask
	}
}

func AllTasks() []Task {
	return []Task{TaskImproveTitle, TaskGenerateSubtitle, TaskGenerateSummary, TaskMetaDescription, TaskSuggestTags, TaskRewriteJournalistic, TaskSocialCall, TaskCheckDuplicate}
}

func improveTitle(title string) string {
	title = strings.TrimSpace(title)
	if title == "" {
		return "Titulo pendente de apuracao"
	}
	if strings.Contains(strings.ToLower(title), "inhumas") {
		return title
	}
	if len([]rune(title)) <= 78 {
		return title + " em Inhumas"
	}
	return title
}

func subtitleFrom(title, summary string) string {
	if summary != "" {
		return truncate(summary, 180)
	}
	return truncate("Pauta local sobre "+strings.TrimSpace(title)+", com revisao editorial antes da publicacao.", 180)
}

func metaDescriptionFrom(title, summary string) string {
	if summary != "" {
		return summary
	}
	return "Leia no Inhumas em Foco: " + strings.TrimSpace(title) + "."
}

func suggestTags(post model.Post, existing []model.Tag) []string {
	text := strings.ToLower(post.Title + " " + post.Excerpt + " " + cleanText(post.Content))
	var matched []string
	seen := map[string]bool{}
	for _, tag := range existing {
		name := strings.TrimSpace(tag.Name)
		if name == "" {
			continue
		}
		if strings.Contains(text, strings.ToLower(name)) && !seen[strings.ToLower(name)] {
			matched = append(matched, name)
			seen[strings.ToLower(name)] = true
		}
	}
	for _, token := range topTokens(text) {
		if len(matched) >= 6 {
			break
		}
		label := titleCase(token)
		key := strings.ToLower(label)
		if seen[key] {
			continue
		}
		matched = append(matched, label)
		seen[key] = true
	}
	if !seen["inhumas"] {
		matched = append(matched, "Inhumas")
	}
	return matched
}

func rewriteJournalistic(title, body string) string {
	body = truncate(body, 900)
	if body == "" {
		return "Texto insuficiente para reescrita. Complete a apuracao antes de usar a sugestao."
	}
	title = strings.TrimSpace(title)
	if title == "" {
		return body
	}
	return strings.TrimSpace(title) + ". " + body
}

func socialCall(title string) string {
	title = strings.TrimSpace(title)
	if title == "" {
		return "Nova pauta local em revisao no Inhumas em Foco."
	}
	return truncate("Nova pauta no Inhumas em Foco: "+title+". Acompanhe a cobertura local.", 220)
}

func duplicateRisk(post model.Post, recent []model.Post) string {
	for _, candidate := range recent {
		if candidate.ID == post.ID {
			continue
		}
		if similarTitle(post.Title, candidate.Title) {
			return "Possivel duplicidade com: " + candidate.Title
		}
	}
	return "Nenhuma duplicidade forte encontrada nos titulos recentes."
}

func guardrailNote(post model.Post) string {
	parts := []string{"Sugestao baseada apenas no titulo, resumo, conteudo e fonte cadastrados; revise fatos antes de publicar."}
	if strings.TrimSpace(post.SourceName) == "" && strings.TrimSpace(post.SourceURL) == "" {
		parts = append(parts, "Fonte original ausente: complete a fonte antes de aprovar.")
	}
	if post.Status != model.StatusDraft {
		parts = append(parts, "A IA nao altera status; use o fluxo editorial humano.")
	}
	return strings.Join(parts, " ")
}

var tagRE = regexp.MustCompile(`<[^>]+>`)

func cleanText(value string) string {
	value = tagRE.ReplaceAllString(value, " ")
	value = html.UnescapeString(value)
	return strings.Join(strings.Fields(value), " ")
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

func topTokens(text string) []string {
	counts := map[string]int{}
	for _, token := range tokens(text) {
		if isStopword(token) {
			continue
		}
		counts[token]++
	}
	type pair struct {
		token string
		count int
	}
	var pairs []pair
	for token, count := range counts {
		pairs = append(pairs, pair{token: token, count: count})
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].count == pairs[j].count {
			return pairs[i].token < pairs[j].token
		}
		return pairs[i].count > pairs[j].count
	})
	limit := 5
	if len(pairs) < limit {
		limit = len(pairs)
	}
	out := make([]string, 0, limit)
	for i := 0; i < limit; i++ {
		out = append(out, pairs[i].token)
	}
	return out
}

func similarTitle(a, b string) bool {
	left := tokens(a)
	right := tokens(b)
	if len(left) < 4 || len(right) < 4 {
		return false
	}
	set := map[string]bool{}
	for _, token := range left {
		set[token] = true
	}
	intersection := 0
	for _, token := range right {
		if set[token] {
			intersection++
		}
	}
	union := len(left) + len(right) - intersection
	return union > 0 && float64(intersection)/float64(union) >= 0.72
}

func tokens(value string) []string {
	seen := map[string]bool{}
	var out []string
	var b strings.Builder
	flush := func() {
		token := strings.TrimSpace(b.String())
		b.Reset()
		if len([]rune(token)) < 3 || seen[token] {
			return
		}
		seen[token] = true
		out = append(out, token)
	}
	for _, r := range strings.ToLower(value) {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			b.WriteRune(r)
			continue
		}
		flush()
	}
	flush()
	sort.Strings(out)
	return out
}

func titleCase(value string) string {
	if value == "" {
		return value
	}
	runes := []rune(value)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

func isStopword(token string) bool {
	switch token {
	case "para", "com", "dos", "das", "por", "uma", "sobre", "mais", "como", "que", "local", "noticia", "noticias", "inhumas":
		return true
	default:
		return false
	}
}
