package courses

import (
	"context"
	"database/sql"
	"strings"

	"react-go-cms-courses-service/internal/platform"
)

type LessonTranslation struct {
	LanguageCode string `json:"languageCode"`
	Title        string `json:"title"`
	Slug         string `json:"slug"`
	Content      string `json:"content"`
}

type CreateLesson struct {
	ParentLessonID *string            `json:"parentLessonId"`
	SortOrder      *int               `json:"sortOrder"`
	AccessLevel    *string            `json:"accessLevel"`
	Translation    *LessonTranslation `json:"translation"`
}

type UpdateLesson struct {
	ParentLessonID *string            `json:"parentLessonId"`
	SortOrder      *int               `json:"sortOrder"`
	AccessLevel    *string            `json:"accessLevel"`
	Translation    *LessonTranslation `json:"translation"`
}

func (s *Store) ListLessons(ctx context.Context, courseID, lang string) ([]Lesson, error) {
	if _, err := s.requireCourse(ctx, courseID); err != nil {
		return nil, err
	}
	rows, err := s.lessonsOf(ctx, courseID)
	if err != nil {
		return nil, err
	}
	out := []Lesson{}
	for _, row := range rows {
		item, ok, err := s.localizeLesson(ctx, row, normalizeLang(lang))
		if err != nil {
			return nil, err
		}
		if ok {
			out = append(out, item)
		}
	}
	return out, nil
}

func (s *Store) GetLesson(ctx context.Context, courseID, lessonID, lang string) (Lesson, error) {
	row, err := s.requireLesson(ctx, courseID, lessonID)
	if err != nil {
		return Lesson{}, err
	}
	item, ok, err := s.localizeLesson(ctx, row, normalizeLang(lang))
	if err != nil {
		return Lesson{}, err
	}
	if !ok {
		return Lesson{}, missing("Lesson translation not found: " + lessonID)
	}
	return item, nil
}

func (s *Store) GetLessonByID(ctx context.Context, lessonID, lang string) (Lesson, error) {
	row, err := s.lessonByID(ctx, lessonID)
	if err != nil {
		return Lesson{}, err
	}
	if _, err := s.requireCourse(ctx, row.courseID); err != nil {
		return Lesson{}, err
	}
	item, ok, err := s.localizeLesson(ctx, row, normalizeLang(lang))
	if err != nil || !ok {
		return Lesson{}, missing("Lesson translation not found: " + lessonID)
	}
	return item, nil
}

func (s *Store) GetLessonBySlug(ctx context.Context, courseID, slug, lang string) (Lesson, error) {
	if _, err := s.requireCourse(ctx, courseID); err != nil {
		return Lesson{}, err
	}
	language := normalizeLang(lang)
	rows, err := s.lessonsOf(ctx, courseID)
	if err != nil {
		return Lesson{}, err
	}
	for _, row := range rows {
		item, ok, err := s.localizeLesson(ctx, row, language)
		if err != nil {
			return Lesson{}, err
		}
		if ok && item.Slug == slug {
			return item, nil
		}
	}
	for _, row := range rows {
		trs, err := s.lessonI18n(ctx, row.id)
		if err != nil {
			return Lesson{}, err
		}
		match := false
		for _, t := range trs {
			if t.Slug == slug {
				match = true
			}
		}
		if match {
			item, ok, err := s.localizeLesson(ctx, row, language)
			if err != nil {
				return Lesson{}, err
			}
			if ok {
				return item, nil
			}
		}
	}
	return Lesson{}, missing("Lesson not found for slug: " + slug)
}

func (s *Store) CreateLesson(ctx context.Context, courseID string, req CreateLesson) (Lesson, error) {
	if _, err := s.requireCourse(ctx, courseID); err != nil {
		return Lesson{}, err
	}
	if req.Translation == nil || strings.TrimSpace(req.Translation.LanguageCode) == "" || strings.TrimSpace(req.Translation.Title) == "" || strings.TrimSpace(req.Translation.Slug) == "" || strings.TrimSpace(req.Translation.Content) == "" {
		return Lesson{}, bad("translation is required")
	}
	if req.ParentLessonID != nil && strings.TrimSpace(*req.ParentLessonID) != "" {
		if _, err := s.requireLesson(ctx, courseID, strings.TrimSpace(*req.ParentLessonID)); err != nil {
			return Lesson{}, err
		}
	}
	id := platform.NewUUID()
	parent := blankPtr(req.ParentLessonID)
	order := 0
	if req.SortOrder != nil {
		order = *req.SortOrder
	}
	access := "premium"
	if req.AccessLevel != nil {
		access = *req.AccessLevel
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Lesson{}, err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO course_lessons (id, course_id, parent_lesson_id, sort_order, access_level)
		VALUES (?, ?, ?, ?, ?)`, id, courseID, parent, order, access); err != nil {
		return Lesson{}, err
	}
	lang := normalizeLang(req.Translation.LanguageCode)
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO course_lesson_i18n (course_lesson_id, language_code, title, slug, content)
		VALUES (?, ?, ?, ?, ?)`, id, lang, req.Translation.Title, req.Translation.Slug, req.Translation.Content); err != nil {
		return Lesson{}, err
	}
	if err := tx.Commit(); err != nil {
		return Lesson{}, err
	}
	return s.GetLesson(ctx, courseID, id, lang)
}

func (s *Store) UpdateLesson(ctx context.Context, courseID, lessonID string, req UpdateLesson, lang string) (Lesson, error) {
	if _, err := s.requireLesson(ctx, courseID, lessonID); err != nil {
		return Lesson{}, err
	}
	return s.applyLessonUpdate(ctx, courseID, lessonID, req, lang)
}

func (s *Store) UpdateLessonByID(ctx context.Context, lessonID string, req UpdateLesson, lang string) (Lesson, error) {
	row, err := s.lessonByID(ctx, lessonID)
	if err != nil {
		return Lesson{}, err
	}
	if _, err := s.requireCourse(ctx, row.courseID); err != nil {
		return Lesson{}, err
	}
	return s.applyLessonUpdate(ctx, row.courseID, lessonID, req, lang)
}

func (s *Store) DeleteLesson(ctx context.Context, courseID, lessonID string) error {
	if _, err := s.requireLesson(ctx, courseID, lessonID); err != nil {
		return err
	}
	_, err := s.db.ExecContext(ctx, `DELETE FROM course_lessons WHERE id = ?`, lessonID)
	return err
}

func (s *Store) DeleteLessonByID(ctx context.Context, lessonID string) error {
	row, err := s.lessonByID(ctx, lessonID)
	if err != nil {
		return err
	}
	if _, err := s.requireCourse(ctx, row.courseID); err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `DELETE FROM course_lessons WHERE id = ?`, lessonID)
	return err
}

func (s *Store) applyLessonUpdate(ctx context.Context, courseID, lessonID string, req UpdateLesson, lang string) (Lesson, error) {
	if req.ParentLessonID != nil {
		parent := strings.TrimSpace(*req.ParentLessonID)
		if parent == "" {
			if _, err := s.db.ExecContext(ctx, `UPDATE course_lessons SET parent_lesson_id = NULL, updated_at = UTC_TIMESTAMP() WHERE id = ?`, lessonID); err != nil {
				return Lesson{}, err
			}
		} else {
			if parent == lessonID {
				return Lesson{}, bad("Lesson cannot be its own parent")
			}
			if _, err := s.requireLesson(ctx, courseID, parent); err != nil {
				return Lesson{}, err
			}
			if _, err := s.db.ExecContext(ctx, `UPDATE course_lessons SET parent_lesson_id = ?, updated_at = UTC_TIMESTAMP() WHERE id = ?`, parent, lessonID); err != nil {
				return Lesson{}, err
			}
		}
	}
	if req.SortOrder != nil {
		if _, err := s.db.ExecContext(ctx, `UPDATE course_lessons SET sort_order = ? WHERE id = ?`, *req.SortOrder, lessonID); err != nil {
			return Lesson{}, err
		}
	}
	if req.AccessLevel != nil {
		if _, err := s.db.ExecContext(ctx, `UPDATE course_lessons SET access_level = ? WHERE id = ?`, *req.AccessLevel, lessonID); err != nil {
			return Lesson{}, err
		}
	}
	responseLang := normalizeLang(lang)
	if req.Translation != nil {
		if err := s.upsertLessonI18n(ctx, lessonID, req.Translation); err != nil {
			return Lesson{}, err
		}
		responseLang = normalizeLang(req.Translation.LanguageCode)
	}
	_, _ = s.db.ExecContext(ctx, `UPDATE course_lessons SET updated_at = UTC_TIMESTAMP() WHERE id = ?`, lessonID)
	return s.GetLesson(ctx, courseID, lessonID, responseLang)
}

func (s *Store) upsertLessonI18n(ctx context.Context, lessonID string, t *LessonTranslation) error {
	if t == nil {
		return nil
	}
	lang := normalizeLang(t.LanguageCode)
	var id int
	err := s.db.QueryRowContext(ctx, `SELECT id FROM course_lesson_i18n WHERE course_lesson_id = ? AND language_code = ?`, lessonID, lang).Scan(&id)
	if err == sql.ErrNoRows {
		_, err = s.db.ExecContext(ctx, `
			INSERT INTO course_lesson_i18n (course_lesson_id, language_code, title, slug, content)
			VALUES (?, ?, ?, ?, ?)`, lessonID, lang, t.Title, t.Slug, t.Content)
		return err
	}
	if err != nil {
		return err
	}
	if t.Title != "" {
		if _, err := s.db.ExecContext(ctx, `UPDATE course_lesson_i18n SET title = ? WHERE id = ?`, t.Title, id); err != nil {
			return err
		}
	}
	if t.Slug != "" {
		if _, err := s.db.ExecContext(ctx, `UPDATE course_lesson_i18n SET slug = ? WHERE id = ?`, t.Slug, id); err != nil {
			return err
		}
	}
	if t.Content != "" {
		if _, err := s.db.ExecContext(ctx, `UPDATE course_lesson_i18n SET content = ? WHERE id = ?`, t.Content, id); err != nil {
			return err
		}
	}
	return nil
}

type lessonRow struct {
	id, courseID, access string
	parent               sql.NullString
	sort                 int
	created, updated     sql.NullTime
}

func (s *Store) lessonsOf(ctx context.Context, courseID string) ([]lessonRow, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, course_id, parent_lesson_id, sort_order, access_level, created_at, updated_at
		FROM course_lessons WHERE course_id = ? ORDER BY sort_order, id`, courseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []lessonRow
	for rows.Next() {
		var r lessonRow
		if err := rows.Scan(&r.id, &r.courseID, &r.parent, &r.sort, &r.access, &r.created, &r.updated); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Store) lessonByID(ctx context.Context, id string) (lessonRow, error) {
	var r lessonRow
	err := s.db.QueryRowContext(ctx, `
		SELECT id, course_id, parent_lesson_id, sort_order, access_level, created_at, updated_at
		FROM course_lessons WHERE id = ?`, id).Scan(&r.id, &r.courseID, &r.parent, &r.sort, &r.access, &r.created, &r.updated)
	if err == sql.ErrNoRows {
		return r, missing("Lesson not found: " + id)
	}
	return r, err
}

func (s *Store) requireLesson(ctx context.Context, courseID, lessonID string) (lessonRow, error) {
	if _, err := s.requireCourse(ctx, courseID); err != nil {
		return lessonRow{}, err
	}
	row, err := s.lessonByID(ctx, lessonID)
	if err != nil {
		return lessonRow{}, err
	}
	if row.courseID != courseID {
		return lessonRow{}, missing("Lesson not found: " + lessonID)
	}
	return row, nil
}

func (s *Store) lessonI18n(ctx context.Context, lessonID string) ([]LessonTranslation, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT language_code, title, slug, content FROM course_lesson_i18n WHERE course_lesson_id = ? ORDER BY id`, lessonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []LessonTranslation
	for rows.Next() {
		var t LessonTranslation
		if err := rows.Scan(&t.LanguageCode, &t.Title, &t.Slug, &t.Content); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (s *Store) localizeLesson(ctx context.Context, row lessonRow, lang string) (Lesson, bool, error) {
	trs, err := s.lessonI18n(ctx, row.id)
	if err != nil {
		return Lesson{}, false, err
	}
	if len(trs) == 0 {
		return Lesson{}, false, nil
	}
	chosen := trs[0]
	found := false
	for _, t := range trs {
		if t.LanguageCode == lang {
			chosen = t
			found = true
			break
		}
	}
	if !found {
		for _, t := range trs {
			if t.LanguageCode == "en" {
				chosen = t
				break
			}
		}
	}
	return Lesson{
		ID: row.id, CourseID: row.courseID, ParentLessonID: platform.StrPtr(row.parent), SortOrder: row.sort,
		AccessLevel: row.access, CreatedAt: platform.InstantFrom(row.created), UpdatedAt: platform.InstantFrom(row.updated),
		LanguageCode: chosen.LanguageCode, Title: chosen.Title, Slug: chosen.Slug, Content: chosen.Content,
	}, true, nil
}

func blankPtr(v *string) any {
	if v == nil || strings.TrimSpace(*v) == "" {
		return nil
	}
	return strings.TrimSpace(*v)
}
