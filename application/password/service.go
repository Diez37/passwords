package password

import (
	"context"
	"errors"
	"github.com/Diez37/passwords/application/blocker"
	"github.com/Diez37/passwords/application/hash"
	"github.com/Diez37/passwords/domain"
	"github.com/Diez37/passwords/infrastructure/config"
	"github.com/Diez37/passwords/infrastructure/repository"
	"github.com/Diez37/passwords/infrastructure/time"
	"github.com/diez37/go-packages/clients/db"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var (
	AlreadyExistError = errors.New("password already exist")
)

type Service interface {
	Add(ctx context.Context, password *domain.Password) error
	Check(ctx context.Context, password *domain.Password) (bool, error)
}

type password struct {
	config     *config.Password
	blocker    blocker.Blocker
	hasher     hash.Hasher
	repository repository.Repository
	tracer     trace.Tracer
}

func NewPassword(config *config.Password, hasher hash.Hasher, repository repository.Repository, tracer trace.Tracer, blocker blocker.Blocker) Service {
	return &password{config: config, hasher: hasher, repository: repository, tracer: tracer, blocker: blocker}
}

func (service *password) Add(ctx context.Context, password *domain.Password) error {
	ctx, span := service.tracer.Start(ctx, "FindByLogin")
	defer span.End()

	span.SetAttributes(attribute.String("service", "password"))

	passwords, err := service.repository.FindByLogin(ctx, password.Login)
	if err != nil && err != db.RecordNotFoundError {
		return err
	}

	for _, pas := range passwords {
		if service.hasher.Check(ctx, password.Login, password.Password, pas.Password) {
			return AlreadyExistError
		}
	}

	passwordHash, err := service.hasher.Password(ctx, password.Login, password.Password)
	if err != nil {
		return err
	}

	ValidUntil := time.NowUTC().Add(service.config.Lifetime)
	if password.ValidUntil != nil {
		ValidUntil = *password.ValidUntil
	}

	_, err = service.repository.Insert(ctx, &repository.Password{
		Login:      password.Login,
		Password:   passwordHash,
		OneTime:    password.OneTime,
		ValidUntil: &ValidUntil,
	})

	return err
}

func (service *password) Check(ctx context.Context, password *domain.Password) (bool, error) {
	ctx, span := service.tracer.Start(ctx, "Check")
	defer span.End()

	span.SetAttributes(attribute.String("service", "password"))

	passwords, err := service.repository.FindActiveByLogin(ctx, password.Login)
	if err != nil && err != db.RecordNotFoundError {
		return false, err
	}

	for _, pas := range passwords {
		if service.hasher.Check(ctx, password.Login, password.Password, pas.Password) {
			if pas.ValidUntil.Sub(time.NowUTC()).Seconds() <= 0 {
				service.blocker.Add(ctx, pas.Uuid)
				return false, nil
			}

			if pas.OneTime {
				service.blocker.Add(ctx, pas.Uuid)
			}

			return true, nil
		}
	}

	return false, nil
}
