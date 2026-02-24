package event

import (
	"context"
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	conn *amqp.Connection
}

func NewPublisher(amqpURL string) (*Publisher, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, err
	}
	return &Publisher{conn: conn}, nil
}

func (p *Publisher) Publish(_ context.Context, queue string, payload map[string]interface{}) {
	ch, err := p.conn.Channel()
	if err != nil {
		log.Printf("[publisher] channel error: %v", err)
		return
	}
	defer ch.Close()

	ch.QueueDeclare(queue, true, false, false, false, nil)

	data, _ := json.Marshal(payload)
	ch.Publish("", queue, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        data,
	})
}

func (p *Publisher) Close() {
	p.conn.Close()
}
