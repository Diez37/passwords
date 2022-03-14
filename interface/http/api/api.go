package api

import (
	"fmt"
	"github.com/Diez37/passwords/application/blocker"
	service "github.com/Diez37/passwords/application/password"
	"github.com/Diez37/passwords/infrastructure/repository"
	"github.com/Diez37/passwords/interface/http/api/v1"
	"github.com/diez37/go-packages/log"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"go.opentelemetry.io/otel/trace"
)

func Router(
	repository repository.Repository,
	tracer trace.Tracer,
	logger log.Logger,
	validator *validator.Validate,
	service service.Service,
	blocker blocker.Blocker,
) chi.Router {
	apiV1 := v1.NewAPI(repository, tracer, logger, validator, service, blocker)

	router := chi.NewRouter()

	router.Route("/v1/password", func(r chi.Router) {
		r.Put("/", apiV1.Add)
		r.Options("/", apiV1.Check)

		r.Route(fmt.Sprintf("/{%s}", v1.UuidFieldName), func(r chi.Router) {
			r.Use(v1.NewUuid(logger).Middleware)
			r.Delete("/", apiV1.Delete)
		})
	})

	return router
}
