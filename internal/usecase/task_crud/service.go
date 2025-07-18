package task_crud

import (
	"context"
	domainTaskCrud "github.com/Vy4cheSlave/grpc_user-manager/internal/domain/task_crud"
	"github.com/pkg/errors"
)

// todo: логи? надо? хз

type TaskCrud struct {
	taskManager TaskManager
	taskReader  TaskReader
}

type TaskManager interface {
	CreateTask(ctx context.Context, userId *string, title *string, description *string) (taskId int, err error)
	UpdateStatusTask(ctx context.Context, taskId int, status *string) (err error)
	DeleteTask(ctx context.Context, taskId int) (err error)
}

type TaskReader interface {
	ReadTasks(ctx context.Context) (*[]domainTaskCrud.Task, error)
	ReadTaskById(ctx context.Context, taskId int) (*domainTaskCrud.Task, error)
	ReadTasksByUserId(ctx context.Context, userId *string) (*[]domainTaskCrud.Task, error)
}

func NewTaskManagerService(taskManager TaskManager, taskReader TaskReader) *TaskCrud {
	return &TaskCrud{
		taskManager: taskManager,
		taskReader:  taskReader,
	}
}

func (t *TaskCrud) CreateTask(ctx context.Context, userId *string, title *string, description *string) (taskId int, err error) {
	const op = "internal/usecase/task_crud/service.TaskCrud.CreateTask"
	taskId, err = t.taskManager.CreateTask(ctx, userId, title, description)
	if err != nil {
		return 0, errors.Wrap(err, op)
	}
	return taskId, nil
}

func (t *TaskCrud) ReadTasks(ctx context.Context) (*[]domainTaskCrud.Task, error) {
	const op = "internal/usecase/task_crud/service.TaskCrud.ReadTasks"
	tasks, err := t.taskReader.ReadTasks(ctx)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}
	return tasks, nil
}

func (t *TaskCrud) ReadTaskById(ctx context.Context, taskId int) (*domainTaskCrud.Task, error) {
	const op = "internal/usecase/task_crud/service.TaskCrud.ReadTaskById"
	task, err := t.taskReader.ReadTaskById(ctx, taskId)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}
	return task, nil
}

func (t *TaskCrud) ReadTasksByUserId(ctx context.Context, userId *string) (*[]domainTaskCrud.Task, error) {
	const op = "internal/usecase/task_crud/service.TaskCrud.ReadTaskById"
	tasks, err := t.taskReader.ReadTasksByUserId(ctx, userId)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}
	return tasks, nil
}

func (t *TaskCrud) UpdateStatusTask(ctx context.Context, taskId int, status *string) (err error) {
	const op = "internal/usecase/task_crud/service.TaskCrud.UpdateStatusTask"
	err = t.taskManager.UpdateStatusTask(ctx, taskId, status)
	if err != nil {
		return errors.Wrap(err, op)
	}
	return nil
}

func (t *TaskCrud) DeleteTask(ctx context.Context, taskId int) (err error) {
	const op = "internal/usecase/task_crud/service.TaskCrud.DeleteTask"
	err = t.taskManager.DeleteTask(ctx, taskId)
	if err != nil {
		return errors.Wrap(err, op)
	}
	return nil
}
