package repository

import (
	"context"
	"fmt"
	"post-service/internal/domain"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostRepository interface {
	CreatePost(ctx context.Context, post *domain.Post) (int64, error)
	GetPost(ctx context.Context, id int64) (*domain.Post, error)
	UpdatePost(ctx context.Context, id int64, title *string, content *string, tags []string) (*domain.Post, error)
	DeletePost(ctx context.Context, id int64) error
	ListPostsFiltered(ctx context.Context, limit, offset int, sort, author, tag string) ([]domain.Post, error)
	SearchPosts(ctx context.Context, query string, limit, offset int) ([]domain.Post, error)
	IncrementView(ctx context.Context, postID int64) error
	IncrementLike(ctx context.Context, postID int64) error
	DecrementLike(ctx context.Context, postID int64) error
	IncrementComments(ctx context.Context, postID int64) error
	DecrementComments(ctx context.Context, postID int64) error
	AddLike(ctx context.Context, postID, userID int64) error
	RemoveLike(ctx context.Context, postID, userID int64) error
	HasLiked(ctx context.Context, postID, userID int64) (bool, error)
	UpdateAuthorInfo(ctx context.Context, authorID int64, username, avatarURL string) error
}

type postRepository struct {
	db *pgxpool.Pool
}

func NewPostRepository(db *pgxpool.Pool) PostRepository {
	return &postRepository{db: db}
}

const postSelectFields = `
	id, title, content, author_id, author_username, author_avatar_url,
	tags, views, likes_count, comments_count, created_at, updated_at
`

func scanPost(row pgx.Row) (*domain.Post, error) {
	var p domain.Post
	err := row.Scan(
		&p.ID, &p.Title, &p.Content,
		&p.AuthorID, &p.AuthorUsername, &p.AuthorAvatarURL,
		&p.Tags, &p.Views, &p.LikesCount, &p.CommentsCount,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func scanPosts(rows pgx.Rows) ([]domain.Post, error) {
	posts := make([]domain.Post, 0)
	for rows.Next() {
		var p domain.Post
		if err := rows.Scan(
			&p.ID, &p.Title, &p.Content,
			&p.AuthorID, &p.AuthorUsername, &p.AuthorAvatarURL,
			&p.Tags, &p.Views, &p.LikesCount, &p.CommentsCount,
			&p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, rows.Err()
}

func (r *postRepository) CreatePost(ctx context.Context, post *domain.Post) (int64, error) {
	query := `
		INSERT INTO posts (title, content, author_id, author_username, author_avatar_url, tags)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`
	var id int64
	var createdAt, updatedAt time.Time
	err := r.db.QueryRow(ctx, query,
		post.Title, post.Content, post.AuthorID,
		post.AuthorUsername, post.AuthorAvatarURL, post.Tags,
	).Scan(&id, &createdAt, &updatedAt)
	if err != nil {
		return 0, err
	}
	post.ID = id
	post.CreatedAt = createdAt
	post.UpdatedAt = updatedAt
	return id, nil
}

func (r *postRepository) GetPost(ctx context.Context, id int64) (*domain.Post, error) {
	query := fmt.Sprintf(`
		SELECT %s FROM posts
		WHERE id = $1 AND deleted_at IS NULL
	`, postSelectFields)
	p, err := scanPost(r.db.QueryRow(ctx, query, id))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("post not found")
		}
		return nil, err
	}
	return p, nil
}

func (r *postRepository) UpdatePost(ctx context.Context, id int64, title *string, content *string, tags []string) (*domain.Post, error) {
	setClauses := []string{"updated_at = NOW()"}
	args := []any{id}
	argIdx := 2

	if title != nil {
		setClauses = append(setClauses, fmt.Sprintf("title = $%d", argIdx))
		args = append(args, *title)
		argIdx++
	}
	if content != nil {
		setClauses = append(setClauses, fmt.Sprintf("content = $%d", argIdx))
		args = append(args, *content)
		argIdx++
	}
	if tags != nil {
		setClauses = append(setClauses, fmt.Sprintf("tags = $%d", argIdx))
		args = append(args, tags)
		argIdx++
	}

	query := fmt.Sprintf(`
		UPDATE posts SET %s
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING %s
	`, strings.Join(setClauses, ", "), postSelectFields)

	p, err := scanPost(r.db.QueryRow(ctx, query, args...))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("post not found")
		}
		return nil, err
	}
	return p, nil
}

func (r *postRepository) DeletePost(ctx context.Context, id int64) error {
	result, err := r.db.Exec(ctx,
		`UPDATE posts SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("post not found")
	}
	return nil
}

func (r *postRepository) ListPostsFiltered(ctx context.Context, limit, offset int, sort, author, tag string) ([]domain.Post, error) {
	conditions := []string{"deleted_at IS NULL"}
	args := []any{}
	argIdx := 1

	if author != "" {
		conditions = append(conditions, fmt.Sprintf("author_username = $%d", argIdx))
		args = append(args, author)
		argIdx++
	}
	if tag != "" {
		conditions = append(conditions, fmt.Sprintf("$%d = ANY(tags)", argIdx))
		args = append(args, tag)
		argIdx++
	}

	orderBy := "created_at DESC"
	switch sort {
	case "hot":
		orderBy = "(likes_count * 0.8 + comments_count * 0.5 + ln(views + 1)) DESC"
	case "top":
		orderBy = "likes_count DESC"
	}

	args = append(args, limit, offset)
	query := fmt.Sprintf(`
		SELECT %s FROM posts
		WHERE %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, postSelectFields,
		strings.Join(conditions, " AND "),
		orderBy, argIdx, argIdx+1)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPosts(rows)
}

func (r *postRepository) SearchPosts(ctx context.Context, query string, limit, offset int) ([]domain.Post, error) {
	q := fmt.Sprintf(`
		SELECT %s FROM posts
		WHERE deleted_at IS NULL
		  AND search_vector @@ plainto_tsquery('russian', $1)
		ORDER BY ts_rank(search_vector, plainto_tsquery('russian', $1)) DESC
		LIMIT $2 OFFSET $3
	`, postSelectFields)

	rows, err := r.db.Query(ctx, q, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPosts(rows)
}

func (r *postRepository) IncrementView(ctx context.Context, postID int64) error {
	_, err := r.db.Exec(ctx, `UPDATE posts SET views = views + 1 WHERE id = $1`, postID)
	return err
}

func (r *postRepository) IncrementLike(ctx context.Context, postID int64) error {
	_, err := r.db.Exec(ctx, `UPDATE posts SET likes_count = likes_count + 1 WHERE id = $1`, postID)
	return err
}

func (r *postRepository) DecrementLike(ctx context.Context, postID int64) error {
	_, err := r.db.Exec(ctx, `UPDATE posts SET likes_count = GREATEST(likes_count - 1, 0) WHERE id = $1`, postID)
	return err
}

func (r *postRepository) IncrementComments(ctx context.Context, postID int64) error {
	_, err := r.db.Exec(ctx, `UPDATE posts SET comments_count = comments_count + 1 WHERE id = $1`, postID)
	return err
}

func (r *postRepository) DecrementComments(ctx context.Context, postID int64) error {
	_, err := r.db.Exec(ctx, `UPDATE posts SET comments_count = GREATEST(comments_count - 1, 0) WHERE id = $1`, postID)
	return err
}

func (r *postRepository) AddLike(ctx context.Context, postID, userID int64) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO post_likes (post_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		postID, userID)
	return err
}

func (r *postRepository) RemoveLike(ctx context.Context, postID, userID int64) error {
	_, err := r.db.Exec(ctx,
		`DELETE FROM post_likes WHERE post_id = $1 AND user_id = $2`,
		postID, userID)
	return err
}

func (r *postRepository) HasLiked(ctx context.Context, postID, userID int64) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM post_likes WHERE post_id = $1 AND user_id = $2)`,
		postID, userID,
	).Scan(&exists)
	return exists, err
}

func (r *postRepository) UpdateAuthorInfo(ctx context.Context, authorID int64, username, avatarURL string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE posts SET author_username = $2, author_avatar_url = $3 WHERE author_id = $1`,
		authorID, username, avatarURL)
	return err
}
