package domain

import (
	"encoding/json"
	"time"
)

// DLQEvent is the raw message consumed from Kafka.
type DLQEvent struct {
	OriginalEvent json.RawMessage `json:"original_event"`
	Error         string          `json:"error"`
	Service       string          `json:"service"`
	RetryCount    int             `json:"retry_count"`
	FailedAt      time.Time       `json:"failed_at"`
}

// DLQEventRecord is the persisted representation.
type DLQEventRecord struct {
	ID            string          `db:"id"`
	EventID       string          `db:"event_id"`
	Topic         string          `db:"topic"`
	Payload       json.RawMessage `db:"payload"`
	Error         string          `db:"error"`
	SourceService string          `db:"source_service"`
	RetryCount    int             `db:"retry_count"`
	CreatedAt     time.Time       `db:"created_at"`
	UpdatedAt     time.Time       `db:"updated_at"`
}
