package lesson

import (
	"context"
	"fmt"
	"strings"

	"react-go-cms-courses-service/internal/domain/apperror"
	"react-go-cms-courses-service/internal/domain/entity"
)

// Create stores a lesson and returns the localized row.
func (s *Service) Create(ctx context.Context, courseID string, req entity.LessonCreate) (entity.Lesson, error) {
	var err error
	err = s.repository.EnsureCourse(ctx, courseID)
	if err != nil {
		return entity.Lesson{}, fmt.Errorf("create lesson: %w", err)
	}
	if req.Translation == nil || strings.TrimSpace(req.Translation.LanguageCode) == "" || strings.TrimSpace(req.Translation.Title) == "" || strings.TrimSpace(req.Translation.Slug) == "" || strings.TrimSpace(req.Translation.Content) == "" {
		return entity.Lesson{}, fmt.Errorf("create lesson: %w", apperror.Invalid("translation is required"))
	}
	parent := ""
	if req.ParentLessonID != nil && strings.TrimSpace(*req.ParentLessonID) != "" {
		parent = strings.TrimSpace(*req.ParentLessonID)
		_, err = s.repository.RequireLesson(ctx, courseID, parent)
		if err != nil {
			return entity.Lesson{}, fmt.Errorf("create lesson: %w", err)
		}
	}
	order := defaultSort
	if req.SortOrder != nil {
		order = *req.SortOrder
	}
	access := defaultAccess
	if req.AccessLevel != nil {
		access = *req.AccessLevel
	}
	translation := *req.Translation
	translation.LanguageCode = normalizeLang(translation.LanguageCode)
	var id string
	id, err = s.repository.Insert(ctx, courseID, parent, order, access, translation)
	if err != nil {
		return entity.Lesson{}, fmt.Errorf("create lesson: %w", err)
	}
	return s.Get(ctx, courseID, id, translation.LanguageCode)
}

// Update stores a partial lesson change inside a course.
func (s *Service) Update(ctx context.Context, courseID, lessonID string, req entity.LessonUpdate, lang string) (entity.Lesson, error) {
	var err error
	_, err = s.repository.RequireLesson(ctx, courseID, lessonID)
	if err != nil {
		return entity.Lesson{}, fmt.Errorf("update lesson: %w", err)
	}
	return s.applyUpdate(ctx, courseID, lessonID, req, lang)
}

// UpdateByID stores a partial lesson change looked up by lesson id.
func (s *Service) UpdateByID(ctx context.Context, lessonID string, req entity.LessonUpdate, lang string) (entity.Lesson, error) {
	var row entity.LessonRecord
	var err error
	row, err = s.repository.FindByID(ctx, lessonID)
	if err != nil {
		return entity.Lesson{}, fmt.Errorf("update lesson: %w", err)
	}
	err = s.repository.EnsureCourse(ctx, row.CourseID)
	if err != nil {
		return entity.Lesson{}, fmt.Errorf("update lesson: %w", err)
	}
	return s.applyUpdate(ctx, row.CourseID, lessonID, req, lang)
}

// Delete removes a lesson from a course.
func (s *Service) Delete(ctx context.Context, courseID, lessonID string) error {
	var err error
	_, err = s.repository.RequireLesson(ctx, courseID, lessonID)
	if err != nil {
		return fmt.Errorf("delete lesson: %w", err)
	}
	err = s.repository.Delete(ctx, lessonID)
	if err != nil {
		return fmt.Errorf("delete lesson: %w", err)
	}
	return nil
}

// DeleteByID removes a lesson looked up by lesson id.
func (s *Service) DeleteByID(ctx context.Context, lessonID string) error {
	var row entity.LessonRecord
	var err error
	row, err = s.repository.FindByID(ctx, lessonID)
	if err != nil {
		return fmt.Errorf("delete lesson: %w", err)
	}
	err = s.repository.EnsureCourse(ctx, row.CourseID)
	if err != nil {
		return fmt.Errorf("delete lesson: %w", err)
	}
	err = s.repository.Delete(ctx, lessonID)
	if err != nil {
		return fmt.Errorf("delete lesson: %w", err)
	}
	return nil
}

func (s *Service) applyUpdate(ctx context.Context, courseID, lessonID string, req entity.LessonUpdate, lang string) (entity.Lesson, error) {
	var err error
	if req.ParentLessonID != nil {
		parent := strings.TrimSpace(*req.ParentLessonID)
		if parent == "" {
			err = s.repository.ClearParent(ctx, lessonID)
			if err != nil {
				return entity.Lesson{}, fmt.Errorf("update lesson: %w", err)
			}
		} else {
			if parent == lessonID {
				return entity.Lesson{}, fmt.Errorf("update lesson: %w", apperror.Invalid("Lesson cannot be its own parent"))
			}
			_, err = s.repository.RequireLesson(ctx, courseID, parent)
			if err != nil {
				return entity.Lesson{}, fmt.Errorf("update lesson: %w", err)
			}
			err = s.repository.SetParent(ctx, lessonID, parent)
			if err != nil {
				return entity.Lesson{}, fmt.Errorf("update lesson: %w", err)
			}
		}
	}
	if req.SortOrder != nil {
		err = s.repository.SetSortOrder(ctx, lessonID, *req.SortOrder)
		if err != nil {
			return entity.Lesson{}, fmt.Errorf("update lesson: %w", err)
		}
	}
	if req.AccessLevel != nil {
		err = s.repository.SetAccessLevel(ctx, lessonID, *req.AccessLevel)
		if err != nil {
			return entity.Lesson{}, fmt.Errorf("update lesson: %w", err)
		}
	}
	responseLang := normalizeLang(lang)
	if req.Translation != nil {
		translation := *req.Translation
		translation.LanguageCode = normalizeLang(translation.LanguageCode)
		err = s.repository.UpsertTranslation(ctx, lessonID, &translation)
		if err != nil {
			return entity.Lesson{}, fmt.Errorf("update lesson: %w", err)
		}
		responseLang = translation.LanguageCode
	}
	s.repository.TouchUpdatedAt(ctx, lessonID)
	return s.Get(ctx, courseID, lessonID, responseLang)
}
