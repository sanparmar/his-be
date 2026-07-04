---
name: new-domain-event
description: "Generate a Kafka domain event publisher and consumer for an existing HIS microservice. Use when: adding a new domain event, wiring cross-service async communication, publishing state changes to Kafka."
---

# New Domain Event

Generate a Kafka domain event publisher and consumer for the following event.

## Inputs

- **Service/Domain**: ${input:domain:Service domain that owns this event (e.g. patient, appointment)}
- **Entity**: ${input:entity:The entity that changed (e.g. Patient, Appointment)}
- **Verb (past tense)**: ${input:verb:What happened (e.g. Registered, Booked, Cancelled, Discharged)}
- **Payload fields**: ${input:fields:Comma-separated fields to include in the event payload}

## Generated Topic Name

```
his.${domain}.${entity:lower}.${verb:lower}
```

## Files to Generate

### 1. Event Type Definition
File: `services/${domain}/internal/domain/events.go` (add to existing file or create)

```go
// ${entity}${verb}Event is published when a ${entity:lower} is ${verb:lower}.
type ${entity}${verb}Event struct {
    EventID       uuid.UUID `json:"event_id"`
    TenantID      uuid.UUID `json:"tenant_id"`
    ${entity}ID   uuid.UUID `json:"${entity:lower}_id"`
    OccurredAt    time.Time `json:"occurred_at"`
    CorrelationID string    `json:"correlation_id"`
    // payload fields derived from input
}

func (e ${entity}${verb}Event) Topic() string {
    return "his.${domain}.${entity:lower}.${verb:lower}"
}
```

### 2. Event Publisher (Kafka Producer)
File: `services/${domain}/internal/infrastructure/kafka/event_publisher.go`

- Use `github.com/segmentio/kafka-go`
- Serialize events to JSON
- Include OTel span propagation in Kafka message headers
- Retry with exponential backoff (max 3 attempts)
- Log successful publish and errors via zerolog
- Never panic; return wrapped errors

```go
func (p *EventPublisher) Publish${entity}${verb}(ctx context.Context, event domain.${entity}${verb}Event) error {
    ctx, span := p.tracer.Start(ctx, "Kafka.Publish.${entity}${verb}")
    defer span.End()
    // ... marshal, inject headers, write message
}
```

### 3. Event Consumer (Kafka Reader) — Optional
File: `services/${domain}/internal/infrastructure/kafka/${entity:lower}_${verb:lower}_consumer.go`

Only generate if this service needs to react to this event from another service.

- Consumer group: `his-${domain}-${entity:lower}-${verb:lower}`
- Idempotent handler (check if already processed using event_id)
- Dead letter topic on repeated failure: `his.${domain}.dlq`
- Commit offset only after successful processing

### 4. Unit Tests
File: `services/${domain}/internal/infrastructure/kafka/event_publisher_test.go`

- Mock `kafka-go` writer interface
- Test: successful publish, serialization, header injection, retry on transient error

## Quality Check

- Event payload must NOT contain PHI fields directly — use reference IDs only
- `EventID` must be a new UUID generated at publish time, not the entity ID
- Always set `OccurredAt` to `time.Now().UTC()`
- Correlation ID must be propagated from the incoming request context
