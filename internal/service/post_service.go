package service

import (
	"context"
	"fmt"
	"post-service/internal/domain"
	"post-service/internal/dto"
	"post-service/internal/event"
	"post-service/internal/repository"
	"time"

	"github.com/redis/go-redis/v9"
)

type PostService interface {
	Create(ctx context.Context, authorID int64, username, avatarURL string, req *dto.CreatePostRequest) (*domain.Post, error)
	GetPost(ctx context.Context, id int64, userID int64) (*domain.Post, error)
	UpdatePost(ctx context.Context, id, userID int64, req *dto.UpdatePostRequest) (*domain.Post, error)
	DeletePost(ctx context.Context, id, userID int64) error
	ListPosts(ctx context.Context, q dto.ListPostsQuery, userID int64) ([]domain.Post, error)
	SearchPosts(ctx context.Context, query string, limit, offset int) ([]domain.Post, error)
	IncrementView(ctx context.Context, postID, userID int64) error
	Like(ctx context.Context, postID, userID int64) error
	Unlike(ctx context.Context, postID, userID int64) error
}

type postService struct {
	repo      repository.PostRepository
	redis     *redis.Client
	publisher *event.Publisher
}

func NewPostService(repo repository.PostRepository, redis *redis.Client, publisher *event.Publisher) PostService {
	return &postService{repo: repo, redis: redis, publisher: publisher}
}

func (s *postService) Create(ctx context.Context, authorID int64, username, avatarURL string, req *dto.CreatePostRequest) (*domain.Post, error) {
	post := &domain.Post{
		Title:           req.Title,
		Content:         req.Content,
		AuthorID:        authorID,
		AuthorUsername:  username,
		AuthorAvatarURL: avatarURL,
		Tags:            req.Tags,
	}
	id, err := s.repo.CreatePost(ctx, post)
	if err != nil {
		return nil, err
	}
	s.publishEvent(ctx, "PostCreated", map[string]interface{}{
		"post_id":   id,
		"author_id": authorID,
		"title":     post.Title,
	})
	return post, nil
}

func (s *postService) GetPost(ctx context.Context, id int64, userID int64) (*domain.Post, error) {
	post, err := s.repo.GetPost(ctx, id)
	if err != nil {
		return nil, err
	}
	if userID != 0 {
		post.IsLikedByMe, _ = s.repo.HasLiked(ctx, id, userID)
	}
	return post, nil
}

func (s *postService) ListPosts(ctx context.Context, q dto.ListPostsQuery, userID int64) ([]domain.Post, error) {
	posts, err := s.repo.ListPostsFiltered(ctx, q.Limit, q.Offset, q.Sort, q.Author, q.Tag)
	if err != nil {
		return nil, err
	}
	// We set IsLikedByMe if the user is authorized
	if userID != 0 {
		for i := range posts {
			posts[i].IsLikedByMe, _ = s.repo.HasLiked(ctx, posts[i].ID, userID)
		}
	}
	return posts, nil
}

func (s *postService) SearchPosts(ctx context.Context, query string, limit, offset int) ([]domain.Post, error) {
	if query == "" {
		return []domain.Post{}, nil
	}
	return s.repo.SearchPosts(ctx, query, limit, offset)
}

func (s *postService) UpdatePost(ctx context.Context, id, userID int64, req *dto.UpdatePostRequest) (*domain.Post, error) {
	existing, err := s.repo.GetPost(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing.AuthorID != userID {
		return nil, fmt.Errorf("forbidden")
	}
	post, err := s.repo.UpdatePost(ctx, id, req.Title, req.Content, req.Tags)
	if err != nil {
		return nil, err
	}
	s.redis.Del(ctx, fmt.Sprintf("post:%d", id))
	s.publishEvent(ctx, "PostUpdated", map[string]interface{}{
		"post_id":   id,
		"author_id": userID,
	})
	return post, nil
}

func (s *postService) DeletePost(ctx context.Context, id, userID int64) error {
	existing, err := s.repo.GetPost(ctx, id)
	if err != nil {
		return err
	}
	if existing.AuthorID != userID {
		return fmt.Errorf("forbidden")
	}
	if err := s.repo.DeletePost(ctx, id); err != nil {
		return err
	}
	s.redis.Del(ctx, fmt.Sprintf("post:%d", id))
	s.publishEvent(ctx, "PostDeleted", map[string]interface{}{
		"post_id":   id,
		"author_id": userID,
	})
	return nil
}

func (s *postService) IncrementView(ctx context.Context, postID, userID int64) error {
	key := fmt.Sprintf("post:%d:viewed_by:%d", postID, userID)
	set, err := s.redis.SetNX(ctx, key, "1", 24*time.Hour).Result()
	if err != nil || !set {
		return err
	}
	return s.repo.IncrementView(ctx, postID)
}

func (s *postService) Like(ctx context.Context, postID, userID int64) error {
	hasLiked, err := s.repo.HasLiked(ctx, postID, userID)
	if err != nil {
		return err
	}
	if hasLiked {
		return fmt.Errorf("already liked")
	}
	if err := s.repo.AddLike(ctx, postID, userID); err != nil {
		return err
	}
	if err := s.repo.IncrementLike(ctx, postID); err != nil {
		return err
	}
	s.publishEvent(ctx, "PostLiked", map[string]interface{}{
		"post_id": postID,
		"user_id": userID,
	})
	return nil
}

func (s *postService) Unlike(ctx context.Context, postID, userID int64) error {
	hasLiked, err := s.repo.HasLiked(ctx, postID, userID)
	if err != nil {
		return err
	}
	if !hasLiked {
		return fmt.Errorf("not liked")
	}
	if err := s.repo.RemoveLike(ctx, postID, userID); err != nil {
		return err
	}
	if err := s.repo.DecrementLike(ctx, postID); err != nil {
		return err
	}
	s.publishEvent(ctx, "PostUnliked", map[string]interface{}{
		"post_id": postID,
		"user_id": userID,
	})
	return nil
}

func (s *postService) publishEvent(ctx context.Context, eventType string, payload map[string]interface{}) {
	payload["event"] = eventType
	s.publisher.Publish(ctx, "post_events", payload)
}
