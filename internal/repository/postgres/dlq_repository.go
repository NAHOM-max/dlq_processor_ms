package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"dlq_processor_ms/internal/domain"
)

type dlqRepository struct {
	db *pgxpool.Pool
}

func NewDLQRepository(db *pgxpool.Pool) domain.DLQRepository {
	return &dlqRepository{db: db}
}

func (r *dlqRepository) Save(ctx context.Context, e domain.DLQEventRecord) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO dlq_events
			(id, event_id, topic, payload, error, source_service, retry_count, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		e.ID, e.EventID, e.Topic, e.Payload, e.Error,
		e.SourceService, e.RetryCount, e.CreatedAt, e.UpdatedAt,
	)
	return err
}

func (r *dlqRepository) List(ctx context.Context, limit int) ([]domain.DLQEventRecord, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, event_id, topic, payload, error, source_service, retry_count, created_at, updated_at
		FROM dlq_events ORDER BY created_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []domain.DLQEventRecord
	for rows.Next() {
		var e domain.DLQEventRecord
		if err := rows.Scan(&e.ID, &e.EventID, &e.Topic, &e.Payload, &e.Error,
			&e.SourceService, &e.RetryCount, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		records = append(records, e)
	}
	return records, rows.Err()
}

func (r *dlqRepository) GetByID(ctx context.Context, id string) (domain.DLQEventRecord, error) {
	var e domain.DLQEventRecord
	err := r.db.QueryRow(ctx, `
		SELECT id, event_id, topic, payload, error, source_service, retry_count, created_at, updated_at
		FROM dlq_events WHERE id = $1`, id).
		Scan(&e.ID, &e.EventID, &e.Topic, &e.Payload, &e.Error,
			&e.SourceService, &e.RetryCount, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return domain.DLQEventRecord{}, fmt.Errorf("dlq event not found: %w", err)
	}
	return e, nil
}
