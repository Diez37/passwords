package blocker

import (
	"context"
	"github.com/Diez37/passwords/infrastructure/repository"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"sync"
)

type Blocker interface {
	Add(context.Context, uuid.UUID)
	Block(context.Context) error
}

type blocker struct {
	mutex *sync.Mutex

	uuids []uuid.UUID

	repository repository.Repository
	tracer     trace.Tracer
}

func NewBlocker(repository repository.Repository, tracer trace.Tracer) Blocker {
	return &blocker{repository: repository, tracer: tracer, mutex: &sync.Mutex{}}
}

func (service *blocker) Add(ctx context.Context, uuid uuid.UUID) {
	_, span := service.tracer.Start(ctx, "Add")
	defer span.End()

	span.SetAttributes(attribute.String("service", "blocker"))

	service.mutex.Lock()
	defer service.mutex.Unlock()

	service.uuids = append(service.uuids, uuid)
}

func (service *blocker) Block(ctx context.Context) error {
	ctx, span := service.tracer.Start(ctx, "Block")
	defer span.End()

	span.SetAttributes(attribute.String("service", "blocker"))

	service.mutex.Lock()

	if len(service.uuids) == 0 {
		service.mutex.Unlock()
		return nil
	}

	uuids := service.uuids
	service.uuids = []uuid.UUID{}
	service.mutex.Unlock()

	_, err := service.repository.DisableByUuids(ctx, uuids...)
	return err
}
