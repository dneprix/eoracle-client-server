package queue

import (
	"context"
	"fmt"
	"log"

	"eoracle-client-server/internal/commands"

	"github.com/streadway/amqp"
)

// RabbitMQQueue implements Queue interface using RabbitMQ
type RabbitMQQueue struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   amqp.Queue
}

// NewRabbitMQQueue creates a new RabbitMQ queue
func NewRabbitMQQueue(url, queueName string) (Queue, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	q, err := ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	return &RabbitMQQueue{
		conn:    conn,
		channel: ch,
		queue:   q,
	}, nil
}

// Publish sends a command to the queue
func (r *RabbitMQQueue) Publish(command commands.Command) error {
	body, err := command.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize command: %w", err)
	}

	err = r.channel.Publish(
		"",           // exchange
		r.queue.Name, // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

// Subscribe listens for messages and handles them
func (r *RabbitMQQueue) Subscribe(ctx context.Context, handler func(commands.Command) error) error {
	msgs, err := r.channel.Consume(
		r.queue.Name, // queue
		"",           // consumer
		false,        // auto-ack
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping queue subscription")
			return nil
		case msg := <-msgs:
			cmd, err := commands.FromJSON(msg.Body)
			if err != nil {
				log.Printf("Failed to parse command: %v", err)
				msg.Nack(false, false) // Do not requeue the message
				continue
			}

			if err := handler(cmd); err != nil {
				log.Printf("Failed to handle command: %v", err)
				msg.Nack(false, true) // Requeue the message
				continue
			}

			msg.Ack(false)
		}
	}
}

// Close closes the connection
func (r *RabbitMQQueue) Close() error {
	if r.channel != nil {
		r.channel.Close()
	}
	if r.conn != nil {
		r.conn.Close()
	}
	return nil
}
