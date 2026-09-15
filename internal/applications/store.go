package applications

import "context"

type ApplicationStore interface {
	GetAll(ctx context.Context, filter applicationFilter) ([]Application, error)
	GetByID(ctx context.Context, id int64) (Application, error)
	Delete(ctx context.Context, id int64) (bool, error)
	UpdateStatus(ctx context.Context, id int64, status string) (Application, error)
}
