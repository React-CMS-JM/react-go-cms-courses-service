package handler

import "react-go-cms-courses-service/internal/domain/entity"

// CoursePage is a JSON page of courses.
type CoursePage struct {
	Items []CourseResponse `json:"items"`
	Page  int              `json:"page"`
	Size  int              `json:"size"`
	Total int64            `json:"total"`
}

// CourseResponse is the public course payload.
type CourseResponse struct {
	ID               string        `json:"id"`
	AuthorID         string        `json:"authorId"`
	ContentTypeID    int           `json:"contentTypeId"`
	FeaturedImageURL *string       `json:"featuredImageUrl"`
	AccessLevel      string        `json:"accessLevel"`
	Status           string        `json:"status"`
	ViewCount        int           `json:"viewCount"`
	PublishedAt      InstantJSON   `json:"publishedAt"`
	CategoryIDs      []int         `json:"categoryIds"`
	TagIDs           []int         `json:"tagIds"`
	CreatedAt        InstantJSON   `json:"createdAt"`
	UpdatedAt        InstantJSON   `json:"updatedAt"`
	LanguageCode     string        `json:"languageCode"`
	Title            string        `json:"title"`
	Slug             string        `json:"slug"`
	Content          string        `json:"content"`
	Excerpt          *string       `json:"excerpt"`
	MetaTitle        *string       `json:"metaTitle"`
	MetaDescription  *string       `json:"metaDescription"`
	Metadata         []MetadataDTO `json:"metadata"`
	LessonCount      int           `json:"lessonCount"`
}

// MetadataDTO is one metadata object in a course response.
type MetadataDTO struct {
	ID        string  `json:"id"`
	PostID    string  `json:"postId"`
	MetaKey   string  `json:"metaKey"`
	MetaValue *string `json:"metaValue"`
}

// CourseStatsResponse is the stats payload.
type CourseStatsResponse struct {
	Total int64 `json:"total"`
}

// CourseTranslationDTO is a course translation in a write body.
type CourseTranslationDTO struct {
	LanguageCode    string  `json:"languageCode"`
	Title           string  `json:"title"`
	Slug            string  `json:"slug"`
	Content         string  `json:"content"`
	Excerpt         *string `json:"excerpt"`
	MetaTitle       *string `json:"metaTitle"`
	MetaDescription *string `json:"metaDescription"`
}

// MetaInDTO is one metadata entry in a write body.
type MetaInDTO struct {
	MetaKey   string  `json:"metaKey"`
	MetaValue *string `json:"metaValue"`
}

// CreateCourseDTO is the create-course body.
type CreateCourseDTO struct {
	AuthorID         *string               `json:"authorId"`
	FeaturedImageURL *string               `json:"featuredImageUrl"`
	AccessLevel      *string               `json:"accessLevel"`
	Status           *string               `json:"status"`
	Translation      *CourseTranslationDTO `json:"translation"`
	Metadata         []MetaInDTO           `json:"metadata"`
}

// UpdateCourseDTO is the update-course body.
type UpdateCourseDTO struct {
	FeaturedImageURL *string               `json:"featuredImageUrl"`
	AccessLevel      *string               `json:"accessLevel"`
	Status           *string               `json:"status"`
	Translation      *CourseTranslationDTO `json:"translation"`
}

// StatusDTO is the patch-status body.
type StatusDTO struct {
	Status string `json:"status"`
}

func toCourseResponse(course entity.Course) CourseResponse {
	return CourseResponse{
		ID:               course.ID,
		AuthorID:         course.AuthorID,
		ContentTypeID:    course.ContentTypeID,
		FeaturedImageURL: course.FeaturedImageURL,
		AccessLevel:      course.AccessLevel,
		Status:           course.Status,
		ViewCount:        course.ViewCount,
		PublishedAt:      toInstant(course.PublishedAt),
		CategoryIDs:      emptyInts(course.CategoryIDs),
		TagIDs:           emptyInts(course.TagIDs),
		CreatedAt:        toInstant(course.CreatedAt),
		UpdatedAt:        toInstant(course.UpdatedAt),
		LanguageCode:     course.LanguageCode,
		Title:            course.Title,
		Slug:             course.Slug,
		Content:          course.Content,
		Excerpt:          course.Excerpt,
		MetaTitle:        course.MetaTitle,
		MetaDescription:  course.MetaDescription,
		Metadata:         toMetadataDTOs(course.Metadata),
		LessonCount:      course.LessonCount,
	}
}

func toCoursePage(page entity.Page[entity.Course]) CoursePage {
	items := make([]CourseResponse, 0, len(page.Items))
	for _, item := range page.Items {
		items = append(items, toCourseResponse(item))
	}
	return CoursePage{Items: items, Page: page.Page, Size: page.Size, Total: page.Total}
}

func toMetadataDTOs(rows []entity.Metadata) []MetadataDTO {
	if rows == nil {
		return []MetadataDTO{}
	}
	out := make([]MetadataDTO, 0, len(rows))
	for _, row := range rows {
		out = append(out, MetadataDTO{ID: row.ID, PostID: row.PostID, MetaKey: row.MetaKey, MetaValue: row.MetaValue})
	}
	return out
}

func emptyInts(values []int) []int {
	if values == nil {
		return []int{}
	}
	return values
}

func toCourseCreate(body CreateCourseDTO) entity.CourseCreate {
	return entity.CourseCreate{
		AuthorID:         body.AuthorID,
		FeaturedImageURL: body.FeaturedImageURL,
		AccessLevel:      body.AccessLevel,
		Status:           body.Status,
		Translation:      toCourseTranslation(body.Translation),
		Metadata:         toMetaInputs(body.Metadata),
	}
}

func toCourseUpdate(body UpdateCourseDTO) entity.CourseUpdate {
	return entity.CourseUpdate{
		FeaturedImageURL: body.FeaturedImageURL,
		AccessLevel:      body.AccessLevel,
		Status:           body.Status,
		Translation:      toCourseTranslation(body.Translation),
	}
}

func toCourseTranslation(body *CourseTranslationDTO) *entity.CourseTranslation {
	if body == nil {
		return nil
	}
	return &entity.CourseTranslation{
		LanguageCode:    body.LanguageCode,
		Title:           body.Title,
		Slug:            body.Slug,
		Content:         body.Content,
		Excerpt:         body.Excerpt,
		MetaTitle:       body.MetaTitle,
		MetaDescription: body.MetaDescription,
	}
}

func toMetaInputs(rows []MetaInDTO) []entity.MetadataInput {
	if rows == nil {
		return nil
	}
	out := make([]entity.MetadataInput, 0, len(rows))
	for _, row := range rows {
		out = append(out, entity.MetadataInput{MetaKey: row.MetaKey, MetaValue: row.MetaValue})
	}
	return out
}
