// Package course applies course use cases.
package course

import (
	"context"
	"time"

	"react-go-cms-courses-service/internal/domain/entity"
)

const (
	// StatusPublished is the public course status.
	StatusPublished = "published"
	statusDraft     = "draft"
	accessPublic    = "public"
	defaultLanguage = "en"
	minPageIndex    = 0
	minPageSize     = 1
	maxPageSize     = 100
)

// Repository is the course persistence port.
type Repository interface {
	ContentTypeID(ctx context.Context) (int, error)
	List(ctx context.Context, status string, onlyPublished bool, page, size int) ([]entity.CourseRecord, int64, error)
	Count(ctx context.Context) (int64, error)
	Require(ctx context.Context, id string) (entity.CourseRecord, error)
	AttachDetails(ctx context.Context, record entity.CourseRecord) (entity.CourseRecord, error)
	FindIDBySlug(ctx context.Context, slug, language string) (string, error)
	Insert(ctx context.Context, typeID int, author string, image *string, access, status string, publishedAt *time.Time, translation entity.CourseTranslation, metadata []entity.MetadataInput) (string, error)
	Update(ctx context.Context, id string, req entity.CourseUpdate) error
	PatchStatus(ctx context.Context, id, status string) error
	Delete(ctx context.Context, id string) error
	ListMetadata(ctx context.Context, id string) ([]entity.Metadata, error)
	ReplaceMetadata(ctx context.Context, id string, items []entity.MetadataInput) error
}

// Service coordinates course commands and queries.
type Service struct {
	repository Repository
}

// New builds a course service.
func New(repository Repository) *Service {
	return &Service{repository: repository}
}
