package finance

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

var (
	ErrCategoryNameRequired = errors.New("category name is required")
	ErrCategoryNameTooLong  = errors.New("category name must be at most 40 characters")
	ErrCategoryBudgetInvalid = errors.New("budget must be greater than 0")
)

type CategoryService struct {
	repo CategoryRepository
}

func NewCategoryService(repo CategoryRepository) *CategoryService {
	return &CategoryService{repo: repo}
}

type CreateCategoryInput struct {
	Name string
	Budget *int64
}

func (s *CategoryService) Create(ctx context.Context, in CreateCategoryInput) (Category, error) {
	if err := validateCategoryName(in.Name); err != nil {
		return Category{}, err
	}

	if err := validateCategoryBudget(in.Budget); err != nil {
		return Category{}, err
	}

	created, err := s.repo.Create(ctx, Category{
		Name: in.Name,
		Budget: in.Budget,
	})

	if err != nil {
		return Category{}, fmt.Errorf("create category: %w", err)
	}

	return created, nil
}

type UpdateCategoryInput struct {
	Name *string
	Budget *int64
}

func (s *CategoryService) Update(ctx context.Context, id uuid.UUID, in UpdateCategoryInput) (Category, error) {
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return Category{}, err
	}

	if in.Name != nil {
		if err := validateCategoryName(*in.Name); err != nil {
			return  Category{}, err
		}
		existing.Name = *in.Name
	}
	if in.Budget!= nil {
		if err := validateCategoryBudget(in.Budget); err != nil {
			return Category{}, err
		}
		existing.Budget = in.Budget
	}

	updated, err := s.repo.Update(ctx, existing)

	if err != nil {
		return Category{}, fmt.Errorf("update category: %w", err)
	}

	return updated, nil
}

func (s *CategoryService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *CategoryService) Get(ctx context.Context, id uuid.UUID) (Category, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *CategoryService) List(ctx context.Context) ([]Category, error) {
	return s.repo.List(ctx)
}

func (s *CategoryService) ListUsage(ctx context.Context) ([]CategoryUsage, error) {
	return s.repo.ListUsage(ctx)
}

// helper functions
func validateCategoryName(name string) error {
	if name == "" {
		return ErrCategoryNameRequired
	}

	if len(name) > 40 {
		return ErrCategoryNameTooLong
	}

	return nil
}

func validateCategoryBudget(budget *int64) error {
	if budget != nil && *budget <= 0 {
		return ErrCategoryBudgetInvalid
	}

	return nil
}