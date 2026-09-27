// Package lesson applies lesson use cases.
package lesson

import (
	"context"

	"react-go-cms-courses-service/internal/domain/entity"
)

const (
	defaultAccess   = "premium"
	defaultSort     = 0
	defaultLanguage = "en"
)

// Repository is the lesson persistence port.
type Repository interface {
	EnsureCourse(ctx context.Context, courseID string) error
	ListRows(ctx context.Context, courseID string) ([]entity.LessonRecord, error)
	RequireLesson(ctx context.Context, courseID, lessonID string) (entity.LessonRecord, error)
	FindByID(ctx context.Context, lessonID string) (entity.LessonRecord, error)
	Translations(ctx context.Context, lessonID string) ([]entity.LessonTranslation, error)
	Insert(ctx context.Context, courseID, parent string, sortOrder int, access string, translation entity.LessonTranslation) (string, error)
	ClearParent(ctx context.Context, lessonID string) error
	SetParent(ctx context.Context, lessonID, parentID string) error
	SetSortOrder(ctx context.Context, lessonID string, sortOrder int) error
	SetAccessLevel(ctx context.Context, lessonID, access string) error
	UpsertTranslation(ctx context.Context, lessonID string, translation *entity.LessonTranslation) error
	TouchUpdatedAt(ctx context.Context, lessonID string)
	Delete(ctx context.Context, lessonID string) error
}

// Service coordinates lesson commands and queries.
type Service struct {
	repository Repository
}

// New builds a lesson service.
func New(repository Repository) *Service {
	return &Service{repository: repository}
}
