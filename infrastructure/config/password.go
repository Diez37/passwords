package config

import "time"

const (
	PasswordLifetimeFieldName = "password.lifetime"

	PasswordLifetimeDefault = 2 * 12 * 30 * 24 * time.Hour
)

type Password struct {
	Lifetime time.Duration
}

func NewPassword() *Password {
	return &Password{}
}
