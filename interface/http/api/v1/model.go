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

type Page struct {
	Meta    *Meta              `json:"meta"`
	Records []*PasswordForPage `json:"records"`
}

type Meta struct {
	Count int64 `json:"count"`
	Page  uint  `json:"page"`
	Limit uint  `json:"limit"`
}

type PasswordForPage struct {
	Uuid       uuid.UUID  `json:"uuid"`
	Login      uuid.UUID  `json:"login"`
	OneTime    bool       `json:"one_time"`
	ValidUntil *time.Time `json:"valid_until"`
	Disabled   bool       `json:"disabled"`
}
