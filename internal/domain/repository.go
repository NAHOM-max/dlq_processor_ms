package domain

import "context"

type DLQRepository interface {
	Save(ctx context.Context, event DLQEventRecord) error
	List(ctx context.Context, limit int) ([]DLQEventRecord, error)
	GetByID(ctx context.Context, id string) (DLQEventRecord, error)
}
