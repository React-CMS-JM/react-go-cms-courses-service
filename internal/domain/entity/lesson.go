package entity

import "time"

// LessonTranslation is one localized lesson text.
type LessonTranslation struct {
	LanguageCode string
	Title        string
	Slug         string
	Content      string
}

// LessonRecord is a stored lesson plus every translation, before language selection.
type LessonRecord struct {
	ID             string
	CourseID       string
	ParentLessonID *string
	SortOrder      int
	AccessLevel    string
	CreatedAt      *time.Time
	UpdatedAt      *time.Time
	Translations   []LessonTranslation
}

// Lesson is a lesson localized to one language.
type Lesson struct {
	ID             string
	CourseID       string
	ParentLessonID *string
	SortOrder      int
	AccessLevel    string
	CreatedAt      *time.Time
	UpdatedAt      *time.Time
	LanguageCode   string
	Title          string
	Slug           string
	Content        string
}

// LessonCreate is the data required to insert a lesson.
type LessonCreate struct {
	ParentLessonID *string
	SortOrder      *int
	AccessLevel    *string
	Translation    *LessonTranslation
}

// LessonUpdate is a partial lesson update.
type LessonUpdate struct {
	ParentLessonID *string
	SortOrder      *int
	AccessLevel    *string
	Translation    *LessonTranslation
}
