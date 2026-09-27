// Package entity holds courses-service domain objects without transport tags.
package entity

import "time"

// Page is an offset page of items.
type Page[T any] struct {
	Items []T
	Page  int
	Size  int
	Total int64
}

// Metadata is one post metadata row.
type Metadata struct {
	ID        string
	PostID    string
	MetaKey   string
	MetaValue *string
}

// MetadataInput is a metadata write payload.
type MetadataInput struct {
	MetaKey   string
	MetaValue *string
}

// CourseTranslation is one localized course text.
type CourseTranslation struct {
	LanguageCode    string
	Title           string
	Slug            string
	Content         string
	Excerpt         *string
	MetaTitle       *string
	MetaDescription *string
}

// CourseRecord is a stored course plus every translation, before language selection.
type CourseRecord struct {
	ID               string
	AuthorID         string
	ContentTypeID    int
	FeaturedImageURL *string
	AccessLevel      string
	Status           string
	ViewCount        int
	PublishedAt      *time.Time
	CreatedAt        *time.Time
	UpdatedAt        *time.Time
	Translations     []CourseTranslation
	Metadata         []Metadata
	LessonCount      int
}

// Course is a course localized to one language.
type Course struct {
	ID               string
	AuthorID         string
	ContentTypeID    int
	FeaturedImageURL *string
	AccessLevel      string
	Status           string
	ViewCount        int
	PublishedAt      *time.Time
	CategoryIDs      []int
	TagIDs           []int
	CreatedAt        *time.Time
	UpdatedAt        *time.Time
	LanguageCode     string
	Title            string
	Slug             string
	Content          string
	Excerpt          *string
	MetaTitle        *string
	MetaDescription  *string
	Metadata         []Metadata
	LessonCount      int
}

// CourseStats is the course count for the admin dashboard.
type CourseStats struct {
	Total int64
}

// CourseCreate is the data required to insert a course.
type CourseCreate struct {
	AuthorID         *string
	FeaturedImageURL *string
	AccessLevel      *string
	Status           *string
	Translation      *CourseTranslation
	Metadata         []MetadataInput
}

// CourseUpdate is a partial course update.
type CourseUpdate struct {
	FeaturedImageURL *string
	AccessLevel      *string
	Status           *string
	Translation      *CourseTranslation
}
