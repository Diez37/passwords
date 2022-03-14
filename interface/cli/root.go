package cli

import (
	"context"
	"github.com/Diez37/passwords/application/blocker"
	"github.com/Diez37/passwords/application/hash"
	"github.com/Diez37/passwords/application/password"
	"github.com/Diez37/passwords/infrastructure/config"
	container2 "github.com/Diez37/passwords/infrastructure/container"
	"github.com/Diez37/passwords/infrastructure/repository"
	"github.com/Diez37/passwords/interface/http"
	"github.com/Diez37/passwords/interface/repeater"
	"github.com/diez37/go-packages/app"
	"github.com/diez37/go-packages/closer"
	"github.com/diez37/go-packages/configurator"
	bindFlags "github.com/diez37/go-packages/configurator/bind_flags"
	"github.com/diez37/go-packages/container"
	"github.com/diez37/go-packages/log"
	"github.com/golang-migrate/migrate/v4"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/errgroup"
)

const (
	// AppName name of application
	AppName = "passwords"
)

// NewRootCommand creating, configuration and return cobra.Command for root command
func NewRootCommand() (*cobra.Command, error) {
	container := container.GetContainer()

	if err := container2.AddProvide(container); err != nil {
		return nil, err
	}

	cmd := &cobra.Command{
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			return container.Invoke(func(
				generalConfig *app.Config,
				configurator configurator.Configurator,
				blockerConfig *config.Blocker,
				hashConfig *config.Hash,
				passwordConfig *config.Password,
			) {
				app.Configuration(generalConfig, configurator, app.WithAppName(AppName))

				configurator.SetDefault(config.BlockerBlockIntervalFieldName, config.BlockerBlockIntervalDefault)
				configurator.SetDefault(config.PasswordLifetimeFieldName, config.PasswordLifetimeDefault)
				configurator.SetDefault(config.HashSaltFieldName, config.HashSaltDefault)

				if blockInterval := configurator.GetDuration(config.BlockerBlockIntervalFieldName); blockerConfig.BlockInterval == config.BlockerBlockIntervalDefault {
					blockerConfig.BlockInterval = blockInterval
				}

				if lifetime := configurator.GetDuration(config.PasswordLifetimeFieldName); passwordConfig.Lifetime == config.PasswordLifetimeDefault {
					passwordConfig.Lifetime = lifetime
				}

				if salt := configurator.GetString(config.HashSaltFieldName); hashConfig.Salt == config.HashSaltDefault {
					hashConfig.Salt = salt
				}
			})
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return container.Invoke(func(
				generalConfig *app.Config,
				logger log.Logger,
				closer closer.Closer,
				tracer trace.Tracer,
				repository repository.Repository,
				hashConfig *config.Hash,
				passwordConfig *config.Password,
				blockerConfig *config.Blocker,
				migrator *migrate.Migrate,
			) error {
				logger.Infof("app: %s started", generalConfig.Name)
				logger.Infof("app: pid - %d", generalConfig.PID)

				if err := migrator.Up(); err != nil && err != migrate.ErrNoChange {
					return err
				}

				hasher := hash.NewCrypto(hashConfig, tracer)
				blocker := blocker.NewBlocker(repository, tracer)
				service := password.NewPassword(passwordConfig, hasher, repository, tracer, blocker)

				ctx, generalCancelFnc := context.WithCancel(closer.GetContext())
				defer generalCancelFnc()

				wg := &errgroup.Group{}
				wg.Go(func() error {
					parentCtx, cancelFunc := context.WithCancel(ctx)
					defer cancelFunc()

					if err := http.Serve(parentCtx, container, logger, service, blocker); err != nil {
						generalCancelFnc()
						return err
					}

					return nil
				})

				wg.Go(func() error {
					repeater.Serve(ctx, blockerConfig, logger, blocker)

					return nil
				})

				return wg.Wait()
			})
		},
	}

	cmd, err := bindFlags.CobraCmd(container, cmd,
		bindFlags.HttpServer,
		bindFlags.Logger,
		bindFlags.Tracer,
	)
	if err != nil {
		return nil, err
	}

	container.Invoke(func(blockerConfig *config.Blocker, hashConfig *config.Hash, passwordConfig *config.Password) {
		cmd.PersistentFlags().DurationVar(&blockerConfig.BlockInterval, config.BlockerBlockIntervalFieldName, config.BlockerBlockIntervalDefault, "")
		cmd.PersistentFlags().DurationVar(&passwordConfig.Lifetime, config.PasswordLifetimeFieldName, config.PasswordLifetimeDefault, "")
		cmd.PersistentFlags().StringVar(&hashConfig.Salt, config.HashSaltFieldName, config.HashSaltDefault, "")
	})

	return cmd, nil
}
