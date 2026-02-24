DROP INDEX IF EXISTS idx_posts_search;
DROP INDEX IF EXISTS idx_posts_tags;
DROP INDEX IF EXISTS idx_posts_hot;

ALTER TABLE posts
    DROP COLUMN IF EXISTS search_vector,
    DROP COLUMN IF EXISTS author_username,
    DROP COLUMN IF EXISTS author_avatar_url;
