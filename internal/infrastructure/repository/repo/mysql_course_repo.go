package repo

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"react-go-cms-courses-service/internal/application/service/course"
	"react-go-cms-courses-service/internal/domain/apperror"
	"react-go-cms-courses-service/internal/domain/entity"
	"react-go-cms-courses-service/internal/infrastructure/identity"
	"react-go-cms-courses-service/internal/infrastructure/repository/model"
)

// CourseRepository persists courses in MySQL.
type CourseRepository struct {
	db *sql.DB
}

// NewCourseRepository builds a course repository.
func NewCourseRepository(db *sql.DB) *CourseRepository {
	return &CourseRepository{db: db}
}

var _ course.Repository = (*CourseRepository)(nil)

type statusExec interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// ContentTypeID returns the course content type id.
func (r *CourseRepository) ContentTypeID(ctx context.Context) (int, error) {
	return contentTypeID(ctx, r.db)
}

func contentTypeID(ctx context.Context, db *sql.DB) (int, error) {
	var id int
	var err error
	err = db.QueryRowContext(ctx, `SELECT id FROM content_types WHERE slug = 'course'`).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, apperror.Invalid("Content type 'course' is not configured")
	}
	return id, err
}

func requireCourse(ctx context.Context, db *sql.DB, id string) (model.CourseRow, error) {
	var typeID int
	var err error
	typeID, err = contentTypeID(ctx, db)
	if err != nil {
		return model.CourseRow{}, err
	}
	var row model.CourseRow
	err = db.QueryRowContext(ctx, `
		SELECT id, author_id, content_type_id, featured_image_url, access_level, status, COALESCE(view_count,0),
		       published_at, created_at, updated_at
		FROM posts WHERE id = ?`, id).Scan(
		&row.ID, &row.AuthorID, &row.ContentTypeID, &row.FeaturedImageURL, &row.AccessLevel, &row.Status, &row.ViewCount,
		&row.PublishedAt, &row.CreatedAt, &row.UpdatedAt,
	)
	if err == sql.ErrNoRows || (err == nil && row.ContentTypeID != typeID) {
		return model.CourseRow{}, apperror.NotFound("Course not found: " + id)
	}
	return row, err
}

// List returns course rows for one page, plus the unfiltered-by-translation total.
func (r *CourseRepository) List(ctx context.Context, status string, onlyPublished bool, page, size int) ([]entity.CourseRecord, int64, error) {
	var typeID int
	var err error
	typeID, err = contentTypeID(ctx, r.db)
	if err != nil {
		return nil, 0, err
	}
	where := ` WHERE content_type_id = ?`
	args := []any{typeID}
	if status != "" {
		where += ` AND status = ?`
		args = append(args, status)
	} else if onlyPublished {
		where += ` AND status = 'published'`
	}
	var total int64
	err = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM posts`+where, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	query := `SELECT id, author_id, content_type_id, featured_image_url, access_level, status, COALESCE(view_count,0), published_at, created_at, updated_at FROM posts` + where + ` ORDER BY created_at DESC LIMIT ? OFFSET ?`
	args = append(args, size, page*size)
	var rows *sql.Rows
	rows, err = r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var posts []model.CourseRow
	for rows.Next() {
		var row model.CourseRow
		err = rows.Scan(&row.ID, &row.AuthorID, &row.ContentTypeID, &row.FeaturedImageURL, &row.AccessLevel, &row.Status, &row.ViewCount, &row.PublishedAt, &row.CreatedAt, &row.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}
		posts = append(posts, row)
	}
	var records []entity.CourseRecord
	records, err = r.hydrate(ctx, posts)
	if err != nil {
		return nil, 0, err
	}
	return records, total, nil
}

// Count returns the number of course posts.
func (r *CourseRepository) Count(ctx context.Context) (int64, error) {
	var typeID int
	var err error
	typeID, err = contentTypeID(ctx, r.db)
	if err != nil {
		return 0, err
	}
	var total int64
	err = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM posts WHERE content_type_id = ?`, typeID).Scan(&total)
	return total, err
}

// Require loads one course row. Translations stay empty until AttachDetails.
func (r *CourseRepository) Require(ctx context.Context, id string) (entity.CourseRecord, error) {
	var row model.CourseRow
	var err error
	row, err = requireCourse(ctx, r.db, id)
	if err != nil {
		return entity.CourseRecord{}, err
	}
	return courseRecord(row), nil
}

// AttachDetails loads translations, metadata, and the lesson count for one course row.
func (r *CourseRepository) AttachDetails(ctx context.Context, record entity.CourseRecord) (entity.CourseRecord, error) {
	var records []entity.CourseRecord
	var err error
	records, err = r.hydrate(ctx, []model.CourseRow{courseModel(record)})
	if err != nil {
		return entity.CourseRecord{}, err
	}
	if len(records) == 0 {
		return record, nil
	}
	return records[0], nil
}

// FindIDBySlug resolves a post id from a translation slug.
func (r *CourseRepository) FindIDBySlug(ctx context.Context, slug, language string) (string, error) {
	var postID string
	var err error
	err = r.db.QueryRowContext(ctx, `SELECT post_id FROM post_i18n WHERE slug = ? AND language_code = ? ORDER BY id LIMIT 1`, slug, language).Scan(&postID)
	if err == sql.ErrNoRows {
		err = r.db.QueryRowContext(ctx, `SELECT post_id FROM post_i18n WHERE slug = ? ORDER BY id LIMIT 1`, slug).Scan(&postID)
	}
	if err == sql.ErrNoRows {
		return "", apperror.NotFound("Course not found for slug: " + slug)
	}
	if err != nil {
		return "", err
	}
	return postID, nil
}

// Insert stores a course, its first translation, and metadata.
func (r *CourseRepository) Insert(ctx context.Context, typeID int, author string, image *string, access, status string, publishedAt *time.Time, translation entity.CourseTranslation, metadata []entity.MetadataInput) (string, error) {
	var id string
	id = identity.NewUUID()
	var published any
	if publishedAt != nil {
		published = *publishedAt
	}
	var tx *sql.Tx
	var err error
	tx, err = r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, `
		INSERT INTO posts (id, author_id, content_type_id, featured_image_url, access_level, status, view_count, published_at)
		VALUES (?, ?, ?, ?, ?, ?, 0, ?)`, id, author, typeID, image, access, status, published)
	if err != nil {
		return "", err
	}
	language := normalizeLang(translation.LanguageCode)
	_, err = tx.ExecContext(ctx, `
		INSERT INTO post_i18n (post_id, language_code, title, slug, content, excerpt, meta_title, meta_description)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		id, language, translation.Title, translation.Slug, translation.Content, translation.Excerpt, translation.MetaTitle, translation.MetaDescription)
	if err != nil {
		return "", err
	}
	err = insertMeta(ctx, tx, id, metadata)
	if err != nil {
		return "", err
	}
	err = tx.Commit()
	if err != nil {
		return "", err
	}
	return id, nil
}

// Update applies a partial course change inside one transaction.
func (r *CourseRepository) Update(ctx context.Context, id string, req entity.CourseUpdate) error {
	var err error
	_, err = requireCourse(ctx, r.db, id)
	if err != nil {
		return err
	}
	var tx *sql.Tx
	tx, err = r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if req.FeaturedImageURL != nil {
		_, err = tx.ExecContext(ctx, `UPDATE posts SET featured_image_url = ? WHERE id = ?`, *req.FeaturedImageURL, id)
		if err != nil {
			return err
		}
	}
	if req.AccessLevel != nil {
		_, err = tx.ExecContext(ctx, `UPDATE posts SET access_level = ? WHERE id = ?`, *req.AccessLevel, id)
		if err != nil {
			return err
		}
	}
	if req.Status != nil {
		err = applyCourseStatus(ctx, tx, id, *req.Status)
		if err != nil {
			return err
		}
	}
	if req.Translation != nil {
		err = upsertPostI18n(ctx, tx, id, req.Translation, false)
		if err != nil {
			return err
		}
	}
	_, err = tx.ExecContext(ctx, `UPDATE posts SET updated_at = UTC_TIMESTAMP() WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// PatchStatus updates the course status and published timestamp when needed.
func (r *CourseRepository) PatchStatus(ctx context.Context, id, status string) error {
	var err error
	_, err = requireCourse(ctx, r.db, id)
	if err != nil {
		return err
	}
	err = applyCourseStatus(ctx, r.db, id, status)
	if err != nil {
		return err
	}
	_, _ = r.db.ExecContext(ctx, `UPDATE posts SET updated_at = UTC_TIMESTAMP() WHERE id = ?`, id)
	return nil
}

// Delete removes a course post.
func (r *CourseRepository) Delete(ctx context.Context, id string) error {
	var err error
	_, err = requireCourse(ctx, r.db, id)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, `DELETE FROM posts WHERE id = ?`, id)
	return err
}

// ListMetadata returns metadata for one course.
func (r *CourseRepository) ListMetadata(ctx context.Context, id string) ([]entity.Metadata, error) {
	var err error
	_, err = requireCourse(ctx, r.db, id)
	if err != nil {
		return nil, err
	}
	rows := loadMeta(ctx, r.db, []any{id})[id]
	if rows == nil {
		return []entity.Metadata{}, nil
	}
	return rows, nil
}

// ReplaceMetadata replaces metadata for one course.
func (r *CourseRepository) ReplaceMetadata(ctx context.Context, id string, items []entity.MetadataInput) error {
	var err error
	_, err = requireCourse(ctx, r.db, id)
	if err != nil {
		return err
	}
	var tx *sql.Tx
	tx, err = r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, `DELETE FROM post_metadata WHERE post_id = ?`, id)
	if err != nil {
		return err
	}
	err = insertMeta(ctx, tx, id, items)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (r *CourseRepository) hydrate(ctx context.Context, posts []model.CourseRow) ([]entity.CourseRecord, error) {
	out := []entity.CourseRecord{}
	if len(posts) == 0 {
		return out, nil
	}
	ids := make([]any, len(posts))
	for index, row := range posts {
		ids[index] = row.ID
	}
	i18n := map[string][]entity.CourseTranslation{}
	var rows *sql.Rows
	var err error
	rows, err = r.db.QueryContext(ctx, `SELECT post_id, language_code, title, slug, content, excerpt, meta_title, meta_description FROM post_i18n WHERE post_id IN (`+ph(len(ids))+`) ORDER BY id`, ids...)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var postID string
		var translation entity.CourseTranslation
		var excerpt, metaTitle, metaDescription sql.NullString
		err = rows.Scan(&postID, &translation.LanguageCode, &translation.Title, &translation.Slug, &translation.Content, &excerpt, &metaTitle, &metaDescription)
		if err != nil {
			rows.Close()
			return nil, err
		}
		translation.Excerpt = StrPtr(excerpt)
		translation.MetaTitle = StrPtr(metaTitle)
		translation.MetaDescription = StrPtr(metaDescription)
		i18n[postID] = append(i18n[postID], translation)
	}
	rows.Close()
	meta := loadMetaMap(ctx, r.db, ids)
	counts := map[string]int{}
	var countRows *sql.Rows
	countRows, err = r.db.QueryContext(ctx, `SELECT course_id FROM course_lessons WHERE course_id IN (`+ph(len(ids))+`)`, ids...)
	if err != nil {
		return nil, err
	}
	for countRows.Next() {
		var id string
		err = countRows.Scan(&id)
		if err != nil {
			countRows.Close()
			return nil, err
		}
		counts[id]++
	}
	countRows.Close()
	for _, row := range posts {
		out = append(out, entity.CourseRecord{
			ID:               row.ID,
			AuthorID:         row.AuthorID,
			ContentTypeID:    row.ContentTypeID,
			FeaturedImageURL: StrPtr(row.FeaturedImageURL),
			AccessLevel:      row.AccessLevel,
			Status:           row.Status,
			ViewCount:        row.ViewCount,
			PublishedAt:      InstantFrom(row.PublishedAt),
			CreatedAt:        InstantFrom(row.CreatedAt),
			UpdatedAt:        InstantFrom(row.UpdatedAt),
			Translations:     i18n[row.ID],
			Metadata:         meta[row.ID],
			LessonCount:      counts[row.ID],
		})
	}
	return out, nil
}

func loadMeta(ctx context.Context, db *sql.DB, ids []any) map[string][]entity.Metadata {
	return loadMetaMap(ctx, db, ids)
}

func loadMetaMap(ctx context.Context, db *sql.DB, ids []any) map[string][]entity.Metadata {
	out := map[string][]entity.Metadata{}
	if len(ids) == 0 {
		return out
	}
	var rows *sql.Rows
	var err error
	rows, err = db.QueryContext(ctx, `SELECT id, post_id, meta_key, meta_value FROM post_metadata WHERE post_id IN (`+ph(len(ids))+`)`, ids...)
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var item entity.Metadata
		var value sql.NullString
		err = rows.Scan(&item.ID, &item.PostID, &item.MetaKey, &value)
		if err != nil {
			return out
		}
		item.MetaValue = StrPtr(value)
		out[item.PostID] = append(out[item.PostID], item)
	}
	return out
}

func insertMeta(ctx context.Context, tx *sql.Tx, postID string, items []entity.MetadataInput) error {
	for _, item := range items {
		if strings.TrimSpace(item.MetaKey) == "" {
			continue
		}
		var err error
		_, err = tx.ExecContext(ctx, `INSERT INTO post_metadata (id, post_id, meta_key, meta_value) VALUES (?, ?, ?, ?)`,
			identity.NewUUID(), postID, strings.TrimSpace(item.MetaKey), item.MetaValue)
		if err != nil {
			return err
		}
	}
	return nil
}

func applyCourseStatus(ctx context.Context, ex statusExec, id, status string) error {
	if strings.TrimSpace(status) == "" {
		return apperror.Invalid("status is required")
	}
	var current string
	var published sql.NullTime
	var err error
	err = ex.QueryRowContext(ctx, `SELECT status, published_at FROM posts WHERE id = ?`, id).Scan(&current, &published)
	if err != nil {
		return err
	}
	if status == "published" && current != "published" && !published.Valid {
		_, err = ex.ExecContext(ctx, `UPDATE posts SET status = ?, published_at = UTC_TIMESTAMP() WHERE id = ?`, status, id)
		return err
	}
	_, err = ex.ExecContext(ctx, `UPDATE posts SET status = ? WHERE id = ?`, status, id)
	return err
}

func upsertPostI18n(ctx context.Context, tx *sql.Tx, postID string, translation *entity.CourseTranslation, required bool) error {
	if translation == nil {
		if required {
			return apperror.Invalid("translation is required")
		}
		return nil
	}
	var id int
	var err error
	language := normalizeLang(translation.LanguageCode)
	err = tx.QueryRowContext(ctx, `SELECT id FROM post_i18n WHERE post_id = ? AND language_code = ?`, postID, language).Scan(&id)
	if err == sql.ErrNoRows {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO post_i18n (post_id, language_code, title, slug, content, excerpt, meta_title, meta_description)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, postID, language, translation.Title, translation.Slug, translation.Content, translation.Excerpt, translation.MetaTitle, translation.MetaDescription)
		return err
	}
	if err != nil {
		return err
	}
	if translation.Title != "" {
		_, err = tx.ExecContext(ctx, `UPDATE post_i18n SET title = ? WHERE id = ?`, translation.Title, id)
		if err != nil {
			return err
		}
	}
	if translation.Slug != "" {
		_, err = tx.ExecContext(ctx, `UPDATE post_i18n SET slug = ? WHERE id = ?`, translation.Slug, id)
		if err != nil {
			return err
		}
	}
	if translation.Content != "" {
		_, err = tx.ExecContext(ctx, `UPDATE post_i18n SET content = ? WHERE id = ?`, translation.Content, id)
		if err != nil {
			return err
		}
	}
	if translation.Excerpt != nil {
		_, err = tx.ExecContext(ctx, `UPDATE post_i18n SET excerpt = ? WHERE id = ?`, *translation.Excerpt, id)
		if err != nil {
			return err
		}
	}
	if translation.MetaTitle != nil {
		_, err = tx.ExecContext(ctx, `UPDATE post_i18n SET meta_title = ? WHERE id = ?`, *translation.MetaTitle, id)
		if err != nil {
			return err
		}
	}
	if translation.MetaDescription != nil {
		_, err = tx.ExecContext(ctx, `UPDATE post_i18n SET meta_description = ? WHERE id = ?`, *translation.MetaDescription, id)
		if err != nil {
			return err
		}
	}
	return nil
}

func courseRecord(row model.CourseRow) entity.CourseRecord {
	return entity.CourseRecord{
		ID:               row.ID,
		AuthorID:         row.AuthorID,
		ContentTypeID:    row.ContentTypeID,
		FeaturedImageURL: StrPtr(row.FeaturedImageURL),
		AccessLevel:      row.AccessLevel,
		Status:           row.Status,
		ViewCount:        row.ViewCount,
		PublishedAt:      InstantFrom(row.PublishedAt),
		CreatedAt:        InstantFrom(row.CreatedAt),
		UpdatedAt:        InstantFrom(row.UpdatedAt),
	}
}

func courseModel(record entity.CourseRecord) model.CourseRow {
	return model.CourseRow{
		ID:               record.ID,
		AuthorID:         record.AuthorID,
		ContentTypeID:    record.ContentTypeID,
		FeaturedImageURL: nullString(record.FeaturedImageURL),
		AccessLevel:      record.AccessLevel,
		Status:           record.Status,
		ViewCount:        record.ViewCount,
		PublishedAt:      nullTime(record.PublishedAt),
		CreatedAt:        nullTime(record.CreatedAt),
		UpdatedAt:        nullTime(record.UpdatedAt),
	}
}

func nullTime(value *time.Time) sql.NullTime {
	if value == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *value, Valid: true}
}

func nullString(value *string) sql.NullString {
	if value == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *value, Valid: true}
}

func normalizeLang(lang string) string {
	lang = strings.TrimSpace(strings.ToLower(lang))
	if lang == "" {
		return "en"
	}
	return lang
}

func ph(n int) string {
	if n <= 0 {
		return ""
	}
	placeholders := strings.Repeat("?,", n)
	return placeholders[:len(placeholders)-1]
}
