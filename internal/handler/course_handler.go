package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"react-go-cms-courses-service/internal/domain/apperror"
	"react-go-cms-courses-service/internal/domain/entity"
	"react-go-cms-courses-service/internal/infrastructure/httpx"
)

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	var page entity.Page[entity.Course]
	var err error
	page, err = h.courses.List(r.Context(), r.URL.Query().Get("status"), r.URL.Query().Get("lang"), authed(r), httpx.QueryInt(r, "page", defaultPage), httpx.QueryInt(r, "size", defaultPageSize))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toCoursePage(page))
}

func (h *Handler) stats(w http.ResponseWriter, r *http.Request) {
	var stats entity.CourseStats
	var err error
	stats, err = h.courses.Stats(r.Context())
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, CourseStatsResponse{Total: stats.Total})
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	var item entity.Course
	var err error
	item, err = h.courses.Get(r.Context(), chi.URLParam(r, "id"), r.URL.Query().Get("lang"), authed(r))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toCourseResponse(item))
}

func (h *Handler) bySlug(w http.ResponseWriter, r *http.Request) {
	var item entity.Course
	var err error
	item, err = h.courses.GetBySlug(r.Context(), chi.URLParam(r, "slug"), r.URL.Query().Get("lang"), authed(r))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toCourseResponse(item))
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var body CreateCourseDTO
	var err error
	err = httpx.DecodeJSONLenient(r, &body)
	if err != nil {
		writeServiceError(w, apperror.Invalid(invalidJSON))
		return
	}
	var item entity.Course
	item, err = h.courses.Create(r.Context(), toCourseCreate(body), subject(r))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, toCourseResponse(item))
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	var body UpdateCourseDTO
	var err error
	err = httpx.DecodeJSONLenient(r, &body)
	if err != nil {
		writeServiceError(w, apperror.Invalid(invalidJSON))
		return
	}
	var item entity.Course
	item, err = h.courses.Update(r.Context(), chi.URLParam(r, "id"), toCourseUpdate(body), r.URL.Query().Get("lang"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toCourseResponse(item))
}

func (h *Handler) patchStatus(w http.ResponseWriter, r *http.Request) {
	var body StatusDTO
	var err error
	err = httpx.DecodeJSONLenient(r, &body)
	if err != nil {
		writeServiceError(w, apperror.Invalid(invalidJSON))
		return
	}
	var item entity.Course
	item, err = h.courses.PatchStatus(r.Context(), chi.URLParam(r, "id"), body.Status, r.URL.Query().Get("lang"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toCourseResponse(item))
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	var err error
	err = h.courses.Delete(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.WriteNoContent(w)
}

func (h *Handler) getMeta(w http.ResponseWriter, r *http.Request) {
	var rows []entity.Metadata
	var err error
	rows, err = h.courses.Metadata(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toMetadataDTOs(rows))
}

func (h *Handler) putMeta(w http.ResponseWriter, r *http.Request) {
	var body []MetaInDTO
	var err error
	err = httpx.DecodeJSONLenient(r, &body)
	if err != nil {
		writeServiceError(w, apperror.Invalid(invalidJSON))
		return
	}
	var rows []entity.Metadata
	rows, err = h.courses.ReplaceMetadata(r.Context(), chi.URLParam(r, "id"), toMetaInputs(body))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toMetadataDTOs(rows))
}
