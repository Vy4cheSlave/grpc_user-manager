package rest

import (
	"encoding/json"
	"errors"
	"github.com/Vy4cheSlave/grpc_user-manager/internal/domain/task_crud"
	"github.com/Vy4cheSlave/grpc_user-manager/internal/infrastructure/db"
	"github.com/Vy4cheSlave/grpc_user-manager/internal/infrastructure/rest/dto/response"
	"github.com/Vy4cheSlave/grpc_user-manager/internal/infrastructure/rest/mocks"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type testCase struct {
	name           string
	pathParameters string
	requestBody    string
	contentType    string
	setupMock      func(*mocks.MockTaskCrud)
	expectedStatus int
	expectedResp   interface{}
}

const (
	defaultContentType = "application/json; charset=utf-8"
)

var (
	voidSetupMock = func(ms *mocks.MockTaskCrud) {}
)

func TestCreateTask(t *testing.T) {
	defaultUserId := "userId"
	// Общие настройки
	mockTaskCrud := mocks.NewMockTaskCrud(t)
	tests := []testCase{
		{
			name:        "Success",
			requestBody: `{"title":"title","description":"description"}`,
			contentType: defaultContentType,
			setupMock: func(ms *mocks.MockTaskCrud) {
				title := "title"
				description := "description"
				ms.On("CreateTask",
					mock.AnythingOfType("*fasthttp.RequestCtx"),
					&defaultUserId,
					&title,
					&description,
				).Return(1, nil).Once()
			},
			expectedStatus: fiber.StatusOK,
			expectedResp: map[string]interface{}{
				"data": map[string]interface{}{
					"task_id": float64(1),
				},
				"status": http.StatusText(fiber.StatusOK),
			},
		},
		{
			name:           "Invalid request body",
			requestBody:    `{invalid_json}`,
			contentType:    defaultContentType,
			setupMock:      voidSetupMock,
			expectedStatus: fiber.StatusBadRequest,
			expectedResp: map[string]interface{}{
				"error": map[string]interface{}{
					"code": response.ErrCodeJsonParsingFailed,
					"desc": "Invalid request body",
				},
				"status": http.StatusText(fiber.StatusBadRequest),
			},
		},
		{
			name:           "Validation Error",
			requestBody:    `{}`,
			contentType:    defaultContentType,
			setupMock:      voidSetupMock,
			expectedStatus: fiber.StatusBadRequest,
			expectedResp: map[string]interface{}{
				"error": map[string]interface{}{
					"code": response.ErrCodeValidationFailed,
					"desc": "Field 'Title' failed validation: rule 'required' (value: )",
				},
				"status": http.StatusText(fiber.StatusBadRequest),
			},
		},
		{
			name:        "User not found",
			requestBody: `{"title":"title","description":"description"}`,
			contentType: defaultContentType,
			setupMock: func(ms *mocks.MockTaskCrud) {
				title := "title"
				description := "description"
				ms.On("CreateTask",
					mock.AnythingOfType("*fasthttp.RequestCtx"),
					&defaultUserId,
					&title,
					&description,
				).Return(0, db.ErrTaskNotFound).Once()
			},
			expectedStatus: fiber.StatusUnauthorized,
			expectedResp: map[string]interface{}{
				"error": map[string]interface{}{
					"code": response.ErrCodeUnauthorized,
					"desc": "User not found",
				},
				"status": http.StatusText(fiber.StatusUnauthorized),
			},
		},
		{
			name:        "Internal Error",
			requestBody: `{"title":"title","description":"description"}`,
			contentType: defaultContentType,
			setupMock: func(ms *mocks.MockTaskCrud) {
				title := "title"
				description := "description"
				ms.On("CreateTask",
					mock.AnythingOfType("*fasthttp.RequestCtx"),
					&defaultUserId,
					&title,
					&description,
				).Return(0, errors.New("internal error")).Once()
			},
			expectedStatus: fiber.StatusInternalServerError,
			expectedResp: map[string]interface{}{
				"error": map[string]interface{}{
					"code": response.ErrCodeInternalServerError,
				},
				"status": http.StatusText(fiber.StatusInternalServerError),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			api := &serverAPI{
				log:      slog.Default(),
				taskCrud: mockTaskCrud,
				validtor: validator.New(),
			}
			mockedCreateTask := func(ctx *fiber.Ctx) error {
				ctx.Locals("userId", defaultUserId)
				return api.CreateTask(ctx)
			}
			app.Post("/tasks", mockedCreateTask)
			defer app.Shutdown()

			// Подготовка мока
			tt.setupMock(mockTaskCrud)

			// Создание запроса
			req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", tt.contentType)

			// Выполнение запроса
			resp, err := app.Test(req) // Я ТОГО РАЗРАБА РУКИ ЕБАЛ
			require.NoError(t, err, "test request")
			defer resp.Body.Close()

			// Проверка статуса
			assert.Equal(t, tt.expectedStatus, resp.StatusCode, "status code")

			// Проверка тела ответа (если ожидается)
			if tt.expectedResp != nil {
				var body interface{}
				err := json.NewDecoder(resp.Body).Decode(&body)
				require.NoError(t, err, "response body")
				assert.Equal(t, tt.expectedResp, body, "expected response body")
			}

			// Проверка, что все ожидания мока выполнены
			mockTaskCrud.AssertExpectations(t)

			// Сброс мока после каждого теста
			mockTaskCrud.ExpectedCalls = nil
			mockTaskCrud.Calls = nil
		})
	}
}

func TestGetTasks(t *testing.T) {
	// Общие настройки
	mockTaskCrud := mocks.NewMockTaskCrud(t)
	api := &serverAPI{
		log:      slog.Default(),
		taskCrud: mockTaskCrud,
		validtor: validator.New(),
	}
	tests := []testCase{
		{
			name:        "Success",
			contentType: defaultContentType,
			setupMock: func(ms *mocks.MockTaskCrud) {
				responseData := &[]task_crud.Task{}
				ms.On("ReadTasks",
					mock.AnythingOfType("*fasthttp.RequestCtx"),
				).Return(responseData, nil).Once()
			},
			expectedStatus: fiber.StatusOK,
			expectedResp: map[string]interface{}{
				"data":   []interface{}{},
				"status": http.StatusText(fiber.StatusOK),
			},
		},
		{
			name:        "Internal Error",
			contentType: defaultContentType,
			setupMock: func(ms *mocks.MockTaskCrud) {
				ms.On("ReadTasks",
					mock.AnythingOfType("*fasthttp.RequestCtx"),
				).Return(nil, errors.New("internal error")).Once()
			},
			expectedStatus: fiber.StatusInternalServerError,
			expectedResp: map[string]interface{}{
				"error": map[string]interface{}{
					"code": response.ErrCodeInternalServerError,
				},
				"status": http.StatusText(fiber.StatusInternalServerError),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			app.Get("/tasks", api.GetTasks)
			defer app.Shutdown()

			// Подготовка мока
			tt.setupMock(mockTaskCrud)

			// Создание запроса
			req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
			req.Header.Set("Content-Type", tt.contentType)

			// Выполнение запроса
			resp, err := app.Test(req)
			require.NoError(t, err, "test request")
			defer resp.Body.Close()

			// Проверка статуса
			assert.Equal(t, tt.expectedStatus, resp.StatusCode, "status code")

			// Проверка тела ответа (если ожидается)
			if tt.expectedResp != nil {
				var body interface{}
				err := json.NewDecoder(resp.Body).Decode(&body)
				require.NoError(t, err, "response body")
				assert.Equal(t, tt.expectedResp, body, "expected response body")
			}

			// Проверка, что все ожидания мока выполнены
			mockTaskCrud.AssertExpectations(t)

			// Сброс мока после каждого теста
			mockTaskCrud.ExpectedCalls = nil
			mockTaskCrud.Calls = nil
		})
	}
}

func TestGetTaskById(t *testing.T) {
	defaultUserId := "defaultUserId"
	defaultTime := time.Now()
	defaultTimeString := defaultTime.Format(time.RFC3339Nano)
	// Общие настройки
	mockTaskCrud := mocks.NewMockTaskCrud(t)
	api := &serverAPI{
		log:      slog.Default(),
		taskCrud: mockTaskCrud,
		validtor: validator.New(),
	}
	tests := []testCase{
		{
			name:           "Success",
			pathParameters: "1",
			contentType:    defaultContentType,
			setupMock: func(ms *mocks.MockTaskCrud) {
				ms.On("ReadTaskById",
					mock.AnythingOfType("*fasthttp.RequestCtx"),
					1,
				).Return(&task_crud.Task{
					Id:          1,
					UserId:      defaultUserId,
					Title:       "title",
					Description: "description",
					Status:      "new",
					CreatedAt:   defaultTime,
					UpdatedAt:   defaultTime,
				}, nil).Once()
			},
			expectedStatus: fiber.StatusOK,
			expectedResp: map[string]interface{}{
				"data": map[string]interface{}{
					"id":          float64(1),
					"user_id":     defaultUserId,
					"title":       "title",
					"description": "description",
					"status":      "new",
					"created_at":  defaultTimeString,
					"updated_at":  defaultTimeString,
				},
				"status": http.StatusText(fiber.StatusOK),
			},
		},
		{
			name:           "Internal Error",
			pathParameters: "1",
			contentType:    defaultContentType,
			setupMock: func(ms *mocks.MockTaskCrud) {
				ms.On("ReadTaskById",
					mock.AnythingOfType("*fasthttp.RequestCtx"),
					1,
				).Return(nil, errors.New("internal error")).Once()
			},
			expectedStatus: fiber.StatusInternalServerError,
			expectedResp: map[string]interface{}{
				"error": map[string]interface{}{
					"code": response.ErrCodeInternalServerError,
				},
				"status": http.StatusText(fiber.StatusInternalServerError),
			},
		},
		{
			name:           "Validation Error",
			pathParameters: "wrong_id",
			contentType:    defaultContentType,
			setupMock:      voidSetupMock,
			expectedStatus: fiber.StatusBadRequest,
			expectedResp: map[string]interface{}{
				"error": map[string]interface{}{
					"code": response.ErrCodeValidationFailed,
					"desc": "id must be a positive integer",
				},
				"status": http.StatusText(fiber.StatusBadRequest),
			},
		},
		{
			name:           "Task Not Found",
			pathParameters: "1",
			contentType:    defaultContentType,
			setupMock: func(ms *mocks.MockTaskCrud) {
				ms.On("ReadTaskById",
					mock.AnythingOfType("*fasthttp.RequestCtx"),
					1,
				).Return(nil, db.ErrTaskNotFound).Once()
			},
			expectedStatus: fiber.StatusNotFound,
			expectedResp: map[string]interface{}{
				"error": map[string]interface{}{
					"code": response.ErrCodeNotFound,
					"desc": "Task not found",
				},
				"status": http.StatusText(fiber.StatusNotFound),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			app.Get("/tasks/:id", api.GetTaskById)
			defer app.Shutdown()

			// Подготовка мока
			tt.setupMock(mockTaskCrud)

			// Создание запроса
			req := httptest.NewRequest(http.MethodGet, "/tasks/"+tt.pathParameters, nil)
			req.Header.Set("Content-Type", tt.contentType)

			// Выполнение запроса
			resp, err := app.Test(req)
			require.NoError(t, err, "test request")
			defer resp.Body.Close()

			// Проверка статуса
			assert.Equal(t, tt.expectedStatus, resp.StatusCode, "status code")

			// Проверка тела ответа (если ожидается)
			if tt.expectedResp != nil {
				var body interface{}
				err := json.NewDecoder(resp.Body).Decode(&body)
				require.NoError(t, err, "response body")
				assert.Equal(t, tt.expectedResp, body, "expected response body")
			}

			// Проверка, что все ожидания мока выполнены
			mockTaskCrud.AssertExpectations(t)

			// Сброс мока после каждого теста
			mockTaskCrud.ExpectedCalls = nil
			mockTaskCrud.Calls = nil
		})
	}
}

func TestGetTasksByUserId(t *testing.T) {
	defaultUserId := uuid.New().String()
	// Общие настройки
	mockTaskCrud := mocks.NewMockTaskCrud(t)
	api := &serverAPI{
		log:      slog.Default(),
		taskCrud: mockTaskCrud,
		validtor: validator.New(),
	}
	tests := []testCase{
		{
			name:           "Success",
			pathParameters: defaultUserId,
			contentType:    defaultContentType,
			setupMock: func(ms *mocks.MockTaskCrud) {
				ms.On("ReadTasksByUserId",
					mock.AnythingOfType("*fasthttp.RequestCtx"),
					&defaultUserId,
				).Return(&[]task_crud.Task{}, nil).Once()
			},
			expectedStatus: fiber.StatusOK,
			expectedResp: map[string]interface{}{
				"data":   []interface{}{},
				"status": http.StatusText(fiber.StatusOK),
			},
		},
		{
			name:           "Internal Error",
			pathParameters: defaultUserId,
			contentType:    defaultContentType,
			setupMock: func(ms *mocks.MockTaskCrud) {
				ms.On("ReadTasksByUserId",
					mock.AnythingOfType("*fasthttp.RequestCtx"),
					&defaultUserId,
				).Return(nil, errors.New("internal error")).Once()
			},
			expectedStatus: fiber.StatusInternalServerError,
			expectedResp: map[string]interface{}{
				"error": map[string]interface{}{
					"code": response.ErrCodeInternalServerError,
				},
				"status": http.StatusText(fiber.StatusInternalServerError),
			},
		},
		{
			name:           "Validation Error",
			pathParameters: "wrong_id",
			contentType:    defaultContentType,
			setupMock:      voidSetupMock,
			expectedStatus: fiber.StatusBadRequest,
			expectedResp: map[string]interface{}{
				"error": map[string]interface{}{
					"code": response.ErrCodeValidationFailed,
					"desc": "Field '' failed validation: rule 'uuid' (value: wrong_id)",
				},
				"status": http.StatusText(fiber.StatusBadRequest),
			},
		},
		{
			name:           "User Not Found",
			pathParameters: defaultUserId,
			contentType:    defaultContentType,
			setupMock: func(ms *mocks.MockTaskCrud) {
				ms.On("ReadTasksByUserId",
					mock.AnythingOfType("*fasthttp.RequestCtx"),
					&defaultUserId,
				).Return(nil, db.ErrUserNotFound).Once()
			},
			expectedStatus: fiber.StatusNotFound,
			expectedResp: map[string]interface{}{
				"error": map[string]interface{}{
					"code": response.ErrCodeNotFound,
					"desc": "User not found",
				},
				"status": http.StatusText(fiber.StatusNotFound),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			app.Get("/tasks/user/:id", api.GetTasksByUserId)
			defer app.Shutdown()

			// Подготовка мока
			tt.setupMock(mockTaskCrud)

			// Создание запроса
			req := httptest.NewRequest(http.MethodGet, "/tasks/user/"+tt.pathParameters, nil)
			req.Header.Set("Content-Type", tt.contentType)

			// Выполнение запроса
			resp, err := app.Test(req)
			require.NoError(t, err, "test request")
			defer resp.Body.Close()

			// Проверка статуса
			assert.Equal(t, tt.expectedStatus, resp.StatusCode, "status code")

			// Проверка тела ответа (если ожидается)
			if tt.expectedResp != nil {
				var body interface{}
				err := json.NewDecoder(resp.Body).Decode(&body)
				require.NoError(t, err, "response body")
				assert.Equal(t, tt.expectedResp, body, "expected response body")
			}

			// Проверка, что все ожидания мока выполнены
			mockTaskCrud.AssertExpectations(t)

			// Сброс мока после каждого теста
			mockTaskCrud.ExpectedCalls = nil
			mockTaskCrud.Calls = nil
		})
	}
}

// БЛЯТЬ, МЕНЯ ЗАЕБАЛО КОНКРЕТНО, ОДИН ХУЙ НИКТО НЕ ПОСМОТРИТ, ОСТАЛЬНЫЕ ПО АНАЛОГИИ С ВЕРХНИМИ, СИЖУ ДРОЧУ КОГОТО, ЛЮДИ ЕБАНУЛИ НА РАБОТАЕТ/НЕРАБОТАЕТ И ЗАЕБИСЬ, А Я СИЖУ-ДРОЧУ ВСЕ УСЛОВИЯ, КАК ЕБЛАН
