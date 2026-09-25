package courses

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"react-go-cms-courses-service/internal/platform"
)

func bad(msg string) error     { return platform.BadRequest("error", msg) }
func missing(msg string) error { return platform.NotFound("error", msg) }

func normalizeLang(lang string) string {
	lang = strings.TrimSpace(strings.ToLower(lang))
	if lang == "" {
		return "en"
	}
	return lang
}

type Page[T any] struct {
	Items []T   `json:"items"`
	Page  int   `json:"page"`
	Size  int   `json:"size"`
	Total int64 `json:"total"`
}

type Metadata struct {
	ID        string  `json:"id"`
	PostID    string  `json:"postId"`
	MetaKey   string  `json:"metaKey"`
	MetaValue *string `json:"metaValue"`
}

type Course struct {
	ID               string               `json:"id"`
	AuthorID         string               `json:"authorId"`
	ContentTypeID    int                  `json:"contentTypeId"`
	FeaturedImageURL *string              `json:"featuredImageUrl"`
	AccessLevel      string               `json:"accessLevel"`
	Status           string               `json:"status"`
	ViewCount        int                  `json:"viewCount"`
	PublishedAt      platform.InstantJSON `json:"publishedAt"`
	CategoryIDs      []int                `json:"categoryIds"`
	TagIDs           []int                `json:"tagIds"`
	CreatedAt        platform.InstantJSON `json:"createdAt"`
	UpdatedAt        platform.InstantJSON `json:"updatedAt"`
	LanguageCode     string               `json:"languageCode"`
	Title            string               `json:"title"`
	Slug             string               `json:"slug"`
	Content          string               `json:"content"`
	Excerpt          *string              `json:"excerpt"`
	MetaTitle        *string              `json:"metaTitle"`
	MetaDescription  *string              `json:"metaDescription"`
	Metadata         []Metadata           `json:"metadata"`
	LessonCount      int                  `json:"lessonCount"`
}

type Lesson struct {
	ID             string               `json:"id"`
	CourseID       string               `json:"courseId"`
	ParentLessonID *string              `json:"parentLessonId"`
	SortOrder      int                  `json:"sortOrder"`
	AccessLevel    string               `json:"accessLevel"`
	CreatedAt      platform.InstantJSON `json:"createdAt"`
	UpdatedAt      platform.InstantJSON `json:"updatedAt"`
	LanguageCode   string               `json:"languageCode"`
	Title          string               `json:"title"`
	Slug           string               `json:"slug"`
	Content        string               `json:"content"`
}

type Translation struct {
	LanguageCode    string  `json:"languageCode"`
	Title           string  `json:"title"`
	Slug            string  `json:"slug"`
	Content         string  `json:"content"`
	Excerpt         *string `json:"excerpt"`
	MetaTitle       *string `json:"metaTitle"`
	MetaDescription *string `json:"metaDescription"`
}

type Store struct{ db *sql.DB }

func NewStore(db *sql.DB) *Store { return &Store{db: db} }

func (s *Store) courseTypeID(ctx context.Context) (int, error) {
	var id int
	err := s.db.QueryRowContext(ctx, `SELECT id FROM content_types WHERE slug = 'course'`).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, bad("Content type 'course' is not configured")
	}
	return id, err
}

type postRow struct {
	id, author, access, status  string
	typeID, views               int
	image                       sql.NullString
	published, created, updated sql.NullTime
}

func (s *Store) requireCourse(ctx context.Context, id string) (postRow, error) {
	typeID, err := s.courseTypeID(ctx)
	if err != nil {
		return postRow{}, err
	}
	var p postRow
	err = s.db.QueryRowContext(ctx, `
		SELECT id, author_id, content_type_id, featured_image_url, access_level, status, COALESCE(view_count,0),
		       published_at, created_at, updated_at
		FROM posts WHERE id = ?`, id).Scan(&p.id, &p.author, &p.typeID, &p.image, &p.access, &p.status, &p.views, &p.published, &p.created, &p.updated)
	if err == sql.ErrNoRows || (err == nil && p.typeID != typeID) {
		return postRow{}, missing("Course not found: " + id)
	}
	return p, err
}

func (s *Store) List(ctx context.Context, status, lang string, includeAll bool, page, size int) (Page[Course], error) {
	typeID, err := s.courseTypeID(ctx)
	if err != nil {
		return Page[Course]{}, err
	}
	if page < 0 {
		page = 0
	}
	if size < 1 {
		size = 1
	}
	if size > 100 {
		size = 100
	}
	where := ` WHERE content_type_id = ?`
	args := []any{typeID}
	if strings.TrimSpace(status) != "" {
		where += ` AND status = ?`
		args = append(args, strings.TrimSpace(status))
	} else if !includeAll {
		where += ` AND status = 'published'`
	}
	var total int64
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM posts`+where, args...).Scan(&total); err != nil {
		return Page[Course]{}, err
	}
	q := `SELECT id, author_id, content_type_id, featured_image_url, access_level, status, COALESCE(view_count,0), published_at, created_at, updated_at FROM posts` + where + ` ORDER BY created_at DESC LIMIT ? OFFSET ?`
	args = append(args, size, page*size)
	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return Page[Course]{}, err
	}
	defer rows.Close()
	var posts []postRow
	for rows.Next() {
		var p postRow
		if err := rows.Scan(&p.id, &p.author, &p.typeID, &p.image, &p.access, &p.status, &p.views, &p.published, &p.created, &p.updated); err != nil {
			return Page[Course]{}, err
		}
		posts = append(posts, p)
	}
	items, err := s.hydrate(ctx, posts, normalizeLang(lang))
	if err != nil {
		return Page[Course]{}, err
	}
	return Page[Course]{Items: items, Page: page, Size: size, Total: total}, nil
}

func (s *Store) Stats(ctx context.Context) (struct {
	Total int64 `json:"total"`
}, error) {
	var out struct {
		Total int64 `json:"total"`
	}
	typeID, err := s.courseTypeID(ctx)
	if err != nil {
		return out, err
	}
	err = s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM posts WHERE content_type_id = ?`, typeID).Scan(&out.Total)
	return out, err
}

func (s *Store) Get(ctx context.Context, id, lang string, includeAll bool) (Course, error) {
	p, err := s.requireCourse(ctx, id)
	if err != nil {
		return Course{}, err
	}
	if !includeAll && p.status != "published" {
		return Course{}, missing("Course not found: " + id)
	}
	items, err := s.hydrate(ctx, []postRow{p}, normalizeLang(lang))
	if err != nil {
		return Course{}, err
	}
	if len(items) == 0 {
		return Course{}, missing("Course translation not found: " + id)
	}
	return items[0], nil
}

func (s *Store) GetBySlug(ctx context.Context, slug, lang string, includeAll bool) (Course, error) {
	language := normalizeLang(lang)
	typeID, err := s.courseTypeID(ctx)
	if err != nil {
		return Course{}, err
	}
	var postID string
	err = s.db.QueryRowContext(ctx, `SELECT post_id FROM post_i18n WHERE slug = ? AND language_code = ? ORDER BY id LIMIT 1`, slug, language).Scan(&postID)
	if err == sql.ErrNoRows {
		err = s.db.QueryRowContext(ctx, `SELECT post_id FROM post_i18n WHERE slug = ? ORDER BY id LIMIT 1`, slug).Scan(&postID)
	}
	if err == sql.ErrNoRows {
		return Course{}, missing("Course not found for slug: " + slug)
	}
	if err != nil {
		return Course{}, err
	}
	p, err := s.requireCourse(ctx, postID)
	if err != nil || p.typeID != typeID {
		return Course{}, missing("Course not found for slug: " + slug)
	}
	if !includeAll && p.status != "published" {
		return Course{}, missing("Course not found for slug: " + slug)
	}
	items, err := s.hydrate(ctx, []postRow{p}, language)
	if err != nil || len(items) == 0 {
		return Course{}, missing("Course translation not found for slug: " + slug)
	}
	return items[0], nil
}

type CreateCourse struct {
	AuthorID         *string      `json:"authorId"`
	FeaturedImageURL *string      `json:"featuredImageUrl"`
	AccessLevel      *string      `json:"accessLevel"`
	Status           *string      `json:"status"`
	Translation      *Translation `json:"translation"`
	Metadata         []MetaIn     `json:"metadata"`
}

type MetaIn struct {
	MetaKey   string  `json:"metaKey"`
	MetaValue *string `json:"metaValue"`
}

func (s *Store) Create(ctx context.Context, req CreateCourse, fallbackAuthor string) (Course, error) {
	if req.Translation == nil || strings.TrimSpace(req.Translation.LanguageCode) == "" || strings.TrimSpace(req.Translation.Title) == "" || strings.TrimSpace(req.Translation.Slug) == "" || strings.TrimSpace(req.Translation.Content) == "" {
		return Course{}, bad("translation is required")
	}
	typeID, err := s.courseTypeID(ctx)
	if err != nil {
		return Course{}, err
	}
	author := fallbackAuthor
	if req.AuthorID != nil && strings.TrimSpace(*req.AuthorID) != "" {
		author = strings.TrimSpace(*req.AuthorID)
	}
	if author == "" {
		return Course{}, bad("authorId is required")
	}
	status := "draft"
	if req.Status != nil {
		status = *req.Status
	}
	access := "public"
	if req.AccessLevel != nil {
		access = *req.AccessLevel
	}
	id := platform.NewUUID()
	var published any
	if status == "published" {
		published = time.Now().UTC()
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Course{}, err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO posts (id, author_id, content_type_id, featured_image_url, access_level, status, view_count, published_at)
		VALUES (?, ?, ?, ?, ?, ?, 0, ?)`, id, author, typeID, req.FeaturedImageURL, access, status, published); err != nil {
		return Course{}, err
	}
	lang := normalizeLang(req.Translation.LanguageCode)
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO post_i18n (post_id, language_code, title, slug, content, excerpt, meta_title, meta_description)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		id, lang, req.Translation.Title, req.Translation.Slug, req.Translation.Content, req.Translation.Excerpt, req.Translation.MetaTitle, req.Translation.MetaDescription); err != nil {
		return Course{}, err
	}
	if err := insertMeta(ctx, tx, id, req.Metadata); err != nil {
		return Course{}, err
	}
	if err := tx.Commit(); err != nil {
		return Course{}, err
	}
	return s.Get(ctx, id, lang, true)
}

type UpdateCourse struct {
	FeaturedImageURL *string      `json:"featuredImageUrl"`
	AccessLevel      *string      `json:"accessLevel"`
	Status           *string      `json:"status"`
	Translation      *Translation `json:"translation"`
}

func (s *Store) Update(ctx context.Context, id string, req UpdateCourse, lang string) (Course, error) {
	if _, err := s.requireCourse(ctx, id); err != nil {
		return Course{}, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Course{}, err
	}
	defer tx.Rollback()
	if req.FeaturedImageURL != nil {
		if _, err := tx.ExecContext(ctx, `UPDATE posts SET featured_image_url = ? WHERE id = ?`, *req.FeaturedImageURL, id); err != nil {
			return Course{}, err
		}
	}
	if req.AccessLevel != nil {
		if _, err := tx.ExecContext(ctx, `UPDATE posts SET access_level = ? WHERE id = ?`, *req.AccessLevel, id); err != nil {
			return Course{}, err
		}
	}
	if req.Status != nil {
		if err := applyCourseStatus(ctx, tx, id, *req.Status); err != nil {
			return Course{}, err
		}
	}
	responseLang := normalizeLang(lang)
	if req.Translation != nil {
		if err := upsertPostI18n(ctx, tx, id, req.Translation, false); err != nil {
			return Course{}, err
		}
		responseLang = normalizeLang(req.Translation.LanguageCode)
	}
	if _, err := tx.ExecContext(ctx, `UPDATE posts SET updated_at = UTC_TIMESTAMP() WHERE id = ?`, id); err != nil {
		return Course{}, err
	}
	if err := tx.Commit(); err != nil {
		return Course{}, err
	}
	return s.Get(ctx, id, responseLang, true)
}

func (s *Store) PatchStatus(ctx context.Context, id, status, lang string) (Course, error) {
	if _, err := s.requireCourse(ctx, id); err != nil {
		return Course{}, err
	}
	if err := applyCourseStatus(ctx, s.db, id, status); err != nil {
		return Course{}, err
	}
	_, _ = s.db.ExecContext(ctx, `UPDATE posts SET updated_at = UTC_TIMESTAMP() WHERE id = ?`, id)
	return s.Get(ctx, id, normalizeLang(lang), true)
}

func (s *Store) Delete(ctx context.Context, id string) error {
	if _, err := s.requireCourse(ctx, id); err != nil {
		return err
	}
	_, err := s.db.ExecContext(ctx, `DELETE FROM posts WHERE id = ?`, id)
	return err
}

func (s *Store) GetMetadata(ctx context.Context, id string) ([]Metadata, error) {
	if _, err := s.requireCourse(ctx, id); err != nil {
		return nil, err
	}
	rows := loadMeta(ctx, s.db, []any{id})[id]
	if rows == nil {
		return []Metadata{}, nil
	}
	return rows, nil
}

func (s *Store) ReplaceMetadata(ctx context.Context, id string, items []MetaIn) ([]Metadata, error) {
	if _, err := s.requireCourse(ctx, id); err != nil {
		return nil, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `DELETE FROM post_metadata WHERE post_id = ?`, id); err != nil {
		return nil, err
	}
	if err := insertMeta(ctx, tx, id, items); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return s.GetMetadata(ctx, id)
}

func (s *Store) hydrate(ctx context.Context, posts []postRow, lang string) ([]Course, error) {
	out := []Course{}
	if len(posts) == 0 {
		return out, nil
	}
	ids := make([]any, len(posts))
	for i, p := range posts {
		ids[i] = p.id
	}
	i18n := map[string][]Translation{}
	rows, err := s.db.QueryContext(ctx, `SELECT post_id, language_code, title, slug, content, excerpt, meta_title, meta_description FROM post_i18n WHERE post_id IN (`+ph(len(ids))+`) ORDER BY id`, ids...)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var postID string
		var t Translation
		var excerpt, mt, md sql.NullString
		if err := rows.Scan(&postID, &t.LanguageCode, &t.Title, &t.Slug, &t.Content, &excerpt, &mt, &md); err != nil {
			rows.Close()
			return nil, err
		}
		t.Excerpt = platform.StrPtr(excerpt)
		t.MetaTitle = platform.StrPtr(mt)
		t.MetaDescription = platform.StrPtr(md)
		i18n[postID] = append(i18n[postID], t)
	}
	rows.Close()
	meta := loadMetaMap(ctx, s.db, ids)
	counts := map[string]int{}
	crows, err := s.db.QueryContext(ctx, `SELECT course_id FROM course_lessons WHERE course_id IN (`+ph(len(ids))+`)`, ids...)
	if err != nil {
		return nil, err
	}
	for crows.Next() {
		var id string
		if err := crows.Scan(&id); err != nil {
			crows.Close()
			return nil, err
		}
		counts[id]++
	}
	crows.Close()
	for _, p := range posts {
		tr := pickTranslation(i18n[p.id], lang)
		if tr == nil {
			continue
		}
		m := meta[p.id]
		if m == nil {
			m = []Metadata{}
		}
		out = append(out, Course{
			ID: p.id, AuthorID: p.author, ContentTypeID: p.typeID, FeaturedImageURL: platform.StrPtr(p.image),
			AccessLevel: p.access, Status: p.status, ViewCount: p.views,
			PublishedAt: platform.InstantFrom(p.published), CreatedAt: platform.InstantFrom(p.created), UpdatedAt: platform.InstantFrom(p.updated),
			CategoryIDs: []int{}, TagIDs: []int{},
			LanguageCode: tr.LanguageCode, Title: tr.Title, Slug: tr.Slug, Content: tr.Content,
			Excerpt: tr.Excerpt, MetaTitle: tr.MetaTitle, MetaDescription: tr.MetaDescription,
			Metadata: m, LessonCount: counts[p.id],
		})
	}
	return out, nil
}

func pickTranslation(rows []Translation, lang string) *Translation {
	if len(rows) == 0 {
		return nil
	}
	for i := range rows {
		if rows[i].LanguageCode == lang {
			return &rows[i]
		}
	}
	for i := range rows {
		if rows[i].LanguageCode == "en" {
			return &rows[i]
		}
	}
	return &rows[0]
}

func loadMeta(ctx context.Context, db *sql.DB, ids []any) map[string][]Metadata {
	return loadMetaMap(ctx, db, ids)
}

func loadMetaMap(ctx context.Context, db *sql.DB, ids []any) map[string][]Metadata {
	out := map[string][]Metadata{}
	if len(ids) == 0 {
		return out
	}
	rows, err := db.QueryContext(ctx, `SELECT id, post_id, meta_key, meta_value FROM post_metadata WHERE post_id IN (`+ph(len(ids))+`)`, ids...)
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var m Metadata
		var val sql.NullString
		if err := rows.Scan(&m.ID, &m.PostID, &m.MetaKey, &val); err != nil {
			return out
		}
		m.MetaValue = platform.StrPtr(val)
		out[m.PostID] = append(out[m.PostID], m)
	}
	return out
}

func insertMeta(ctx context.Context, tx *sql.Tx, postID string, items []MetaIn) error {
	for _, item := range items {
		if strings.TrimSpace(item.MetaKey) == "" {
			continue
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO post_metadata (id, post_id, meta_key, meta_value) VALUES (?, ?, ?, ?)`,
			platform.NewUUID(), postID, strings.TrimSpace(item.MetaKey), item.MetaValue); err != nil {
			return err
		}
	}
	return nil
}

type statusExec interface {
	ExecContext(ctx context.Context, q string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, q string, args ...any) *sql.Row
}

func applyCourseStatus(ctx context.Context, ex statusExec, id, status string) error {
	if strings.TrimSpace(status) == "" {
		return bad("status is required")
	}
	var current string
	var published sql.NullTime
	if err := ex.QueryRowContext(ctx, `SELECT status, published_at FROM posts WHERE id = ?`, id).Scan(&current, &published); err != nil {
		return err
	}
	if status == "published" && current != "published" && !published.Valid {
		_, err := ex.ExecContext(ctx, `UPDATE posts SET status = ?, published_at = UTC_TIMESTAMP() WHERE id = ?`, status, id)
		return err
	}
	_, err := ex.ExecContext(ctx, `UPDATE posts SET status = ? WHERE id = ?`, status, id)
	return err
}

func upsertPostI18n(ctx context.Context, tx *sql.Tx, postID string, t *Translation, required bool) error {
	if t == nil {
		if required {
			return bad("translation is required")
		}
		return nil
	}
	lang := normalizeLang(t.LanguageCode)
	var id int
	err := tx.QueryRowContext(ctx, `SELECT id FROM post_i18n WHERE post_id = ? AND language_code = ?`, postID, lang).Scan(&id)
	if err == sql.ErrNoRows {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO post_i18n (post_id, language_code, title, slug, content, excerpt, meta_title, meta_description)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, postID, lang, t.Title, t.Slug, t.Content, t.Excerpt, t.MetaTitle, t.MetaDescription)
		return err
	}
	if err != nil {
		return err
	}
	if t.Title != "" {
		if _, err := tx.ExecContext(ctx, `UPDATE post_i18n SET title = ? WHERE id = ?`, t.Title, id); err != nil {
			return err
		}
	}
	if t.Slug != "" {
		if _, err := tx.ExecContext(ctx, `UPDATE post_i18n SET slug = ? WHERE id = ?`, t.Slug, id); err != nil {
			return err
		}
	}
	if t.Content != "" {
		if _, err := tx.ExecContext(ctx, `UPDATE post_i18n SET content = ? WHERE id = ?`, t.Content, id); err != nil {
			return err
		}
	}
	if t.Excerpt != nil {
		if _, err := tx.ExecContext(ctx, `UPDATE post_i18n SET excerpt = ? WHERE id = ?`, *t.Excerpt, id); err != nil {
			return err
		}
	}
	if t.MetaTitle != nil {
		if _, err := tx.ExecContext(ctx, `UPDATE post_i18n SET meta_title = ? WHERE id = ?`, *t.MetaTitle, id); err != nil {
			return err
		}
	}
	if t.MetaDescription != nil {
		if _, err := tx.ExecContext(ctx, `UPDATE post_i18n SET meta_description = ? WHERE id = ?`, *t.MetaDescription, id); err != nil {
			return err
		}
	}
	return nil
}

func ph(n int) string {
	if n <= 0 {
		return ""
	}
	b := strings.Repeat("?,", n)
	return b[:len(b)-1]
}
