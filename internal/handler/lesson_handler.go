package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"react-go-cms-courses-service/internal/domain/apperror"
	"react-go-cms-courses-service/internal/domain/entity"
	"react-go-cms-courses-service/internal/infrastructure/httpx"
)

func (h *Handler) listLessons(w http.ResponseWriter, r *http.Request) {
	var rows []entity.Lesson
	var err error
	rows, err = h.lessons.List(r.Context(), chi.URLParam(r, "courseId"), r.URL.Query().Get("lang"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toLessonResponses(rows))
}

func (h *Handler) getLesson(w http.ResponseWriter, r *http.Request) {
	var item entity.Lesson
	var err error
	item, err = h.lessons.Get(r.Context(), chi.URLParam(r, "courseId"), chi.URLParam(r, "id"), r.URL.Query().Get("lang"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toLessonResponse(item))
}

func (h *Handler) lessonBySlug(w http.ResponseWriter, r *http.Request) {
	var item entity.Lesson
	var err error
	item, err = h.lessons.GetBySlug(r.Context(), chi.URLParam(r, "courseId"), chi.URLParam(r, "slug"), r.URL.Query().Get("lang"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toLessonResponse(item))
}

func (h *Handler) createLesson(w http.ResponseWriter, r *http.Request) {
	var body CreateLessonDTO
	var err error
	err = httpx.DecodeJSONLenient(r, &body)
	if err != nil {
		writeServiceError(w, apperror.Invalid(invalidJSON))
		return
	}
	var item entity.Lesson
	item, err = h.lessons.Create(r.Context(), chi.URLParam(r, "courseId"), toLessonCreate(body))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, toLessonResponse(item))
}

func (h *Handler) updateLesson(w http.ResponseWriter, r *http.Request) {
	var body UpdateLessonDTO
	var err error
	err = httpx.DecodeJSONLenient(r, &body)
	if err != nil {
		writeServiceError(w, apperror.Invalid(invalidJSON))
		return
	}
	var item entity.Lesson
	item, err = h.lessons.Update(r.Context(), chi.URLParam(r, "courseId"), chi.URLParam(r, "id"), toLessonUpdate(body), r.URL.Query().Get("lang"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toLessonResponse(item))
}

func (h *Handler) deleteLesson(w http.ResponseWriter, r *http.Request) {
	var err error
	err = h.lessons.Delete(r.Context(), chi.URLParam(r, "courseId"), chi.URLParam(r, "id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.WriteNoContent(w)
}

func (h *Handler) updateLessonAlias(w http.ResponseWriter, r *http.Request) {
	var body UpdateLessonDTO
	var err error
	err = httpx.DecodeJSONLenient(r, &body)
	if err != nil {
		writeServiceError(w, apperror.Invalid(invalidJSON))
		return
	}
	var item entity.Lesson
	item, err = h.lessons.UpdateByID(r.Context(), chi.URLParam(r, "id"), toLessonUpdate(body), r.URL.Query().Get("lang"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toLessonResponse(item))
}

func (h *Handler) deleteLessonAlias(w http.ResponseWriter, r *http.Request) {
	var err error
	err = h.lessons.DeleteByID(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.WriteNoContent(w)
}
