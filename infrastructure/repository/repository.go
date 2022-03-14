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

type Repository interface {
	Finder
	Saver
	Blocker
}
