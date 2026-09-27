package lesson

import (
	"context"
	"fmt"
	"strings"

	"react-go-cms-courses-service/internal/domain/apperror"
	"react-go-cms-courses-service/internal/domain/entity"
)

// List returns localized lessons for a course.
func (s *Service) List(ctx context.Context, courseID, lang string) ([]entity.Lesson, error) {
	var err error
	err = s.repository.EnsureCourse(ctx, courseID)
	if err != nil {
		return nil, fmt.Errorf("list lessons: %w", err)
	}
	var rows []entity.LessonRecord
	rows, err = s.repository.ListRows(ctx, courseID)
	if err != nil {
		return nil, fmt.Errorf("list lessons: %w", err)
	}
	out := []entity.Lesson{}
	for _, row := range rows {
		var item entity.Lesson
		var ok bool
		item, ok, err = s.localize(ctx, row, normalizeLang(lang))
		if err != nil {
			return nil, fmt.Errorf("list lessons: %w", err)
		}
		if ok {
			out = append(out, item)
		}
	}
	return out, nil
}

// Get returns one lesson in a course.
func (s *Service) Get(ctx context.Context, courseID, lessonID, lang string) (entity.Lesson, error) {
	var row entity.LessonRecord
	var err error
	row, err = s.repository.RequireLesson(ctx, courseID, lessonID)
	if err != nil {
		return entity.Lesson{}, fmt.Errorf("get lesson: %w", err)
	}
	var item entity.Lesson
	var ok bool
	item, ok, err = s.localize(ctx, row, normalizeLang(lang))
	if err != nil {
		return entity.Lesson{}, fmt.Errorf("get lesson: %w", err)
	}
	if !ok {
		return entity.Lesson{}, fmt.Errorf("get lesson: %w", apperror.NotFound("Lesson translation not found: "+lessonID))
	}
	return item, nil
}

// GetBySlug returns a lesson whose translation slug matches.
func (s *Service) GetBySlug(ctx context.Context, courseID, slug, lang string) (entity.Lesson, error) {
	var err error
	err = s.repository.EnsureCourse(ctx, courseID)
	if err != nil {
		return entity.Lesson{}, fmt.Errorf("get lesson by slug: %w", err)
	}
	language := normalizeLang(lang)
	var rows []entity.LessonRecord
	rows, err = s.repository.ListRows(ctx, courseID)
	if err != nil {
		return entity.Lesson{}, fmt.Errorf("get lesson by slug: %w", err)
	}
	for _, row := range rows {
		var item entity.Lesson
		var ok bool
		item, ok, err = s.localize(ctx, row, language)
		if err != nil {
			return entity.Lesson{}, fmt.Errorf("get lesson by slug: %w", err)
		}
		if ok && item.Slug == slug {
			return item, nil
		}
	}
	for _, row := range rows {
		var translations []entity.LessonTranslation
		translations, err = s.repository.Translations(ctx, row.ID)
		if err != nil {
			return entity.Lesson{}, fmt.Errorf("get lesson by slug: %w", err)
		}
		match := false
		for _, translation := range translations {
			if translation.Slug == slug {
				match = true
			}
		}
		if match {
			var item entity.Lesson
			var ok bool
			item, ok, err = s.localize(ctx, row, language)
			if err != nil {
				return entity.Lesson{}, fmt.Errorf("get lesson by slug: %w", err)
			}
			if ok {
				return item, nil
			}
		}
	}
	return entity.Lesson{}, fmt.Errorf("get lesson by slug: %w", apperror.NotFound("Lesson not found for slug: "+slug))
}

func (s *Service) localize(ctx context.Context, row entity.LessonRecord, language string) (entity.Lesson, bool, error) {
	var translations []entity.LessonTranslation
	var err error
	translations, err = s.repository.Translations(ctx, row.ID)
	if err != nil {
		return entity.Lesson{}, false, err
	}
	if len(translations) == 0 {
		return entity.Lesson{}, false, nil
	}
	chosen := translations[0]
	found := false
	for _, translation := range translations {
		if translation.LanguageCode == language {
			chosen = translation
			found = true
			break
		}
	}
	if !found {
		for _, translation := range translations {
			if translation.LanguageCode == defaultLanguage {
				chosen = translation
				break
			}
		}
	}
	return entity.Lesson{
		ID:             row.ID,
		CourseID:       row.CourseID,
		ParentLessonID: row.ParentLessonID,
		SortOrder:      row.SortOrder,
		AccessLevel:    row.AccessLevel,
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
		LanguageCode:   chosen.LanguageCode,
		Title:          chosen.Title,
		Slug:           chosen.Slug,
		Content:        chosen.Content,
	}, true, nil
}

func normalizeLang(lang string) string {
	lang = strings.TrimSpace(strings.ToLower(lang))
	if lang == "" {
		return defaultLanguage
	}
	return lang
}
