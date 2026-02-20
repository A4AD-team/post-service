CREATE TABLE posts (
    id             BIGSERIAL PRIMARY KEY,
    title          VARCHAR(300)  NOT NULL,
    content        TEXT          NOT NULL,
    author_id      BIGINT        NOT NULL,
    tags           TEXT[],
    views          BIGINT        DEFAULT 0   NOT NULL,
    likes_count    BIGINT        DEFAULT 0   NOT NULL,
    comments_count BIGINT        DEFAULT 0   NOT NULL,
    created_at     TIMESTAMPTZ   DEFAULT NOW() NOT NULL,
    updated_at     TIMESTAMPTZ   DEFAULT NOW() NOT NULL,
    deleted_at     TIMESTAMPTZ   DEFAULT NULL
);

CREATE INDEX idx_posts_author_id  ON posts(author_id);
CREATE INDEX idx_posts_created_at ON posts(created_at DESC);
