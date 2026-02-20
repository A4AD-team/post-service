package dto

type CreatePostRequest struct {
	Title   string   `json:"title"   validate:"required,min=3,max=300"`
	Content string   `json:"content" validate:"required,min=10,max=50000"`
	Tags    []string `json:"tags,omitempty"`
}

type UpdatePostRequest struct {
	Title   *string  `json:"title,omitempty"`
	Content *string  `json:"content,omitempty"`
	Tags    []string `json:"tags,omitempty"`
}

type ListPostsQuery struct {
	Limit  int    `query:"limit"`
	Offset int    `query:"offset"`
	Sort   string `query:"sort"`   // new | hot | top
	Author string `query:"author"` // фильтр по username
	Tag    string `query:"tag"`    // фильтр по тегу
}
