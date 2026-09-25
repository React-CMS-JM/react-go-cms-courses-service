package courses

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"react-go-cms-courses-service/internal/platform"
)

type Handler struct {
	store  *Store
	secret string
	issuer string
}

func NewHandler(store *Store, secret, issuer string) *Handler {
	return &Handler{store: store, secret: secret, issuer: issuer}
}

func (h *Handler) Register(mux chi.Router) {
	mux.Get("/api/courses/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("courses-service-ok"))
	})

	mux.Get("/api/courses/stats", h.auth(h.stats))
	mux.Get("/api/courses/by-slug/{slug}", h.optional(h.bySlug))
	mux.Get("/api/courses/{id}/metadata", h.getMeta)
	mux.Put("/api/courses/{id}/metadata", h.auth(h.putMeta))
	mux.Patch("/api/courses/{id}/status", h.auth(h.patchStatus))
	mux.Get("/api/courses/{id}", h.optional(h.get))
	mux.Put("/api/courses/{id}", h.auth(h.update))
	mux.Delete("/api/courses/{id}", h.auth(h.delete))
	mux.Get("/api/courses", h.optional(h.list))
	mux.Post("/api/courses", h.auth(h.create))

	mux.Get("/api/courses/{courseId}/lessons/by-slug/{slug}", h.lessonBySlug)
	mux.Get("/api/courses/{courseId}/lessons/{id}", h.getLesson)
	mux.Put("/api/courses/{courseId}/lessons/{id}", h.auth(h.updateLesson))
	mux.Delete("/api/courses/{courseId}/lessons/{id}", h.auth(h.deleteLesson))
	mux.Get("/api/courses/{courseId}/lessons", h.listLessons)
	mux.Post("/api/courses/{courseId}/lessons", h.auth(h.createLesson))

	mux.Put("/api/lessons/{id}", h.auth(h.updateLessonAlias))
	mux.Delete("/api/lessons/{id}", h.auth(h.deleteLessonAlias))
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	page, err := h.store.List(r.Context(), r.URL.Query().Get("status"), r.URL.Query().Get("lang"), authed(r), platform.QueryInt(r, "page", 0), platform.QueryInt(r, "size", 20))
	if err != nil {
		platform.WriteError(w, err, "error")
		return
	}
	platform.WriteJSON(w, http.StatusOK, page)
}

func (h *Handler) stats(w http.ResponseWriter, r *http.Request) {
	st, err := h.store.Stats(r.Context())
	if err != nil {
		platform.WriteError(w, err, "error")
		return
	}
	platform.WriteJSON(w, http.StatusOK, st)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	item, err := h.store.Get(r.Context(), chi.URLParam(r, "id"), r.URL.Query().Get("lang"), authed(r))
	if err != nil {
		platform.WriteError(w, err, "error")
		return
	}
	platform.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) bySlug(w http.ResponseWriter, r *http.Request) {
	item, err := h.store.GetBySlug(r.Context(), chi.URLParam(r, "slug"), r.URL.Query().Get("lang"), authed(r))
	if err != nil {
		platform.WriteError(w, err, "error")
		return
	}
	platform.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var body CreateCourse
	if err := platform.DecodeJSONLenient(r, &body); err != nil {
		platform.WriteError(w, bad("invalid JSON body"), "error")
		return
	}
	item, err := h.store.Create(r.Context(), body, subject(r))
	if err != nil {
		platform.WriteError(w, err, "error")
		return
	}
	platform.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	var body UpdateCourse
	if err := platform.DecodeJSONLenient(r, &body); err != nil {
		platform.WriteError(w, bad("invalid JSON body"), "error")
		return
	}
	item, err := h.store.Update(r.Context(), chi.URLParam(r, "id"), body, r.URL.Query().Get("lang"))
	if err != nil {
		platform.WriteError(w, err, "error")
		return
	}
	platform.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) patchStatus(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Status string `json:"status"`
	}
	if err := platform.DecodeJSONLenient(r, &body); err != nil {
		platform.WriteError(w, bad("invalid JSON body"), "error")
		return
	}
	item, err := h.store.PatchStatus(r.Context(), chi.URLParam(r, "id"), body.Status, r.URL.Query().Get("lang"))
	if err != nil {
		platform.WriteError(w, err, "error")
		return
	}
	platform.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	if err := h.store.Delete(r.Context(), chi.URLParam(r, "id")); err != nil {
		platform.WriteError(w, err, "error")
		return
	}
	platform.WriteNoContent(w)
}

func (h *Handler) getMeta(w http.ResponseWriter, r *http.Request) {
	rows, err := h.store.GetMetadata(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		platform.WriteError(w, err, "error")
		return
	}
	platform.WriteJSON(w, http.StatusOK, rows)
}

func (h *Handler) putMeta(w http.ResponseWriter, r *http.Request) {
	var body []MetaIn
	if err := platform.DecodeJSONLenient(r, &body); err != nil {
		platform.WriteError(w, bad("invalid JSON body"), "error")
		return
	}
	rows, err := h.store.ReplaceMetadata(r.Context(), chi.URLParam(r, "id"), body)
	if err != nil {
		platform.WriteError(w, err, "error")
		return
	}
	platform.WriteJSON(w, http.StatusOK, rows)
}

func (h *Handler) listLessons(w http.ResponseWriter, r *http.Request) {
	rows, err := h.store.ListLessons(r.Context(), chi.URLParam(r, "courseId"), r.URL.Query().Get("lang"))
	if err != nil {
		platform.WriteError(w, err, "error")
		return
	}
	platform.WriteJSON(w, http.StatusOK, rows)
}

func (h *Handler) getLesson(w http.ResponseWriter, r *http.Request) {
	item, err := h.store.GetLesson(r.Context(), chi.URLParam(r, "courseId"), chi.URLParam(r, "id"), r.URL.Query().Get("lang"))
	if err != nil {
		platform.WriteError(w, err, "error")
		return
	}
	platform.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) lessonBySlug(w http.ResponseWriter, r *http.Request) {
	item, err := h.store.GetLessonBySlug(r.Context(), chi.URLParam(r, "courseId"), chi.URLParam(r, "slug"), r.URL.Query().Get("lang"))
	if err != nil {
		platform.WriteError(w, err, "error")
		return
	}
	platform.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) createLesson(w http.ResponseWriter, r *http.Request) {
	var body CreateLesson
	if err := platform.DecodeJSONLenient(r, &body); err != nil {
		platform.WriteError(w, bad("invalid JSON body"), "error")
		return
	}
	item, err := h.store.CreateLesson(r.Context(), chi.URLParam(r, "courseId"), body)
	if err != nil {
		platform.WriteError(w, err, "error")
		return
	}
	platform.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) updateLesson(w http.ResponseWriter, r *http.Request) {
	var body UpdateLesson
	if err := platform.DecodeJSONLenient(r, &body); err != nil {
		platform.WriteError(w, bad("invalid JSON body"), "error")
		return
	}
	item, err := h.store.UpdateLesson(r.Context(), chi.URLParam(r, "courseId"), chi.URLParam(r, "id"), body, r.URL.Query().Get("lang"))
	if err != nil {
		platform.WriteError(w, err, "error")
		return
	}
	platform.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) deleteLesson(w http.ResponseWriter, r *http.Request) {
	if err := h.store.DeleteLesson(r.Context(), chi.URLParam(r, "courseId"), chi.URLParam(r, "id")); err != nil {
		platform.WriteError(w, err, "error")
		return
	}
	platform.WriteNoContent(w)
}

func (h *Handler) updateLessonAlias(w http.ResponseWriter, r *http.Request) {
	var body UpdateLesson
	if err := platform.DecodeJSONLenient(r, &body); err != nil {
		platform.WriteError(w, bad("invalid JSON body"), "error")
		return
	}
	item, err := h.store.UpdateLessonByID(r.Context(), chi.URLParam(r, "id"), body, r.URL.Query().Get("lang"))
	if err != nil {
		platform.WriteError(w, err, "error")
		return
	}
	platform.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) deleteLessonAlias(w http.ResponseWriter, r *http.Request) {
	if err := h.store.DeleteLessonByID(r.Context(), chi.URLParam(r, "id")); err != nil {
		platform.WriteError(w, err, "error")
		return
	}
	platform.WriteNoContent(w)
}

type ctxKey int

const (
	subjectKey ctxKey = iota
	authedKey
)

func subject(r *http.Request) string {
	v, _ := r.Context().Value(subjectKey).(string)
	return v
}

func authed(r *http.Request) bool {
	v, _ := r.Context().Value(authedKey).(bool)
	return v
}

func (h *Handler) auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, err := h.parse(r)
		if err != nil {
			platform.WriteError(w, platform.Unauthorized("Unauthorized"), "error")
			return
		}
		ctx := context.WithValue(r.Context(), subjectKey, claims.Subject)
		ctx = context.WithValue(ctx, authedKey, true)
		next(w, r.WithContext(ctx))
	}
}

func (h *Handler) optional(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		raw := platform.BearerToken(r)
		if raw == "" {
			next(w, r)
			return
		}
		claims, err := platform.ParseHS256(h.secret, h.issuer, raw)
		if err != nil || claims.Subject == "" {
			platform.WriteError(w, platform.Unauthorized("Unauthorized"), "error")
			return
		}
		ctx := context.WithValue(r.Context(), subjectKey, claims.Subject)
		ctx = context.WithValue(ctx, authedKey, true)
		next(w, r.WithContext(ctx))
	}
}

func (h *Handler) parse(r *http.Request) (*platform.Claims, error) {
	raw := platform.BearerToken(r)
	if raw == "" {
		return nil, platform.Unauthorized("Unauthorized")
	}
	claims, err := platform.ParseHS256(h.secret, h.issuer, raw)
	if err != nil || claims.Subject == "" {
		return nil, platform.Unauthorized("Unauthorized")
	}
	return claims, nil
}
