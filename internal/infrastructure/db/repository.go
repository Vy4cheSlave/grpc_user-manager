package db

import (
	"context"
	domainAuth "github.com/Vy4cheSlave/grpc_user-manager/internal/domain/auth"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
)

var (
	ErrUserExists   = errors.New("user already exists")
	ErrUserNotFound = errors.New("user not found")
)

// SQL-запрос на вставку задачи
const (
	getUserByUsernameQuery        = `select id, email, username, password_hash, first_name, last_name, is_active, role from users where username = $1;`
	setLastLoginAtByUsernameQuery = `update users set last_login_at = now() where username = $1;`
	checkUserByUsenameQuery       = `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1);`
	insertUserQuery               = `insert into users (email, username, password_hash) values ($1, $2, $3) returning id;`

	//insertTaskQuery           = `INSERT INTO tasks (title, description, user_id) VALUES ($1, $2, $3) RETURNING id;`
	//getTasksQuery             = `select id, user_id, title, description, status, created_at, updated_at from tasks;`
	//getTaskByIdQuery          = `SELECT id, user_id, title, description, status, created_at, updated_at FROM tasks WHERE id = $1;`
	//updateStatusTaskByIdQuery = `update tasks set status = $1, updated_at = now() where id = $2 returning id;`
	//deleteTaskByIdQuery       = `delete from tasks where id = $1;`
	//getTasksByUserIdQuery     = `select id, user_id, title, description, status, created_at, updated_at from tasks where user_id = $1;`
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

//func (r *Repository) CreateTask(ctx context.Context, task *Task, userId int) (int, error) {
//	var id int
//	tx, err := r.pool.Begin(ctx)
//	if err != nil {
//		return 0, errors.Wrap(err, "failed to begin transaction")
//	}
//	var foundID int
//	err = tx.QueryRow(ctx, checkUserByIdQuery, userId).Scan(&foundID)
//	if err != nil {
//		tx.Rollback(ctx)
//		return 0, customErrors.New(
//			fiber.StatusNotFound,
//			errors.Wrap(err, "failed to check user by id"))
//	}
//	err = tx.QueryRow(ctx, insertTaskQuery, task.Title, task.Description, userId).Scan(&id)
//	if err != nil {
//		tx.Rollback(ctx)
//		return 0, errors.Wrap(err, "failed to insert task")
//	}
//	err = tx.Commit(ctx)
//	if err != nil {
//		return 0, errors.Wrap(err, "failed to commit transaction")
//	}
//	return id, nil
//}
//
//func (r *Repository) GetTasks(ctx context.Context) (*[]Task, error) {
//	rows, err := r.pool.Query(ctx, getTasksQuery)
//	if err != nil {
//		return nil, errors.Wrap(err, "failed to query tasks")
//	}
//	defer rows.Close()
//	tasks, err := pgx.CollectRows(rows, pgx.RowToStructByName[Task])
//	if err != nil {
//		return nil, errors.Wrap(err, "failed to get tasks")
//	}
//	return &tasks, nil
//}
//
//func (r *Repository) GetTaskById(ctx context.Context, id int) (*Task, error) {
//	task := new(Task)
//	err := r.pool.QueryRow(ctx, getTaskByIdQuery, id).Scan(&task.Id, &task.UserId, &task.Title, &task.Description, &task.Status, &task.CreatedAt, &task.UpdatedAt)
//	if err != nil {
//		if errors.Is(err, pgx.ErrNoRows) {
//			return nil, customErrors.New(fiber.StatusBadRequest, errors.Wrap(err, "failed to find task row with current id"))
//		}
//		return nil, errors.Wrap(err, "failed to get task")
//	}
//	return task, nil
//}
//
//func (r *Repository) UpdateStatusTaskById(ctx context.Context, id int, status string) error {
//	var updatedID int
//	err := r.pool.QueryRow(ctx, updateStatusTaskByIdQuery, status, id).Scan(&updatedID)
//	if err != nil {
//		if errors.Is(err, pgx.ErrNoRows) {
//			return customErrors.New(fiber.StatusBadRequest, errors.Wrap(err, "failed to find task row with current id"))
//		}
//		return customErrors.New(fiber.StatusBadRequest, errors.Wrap(err, "failed to find task row with current id or status not appended"))
//	}
//	return nil
//}
//
//func (r *Repository) DeleteTaskById(ctx context.Context, id int) error {
//	result, err := r.pool.Exec(ctx, deleteTaskByIdQuery, id)
//	if err != nil {
//		return errors.Wrap(err, "failed to delete task")
//	}
//	rowsAffected := result.RowsAffected()
//	if rowsAffected == 0 {
//		return customErrors.New(fiber.StatusBadRequest, errors.Wrap(err, "failed to find task row with current id"))
//	}
//	return nil
//}
//
//func (r *Repository) GetTasksByUserId(ctx context.Context, userId int) (*[]Task, error) {
//	tx, err := r.pool.Begin(ctx)
//	if err != nil {
//		return nil, errors.Wrap(err, "failed to begin transaction")
//	}
//	var foundID int
//	err = tx.QueryRow(ctx, checkUserByIdQuery, userId).Scan(&foundID)
//	if err != nil {
//		tx.Rollback(ctx)
//		return nil, customErrors.New(
//			fiber.StatusNotFound,
//			errors.Wrap(err, "failed to check user by id"))
//	}
//	rows, err := tx.Query(ctx, getTasksByUserIdQuery, userId)
//	if err != nil {
//		tx.Rollback(ctx)
//		//if errors.Is(err, pgx.ErrNoRows) {
//		//	return nil, customErrors.New(fiber.StatusBadRequest, errors.Wrap(err, "failed to find task rows with current user_id, or tasks not found"))
//		//}
//		return nil, errors.Wrap(err, "failed to get tasks by user_id")
//	}
//	defer rows.Close()
//	tasks, err := pgx.CollectRows(rows, pgx.RowToStructByName[Task])
//	if err != nil {
//		return nil, errors.Wrap(err, "failed to get tasks")
//	}
//	tx.Commit(ctx)
//	return &tasks, nil
//}
