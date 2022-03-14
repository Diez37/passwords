package v1

import (
	"github.com/google/uuid"
	"time"
)

type Password struct {
	Login      uuid.UUID  `json:"login" validate:"required"`
	Password   string     `json:"password" validate:"required"`
	OneTime    bool       `json:"one_time" validate:"-"`
	ValidUntil *time.Time `json:"valid_until" validate:"-"`
}
