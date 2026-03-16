package repository

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"

	"pz1.2/services/tasks/internal/service"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(dsn string) (*PostgresRepository, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}
	if err := migrate(db); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return &PostgresRepository{db: db}, nil
}

func migrate(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS tasks (
		id          TEXT PRIMARY KEY,
		title       TEXT NOT NULL,
		description TEXT DEFAULT '',
		due_date    TEXT DEFAULT '',
		done        BOOLEAN DEFAULT FALSE,
		created_at  TEXT NOT NULL
	);`
	_, err := db.Exec(query)
	return err
}

func (r *PostgresRepository) Close() error {
	return r.db.Close()
}

func (r *PostgresRepository) Create(task *service.Task) error {
	_, err := r.db.Exec(
		"INSERT INTO tasks (id, title, description, due_date, done, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
		task.ID, task.Title, task.Description, task.DueDate, task.Done, task.CreatedAt,
	)
	return err
}

func (r *PostgresRepository) GetAll() ([]*service.Task, error) {
	rows, err := r.db.Query("SELECT id, title, description, due_date, done, created_at FROM tasks ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTasks(rows)
}

func (r *PostgresRepository) GetByID(id string) (*service.Task, error) {
	row := r.db.QueryRow("SELECT id, title, description, due_date, done, created_at FROM tasks WHERE id = $1", id)
	t := &service.Task{}
	err := row.Scan(&t.ID, &t.Title, &t.Description, &t.DueDate, &t.Done, &t.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, service.ErrTaskNotFound
	}
	return t, err
}

func (r *PostgresRepository) Update(task *service.Task) error {
	res, err := r.db.Exec(
		"UPDATE tasks SET title=$1, description=$2, due_date=$3, done=$4 WHERE id=$5",
		task.Title, task.Description, task.DueDate, task.Done, task.ID,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return service.ErrTaskNotFound
	}
	return nil
}

func (r *PostgresRepository) Delete(id string) error {
	res, err := r.db.Exec("DELETE FROM tasks WHERE id = $1", id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return service.ErrTaskNotFound
	}
	return nil
}

func (r *PostgresRepository) SearchByTitle(title string) ([]*service.Task, error) {
	rows, err := r.db.Query(
		"SELECT id, title, description, due_date, done, created_at FROM tasks WHERE title = $1",
		title,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTasks(rows)
}

func scanTasks(rows *sql.Rows) ([]*service.Task, error) {
	var tasks []*service.Task
	for rows.Next() {
		t := &service.Task{}
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.DueDate, &t.Done, &t.CreatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}
