package kafka

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/IBM/sarama"

	"dlq_processor_ms/internal/domain"
	"dlq_processor_ms/internal/usecase"
)

const (
	Topic         = "delivery.confirmed.dlq"
	ConsumerGroup = "dlq-processor-group"
)

type DLQConsumerHandler struct {
	uc     *usecase.DLQUseCase
	logger *slog.Logger
}

func NewDLQConsumerHandler(uc *usecase.DLQUseCase, logger *slog.Logger) *DLQConsumerHandler {
	return &DLQConsumerHandler{uc: uc, logger: logger}
}

// Setup is called at the beginning of a new session.
func (h *DLQConsumerHandler) Setup(_ sarama.ConsumerGroupSession) error { return nil }

// Cleanup is called at the end of a session.
func (h *DLQConsumerHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

// ConsumeClaim processes messages from a partition claim.
func (h *DLQConsumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		h.process(session.Context(), msg)
		session.MarkMessage(msg, "")
	}
	return nil
}

func (h *DLQConsumerHandler) process(ctx context.Context, msg *sarama.ConsumerMessage) {
	var event domain.DLQEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		h.logger.ErrorContext(ctx, "failed to deserialize dlq message",
			"topic", msg.Topic,
			"partition", msg.Partition,
			"offset", msg.Offset,
			"error", err,
		)
		return
	}

	eventID := extractHeader(msg.Headers, "event_id")
	correlationID := extractHeader(msg.Headers, "correlation_id")

	ctx = context.WithValue(ctx, correlationIDKey{}, correlationID)

	if err := h.uc.Handle(ctx, event, msg.Topic, eventID); err != nil {
		h.logger.ErrorContext(ctx, "use case failed to handle dlq event",
			"event_id", eventID,
			"error", err,
		)
	}
}

type correlationIDKey struct{}

func extractHeader(headers []*sarama.RecordHeader, key string) string {
	for _, h := range headers {
		if string(h.Key) == key {
			return string(h.Value)
		}
	}
	return ""
}
