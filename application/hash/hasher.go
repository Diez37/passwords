package hash

import (
	"context"
	"github.com/google/uuid"
)

type Hasher interface {
	Password(ctx context.Context, login uuid.UUID, password string) (string, error)
	Check(ctx context.Context, login uuid.UUID, password string, hash string) bool
}
