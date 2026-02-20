package event

import (
	"context"
	"encoding/json"
	"log"
	"post-service/internal/repository"

	"github.com/redis/go-redis/v9"
)

type incomingEvent struct {
	Event     string `json:"event"`
	PostID    int64  `json:"post_id"`
	UserID    int64  `json:"user_id"`
	Username  string `json:"username"`
	AvatarURL string `json:"avatar_url"`
}

type Consumer struct {
	redis *redis.Client
	repo  repository.PostRepository
}

func NewConsumer(redis *redis.Client, repo repository.PostRepository) *Consumer {
	return &Consumer{redis: redis, repo: repo}
}

func (c *Consumer) Start(ctx context.Context) {
	pubsub := c.redis.Subscribe(ctx, "comment_events", "profile_events")
	defer pubsub.Close()

	log.Println("[consumer] subscribed to comment_events, profile_events")

	for {
		select {
		case <-ctx.Done():
			log.Println("[consumer] stopped")
			return
		case msg, ok := <-pubsub.Channel():
			if !ok {
				return
			}
			c.handle(ctx, msg.Channel, msg.Payload)
		}
	}
}

func (c *Consumer) handle(ctx context.Context, channel, payload string) {
	var evt incomingEvent
	if err := json.Unmarshal([]byte(payload), &evt); err != nil {
		log.Printf("[consumer] bad payload on %s: %v", channel, err)
		return
	}

	switch evt.Event {
	case "CommentCreated":
		if err := c.repo.IncrementComments(ctx, evt.PostID); err != nil {
			log.Printf("[consumer] IncrementComments(%d): %v", evt.PostID, err)
		}
	case "CommentDeleted":
		if err := c.repo.DecrementComments(ctx, evt.PostID); err != nil {
			log.Printf("[consumer] DecrementComments(%d): %v", evt.PostID, err)
		}
	case "ProfileUpdated":
		if err := c.repo.UpdateAuthorInfo(ctx, evt.UserID, evt.Username, evt.AvatarURL); err != nil {
			log.Printf("[consumer] UpdateAuthorInfo(%d): %v", evt.UserID, err)
		}
	}
}
