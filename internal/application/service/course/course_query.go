package course

import (
	"context"
	"fmt"
	"strings"

	"react-go-cms-courses-service/internal/domain/apperror"
	"react-go-cms-courses-service/internal/domain/entity"
)

// List returns one page of courses visible to the caller.
func (s *Service) List(ctx context.Context, status, lang string, includeAll bool, page, size int) (entity.Page[entity.Course], error) {
	if page < minPageIndex {
		page = minPageIndex
	}
	if size < minPageSize {
		size = minPageSize
	}
	if size > maxPageSize {
		size = maxPageSize
	}
	filter := strings.TrimSpace(status)
	onlyPublished := filter == "" && !includeAll
	var records []entity.CourseRecord
	var total int64
	var err error
	records, total, err = s.repository.List(ctx, filter, onlyPublished, page, size)
	if err != nil {
		return entity.Page[entity.Course]{}, fmt.Errorf("list courses: %w", err)
	}
	items := []entity.Course{}
	for _, record := range records {
		course, ok := localizeCourse(record, normalizeLang(lang))
		if ok {
			items = append(items, course)
		}
	}
	return entity.Page[entity.Course]{Items: items, Page: page, Size: size, Total: total}, nil
}

// Stats returns the course count.
func (s *Service) Stats(ctx context.Context) (entity.CourseStats, error) {
	var total int64
	var err error
	total, err = s.repository.Count(ctx)
	if err != nil {
		return entity.CourseStats{}, fmt.Errorf("course stats: %w", err)
	}
	return entity.CourseStats{Total: total}, nil
}

// Get returns one course when the caller may see its status.
func (s *Service) Get(ctx context.Context, id, lang string, includeAll bool) (entity.Course, error) {
	var record entity.CourseRecord
	var err error
	record, err = s.repository.Require(ctx, id)
	if err != nil {
		return entity.Course{}, fmt.Errorf("get course: %w", err)
	}
	if !includeAll && record.Status != StatusPublished {
		return entity.Course{}, fmt.Errorf("get course: %w", apperror.NotFound("Course not found: "+id))
	}
	record, err = s.repository.AttachDetails(ctx, record)
	if err != nil {
		return entity.Course{}, fmt.Errorf("get course: %w", err)
	}
	course, ok := localizeCourse(record, normalizeLang(lang))
	if !ok {
		return entity.Course{}, fmt.Errorf("get course: %w", apperror.NotFound("Course translation not found: "+id))
	}
	return course, nil
}

// GetBySlug returns one course by translation slug.
func (s *Service) GetBySlug(ctx context.Context, slug, lang string, includeAll bool) (entity.Course, error) {
	language := normalizeLang(lang)
	var postID string
	var err error
	postID, err = s.repository.FindIDBySlug(ctx, slug, language)
	if err != nil {
		return entity.Course{}, fmt.Errorf("get course by slug: %w", err)
	}
	var record entity.CourseRecord
	record, err = s.repository.Require(ctx, postID)
	if err != nil {
		return entity.Course{}, fmt.Errorf("get course by slug: %w", apperror.NotFound("Course not found for slug: "+slug))
	}
	if !includeAll && record.Status != StatusPublished {
		return entity.Course{}, fmt.Errorf("get course by slug: %w", apperror.NotFound("Course not found for slug: "+slug))
	}
	record, err = s.repository.AttachDetails(ctx, record)
	if err != nil {
		return entity.Course{}, fmt.Errorf("get course by slug: %w", apperror.NotFound("Course translation not found for slug: "+slug))
	}
	course, ok := localizeCourse(record, language)
	if !ok {
		return entity.Course{}, fmt.Errorf("get course by slug: %w", apperror.NotFound("Course translation not found for slug: "+slug))
	}
	return course, nil
}

// Metadata returns course metadata.
func (s *Service) Metadata(ctx context.Context, id string) ([]entity.Metadata, error) {
	var rows []entity.Metadata
	var err error
	rows, err = s.repository.ListMetadata(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("course metadata: %w", err)
	}
	return rows, nil
}

func localizeCourse(record entity.CourseRecord, language string) (entity.Course, bool) {
	chosen := pickCourseTranslation(record.Translations, language)
	if chosen == nil {
		return entity.Course{}, false
	}
	metadata := record.Metadata
	if metadata == nil {
		metadata = []entity.Metadata{}
	}
	return entity.Course{
		ID:               record.ID,
		AuthorID:         record.AuthorID,
		ContentTypeID:    record.ContentTypeID,
		FeaturedImageURL: record.FeaturedImageURL,
		AccessLevel:      record.AccessLevel,
		Status:           record.Status,
		ViewCount:        record.ViewCount,
		PublishedAt:      record.PublishedAt,
		CategoryIDs:      []int{},
		TagIDs:           []int{},
		CreatedAt:        record.CreatedAt,
		UpdatedAt:        record.UpdatedAt,
		LanguageCode:     chosen.LanguageCode,
		Title:            chosen.Title,
		Slug:             chosen.Slug,
		Content:          chosen.Content,
		Excerpt:          chosen.Excerpt,
		MetaTitle:        chosen.MetaTitle,
		MetaDescription:  chosen.MetaDescription,
		Metadata:         metadata,
		LessonCount:      record.LessonCount,
	}, true
}

func pickCourseTranslation(rows []entity.CourseTranslation, language string) *entity.CourseTranslation {
	if len(rows) == 0 {
		return nil
	}
	for index := range rows {
		if rows[index].LanguageCode == language {
			return &rows[index]
		}
	}
	for index := range rows {
		if rows[index].LanguageCode == defaultLanguage {
			return &rows[index]
		}
	}
	return &rows[0]
}

func normalizeLang(lang string) string {
	lang = strings.TrimSpace(strings.ToLower(lang))
	if lang == "" {
		return defaultLanguage
	}
	return lang
}
