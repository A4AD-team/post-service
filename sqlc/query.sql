-- name: CreatePost :one
INSERT INTO posts (title, content, author_id, tags)
VALUES ($1, $2, $3, $4)
RETURNING id, title, content, author_id, tags, views, likes_count, comments_count, created_at, updated_at;

-- name: GetPostByID :one
SELECT id, title, content, author_id, tags, views, likes_count, comments_count, created_at, updated_at
FROM posts
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListPostsNew :many
SELECT id, title, content, author_id, tags, views, likes_count, comments_count, created_at, updated_at
FROM posts
WHERE deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: IncrementView :exec
UPDATE posts SET views = views + 1 WHERE id = $1;

-- name: IncrementLike :exec
UPDATE posts SET likes_count = likes_count + 1 WHERE id = $1;

-- name: DecrementLike :exec
UPDATE posts SET likes_count = likes_count - 1 WHERE id = $1 AND likes_count > 0;