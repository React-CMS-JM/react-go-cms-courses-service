package handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"react-go-cms-courses-service/internal/application/service/course"
	"react-go-cms-courses-service/internal/application/service/lesson"
	"react-go-cms-courses-service/internal/domain/apperror"
	"react-go-cms-courses-service/internal/infrastructure/httpx"
)

const (
	healthBody      = "courses-service-ok"
	errorField      = "error"
	invalidJSON     = "invalid JSON body"
	defaultPage     = 0
	defaultPageSize = 20
)

// Handler exposes course and lesson HTTP endpoints.
type Handler struct {
	courses *course.Service
	lessons *lesson.Service
	secret  string
	issuer  string
}

// New builds the HTTP handler.
func New(courses *course.Service, lessons *lesson.Service, secret, issuer string) *Handler {
	return &Handler{courses: courses, lessons: lessons, secret: secret, issuer: issuer}
}

// Register wires the courses service routes.
func (h *Handler) Register(mux chi.Router) {
	mux.Get("/api/courses/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte(healthBody))
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

func writeServiceError(w http.ResponseWriter, err error) {
	var domainErr *apperror.Error
	if errors.As(err, &domainErr) {
		status := http.StatusBadRequest
		if errors.Is(err, apperror.ErrNotFound) {
			status = http.StatusNotFound
		}
		httpx.WriteJSON(w, status, map[string]string{errorField: domainErr.Message})
		return
	}
	httpx.WriteError(w, err, errorField)
}
