package container

import (
	"github.com/Diez37/passwords/infrastructure/config"
	"github.com/Diez37/passwords/infrastructure/repository"
	"github.com/diez37/go-packages/container"
	"github.com/go-playground/validator/v10"
)

func AddProvide(container container.Container) error {
	return container.Provides(
		repository.NewSql,
		config.NewHash,
		config.NewPassword,
		config.NewBlocker,
		validator.New,
	)
}
