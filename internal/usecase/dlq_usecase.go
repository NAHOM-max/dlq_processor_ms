package usecase

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"dlq_processor_ms/internal/domain"
)

type DLQUseCase struct {
	repo   domain.DLQRepository
	logger *slog.Logger
}

func NewDLQUseCase(repo domain.DLQRepository, logger *slog.Logger) *DLQUseCase {
	return &DLQUseCase{repo: repo, logger: logger}
}

func (uc *DLQUseCase) Handle(ctx context.Context, event domain.DLQEvent, topic string, eventID string) error {
	now := time.Now().UTC()
	record := domain.DLQEventRecord{
		ID:            uuid.NewString(),
		EventID:       eventID,
		Topic:         topic,
		Payload:       event.OriginalEvent,
		Error:         event.Error,
		SourceService: event.Service,
		RetryCount:    event.RetryCount,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	uc.logger.InfoContext(ctx, "dlq event received",
		"event_id", eventID,
		"source_service", event.Service,
		"error_reason", event.Error,
		"retry_count", event.RetryCount,
		"topic", topic,
	)

	if err := uc.repo.Save(ctx, record); err != nil {
		uc.logger.ErrorContext(ctx, "failed to persist dlq event",
			"event_id", eventID,
			"error", err,
		)
		return err
	}

	return nil
}
