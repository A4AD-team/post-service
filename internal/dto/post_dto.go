package dto

type CreatePostRequest struct {
	Title    string   `json:"title"   validate:"required,min=3,max=300"`
	Content  string   `json:"content" validate:"required,min=10,max=50000"`
	Tags     []string `json:"tags,omitempty"`
	AuthorID int64    `json:"author_id" validate:"required"`
}

type UpdatePostRequest struct {
	Title    *string  `json:"title,omitempty"`
	Content  *string  `json:"content,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	AuthorID int64    `json:"author_id" validate:"required"`
}

type ListPostsQuery struct {
	Limit  int    `query:"limit"`
	Offset int    `query:"offset"`
	Sort   string `query:"sort"`
	Author string `query:"author"`
	Tag    string `query:"tag"`
}
