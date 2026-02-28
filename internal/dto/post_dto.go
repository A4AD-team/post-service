package dto

type CreatePostRequest struct {
	Title    string   `json:"title"`
	Content  string   `json:"content"`
	Tags     []string `json:"tags,omitempty"`
	AuthorID int64    `json:"-"`
}

type UpdatePostRequest struct {
	Title    *string  `json:"title,omitempty"`
	Content  *string  `json:"content,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	AuthorID int64    `json:"author_id"`
}

type ListPostsQuery struct {
	Limit  int    `query:"limit"`
	Offset int    `query:"offset"`
	Sort   string `query:"sort"`
	Author string `query:"author"`
	Tag    string `query:"tag"`
}
