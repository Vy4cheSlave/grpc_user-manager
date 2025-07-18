package rest

import (
	"context"
	"encoding/json"
	domainTaskCrud "github.com/Vy4cheSlave/grpc_user-manager/internal/domain/task_crud"
	"github.com/Vy4cheSlave/grpc_user-manager/internal/infrastructure/db"
	"github.com/Vy4cheSlave/grpc_user-manager/internal/infrastructure/rest/dto/request"
	"github.com/Vy4cheSlave/grpc_user-manager/internal/infrastructure/rest/dto/response"
	"github.com/Vy4cheSlave/grpc_user-manager/internal/infrastructure/rest/middleware"
	"github.com/Vy4cheSlave/grpc_user-manager/internal/infrastructure/rest/validation"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/pkg/errors"
	"log/slog"
	"strings"
)

type TaskCrud interface {
	CreateTask(ctx context.Context, userId *string, title *string, description *string) (taskId int, err error)
	ReadTasks(ctx context.Context) (*[]domainTaskCrud.Task, error)
	ReadTaskById(ctx context.Context, taskId int) (*domainTaskCrud.Task, error)
	ReadTasksByUserId(ctx context.Context, userId *string) (*[]domainTaskCrud.Task, error)
	UpdateStatusTask(ctx context.Context, taskId int, status *string) (err error)
	DeleteTask(ctx context.Context, taskId int) (err error)
}

type Server struct {
	log        *slog.Logger
	service    TaskCrud
	restServer *fiber.App
	addr       string
}

func NewServer(log *slog.Logger, service TaskCrud, addr *string, tokenSecret *[]byte) *Server {
	restServer := NewRestServer(&serverAPI{log: log, taskCrud: service, validtor: validator.New()}, tokenSecret)
	return &Server{
		log:        log,
		restServer: restServer,
		service:    service,
		addr:       *addr,
	}
}

func NewRestServer(api *serverAPI, tokenSecret *[]byte) *fiber.App {
	app := fiber.New()

	// Настройка CORS (разрешенные методы, заголовки, авторизация)
	app.Use(cors.New(cors.Config{
		AllowMethods:  "GET, POST, PUT, DELETE",
		AllowHeaders:  "Accept, Authorization, Content-Type, X-CSRF-Token, X-REQUEST-ID",
		ExposeHeaders: "Link",
		MaxAge:        300,
	}))

	app.Use(middleware.LoggingMiddleware(api.log))

	{
		apiGroup := app.Group("/task_crud")

		apiGroup.Get("/tasks", api.GetTasks)
		apiGroup.Get("/tasks/user/:id", api.GetTasksByUserId)
		apiGroup.Get("/tasks/:id", api.GetTaskById)

		// Middleware auth
		{
			authGroup := apiGroup.Group("/auth")
			authGroup.Use(middleware.JWTAuthMiddleware(tokenSecret))

			authGroup.Post("/tasks", api.CreateTask)
			authGroup.Put("/tasks/:id", api.UpdateStatusTaskById)
			authGroup.Delete("/tasks/:id", api.DeleteTaskById)
		}
	}

	return app
}

func (s *Server) Run() error {
	const op = "internal/infrastructure/rest/handler.Server.Run"
	log := s.log.With(slog.String("operation", op), slog.String("addr", s.addr))
	log.Info("grpc server is running")

	if err := s.restServer.Listen(s.addr); err != nil {
		return errors.Wrap(err, strings.Join([]string{op, "failed to serve rest server"}, ": "))
	}

	return nil
}

type serverAPI struct {
	log      *slog.Logger
	taskCrud TaskCrud
	validtor *validator.Validate
}

func (s *serverAPI) CreateTask(ctx *fiber.Ctx) error {
	var req request.TaskCreateRequest

	// Десериализация JSON-запроса
	if err := json.Unmarshal(ctx.Body(), &req); err != nil {
		return response.ReturnResponse(
			ctx,
			fiber.StatusBadRequest,
			response.WithError(response.ErrCodeJsonParsingFailed, "Invalid request body"),
			response.WithMeta(err),
		)
	}

	// валидация данных
	if err := s.validtor.StructCtx(ctx.Context(), &req); err != nil {
		var verrs validator.ValidationErrors
		if errors.As(err, &verrs) {
			descriptionError := validation.GetValidationErrorMessage(&verrs)
			return response.ReturnResponse(
				ctx,
				fiber.StatusBadRequest,
				response.WithError(response.ErrCodeValidationFailed, descriptionError),
				response.WithMeta(verrs),
			)
		}
		return response.ReturnResponse(
			ctx,
			fiber.StatusBadRequest,
			response.WithError(response.ErrCodeValidationFailed, ""),
			response.WithMeta(err),
		)
	}

	userId, ok := ctx.Locals("userId").(string)
	if !ok {
		// случай которого, в теории, никогда быть не должно
		return response.ReturnResponse(
			ctx,
			fiber.StatusInternalServerError,
			response.WithError(response.ErrCodeInternalServerError, "Server misconfiguration: userID not set"),
		)
	}

	// Вставка задачи в сервис
	taskId, err := s.taskCrud.CreateTask(ctx.Context(), &userId, &req.Title, req.Description)
	if err != nil {
		if errors.Is(err, db.ErrTaskNotFound) {
			return response.ReturnResponse(
				ctx,
				fiber.StatusUnauthorized,
				response.WithError(response.ErrCodeUnauthorized, "User not found"),
				response.WithMeta(err),
			)
		}
		return response.ReturnResponse(
			ctx,
			fiber.StatusInternalServerError,
			response.WithError(response.ErrCodeInternalServerError, ""),
			response.WithMeta(err),
		)
	}

	// Формирование ответа
	return response.ReturnResponse(
		ctx,
		fiber.StatusOK,
		response.WithData(&response.TaskCreateResponse{TaskId: taskId}),
	)
}

func (s *serverAPI) GetTasks(ctx *fiber.Ctx) error {
	tasks, err := s.taskCrud.ReadTasks(ctx.Context())
	if err != nil {
		return response.ReturnResponse(
			ctx,
			fiber.StatusInternalServerError,
			response.WithError(response.ErrCodeInternalServerError, ""),
			response.WithMeta(err),
		)
	}

	return response.ReturnResponse(
		ctx,
		fiber.StatusOK,
		response.WithData(tasks),
	)
}

func (s *serverAPI) GetTaskById(ctx *fiber.Ctx) error {
	idTask, err := ctx.ParamsInt("id", -1)
	if err != nil || idTask < 0 {
		return response.ReturnResponse(
			ctx,
			fiber.StatusBadRequest,
			response.WithError(response.ErrCodeValidationFailed, "id must be a positive integer"),
			response.WithMeta(err),
		)
	}

	task, err := s.taskCrud.ReadTaskById(ctx.Context(), idTask)
	if err != nil {
		if errors.Is(err, db.ErrTaskNotFound) {
			return response.ReturnResponse(
				ctx,
				fiber.StatusNotFound,
				response.WithError(response.ErrCodeNotFound, "Task not found"),
				response.WithMeta(err),
			)
		}
		return response.ReturnResponse(
			ctx,
			fiber.StatusInternalServerError,
			response.WithError(response.ErrCodeInternalServerError, ""),
			response.WithMeta(err),
		)
	}

	return response.ReturnResponse(
		ctx,
		fiber.StatusOK,
		response.WithData(task),
	)
}

func (s *serverAPI) UpdateStatusTaskById(ctx *fiber.Ctx) error {
	var req request.TaskUpdateRequest

	if err := json.Unmarshal(ctx.Body(), &req); err != nil {
		return response.ReturnResponse(
			ctx,
			fiber.StatusBadRequest,
			response.WithError(response.ErrCodeJsonParsingFailed, "Invalid request body"),
			response.WithMeta(err),
		)
	}

	// валидация данных
	if err := s.validtor.StructCtx(ctx.Context(), &req); err != nil {
		var verrs validator.ValidationErrors
		if errors.As(err, &verrs) {
			descriptionError := validation.GetValidationErrorMessage(&verrs)
			return response.ReturnResponse(
				ctx,
				fiber.StatusBadRequest,
				response.WithError(response.ErrCodeValidationFailed, descriptionError),
				response.WithMeta(verrs),
			)
		}
		return response.ReturnResponse(
			ctx,
			fiber.StatusBadRequest,
			response.WithError(response.ErrCodeValidationFailed, ""),
			response.WithMeta(err),
		)
	}
	idTask, err := ctx.ParamsInt("id", -1)
	if err != nil || idTask < 0 {
		return response.ReturnResponse(
			ctx,
			fiber.StatusBadRequest,
			response.WithError(response.ErrCodeValidationFailed, "id must be a positive integer"),
			response.WithMeta(err),
		)
	}

	err = s.taskCrud.UpdateStatusTask(ctx.Context(), idTask, &req.Status)
	if err != nil {
		if errors.Is(err, db.ErrTaskNotFound) {
			return response.ReturnResponse(
				ctx,
				fiber.StatusNotFound,
				response.WithError(response.ErrCodeNotFound, "Task not found"),
				response.WithMeta(err),
			)
		}
		return response.ReturnResponse(
			ctx,
			fiber.StatusInternalServerError,
			response.WithError(response.ErrCodeInternalServerError, ""),
			response.WithMeta(err),
		)
	}

	return response.ReturnResponse(
		ctx,
		fiber.StatusOK,
	)
}

func (s *serverAPI) DeleteTaskById(ctx *fiber.Ctx) error {
	idTask, err := ctx.ParamsInt("id", -1)
	if err != nil || idTask < 0 {
		return response.ReturnResponse(
			ctx,
			fiber.StatusBadRequest,
			response.WithError(response.ErrCodeValidationFailed, "id must be a positive integer"),
			response.WithMeta(err),
		)
	}

	err = s.taskCrud.DeleteTask(ctx.Context(), idTask)
	if err != nil {
		return response.ReturnResponse(
			ctx,
			fiber.StatusInternalServerError,
			response.WithError(response.ErrCodeInternalServerError, ""),
			response.WithMeta(err),
		)
	}

	return response.ReturnResponse(
		ctx,
		fiber.StatusOK,
	)
}

func (s *serverAPI) GetTasksByUserId(ctx *fiber.Ctx) error {
	idUser := ctx.Params("id")
	if err := s.validtor.Var(&idUser, "uuid"); err != nil {
		var verrs validator.ValidationErrors
		if errors.As(err, &verrs) {
			descriptionError := validation.GetValidationErrorMessage(&verrs)
			return response.ReturnResponse(
				ctx,
				fiber.StatusBadRequest,
				response.WithError(response.ErrCodeValidationFailed, descriptionError),
				response.WithMeta(verrs),
			)
		}
		return response.ReturnResponse(
			ctx,
			fiber.StatusBadRequest,
			response.WithError(response.ErrCodeValidationFailed, ""),
			response.WithMeta(err),
		)
	}

	tasks, err := s.taskCrud.ReadTasksByUserId(ctx.Context(), &idUser)
	if err != nil {
		if errors.Is(err, db.ErrUserNotFound) {
			return response.ReturnResponse(
				ctx,
				fiber.StatusNotFound,
				response.WithError(response.ErrCodeNotFound, "User not found"),
				response.WithMeta(err),
			)
		}
		return response.ReturnResponse(
			ctx,
			fiber.StatusInternalServerError,
			response.WithError(response.ErrCodeInternalServerError, ""),
			response.WithMeta(err),
		)
	}

	return response.ReturnResponse(
		ctx,
		fiber.StatusOK,
		response.WithData(tasks),
	)
}
