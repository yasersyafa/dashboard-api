package finance

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresCategoryRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresCategoryRepository(pool *pgxpool.Pool) *PostgresCategoryRepository {
	return &PostgresCategoryRepository{pool: pool}
}

func (r *PostgresCategoryRepository) List(ctx context.Context) ([]Category, error) {
	query := `
		SELECT id, name, name_key, budget, created_at, updated_at
		FROM categories
		ORDER BY name ASC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return  nil, fmt.Errorf("query categories: %w", err)
	}
	defer rows.Close()

	categories := []Category{}
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.ID, &c.Name, &c.NameKey, &c.Budget, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}
		categories = append(categories, c)
	}

	return categories, rows.Err()
}

func (r *PostgresCategoryRepository) ListUsage(ctx context.Context, id uuid.UUID) ([]CategoryUsage, error) {
	query := `
		SELECT c.id, c.name, c.name_key, c.budget, c.created_at, c.updated_at,
		COALESCE(SUM(t.amount) FILTER (
			WHERE t.type = 'EXPENSE'
			AND date_trunc('month', t.occured_at) = date_trunc('month', now())
		), 0) AS spent
		FROM categories c
		LEFT JOIN transactions t ON t.category_id = c.id
		GROUP BY c.id
		ORDER BY c.name ASC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query categories usage: %w", err)
	}
	defer rows.Close()

	categoryUsage := []CategoryUsage{}
	for rows.Next() {
		var c CategoryUsage
		if err := rows.Scan(&c.ID, &c.Name, &c.NameKey, &c.Budget, &c.CreatedAt, &c.UpdatedAt, &c.Spent); err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}

		if c.Budget != nil && *c.Budget > 0 {
			c.PercentUsed = float64(c.Spent) / float64(*c.Budget) * 100
			c.OverBudget = c.Spent > *c.Budget
		}
		categoryUsage = append(categoryUsage, c)
	}

	return categoryUsage, rows.Err()
}

func (r *PostgresCategoryRepository) FindByID(ctx context.Context, id uuid.UUID) (Category, error) {
	query := `
		SELECT id, name, name_key, budget, created_at, updated_at 
		FROM categories
		WHERE id = $1
	`
	var c Category
	err := r.pool.QueryRow(ctx, query, id).Scan(&c.ID, &c.Name, &c.NameKey, &c.Budget, &c.CreatedAt, &c.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Category{}, ErrCategoryNotFound
	}
	if err != nil {
		return Category{}, fmt.Errorf("query categories: %w", err)
	}

	return c, nil
}

func (r *PostgresCategoryRepository) Create(ctx context.Context, c Category) (Category, error) {
	nameKey := strings.ToLower(c.Name)
	var created Category
	query := `
		INSERT INTO categories (name, name_key, budget)
		VALUES ($1, $2, $3)
		RETURNING id, name, name_key, budget, created_at, updated_at
	`

	err := r.pool.QueryRow(ctx, query, c.Name, nameKey, c.Budget).Scan(&created.ID, &created.Name, &created.NameKey, &created.Budget, &created.CreatedAt, &created.UpdatedAt)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return Category{}, ErrCategoryNameTaken
	}

	if err != nil {
		return Category{}, fmt.Errorf("create category: %w", err)
	}

	return created, nil
}

func (r *PostgresCategoryRepository) Update(ctx context.Context, c Category) (Category, error) {
	nameKey := strings.ToLower(c.Name)
	var updated Category

	query := `
		UPDATE categories SET name = $1, name_key = $2, budget = $3, updated_at = now()
		WHERE id = $4
		RETURNING id, name, name_key, budget, created_at, updated_at
	`
	err := r.pool.QueryRow(ctx, query, c.Name, nameKey, c.Budget, c.ID).Scan(&updated.ID, &updated.Name, &updated.NameKey, &updated.Budget, &updated.CreatedAt, &updated.UpdatedAt)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" { // 23505 = unique violation
		return Category{}, ErrCategoryNameTaken
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return Category{}, ErrCategoryNotFound
	}
	if err != nil {
		return Category{}, fmt.Errorf("update categories: %w", err)
	}

	return updated, nil
}

func (r *PostgresCategoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM categories WHERE id = $1`
	tag, err := r.pool.Exec(ctx, query, id)
	
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23503" { // 23503 has foreign key violation
		return ErrCategoryHasTransactions
	}
	if err != nil {
		return fmt.Errorf("delete categories: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrCategoryNotFound
	}

	return nil
}