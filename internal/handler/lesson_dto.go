package handler

import "react-go-cms-courses-service/internal/domain/entity"

// LessonResponse is the public lesson payload.
type LessonResponse struct {
	ID             string      `json:"id"`
	CourseID       string      `json:"courseId"`
	ParentLessonID *string     `json:"parentLessonId"`
	SortOrder      int         `json:"sortOrder"`
	AccessLevel    string      `json:"accessLevel"`
	CreatedAt      InstantJSON `json:"createdAt"`
	UpdatedAt      InstantJSON `json:"updatedAt"`
	LanguageCode   string      `json:"languageCode"`
	Title          string      `json:"title"`
	Slug           string      `json:"slug"`
	Content        string      `json:"content"`
}

// LessonTranslationDTO is a lesson translation in a write body.
type LessonTranslationDTO struct {
	LanguageCode string `json:"languageCode"`
	Title        string `json:"title"`
	Slug         string `json:"slug"`
	Content      string `json:"content"`
}

// CreateLessonDTO is the create-lesson body.
type CreateLessonDTO struct {
	ParentLessonID *string               `json:"parentLessonId"`
	SortOrder      *int                  `json:"sortOrder"`
	AccessLevel    *string               `json:"accessLevel"`
	Translation    *LessonTranslationDTO `json:"translation"`
}

// UpdateLessonDTO is the update-lesson body.
type UpdateLessonDTO struct {
	ParentLessonID *string               `json:"parentLessonId"`
	SortOrder      *int                  `json:"sortOrder"`
	AccessLevel    *string               `json:"accessLevel"`
	Translation    *LessonTranslationDTO `json:"translation"`
}

func toLessonResponse(lesson entity.Lesson) LessonResponse {
	return LessonResponse{
		ID:             lesson.ID,
		CourseID:       lesson.CourseID,
		ParentLessonID: lesson.ParentLessonID,
		SortOrder:      lesson.SortOrder,
		AccessLevel:    lesson.AccessLevel,
		CreatedAt:      toInstant(lesson.CreatedAt),
		UpdatedAt:      toInstant(lesson.UpdatedAt),
		LanguageCode:   lesson.LanguageCode,
		Title:          lesson.Title,
		Slug:           lesson.Slug,
		Content:        lesson.Content,
	}
}

func toLessonResponses(lessons []entity.Lesson) []LessonResponse {
	if lessons == nil {
		return []LessonResponse{}
	}
	out := make([]LessonResponse, 0, len(lessons))
	for _, lesson := range lessons {
		out = append(out, toLessonResponse(lesson))
	}
	return out
}

func toLessonCreate(body CreateLessonDTO) entity.LessonCreate {
	return entity.LessonCreate{
		ParentLessonID: body.ParentLessonID,
		SortOrder:      body.SortOrder,
		AccessLevel:    body.AccessLevel,
		Translation:    toLessonTranslation(body.Translation),
	}
}

func toLessonUpdate(body UpdateLessonDTO) entity.LessonUpdate {
	return entity.LessonUpdate{
		ParentLessonID: body.ParentLessonID,
		SortOrder:      body.SortOrder,
		AccessLevel:    body.AccessLevel,
		Translation:    toLessonTranslation(body.Translation),
	}
}

func toLessonTranslation(body *LessonTranslationDTO) *entity.LessonTranslation {
	if body == nil {
		return nil
	}
	return &entity.LessonTranslation{
		LanguageCode: body.LanguageCode,
		Title:        body.Title,
		Slug:         body.Slug,
		Content:      body.Content,
	}
}
