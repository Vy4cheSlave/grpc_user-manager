package auth

import (
	"context"
	domainAuth "github.com/Vy4cheSlave/grpc_user-manager/internal/domain/auth"
	dbImpl "github.com/Vy4cheSlave/grpc_user-manager/internal/infrastructure/db"
	//grpcImpl "github.com/Vy4cheSlave/grpc_user-manager/internal/infrastructure/grpc"
	"github.com/Vy4cheSlave/grpc_user-manager/internal/jwt"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"time"
)

type Auth struct {
	log         *slog.Logger
	usrSaver    UserSaver
	usrProvider UserProvider
	tokenTTL    time.Duration
	tokenSecret []byte
}

type UserSaver interface {
	SaveUser(
		ctx context.Context,
		username string,
		passwordHash []byte,
	) (userId *string, err error)
}

type UserProvider interface {
	GetUser(ctx context.Context, username string) (*domainAuth.User, error)
	GetUserNameByUserId(ctx context.Context, userId *string) (*string, error)
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

func NewAuthService(log *slog.Logger, usrSaver UserSaver, usrProvider UserProvider, tokenTTL time.Duration, tokenSecret []byte) *Auth {
	return &Auth{
		log:         log,
		usrSaver:    usrSaver,
		usrProvider: usrProvider,
		tokenTTL:    tokenTTL,
		tokenSecret: tokenSecret,
	}
}

func (a *Auth) Login(ctx context.Context, username string, password string) (token string, err error) {
	const op = "internal/usecase/auth/service.Auth.Login"
	log := a.log.With(slog.String("operation", op))
	log.Info("attempting to login user")

	user, err := a.usrProvider.GetUser(ctx, username)
	if err != nil {
		if errors.Is(err, dbImpl.ErrUserNotFound) {
			a.log.Warn("user not found", slog.String("error", err.Error()))
			return "", errors.Wrap(ErrInvalidCredentials, op)
		}

		a.log.Error("failed to get user", slog.String("error", err.Error()))
		return "", errors.Wrap(err, op)
	}

	if err = bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password)); err != nil {
		a.log.Info("invalid credentials", slog.String("error", err.Error()))
		return "", errors.Wrap(ErrInvalidCredentials, op)
	}

	log.Info("successfully logged in")

	token, err = jwt.NewToken(user, a.tokenTTL, a.tokenSecret)
	if err != nil {
		a.log.Error("failed to create token", slog.String("error", err.Error()))
		return "", errors.Wrap(err, op)
	}

	return token, nil
}

func (a *Auth) Register(ctx context.Context, username string, password string) (userId *string, err error) {
	const op = "internal/usecase/auth/service.Auth.Register"
	log := a.log.With(slog.String("operation", op))
	log.Info("registering user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to hash password hash", slog.String("error", err.Error()))
		return nil, errors.Wrap(err, op)
	}

	id, err := a.usrSaver.SaveUser(ctx, username, passHash)
	if err != nil {
		if errors.Is(err, dbImpl.ErrUserExists) {
			log.Warn("user already exists", slog.String("error", err.Error()))
			return nil, errors.Wrap(ErrInvalidCredentials, op)
		}

		log.Error("failed to save user", slog.String("error", err.Error()))
		return nil, errors.Wrap(err, op)
	}

	log.Info("user registered successfully")

	return id, nil
}

func (a *Auth) CheckUser(ctx context.Context, username string) (userId *string, err error) {
	const op = "internal/usecase/auth/service.Auth.CheckUser"
	log := a.log.With(slog.String("operation", op))
	log.Info("checking user")
	user, err := a.usrProvider.GetUserNameByUserId(ctx, &username)
	if err != nil {
		if errors.Is(err, dbImpl.ErrUserNotFound) {
			a.log.Warn("user not found", slog.String("error", err.Error()))
			return nil, errors.Wrap(ErrInvalidCredentials, op)
		}
		a.log.Error("failed to get user", slog.String("error", err.Error()))
		return nil, errors.Wrap(err, op)
	}
	log.Info("successfully checked user")
	return user, nil
}
