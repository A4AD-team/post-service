package domain

import "time"

type Post struct {
	ID              int64     `json:"id"`
	Title           string    `json:"title"`
	Content         string    `json:"content"`
	AuthorID        int64     `json:"authorId"`
	AuthorUsername  string    `json:"authorUsername"`
	AuthorAvatarURL string    `json:"authorAvatarUrl"`
	Tags            []string  `json:"tags,omitempty"`
	Views           int64     `json:"views"`
	LikesCount      int64     `json:"likesCount"`
	CommentsCount   int64     `json:"commentsCount"`
	IsLikedByMe     bool      `json:"isLikedByMe"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}
