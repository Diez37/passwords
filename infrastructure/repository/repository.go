package repository

import (
	"context"
	"github.com/google/uuid"
)

type Finder interface {
	FindByLogin(context.Context, uuid.UUID) ([]*Password, error)
	FindActiveByLogin(context.Context, uuid.UUID) ([]*Password, error)
}

type Saver interface {
	Insert(context.Context, *Password) (*Password, error)
}

type Blocker interface {
	DisableByUuids(context.Context, ...uuid.UUID) (bool, error)
}

type Paginator interface {
	Count(ctx context.Context) (int64, error)
	Page(ctx context.Context, page uint, limit uint, login uuid.UUID) ([]*Password, error)
}

type Repository interface {
	Finder
	Saver
	Blocker
	Paginator
}
