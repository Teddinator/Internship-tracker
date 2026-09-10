package applications

import "context"

type ApplicationStore interface {
	GetByID(ctx context.Context, id int64) (Application, error)
	Delete(ctx context.Context, id int64) (bool, error)
}
