package response

import (
	"github.com/gofiber/fiber/v2"
	"net/http"
)

// Error Codes
const (
	ErrCodeInvalidToken        = "INVALID_TOKEN"
	ErrCodeMissingAuthHeader   = "MISSING_AUTH_HEADER"
	ErrCodeValidationFailed    = "VALIDATION_FAILED"
	ErrCodeJsonParsingFailed   = "JSON_PARSING_FAILED"
	ErrCodeInternalServerError = "INTERNAL_SERVER_ERROR"
	ErrCodeUnauthorized        = "UNAUTHORIZED"
	ErrCodeNotFound            = "NOT_FOUND"
)

type Response struct {
	Status string `json:"status"`
	Error  *Error `json:"error,omitempty"`
	Data   any    `json:"data,omitempty"`
	Meta   error  `json:"-"`
}

type Error struct {
	Code string `json:"code"`
	Desc string `json:"desc,omitempty"`
}

//func (e *Error) Error() string {
//	return strings.Join([]string{e.Code, e.Desc}, ": ")
//}

func ReturnResponse(ctx *fiber.Ctx, httpStatus int, opts ...Option) error {
	resp := &Response{Status: http.StatusText(httpStatus)}

	for _, opt := range opts {
		opt(resp)
	}

	// не знаю насколько я имбулечку придумал или хуйню галимую (вызывается в middleware/logg.go)
	// Пояснение: ответ всегда приходит пользователю обработанный и не перехватывается другими middleware,
	// а ошибка на стороне сервера логируется в одном месте(почти) и не перехватывается другими middleware
	if resp.Meta != nil {
		ctx.Locals("handlerError", resp.Meta)
	}

	return ctx.Status(httpStatus).JSON(resp)
}

type Option func(*Response)

//func WithStatus(status string) Option {
//	return func(r *Response) {
//		r.Status = status
//	}
//}

func WithError(code, desc string) Option {
	return func(r *Response) {
		r.Error = &Error{Code: code, Desc: desc}
	}
}

func WithData(data any) Option {
	return func(r *Response) {
		r.Data = data
	}
}

func WithMeta(meta error) Option {
	return func(r *Response) {
		r.Meta = meta
	}
}

// схемы
type TaskCreateResponse struct {
	TaskId int `json:"task_id"`
}
