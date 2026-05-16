package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"inhumas-em-foco/internal/auth"
	"inhumas-em-foco/internal/model"
	"inhumas-em-foco/internal/slug"
	"inhumas-em-foco/internal/storage"
)

func (h *Handler) AdminEvents(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requirePermission(w, r, auth.PermEventsManage); !ok {
		return
	}
	events, _ := h.repo.EventList(r.Context(), false, 300)
	h.Render(w, r, "admin_events.html", map[string]any{
		"Title":   "Eventos",
		"Active":  "events",
		"Events":  events,
		"Summary": localCommercialSummaryEvents(events),
	})
}

func (h *Handler) AdminEventNew(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requirePermission(w, r, auth.PermEventsManage); !ok {
		return
	}
	h.renderEventForm(w, r, nil, "")
}

func (h *Handler) AdminEventCreate(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requirePermission(w, r, auth.PermEventsManage)
	if !ok {
		return
	}
	if err := h.parseMultipart(w, r); err != nil {
		h.renderEventForm(w, r, nil, "Tamanho maximo excedido")
		return
	}
	event := h.eventFromRequest(r, nil)
	if strings.TrimSpace(event.Title) == "" {
		h.renderEventForm(w, r, event, "Titulo e obrigatorio")
		return
	}
	if event.StartAt.IsZero() {
		h.renderEventForm(w, r, event, "Informe data e horario de inicio")
		return
	}
	event.Slug = slug.Unique(event.Title, func(s string) bool {
		return h.repo.EventSlugExists(r.Context(), s)
	})
	h.applyEventMediaSelection(r, event)
	h.applyEventUpload(r, user, event)
	if err := h.repo.EventCreate(r.Context(), event); err != nil {
		h.renderEventForm(w, r, event, "Nao foi possivel criar o evento")
		return
	}
	h.auditAdminAction(r, user, "create", "event", auditEntityID(event.ID), map[string]any{
		"title":       event.Title,
		"slug":        event.Slug,
		"status":      event.Status,
		"start_at":    event.StartAt.Format(time.RFC3339),
		"is_featured": event.IsFeatured,
	})
	http.Redirect(w, r, h.cfg.AdminPathPrefix+"/events", http.StatusSeeOther)
}

func (h *Handler) AdminEventEdit(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requirePermission(w, r, auth.PermEventsManage); !ok {
		return
	}
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	event, err := h.repo.EventGetByID(r.Context(), id)
	if err != nil || event == nil {
		http.NotFound(w, r)
		return
	}
	h.renderEventForm(w, r, event, "")
}

func (h *Handler) AdminEventUpdate(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requirePermission(w, r, auth.PermEventsManage)
	if !ok {
		return
	}
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	event, err := h.repo.EventGetByID(r.Context(), id)
	if err != nil || event == nil {
		http.NotFound(w, r)
		return
	}
	if err := h.parseMultipart(w, r); err != nil {
		h.renderEventForm(w, r, event, "Tamanho maximo excedido")
		return
	}
	oldSlug := event.Slug
	event = h.eventFromRequest(r, event)
	if strings.TrimSpace(event.Title) == "" {
		h.renderEventForm(w, r, event, "Titulo e obrigatorio")
		return
	}
	if event.StartAt.IsZero() {
		h.renderEventForm(w, r, event, "Informe data e horario de inicio")
		return
	}
	newSlug := slug.Generate(event.Title)
	if newSlug != oldSlug {
		event.Slug = slug.Unique(event.Title, func(s string) bool {
			return h.repo.EventSlugExists(r.Context(), s)
		})
	}
	if r.FormValue("remove_image") == "on" {
		event.ImageKey = ""
	}
	h.applyEventMediaSelection(r, event)
	h.applyEventUpload(r, user, event)
	if err := h.repo.EventUpdate(r.Context(), event); err != nil {
		h.renderEventForm(w, r, event, "Nao foi possivel atualizar o evento")
		return
	}
	h.auditAdminAction(r, user, "update", "event", auditEntityID(event.ID), map[string]any{
		"title":       event.Title,
		"slug":        event.Slug,
		"status":      event.Status,
		"start_at":    event.StartAt.Format(time.RFC3339),
		"is_featured": event.IsFeatured,
	})
	http.Redirect(w, r, h.cfg.AdminPathPrefix+"/events", http.StatusSeeOther)
}

func (h *Handler) AdminEventDelete(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requirePermission(w, r, auth.PermEventsManage)
	if !ok {
		return
	}
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	event, _ := h.repo.EventGetByID(r.Context(), id)
	if err := h.repo.EventDelete(r.Context(), id); err != nil {
		h.Render(w, r, "admin_events.html", map[string]any{"Title": "Eventos", "Active": "events", "Error": "Nao foi possivel excluir o evento"})
		return
	}
	h.auditAdminAction(r, user, "delete", "event", auditEntityID(id), map[string]any{
		"title": eventTitle(event),
	})
	http.Redirect(w, r, h.cfg.AdminPathPrefix+"/events", http.StatusSeeOther)
}

func (h *Handler) renderEventForm(w http.ResponseWriter, r *http.Request, event *model.Event, errorMessage string) {
	title := "Novo Evento"
	if event != nil && event.ID > 0 {
		title = "Editar Evento"
	}
	h.Render(w, r, "admin_event_form.html", map[string]any{
		"Title":       title,
		"Active":      "events",
		"Event":       event,
		"MediaAssets": h.mediaAssetsForForms(r.Context()),
		"Error":       errorMessage,
	})
}

func (h *Handler) eventFromRequest(r *http.Request, event *model.Event) *model.Event {
	if event == nil {
		event = &model.Event{}
	}
	event.Title = strings.TrimSpace(r.FormValue("title"))
	event.Description = strings.TrimSpace(r.FormValue("description"))
	event.Location = strings.TrimSpace(r.FormValue("location"))
	event.Organizer = strings.TrimSpace(r.FormValue("organizer"))
	event.TicketURL = strings.TrimSpace(r.FormValue("ticket_url"))
	event.PriceDisplay = strings.TrimSpace(r.FormValue("price_display"))
	event.Status = normalizeEventStatusForHandler(r.FormValue("status"))
	event.IsFeatured = r.FormValue("is_featured") == "on"
	event.IsSponsored = r.FormValue("is_sponsored") == "on"
	event.MetaTitle = strings.TrimSpace(r.FormValue("meta_title"))
	event.MetaDescription = strings.TrimSpace(r.FormValue("meta_description"))
	event.StartAt = parseDateTimeLocal(r.FormValue("start_at"))
	endAt := parseDateTimeLocal(r.FormValue("end_at"))
	if endAt.IsZero() {
		event.EndAt = nil
	} else {
		event.EndAt = &endAt
	}
	return event
}

func (h *Handler) applyEventMediaSelection(r *http.Request, event *model.Event) {
	if key := h.mediaKeyFromRequest(r, "image_media_key"); key != "" {
		event.ImageKey = key
	}
}

func (h *Handler) applyEventUpload(r *http.Request, user *model.User, event *model.Event) {
	file, header, err := r.FormFile("image")
	if err != nil {
		return
	}
	defer file.Close()
	key := storage.GenerateKey(header.Filename)
	if info, err := h.storage.Upload(r.Context(), key, file, header.Header.Get("Content-Type")); err == nil {
		event.ImageKey = info.Key
		h.recordMediaAssetFromUpload(r, user, header, info, event.Title)
	}
}

func parseDateTimeLocal(value string) time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}
	}
	for _, layout := range []string{"2006-01-02T15:04", "2006-01-02 15:04", "2006-01-02"} {
		if parsed, err := time.ParseInLocation(layout, value, time.Local); err == nil {
			return parsed
		}
	}
	return time.Time{}
}

func normalizeEventStatusForHandler(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "draft", "archived":
		return strings.ToLower(strings.TrimSpace(status))
	default:
		return "active"
	}
}
