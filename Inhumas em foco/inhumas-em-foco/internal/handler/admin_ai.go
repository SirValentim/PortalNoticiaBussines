package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"inhumas-em-foco/internal/editorialai"
	"inhumas-em-foco/internal/model"
)

func (h *Handler) AdminPostAIAction(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requireCurrentUser(w, r)
	if !ok {
		return
	}
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	post, err := h.repo.PostGetByID(r.Context(), id)
	if err != nil || post == nil {
		http.NotFound(w, r)
		return
	}
	if !h.postSvc.CanEdit(user, post) {
		http.Error(w, "Acesso negado", http.StatusForbidden)
		return
	}

	task := editorialai.Task(r.PathValue("action"))
	if err := task.Validate(); err != nil {
		http.Error(w, "Acao de IA invalida", http.StatusBadRequest)
		return
	}

	tags, _ := h.repo.TagList(r.Context(), true)
	recent, _ := h.repo.AutomationRecentPostTitles(r.Context(), 150)
	suggestion, err := h.aiSvc.Suggest(r.Context(), editorialai.Request{
		Task:        task,
		Post:        *post,
		Tags:        tags,
		RecentPosts: recent,
	})
	if err != nil {
		data := h.adminPostFormData(r.Context(), user, post, "Nao foi possivel gerar sugestao editorial")
		h.Render(w, r, "admin_post_form.html", data)
		return
	}

	raw, _ := json.Marshal(suggestion)
	userID := user.ID
	postID := post.ID
	_ = h.repo.AIUsageLogCreate(r.Context(), &model.AIUsageLog{
		PostID:     &postID,
		UserID:     &userID,
		Action:     string(task),
		Provider:   h.aiSvc.Name(),
		InputTitle: post.Title,
		Output:     string(raw),
		SourceName: post.SourceName,
		SourceURL:  post.SourceURL,
	})
	h.auditAdminAction(r, user, "ai_"+string(task), "post", auditEntityID(post.ID), map[string]any{
		"provider": h.aiSvc.Name(),
		"title":    post.Title,
	})

	data := h.adminPostFormData(r.Context(), user, post, "")
	data["AISuggestion"] = suggestion
	data["AIAction"] = string(task)
	data["AIActionLabel"] = aiActionLabel(task)
	h.Render(w, r, "admin_post_form.html", data)
}

func aiActionLabel(value any) string {
	task, ok := value.(editorialai.Task)
	if !ok {
		if raw, ok := value.(string); ok {
			task = editorialai.Task(raw)
		}
	}
	switch task {
	case editorialai.TaskImproveTitle:
		return "Melhorar titulo"
	case editorialai.TaskGenerateSubtitle:
		return "Criar subtitulo"
	case editorialai.TaskGenerateSummary:
		return "Gerar resumo"
	case editorialai.TaskMetaDescription:
		return "Gerar meta description"
	case editorialai.TaskSuggestTags:
		return "Sugerir tags"
	case editorialai.TaskRewriteJournalistic:
		return "Reescrever em tom jornalistico"
	case editorialai.TaskSocialCall:
		return "Criar chamada social"
	case editorialai.TaskCheckDuplicate:
		return "Verificar duplicidade"
	default:
		return "Assistente editorial"
	}
}
