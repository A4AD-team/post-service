ALTER TABLE posts
    ADD COLUMN author_username   VARCHAR(100) NOT NULL DEFAULT '',
    ADD COLUMN author_avatar_url TEXT         NOT NULL DEFAULT '';

ALTER TABLE posts
    ADD COLUMN search_vector tsvector
        GENERATED ALWAYS AS (
            to_tsvector('russian', coalesce(title, '')) ||
            to_tsvector('russian', coalesce(content, ''))
        ) STORED;

CREATE INDEX idx_posts_search ON posts USING GIN(search_vector);
CREATE INDEX idx_posts_tags   ON posts USING GIN(tags);
CREATE INDEX idx_posts_hot    ON posts (likes_count DESC, comments_count DESC, views DESC)
    WHERE deleted_at IS NULL;
