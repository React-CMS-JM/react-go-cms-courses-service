package repo

import (
	"context"
	"database/sql"
	"strings"

	"react-go-cms-courses-service/internal/application/service/lesson"
	"react-go-cms-courses-service/internal/domain/apperror"
	"react-go-cms-courses-service/internal/domain/entity"
	"react-go-cms-courses-service/internal/infrastructure/identity"
	"react-go-cms-courses-service/internal/infrastructure/repository/model"
)

// LessonRepository persists course lessons in MySQL.
type LessonRepository struct {
	db *sql.DB
}

// NewLessonRepository builds a lesson repository.
func NewLessonRepository(db *sql.DB) *LessonRepository {
	return &LessonRepository{db: db}
}

var _ lesson.Repository = (*LessonRepository)(nil)

// EnsureCourse reports whether the course post exists.
func (r *LessonRepository) EnsureCourse(ctx context.Context, courseID string) error {
	_, err := requireCourse(ctx, r.db, courseID)
	return err
}

// ListRows returns lesson rows for a course, without translations.
func (r *LessonRepository) ListRows(ctx context.Context, courseID string) ([]entity.LessonRecord, error) {
	var rows []model.LessonRow
	var err error
	rows, err = r.lessonsOf(ctx, courseID)
	if err != nil {
		return nil, err
	}
	out := make([]entity.LessonRecord, 0, len(rows))
	for _, row := range rows {
		out = append(out, lessonRecord(row))
	}
	return out, nil
}

// RequireLesson loads one lesson that belongs to the course.
func (r *LessonRepository) RequireLesson(ctx context.Context, courseID, lessonID string) (entity.LessonRecord, error) {
	var err error
	err = r.EnsureCourse(ctx, courseID)
	if err != nil {
		return entity.LessonRecord{}, err
	}
	var row model.LessonRow
	row, err = r.lessonByID(ctx, lessonID)
	if err != nil {
		return entity.LessonRecord{}, err
	}
	if row.CourseID != courseID {
		return entity.LessonRecord{}, apperror.NotFound("Lesson not found: " + lessonID)
	}
	return lessonRecord(row), nil
}

// FindByID loads a lesson row by its id.
func (r *LessonRepository) FindByID(ctx context.Context, lessonID string) (entity.LessonRecord, error) {
	var row model.LessonRow
	var err error
	row, err = r.lessonByID(ctx, lessonID)
	if err != nil {
		return entity.LessonRecord{}, err
	}
	return lessonRecord(row), nil
}

// Translations returns every stored lesson translation.
func (r *LessonRepository) Translations(ctx context.Context, lessonID string) ([]entity.LessonTranslation, error) {
	var rows *sql.Rows
	var err error
	rows, err = r.db.QueryContext(ctx, `SELECT language_code, title, slug, content FROM course_lesson_i18n WHERE course_lesson_id = ? ORDER BY id`, lessonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []entity.LessonTranslation
	for rows.Next() {
		var translation entity.LessonTranslation
		err = rows.Scan(&translation.LanguageCode, &translation.Title, &translation.Slug, &translation.Content)
		if err != nil {
			return nil, err
		}
		out = append(out, translation)
	}
	return out, rows.Err()
}

// Insert stores a lesson and its first translation.
func (r *LessonRepository) Insert(ctx context.Context, courseID, parent string, sortOrder int, access string, translation entity.LessonTranslation) (string, error) {
	var id string
	id = identity.NewUUID()
	var parentValue any
	parentValue = blankParent(parent)
	var tx *sql.Tx
	var err error
	tx, err = r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, `
		INSERT INTO course_lessons (id, course_id, parent_lesson_id, sort_order, access_level)
		VALUES (?, ?, ?, ?, ?)`, id, courseID, parentValue, sortOrder, access)
	if err != nil {
		return "", err
	}
	language := normalizeLang(translation.LanguageCode)
	_, err = tx.ExecContext(ctx, `
		INSERT INTO course_lesson_i18n (course_lesson_id, language_code, title, slug, content)
		VALUES (?, ?, ?, ?, ?)`, id, language, translation.Title, translation.Slug, translation.Content)
	if err != nil {
		return "", err
	}
	err = tx.Commit()
	if err != nil {
		return "", err
	}
	return id, nil
}

// ClearParent sets parent_lesson_id to NULL.
func (r *LessonRepository) ClearParent(ctx context.Context, lessonID string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE course_lessons SET parent_lesson_id = NULL, updated_at = UTC_TIMESTAMP() WHERE id = ?`, lessonID)
	return err
}

// SetParent points a lesson at another lesson in the same course.
func (r *LessonRepository) SetParent(ctx context.Context, lessonID, parentID string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE course_lessons SET parent_lesson_id = ?, updated_at = UTC_TIMESTAMP() WHERE id = ?`, parentID, lessonID)
	return err
}

// SetSortOrder updates sort_order.
func (r *LessonRepository) SetSortOrder(ctx context.Context, lessonID string, sortOrder int) error {
	_, err := r.db.ExecContext(ctx, `UPDATE course_lessons SET sort_order = ? WHERE id = ?`, sortOrder, lessonID)
	return err
}

// SetAccessLevel updates access_level.
func (r *LessonRepository) SetAccessLevel(ctx context.Context, lessonID, access string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE course_lessons SET access_level = ? WHERE id = ?`, access, lessonID)
	return err
}

// UpsertTranslation inserts or patches one lesson translation.
func (r *LessonRepository) UpsertTranslation(ctx context.Context, lessonID string, translation *entity.LessonTranslation) error {
	if translation == nil {
		return nil
	}
	language := normalizeLang(translation.LanguageCode)
	var id int
	var err error
	err = r.db.QueryRowContext(ctx, `SELECT id FROM course_lesson_i18n WHERE course_lesson_id = ? AND language_code = ?`, lessonID, language).Scan(&id)
	if err == sql.ErrNoRows {
		_, err = r.db.ExecContext(ctx, `
			INSERT INTO course_lesson_i18n (course_lesson_id, language_code, title, slug, content)
			VALUES (?, ?, ?, ?, ?)`, lessonID, language, translation.Title, translation.Slug, translation.Content)
		return err
	}
	if err != nil {
		return err
	}
	if translation.Title != "" {
		_, err = r.db.ExecContext(ctx, `UPDATE course_lesson_i18n SET title = ? WHERE id = ?`, translation.Title, id)
		if err != nil {
			return err
		}
	}
	if translation.Slug != "" {
		_, err = r.db.ExecContext(ctx, `UPDATE course_lesson_i18n SET slug = ? WHERE id = ?`, translation.Slug, id)
		if err != nil {
			return err
		}
	}
	if translation.Content != "" {
		_, err = r.db.ExecContext(ctx, `UPDATE course_lesson_i18n SET content = ? WHERE id = ?`, translation.Content, id)
		if err != nil {
			return err
		}
	}
	return nil
}

// TouchUpdatedAt sets updated_at, ignoring a failed write the same way as before.
func (r *LessonRepository) TouchUpdatedAt(ctx context.Context, lessonID string) {
	_, _ = r.db.ExecContext(ctx, `UPDATE course_lessons SET updated_at = UTC_TIMESTAMP() WHERE id = ?`, lessonID)
}

// Delete removes a lesson row.
func (r *LessonRepository) Delete(ctx context.Context, lessonID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM course_lessons WHERE id = ?`, lessonID)
	return err
}

func (r *LessonRepository) lessonsOf(ctx context.Context, courseID string) ([]model.LessonRow, error) {
	var rows *sql.Rows
	var err error
	rows, err = r.db.QueryContext(ctx, `
		SELECT id, course_id, parent_lesson_id, sort_order, access_level, created_at, updated_at
		FROM course_lessons WHERE course_id = ? ORDER BY sort_order, id`, courseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.LessonRow
	for rows.Next() {
		var row model.LessonRow
		err = rows.Scan(&row.ID, &row.CourseID, &row.ParentLessonID, &row.SortOrder, &row.AccessLevel, &row.CreatedAt, &row.UpdatedAt)
		if err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (r *LessonRepository) lessonByID(ctx context.Context, id string) (model.LessonRow, error) {
	var row model.LessonRow
	var err error
	err = r.db.QueryRowContext(ctx, `
		SELECT id, course_id, parent_lesson_id, sort_order, access_level, created_at, updated_at
		FROM course_lessons WHERE id = ?`, id).Scan(&row.ID, &row.CourseID, &row.ParentLessonID, &row.SortOrder, &row.AccessLevel, &row.CreatedAt, &row.UpdatedAt)
	if err == sql.ErrNoRows {
		return row, apperror.NotFound("Lesson not found: " + id)
	}
	return row, err
}

func lessonRecord(row model.LessonRow) entity.LessonRecord {
	return entity.LessonRecord{
		ID:             row.ID,
		CourseID:       row.CourseID,
		ParentLessonID: StrPtr(row.ParentLessonID),
		SortOrder:      row.SortOrder,
		AccessLevel:    row.AccessLevel,
		CreatedAt:      InstantFrom(row.CreatedAt),
		UpdatedAt:      InstantFrom(row.UpdatedAt),
	}
}

func blankParent(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return strings.TrimSpace(value)
}
