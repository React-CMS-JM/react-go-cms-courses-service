// Package lesson applies lesson use cases.
package lesson

const (
	defaultAccess   = "premium"
	defaultSort     = 0
	defaultLanguage = "en"
)

// Service coordinates lesson commands and queries.
type Service struct {
	repository Repository
}

// New builds a lesson service.
func New(repository Repository) *Service {
	return &Service{repository: repository}
}
