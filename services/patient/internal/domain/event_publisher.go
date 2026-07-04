package domain

import "context"

// EventPublisher is the port for publishing domain events to Kafka.
type EventPublisher interface {
	Publish(ctx context.Context, event DomainEvent) error
}
