package server

import (
	"fmt"
	gw "github.com/Vy4cheSlave/grpc_user-manager/gen/go/auth"
	"github.com/Vy4cheSlave/grpc_user-manager/internal/infrastructure/getway/server/middleware"
	"github.com/Vy4cheSlave/grpc_user-manager/internal/infrastructure/rest/dto/response"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/proxy"
	"github.com/pkg/errors"
	"log/slog"
	"net/http"
	"strings"
)

const (
	scheme = "http"
)

type Server struct {
	log             *slog.Logger
	restServer      *fiber.App
	addr            string
	authServiceAddr string
	crudServiceAddr string
}

func NewServer(log *slog.Logger, addr *string, tokenSecret *[]byte, authServiceAddr *string, crudServiceAddr *string, gatewayHandler http.Handler, authClient gw.AuthClient) *Server {
	restServer := NewRestServer(&serverAPI{
		log:             log,
		gatewayHandler:  gatewayHandler,
		crudServiceAddr: *crudServiceAddr,
		authServiceAddr: *authServiceAddr,
		authClient:      authClient,
	}, tokenSecret)
	return &Server{
		log:             log,
		restServer:      restServer,
		addr:            *addr,
		authServiceAddr: *authServiceAddr,
		crudServiceAddr: *crudServiceAddr,
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

	// auth grpc
	app.All("/v1/auth/*", adaptor.HTTPHandler(api.gatewayHandler))
	// task_crud rest
	apiGroup := app.Group("/task_crud")

	apiGroup.Get("/tasks", api.DefaultHandler(api.crudServiceAddr))
	apiGroup.Get("/tasks/user/:id", api.DefaultHandler(api.crudServiceAddr))
	apiGroup.Get("/tasks/:id", api.DefaultHandler(api.crudServiceAddr))

	// jwt
	authGroup := apiGroup.Group("/auth")
	authGroup.Use(middleware.JWTAuthMiddleware(tokenSecret))

	authGroup.Post("/tasks", api.AuthHandler(api.crudServiceAddr))
	authGroup.Put("/tasks/:id", api.AuthHandler(api.crudServiceAddr))
	authGroup.Delete("/tasks/:id", api.AuthHandler(api.crudServiceAddr))

	return app
}

func (s *Server) Run() error {
	const op = "internal/infrastructure/getway/server/handler.Server.Run"
	log := s.log.With(slog.String("operation", op), slog.String("addr", s.addr))
	log.Info("grpc server is running")

	if err := s.restServer.Listen(s.addr); err != nil {
		return errors.Wrap(err, strings.Join([]string{op, "failed to serve rest server"}, ": "))
	}

	return nil
}

type serverAPI struct {
	log             *slog.Logger
	gatewayHandler  http.Handler
	crudServiceAddr string
	authServiceAddr string
	authClient      gw.AuthClient
}

func (s *serverAPI) DefaultHandler(serviceAddr string) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		return proxy.Do(ctx, fmt.Sprintf("%s://%s%s", scheme, serviceAddr, ctx.OriginalURL()))
	}
}

func (s *serverAPI) AuthHandler(serviceAddr string) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		userId, ok := ctx.Locals("userId").(string)
		if !ok {
			return response.ReturnResponse(
				ctx,
				fiber.StatusInternalServerError,
				response.WithError(response.ErrCodeInternalServerError, "Server misconfiguration: userID not set"),
			)
		}
		_, err := s.authClient.CheckUser(ctx.Context(), &gw.CheckUserRequest{
			UserId: userId,
		})
		if err != nil {
			return response.ReturnResponse(
				ctx,
				fiber.StatusInternalServerError,
				response.WithError(response.ErrCodeInternalServerError, errors.Wrap(err, "").Error()),
			)
		}

		ctx.Request().Header.Del("Authorization")
		ctx.Request().Header.Add("X-User-Id", userId)

		return proxy.Do(ctx, fmt.Sprintf("%s://%s%s", scheme, serviceAddr, ctx.OriginalURL()))
	}
}
