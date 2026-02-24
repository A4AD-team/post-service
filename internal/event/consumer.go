package event

import (
	"context"
	"encoding/json"
	"log"
	"post-service/internal/repository"

	amqp "github.com/rabbitmq/amqp091-go"
)

type incomingEvent struct {
	Event     string `json:"event"`
	PostID    int64  `json:"post_id"`
	UserID    int64  `json:"user_id"`
	Username  string `json:"username"`
	AvatarURL string `json:"avatar_url"`
}

type Consumer struct {
	conn *amqp.Connection
	repo repository.PostRepository
}

func NewConsumer(amqpURL string, repo repository.PostRepository) (*Consumer, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, err
	}
	return &Consumer{conn: conn, repo: repo}, nil
}

func (c *Consumer) Start(ctx context.Context) {
	ch, err := c.conn.Channel()
	if err != nil {
		log.Fatalf("[consumer] channel error: %v", err)
	}
	defer ch.Close()

	queues := []string{"comment_events", "profile_events"}
	for _, q := range queues {
		ch.QueueDeclare(q, true, false, false, false, nil)
	}

	commentMsgs, _ := ch.Consume("comment_events", "", true, false, false, false, nil)
	profileMsgs, _ := ch.Consume("profile_events", "", true, false, false, false, nil)

	log.Println("[consumer] subscribed to comment_events, profile_events")

	for {
		select {
		case <-ctx.Done():
			log.Println("[consumer] stopped")
			return
		case msg := <-commentMsgs:
			c.handle(ctx, "comment_events", string(msg.Body))
		case msg := <-profileMsgs:
			c.handle(ctx, "profile_events", string(msg.Body))
		}
	}
}

func (c *Consumer) Close() {
	c.conn.Close()
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
