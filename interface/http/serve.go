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
	"net/http"
)

// Serve configuration and running http server
func Serve(ctx context.Context, container container.Container, logger log.Logger, service password.Service, blocker blocker.Blocker) error {
	return container.Invoke(func(
		server *http.Server,
		config *httpServer.Config,
		router chi.Router,

		repository repository.Repository,
		tracer trace.Tracer,
		validator *validator.Validate,
	) error {
		router.Mount("/api", api.Router(
			repository,
			tracer,
			logger,
			validator,
			service,
			blocker,
		))

		errGroup := &errgroup.Group{}

		httpCtx, httpCancelFunc := context.WithCancel(ctx)
		defer httpCancelFunc()

		errGroup.Go(func() error {
			logger.Infof("http server: started")
			if err := server.ListenAndServe(); err != http.ErrServerClosed {
				httpCancelFunc()
				return err
			}

			return nil
		})

		errGroup.Go(func() error {
			<-httpCtx.Done()

			logger.Infof("http server: shutdown")

			ctxTimeout, cancelFnc := context.WithTimeout(context.Background(), config.ShutdownTimeout)
			defer cancelFnc()

			return server.Shutdown(ctxTimeout)
		})

		return errGroup.Wait()
	})
}
