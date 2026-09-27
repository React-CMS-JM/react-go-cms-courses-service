package course

import (
	"context"
	"fmt"
	"strings"
	"time"

	"react-go-cms-courses-service/internal/domain/apperror"
	"react-go-cms-courses-service/internal/domain/entity"
)

// Create stores a course and returns it in the translation language.
func (s *Service) Create(ctx context.Context, req entity.CourseCreate, fallbackAuthor string) (entity.Course, error) {
	if req.Translation == nil || strings.TrimSpace(req.Translation.LanguageCode) == "" || strings.TrimSpace(req.Translation.Title) == "" || strings.TrimSpace(req.Translation.Slug) == "" || strings.TrimSpace(req.Translation.Content) == "" {
		return entity.Course{}, fmt.Errorf("create course: %w", apperror.Invalid("translation is required"))
	}
	var typeID int
	var err error
	typeID, err = s.repository.ContentTypeID(ctx)
	if err != nil {
		return entity.Course{}, fmt.Errorf("create course: %w", err)
	}
	author := fallbackAuthor
	if req.AuthorID != nil && strings.TrimSpace(*req.AuthorID) != "" {
		author = strings.TrimSpace(*req.AuthorID)
	}
	if author == "" {
		return entity.Course{}, fmt.Errorf("create course: %w", apperror.Invalid("authorId is required"))
	}
	status := statusDraft
	if req.Status != nil {
		status = *req.Status
	}
	access := accessPublic
	if req.AccessLevel != nil {
		access = *req.AccessLevel
	}
	var publishedAt *time.Time
	if status == StatusPublished {
		now := time.Now().UTC()
		publishedAt = &now
	}
	translation := *req.Translation
	translation.LanguageCode = normalizeLang(translation.LanguageCode)
	var id string
	id, err = s.repository.Insert(ctx, typeID, author, req.FeaturedImageURL, access, status, publishedAt, translation, req.Metadata)
	if err != nil {
		return entity.Course{}, fmt.Errorf("create course: %w", err)
	}
	return s.Get(ctx, id, translation.LanguageCode, true)
}

// Update stores a partial course change and reloads it.
func (s *Service) Update(ctx context.Context, id string, req entity.CourseUpdate, lang string) (entity.Course, error) {
	responseLang := normalizeLang(lang)
	update := req
	if update.Translation != nil {
		translation := *update.Translation
		translation.LanguageCode = normalizeLang(translation.LanguageCode)
		update.Translation = &translation
		responseLang = translation.LanguageCode
	}
	var err error
	err = s.repository.Update(ctx, id, update)
	if err != nil {
		return entity.Course{}, fmt.Errorf("update course: %w", err)
	}
	return s.Get(ctx, id, responseLang, true)
}

// PatchStatus changes publication state and reloads the course.
func (s *Service) PatchStatus(ctx context.Context, id, status, lang string) (entity.Course, error) {
	var err error
	err = s.repository.PatchStatus(ctx, id, status)
	if err != nil {
		return entity.Course{}, fmt.Errorf("patch course status: %w", err)
	}
	return s.Get(ctx, id, normalizeLang(lang), true)
}

// Delete removes a course.
func (s *Service) Delete(ctx context.Context, id string) error {
	var err error
	err = s.repository.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("delete course: %w", err)
	}
	return nil
}

// ReplaceMetadata replaces course metadata and returns the stored rows.
func (s *Service) ReplaceMetadata(ctx context.Context, id string, items []entity.MetadataInput) ([]entity.Metadata, error) {
	var err error
	err = s.repository.ReplaceMetadata(ctx, id, items)
	if err != nil {
		return nil, fmt.Errorf("replace course metadata: %w", err)
	}
	return s.Metadata(ctx, id)
}
