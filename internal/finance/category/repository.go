package finance

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var (
	ErrCategoryNotFound = errors.New("category not found")
	ErrCategoryNameTaken = errors.New("category name already taken")
	ErrCategoryHasTransactions = errors.New("category still has transactions")
)

type CategoryRepository interface {
	List(ctx context.Context) ([]Category, error)
	ListUsage(ctx context.Context) ([]CategoryUsage, error)
	FindByID(ctx context.Context, id uuid.UUID) (Category, error)
	Create(ctx context.Context, c Category) (Category, error)
	Update(ctx context.Context, c Category) (Category, error)
	Delete(ctx context.Context, id uuid.UUID) error
}