package task

import (
	"time"

	"github.com/google/uuid"
)

type Priority string

const (
	PriorityLow Priority = "LOW"
	PriorityMedium Priority = "MEDIUM"
	PriorityHigh Priority = "HIGH"
)

type Task struct {
	ID uuid.UUID
	Title string
	Notes *string // buat field yang nullable
	Done bool
	Priority Priority
	DueDate *time.Time // buat field yang nullable
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Filter string

const (
	FilterAll Filter = "all"
	FilterActive Filter = "active"
	FilterDone Filter = "done"
)

type Stats struct {
	Total int
	Done int
	Active int
	Overdue int
	PercentDone float64
}