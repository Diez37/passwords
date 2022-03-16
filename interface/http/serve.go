package http

import (
	"context"
	"github.com/Diez37/passwords/application/blocker"
	"github.com/Diez37/passwords/application/password"
	"github.com/Diez37/passwords/infrastructure/repository"
	"github.com/Diez37/passwords/interface/http/api"
	"github.com/diez37/go-packages/container"
	"github.com/diez37/go-packages/log"
	httpServer "github.com/diez37/go-packages/server/http"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/errgroup"
	"net"
	"net/http"
)

// Serve configuration and running http server
func Serve(
	ctx context.Context,
	container container.Container,
	logger log.Logger,
	service password.Service,
	blocker blocker.Blocker,
) error {
	ctx, cancelFunc := context.WithCancel(ctx)
	defer cancelFunc()

	errGroup := &errgroup.Group{}

	err := container.Invoke(func(
		server *http.Server,
		config *httpServer.Config,
		router chi.Router,

		repository repository.Repository,
		tracer trace.Tracer,
		validator *validator.Validate,
	) {
		router.Mount("/api", api.Router(
			repository,
			tracer,
			logger,
			validator,
			service,
			blocker,
		))

		errGroup.Go(func() error {
			defer cancelFunc()

			logger.Infof("http server: started")

			server.BaseContext = func(_ net.Listener) context.Context {
				return ctx
			}

			if err := server.ListenAndServe(); err != http.ErrServerClosed {
				return err
			}

			return nil
		})

		errGroup.Go(func() error {
			<-ctx.Done()

			logger.Infof("http server: shutdown")

			return server.Close()
		})
	})

	if err != nil {
		return err
	}

	return errGroup.Wait()
}
