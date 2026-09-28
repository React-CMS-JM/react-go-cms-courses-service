package course

import (
	"context"
	"time"

	"react-go-cms-courses-service/internal/domain/entity"
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
