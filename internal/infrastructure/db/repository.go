package db

import (
	"context"
	domainAuth "github.com/Vy4cheSlave/grpc_user-manager/internal/domain/auth"
	domainTaskCrud "github.com/Vy4cheSlave/grpc_user-manager/internal/domain/task_crud"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
)

var (
	ErrUserExists   = errors.New("user already exists")
	ErrUserNotFound = errors.New("user not found")
	ErrTaskNotFound = errors.New("task not found")
)

// SQL-запрос на вставку задачи
const (
	getUserByUsernameQuery        = `select id, email, username, password_hash, first_name, last_name, is_active, role from users where username = $1;`
	setLastLoginAtByUsernameQuery = `update users set last_login_at = now() where username = $1;`
	checkUserByUsenameQuery       = `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1);`
	insertUserQuery               = `insert into users (email, username, password_hash) values ($1, $2, $3) returning id;`
	checkUserByIdQuery            = `select id from users where id = $1;`
	insertTaskQuery               = `INSERT INTO tasks (title, description, user_id) VALUES ($1, $2, $3) RETURNING id;`
	getTasksQuery                 = `select id, user_id, title, description, status, created_at, updated_at from tasks;`
	getTaskByIdQuery              = `SELECT id, user_id, title, description, status, created_at, updated_at FROM tasks WHERE id = $1;`
	updateStatusTaskByIdQuery     = `update tasks set status = $1, updated_at = now() where id = $2 returning id;`
	deleteTaskByIdQuery           = `delete from tasks where id = $1;`
	getTasksByUserIdQuery         = `select id, user_id, title, description, status, created_at, updated_at from tasks where user_id = $1;`
)

type Repository struct {
	pool *pgxpool.Pool
}

func (r *Repository) SaveUser(ctx context.Context, username string, passwordHash []byte) (userId *string, err error) {
	const op = "internal/infrastructure/db/repository.Repository.SaveUser"

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to begin transaction")
	}
	var exist bool
	err = tx.QueryRow(ctx, checkUserByUsenameQuery, username).Scan(&exist)
	if err != nil {
		tx.Rollback(ctx)
		return nil, errors.Wrap(err, op)
	}
	if exist {
		tx.Rollback(ctx)
		return nil, errors.Wrap(ErrUserExists, op)
	}
	err = tx.QueryRow(ctx, insertUserQuery, uuid.New().String(), username, passwordHash).Scan(&userId)
	if err != nil {
		tx.Rollback(ctx)
		return nil, errors.Wrap(err, op)
	}
	err = tx.Commit(ctx)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	return userId, nil
}

func (r *Repository) GetUser(ctx context.Context, username string) (*domainAuth.User, error) {
	const op = "internal/infrastructure/db/repository.Repository.GetUser"

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to begin transaction")
	}
	var user domainAuth.User
	err = tx.QueryRow(ctx, getUserByUsernameQuery, username).Scan(&user.Id, &user.Email, &user.Username, &user.PasswordHash, &user.FirstName, &user.LastName, &user.IsActive, &user.Role)
	if err != nil {
		tx.Rollback(ctx)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.Wrap(ErrUserNotFound, op)
		}
		return nil, errors.Wrap(err, op)
	}
	_, err = tx.Exec(ctx, setLastLoginAtByUsernameQuery, username)
	if err != nil {
		tx.Rollback(ctx)
		return nil, errors.Wrap(err, op)
	}
	err = tx.Commit(ctx)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	return &user, nil
}

func (r *Repository) CreateTask(ctx context.Context, userId *string, title *string, description *string) (taskId int, err error) {
	const op = "internal/infrastructure/db/repository.Repository.CreateTask"
	var id int
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, errors.Wrap(err, op)
	}
	var foundID string
	err = tx.QueryRow(ctx, checkUserByIdQuery, userId).Scan(&foundID)
	if err != nil {
		tx.Rollback(ctx)
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, errors.Wrap(ErrUserNotFound, op)
		}
		return 0, errors.Wrap(err, op)
	}
	err = tx.QueryRow(ctx, insertTaskQuery, title, description, userId).Scan(&id)
	if err != nil {
		tx.Rollback(ctx)
		return 0, errors.Wrap(err, op)
	}
	err = tx.Commit(ctx)
	if err != nil {
		return 0, errors.Wrap(err, op)
	}
	return id, nil
}

func (r *Repository) ReadTasks(ctx context.Context) (*[]domainTaskCrud.Task, error) {
	const op = "internal/infrastructure/db/repository.Repository.ReadTasks"
	rows, err := r.pool.Query(ctx, getTasksQuery)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}
	defer rows.Close()
	tasks, err := pgx.CollectRows(rows, pgx.RowToStructByName[domainTaskCrud.Task])
	if err != nil {
		return nil, errors.Wrap(err, op)
	}
	return &tasks, nil
}

func (r *Repository) ReadTaskById(ctx context.Context, taskId int) (*domainTaskCrud.Task, error) {
	const op = "internal/infrastructure/db/repository.Repository.ReadTaskById"
	var task domainTaskCrud.Task
	err := r.pool.QueryRow(ctx, getTaskByIdQuery, taskId).Scan(&task.Id, &task.UserId, &task.Title, &task.Description, &task.Status, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.Wrap(ErrTaskNotFound, op)
		}
		return nil, errors.Wrap(err, op)
	}
	return &task, nil
}

func (r *Repository) ReadTasksByUserId(ctx context.Context, userId *string) (*[]domainTaskCrud.Task, error) {
	const op = "internal/infrastructure/db/repository.Repository.ReadTaskById"
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}
	var foundID string
	err = tx.QueryRow(ctx, checkUserByIdQuery, userId).Scan(&foundID)
	if err != nil {
		tx.Rollback(ctx)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.Wrap(ErrUserNotFound, op)
		}
		return nil, errors.Wrap(err, op)
	}
	rows, err := tx.Query(ctx, getTasksByUserIdQuery, userId)
	if err != nil {
		tx.Rollback(ctx)
		return nil, errors.Wrap(err, op)
	}
	defer rows.Close()
	tasks, err := pgx.CollectRows(rows, pgx.RowToStructByName[domainTaskCrud.Task])
	if err != nil {
		return nil, errors.Wrap(err, op)
	}
	tx.Commit(ctx)
	return &tasks, nil
}

func (r *Repository) UpdateStatusTask(ctx context.Context, taskId int, status *string) (err error) {
	const op = "internal/infrastructure/db/repository.Repository.UpdateStatusTask"
	var updatedID int
	err = r.pool.QueryRow(ctx, updateStatusTaskByIdQuery, status, taskId).Scan(&updatedID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.Wrap(ErrTaskNotFound, op)
		}
		return errors.Wrap(err, op)
	}
	return nil
}

func (r *Repository) DeleteTask(ctx context.Context, taskId int) (err error) {
	const op = "internal/infrastructure/db/repository.Repository.DeleteTask"
	_, err = r.pool.Exec(ctx, deleteTaskByIdQuery, taskId)
	if err != nil {
		return errors.Wrap(err, op)
	}
	return nil
}
