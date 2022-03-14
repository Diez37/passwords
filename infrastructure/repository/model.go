package repository

import (
	"github.com/google/uuid"
	"time"
)

type Password struct {
	Id         int        `db:"-"`
	Uuid       uuid.UUID  `db:"uuid"`
	Login      uuid.UUID  `db:"login"`
	Password   string     `db:"password"`
	Disabled   bool       `db:"disabled"`
	OneTime    bool       `db:"one_time"`
	CreatedAt  *time.Time `db:"created_at"`
	UpdateAt   *time.Time `db:"update_at"`
	ValidUntil *time.Time `db:"valid_until"`
}
