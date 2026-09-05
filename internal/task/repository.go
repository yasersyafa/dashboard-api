package task

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	List(c context.Context, filter Filter) ([]Task, error)
	ListRecent(c context.Context, take int) ([]Task, error)
	Stats(c context.Context) (Stats, error)
	FindByID(c context.Context, id uuid.UUID) (Task, error)
	Create(c context.Context, t Task) (Task, error)
	Update(c context.Context, t Task) (Task, error)
	Delete(c context.Context, id uuid.UUID) error
	DeleteCompleted(c context.Context) (int, error)
}