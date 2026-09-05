package task

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var (
	ErrTitleRequired = errors.New("title is required")
	ErrTitleTooLong = errors.New("title must be at most 120 characters")
	ErrNotesTooLong = errors.New("notes must be at most 500 characters")
	ErrInvalidPriority = errors.New("priority must be LOW, MEDIUM, or HIGH")
)

type Service struct {
	r Repository
}

func NewService(r Repository) *Service {
	return &Service{r: r}
}

// Create
type CreateInput struct {
	Title string
	Notes *string
	Priority *Priority
	DueDate *time.Time
}

func (s *Service) Create(ctx context.Context, in CreateInput) (Task, error) {
	if err := validateTitle(in.Title); err != nil {
		return Task{}, err
	}
	if err := validateNotes(in.Notes); err != nil {
		return Task{}, err
	}

	priority := PriorityMedium
	if in.Priority != nil {
		if err := validatePriority(*in.Priority); err != nil {
			return Task{}, err
		}
		priority = *in.Priority
	}
	
	t := Task {
		Title: in.Title,
		Notes: in.Notes,
		Priority: priority,
		DueDate: in.DueDate,
		Done: false,
	}

	created, err := s.r.Create(ctx, t)
	if err != nil {
		return Task{}, fmt.Errorf("create task: %w", err)
	}

	return created, nil
}

// Update
type UpdateInput struct {
	Title *string
	Notes *string
	Priority *Priority
	DueDate *time.Time
	Done *bool
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, in UpdateInput) (Task, error) {
	existing, err := s.r.FindByID(ctx, id)
	if err != nil {
		return Task{}, err
	}

	if in.Title != nil {
		if err := validateTitle(*in.Title); err != nil {
			return Task{}, err
		}
		existing.Title = *in.Title
	}
	if in.Notes != nil {
		if err := validateNotes(in.Notes); err != nil {
			return Task{}, err
		}
		existing.Notes = in.Notes
	}
	if in.Priority != nil {
		if err := validatePriority(*in.Priority); err != nil {
			return  Task{}, err
		}
		existing.Priority = *in.Priority
	}
	if in.DueDate != nil {
		existing.DueDate = in.DueDate
	}
	if in.Done != nil {
		existing.Done = *in.Done
	}

	updated, err := s.r.Update(ctx, existing)
	if err != nil {
		return Task{}, fmt.Errorf("update task: %w", err)
	}
	return updated, nil
}

func (s *Service) Toggle(ctx context.Context, id uuid.UUID) (Task, error) {
	existing, err := s.r.FindByID(ctx, id)
	if err != nil {
		return Task{}, err
	}

	existing.Done = !existing.Done

	updated, err := s.r.Update(ctx, existing)
	if err != nil {
		return Task{}, fmt.Errorf("toggle task: %w", err)
	}
	return  updated, nil
}

// GETTING DATA TASK
func (s *Service) List(ctx context.Context, filter Filter) ([]Task, error) {
	return s.r.List(ctx, filter)
}

func (s *Service) ListRecent(ctx context.Context, take int) ([]Task, error) {
	return s.r.ListRecent(ctx, take)
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (Task, error) {
	return s.r.FindByID(ctx, id)
}

func (s *Service) Stats(ctx context.Context) (Stats, error) {
	return s.r.Stats(ctx)
}

// DELETE
func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.r.Delete(ctx, id)
}

func (s *Service) DeleteCompleted(ctx context.Context) (int, error) {
	return s.r.DeleteCompleted(ctx)
}

// helper functions
func validateTitle(title string) error {
	if title == "" {
		return ErrTitleRequired
	}

	if len(title) > 120 {
		return ErrTitleTooLong
	}

	return nil
}

func validateNotes(notes *string) error {
	if notes != nil && len(*notes) > 500 {
		return ErrNotesTooLong
	}

	return nil
}

func validatePriority(priority Priority) error {
	switch priority {
	case PriorityLow, PriorityMedium, PriorityHigh:
		return nil
	default:
		return ErrInvalidPriority
	}
}