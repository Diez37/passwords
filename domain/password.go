package domain

import (
	"github.com/google/uuid"
	"time"
)

type Password struct {
	Uuid       uuid.UUID
	Login      uuid.UUID
	Password   string
	Disabled   bool
	OneTime    bool
	CreatedAt  *time.Time
	UpdateAt   *time.Time
	ValidUntil *time.Time
}
