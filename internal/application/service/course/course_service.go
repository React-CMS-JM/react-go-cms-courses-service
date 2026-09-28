// Package course applies course use cases.
package course

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

// Service coordinates course commands and queries.
type Service struct {
	repository Repository
}

// New builds a course service.
func New(repository Repository) *Service {
	return &Service{repository: repository}
}
