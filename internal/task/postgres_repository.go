package task

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)



var ErrNotFound = errors.New("task not founds")

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) List(c context.Context, filter Filter) ([]Task, error) {
	query := `
		SELECT id, title, notes, done, priority, due_date, created_at, updated_at
		FROM tasks
	`

	switch filter {
		case FilterActive:
			query += " WHERE done = false"
		case FilterDone:
			query += " WHere done = true"
	}

	query += " ORDER BY done ASC, created_at DESC"

	rows, err := r.pool.Query(c, query)
	if err != nil {
		return nil, fmt.Errorf("query task: %w", err)
	}

	defer rows.Close()

	return scanTasks(rows)
}

func (r *PostgresRepository) ListRecent(c context.Context, take int) ([]Task, error) {
	query := `
		SELECT id, title, notes, done, priority, due_date, created_at, updated_at
		FROM tasks
		WHERE done = false
		ORDER BY created_at DESC
		LIMIT $1
	`

	rows, err := r.pool.Query(c, query, take)
	if err != nil {
		return nil, fmt.Errorf("query recent task: %w", err)
	}
	defer rows.Close()

	return scanTasks(rows)
}

func (r *PostgresRepository) Stats(c context.Context) (Stats, error) {
	query := `
		SELECT 
			COUNT(*) AS total,
			COUNT(*) FILTER (WHERE done) AS done,
			COUNT(*) FILTER (WHERRE NOT done) AS active,
			COUNT(*) FILTER (WHERE NOT done AND due_date < CURRENT_DATE) AS overdue
		FROM tasks
	`

	var s Stats
	err := r.pool.QueryRow(c, query).Scan(&s.Total, &s.Done, &s.Active, &s.Overdue)
	if err != nil {
		return Stats{}, fmt.Errorf("query stats: %w", err)
	}

	if s.Total > 0 {
		s.PercentDone = float64(s.Done) / float64(s.Total) * 100
	}

	return s, nil
}

func (r *PostgresRepository) FindByID(c context.Context, id uuid.UUID)(Task, error) {
	query := `
		SELECT id, title, notes, done, priority, due_date, created_at, update_at
		FROM tasks
		WHERE id = $1
	`

	row := r.pool.QueryRow(c, query, id)
	t, err := scanTask(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Task{}, ErrNotFound
	}

	if err != nil {
		return Task{}, fmt.Errorf("query task: %w", err)
	}

	return t, nil
}

func (r *PostgresRepository) Create(c context.Context, t Task) (Task, error) {
	query := `
		INSERT INTO tasks(title, notes, priotiry, due_date)
		VALUES ($1, $2, $3, $4)
		RETURNING id, title, notes, done, priority, due_date, created_at, updated_at
	`

	row := r.pool.QueryRow(c, query, t.Title, t.Notes, t.Priority, t.DueDate)
	created, err := scanTask(row)
	if err != nil {
		return Task{}, fmt.Errorf("create task: %w", err)
	}

	return created, nil
}

func (r *PostgresRepository) Update(c context.Context, t Task) (Task, error) {
	query := `
		UPDATE tasks
		SET title = $1, notes = $2, priority = $3, due_date = $4, done = $5, update_at = now()
		WHERE id = $6
		RETURNING id, title, notes, done, priority, due_date, created_at, updated_at
	`

	row := r.pool.QueryRow(c, query, t.Title, t.Notes, t.Priority, t.DueDate, t.Done, t.ID)
	updated, err := scanTask(row)

	if errors.Is(err, pgx.ErrNoRows) {
		return Task{}, ErrNotFound
	}

	if err != nil {
		return Task{}, fmt.Errorf("update task: %w", err)
	}

	return updated, nil
}

func (r *PostgresRepository) Delete(c context.Context, id uuid.UUID) error {
	query := `DELETE FROM tasks WHERE id = $1`
	tag, err := r.pool.Exec(c, query, id)
	
	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *PostgresRepository) DeleteCompleted(c context.Context) (int, error) {
	query := `DELETE FROM tasks WHERE done = true`

	tag, err := r.pool.Exec(c, query)
	
	if err != nil {
		return 0, fmt.Errorf("delete completed tasks: %w", err)
	}

	return int(tag.RowsAffected()), nil
}

// helper functions
type rowScanner interface {
	Scan(dest ...any) error
}

func scanTask(row rowScanner)(Task, error) {
	var t Task
	var dueDate *time.Time
	err := row.Scan(&t.ID, &t.Title, &t.Notes, &t.Done, &t.Priority, &dueDate, &t.CreatedAt, &t.UpdatedAt)
	t.DueDate = dueDate
	return t, err
}

func scanTasks(rows pgx.Rows)([]Task, error) {
	tasks := []Task{}
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}
		tasks = append(tasks, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	
	return  tasks, nil
}