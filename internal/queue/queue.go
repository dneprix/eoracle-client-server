package queue

import (
	"context"
	"eoracle-client-server/internal/commands"
)

// Queue interface defines the operations for message queue
type Queue interface {
	Publish(command commands.Command) error
	Subscribe(ctx context.Context, handleCommand func(commands.Command) error) error
	Close() error
}
