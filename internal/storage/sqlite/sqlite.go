package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"todo-api/internal/models/todo"
	"todo-api/internal/storage" // ErrTaskNotFound
)

type Storage struct {
	db *sql.DB
}

const tsLayout = time.RFC3339

func New(path string) (*Storage, error) {
	const op = "sqlite.New"

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("%s: open: %w", op, err)
	}

	ddl := `
CREATE TABLE IF NOT EXISTS tasks (
	id               INTEGER PRIMARY KEY,
	name             TEXT    NOT NULL,
	description      TEXT,
	time_of_create   TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now')),
	time_of_complete TEXT,
	complete         INTEGER NOT NULL DEFAULT 0 CHECK (complete IN (0,1))
);`
	if _, err := db.Exec(ddl); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("%s: ddl: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) Close() error {
	const op = "sqlite.Close"
	if s == nil || s.db == nil {
		return nil
	}
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

// Создать задачу; вернёт id вставленной записи.
func (s *Storage) SaveTask(ctx context.Context, name, description string) (int64, error) {
	const op = "sqlite.SaveTask"

	res, err := s.db.ExecContext(ctx,
		`INSERT INTO tasks(name, description) VALUES(?, ?)`,
		name, description,
	)
	if err != nil {
		return 0, fmt.Errorf("%s: insert: %w", op, err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: lastInsertId: %w", op, err)
	}
	return id, nil
}

// Получить по id. Если нет — storage.ErrTaskNotFound.
func (s *Storage) GetTaskByID(ctx context.Context, id int64) (*todo.Task, error) {
	const op = "sqlite.GetTaskByID"

	row := s.db.QueryRowContext(ctx, `
		SELECT name, description, time_of_create, time_of_complete, complete
		FROM tasks WHERE id = ?`, id)

	var (
		name, createdStr string
		desc             sql.NullString
		doneAtStr        sql.NullString
		completeInt      int
	)
	if err := row.Scan(&name, &desc, &createdStr, &doneAtStr, &completeInt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, storage.ErrTaskNotFound
		}
		return nil, fmt.Errorf("%s: scan: %w", op, err)
	}

	createdAt, err := time.Parse(tsLayout, createdStr)
	if err != nil {
		return nil, fmt.Errorf("%s: parse created: %w", op, err)
	}

	var doneAt *time.Time
	if doneAtStr.Valid && doneAtStr.String != "" {
		t, err := time.Parse(tsLayout, doneAtStr.String)
		if err != nil {
			return nil, fmt.Errorf("%s: parse completed: %w", op, err)
		}
		doneAt = &t
	}

	return &todo.Task{
		Description:    desc.String,
		Name:           name,
		TimeOfCreate:   createdAt,
		TimeOfComplete: doneAt,
		Complete:       completeInt != 0,
	}, nil
}

// Пометить выполненной (идемпотентно).
func (s *Storage) CompleteByID(ctx context.Context, id int64) error {
	const op = "sqlite.CompleteByID"

	res, err := s.db.ExecContext(ctx, `
		UPDATE tasks
		SET complete = 1,
		    time_of_complete = strftime('%Y-%m-%dT%H:%M:%SZ','now')
		WHERE id = ? AND complete = 0`, id)
	if err != nil {
		return fmt.Errorf("%s: update: %w", op, err)
	}

	n, _ := res.RowsAffected()
	if n > 0 {
		return nil
	}

	var exists int
	if err := s.db.QueryRowContext(ctx, `SELECT 1 FROM tasks WHERE id = ?`, id).Scan(&exists); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return storage.ErrTaskNotFound
		}
		return fmt.Errorf("%s: exists-check: %w", op, err)
	}
	// уже complete = 1
	return nil
}

// Получить все задачи.
func (s *Storage) GetAllTasks(ctx context.Context) ([]todo.Task, error) {
	const op = "sqlite.GetAllTasks"

	rows, err := s.db.QueryContext(ctx, `
		SELECT description, name, time_of_create, time_of_complete, complete
		FROM tasks`)
	if err != nil {
		return nil, fmt.Errorf("%s: query: %w", op, err)
	}
	defer rows.Close()

	tasks, err := forRowsUtilite(rows, op)
	if err != nil {
		return nil, fmt.Errorf("%s: query: %w", op, err)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: rows: %w", op, err)
	}
	return tasks, err
}

func (s *Storage) GetCompleteTasks(ctx context.Context) ([]todo.Task, error) {
	const op = "sqlite.get_complete_tasks"
	rows, err := s.db.QueryContext(ctx, `
		SELECT description, name, time_of_create, time_of_complete, complete
		FROM tasks 
		WHERE complete = 1`)
	if err != nil {
		return nil, fmt.Errorf("%s: query: %w", op, err)
	}
	defer rows.Close()

	tasks, err := forRowsUtilite(rows, op)
	if err != nil {
		return nil, fmt.Errorf("%s: query: %w", op, err)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: rows: %w", op, err)
	}
	return tasks, err
}

func (s *Storage) GetNotCompleteTasks(ctx context.Context) ([]todo.Task, error) {
	const op = "sqlite.get_not_complete_tasks"
	rows, err := s.db.QueryContext(ctx, `
		SELECT description, name, time_of_create, time_of_complete, complete
		FROM tasks 
		WHERE complete = 0`)
	if err != nil {
		return nil, fmt.Errorf("%s: query: %w", op, err)
	}
	defer rows.Close()
	tasks, err := forRowsUtilite(rows, op)
	if err != nil {
		return nil, fmt.Errorf("%s: query: %w", op, err)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: rows: %w", op, err)
	}
	return tasks, err

}

func forRowsUtilite(rows *sql.Rows, op string) ([]todo.Task, error) {
	var out []todo.Task
	for rows.Next() {
		var (
			descStr    sql.NullString
			name       string
			createdStr string
			doneStr    sql.NullString
			compInt    int
		)
		if err := rows.Scan(&descStr, &name, &createdStr, &doneStr, &compInt); err != nil {
			return nil, fmt.Errorf("%s: scan: %w", op, err)
		}

		createdAt, err := time.Parse(tsLayout, createdStr)
		if err != nil {
			return nil, fmt.Errorf("%s: parse created: %w", op, err)
		}

		var doneAt *time.Time
		if doneStr.Valid && doneStr.String != "" {
			t, err := time.Parse(tsLayout, doneStr.String)
			if err != nil {
				return nil, fmt.Errorf("%s: parse completed: %w", op, err)
			}
			doneAt = &t
		}

		out = append(out, todo.Task{
			Description:    descStr.String,
			Name:           name,
			TimeOfCreate:   createdAt,
			TimeOfComplete: doneAt,
			Complete:       compInt != 0,
		})
	}

	return out, nil
}

func (s *Storage) DeleteTask(ctx context.Context, id int64) error {
	const op = "sqlite.delete_task"

	res, err := s.db.ExecContext(ctx, `DELETE FROM tasks WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("%s: parse completed: %w", op, err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return storage.ErrTaskNotFound
	}
	return nil
}
