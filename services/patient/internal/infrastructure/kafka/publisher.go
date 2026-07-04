package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/segmentio/kafka-go"

	"github.com/deloitte-us-consulting/his-be/services/patient/internal/domain"
)

// EventPublisher publishes domain events to Kafka.
type EventPublisher struct {
	writer *kafka.Writer
}

// NewEventPublisher creates a new Kafka event publisher.
func NewEventPublisher(writer *kafka.Writer) *EventPublisher {
	return &EventPublisher{writer: writer}
}

// Publish publishes a domain event to Kafka.
func (p *EventPublisher) Publish(ctx context.Context, event domain.DomainEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	message := kafka.Message{
		Topic: "his.patient.patient." + extractEventSuffix(event.EventType()),
		Key:   []byte(event.AggregateID().String()),
		Value: payload,
		Headers: []kafka.Header{
			{
				Key:   "X-Event-Type",
				Value: []byte(event.EventType()),
			},
			{
				Key:   "X-Aggregate-ID",
				Value: []byte(event.AggregateID().String()),
			},
			{
				Key:   "X-Tenant-ID",
				Value: []byte(event.AggregateTenantID().String()),
			},
		},
	}

	if err := p.writer.WriteMessages(ctx, message); err != nil {
		return fmt.Errorf("publish event: %w", err)
	}

	return nil
}

// Close closes the Kafka writer.
func (p *EventPublisher) Close() error {
	return p.writer.Close()
}

// Helper to extract event suffix from event type.
func extractEventSuffix(eventType string) string {
	// his.patient.patient.registered -> registered
	parts := len(eventType)
	lastDot := 0
	for i := len(eventType) - 1; i >= 0; i-- {
		if eventType[i] == '.' {
			lastDot = i + 1
			break
		}
	}
	if lastDot > 0 && lastDot < parts {
		return eventType[lastDot:]
	}
	return "unknown"
}
