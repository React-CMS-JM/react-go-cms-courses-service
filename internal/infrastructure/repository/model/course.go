// Package model holds database row shapes for the courses service.
package model

import "database/sql"

// CourseRow is one posts row whose content type is course.
type CourseRow struct {
	ID               string         `db:"id"`
	AuthorID         string         `db:"author_id"`
	ContentTypeID    int            `db:"content_type_id"`
	FeaturedImageURL sql.NullString `db:"featured_image_url"`
	AccessLevel      string         `db:"access_level"`
	Status           string         `db:"status"`
	ViewCount        int            `db:"view_count"`
	PublishedAt      sql.NullTime   `db:"published_at"`
	CreatedAt        sql.NullTime   `db:"created_at"`
	UpdatedAt        sql.NullTime   `db:"updated_at"`
}

// LessonRow is one course_lessons row.
type LessonRow struct {
	ID             string         `db:"id"`
	CourseID       string         `db:"course_id"`
	ParentLessonID sql.NullString `db:"parent_lesson_id"`
	SortOrder      int            `db:"sort_order"`
	AccessLevel    string         `db:"access_level"`
	CreatedAt      sql.NullTime   `db:"created_at"`
	UpdatedAt      sql.NullTime   `db:"updated_at"`
}
