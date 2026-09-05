package finance

import (
	"time"

	"github.com/google/uuid"
)

type Category struct {
	ID uuid.UUID
	Name string
	NameKey string
	Budget *int64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CategoryUsage struct {
	Category
	Spent int64
	PercentUsed float64
	OverBudget bool
}