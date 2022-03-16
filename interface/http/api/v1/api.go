package v1

import (
	"encoding/json"
	"github.com/Diez37/passwords/application/blocker"
	service "github.com/Diez37/passwords/application/password"
	"github.com/Diez37/passwords/domain"
	"github.com/Diez37/passwords/infrastructure/repository"
	"github.com/diez37/go-packages/clients/db"
	"github.com/diez37/go-packages/log"
	"github.com/diez37/go-packages/router/middlewares"
	"github.com/go-http-utils/headers"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/ldez/mimetype"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/errgroup"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
)

type API struct {
	repository repository.Repository
	tracer     trace.Tracer
	logger     log.Logger
	validator  *validator.Validate

	service service.Service
	blocker blocker.Blocker
}

func NewAPI(
	repository repository.Repository,
	tracer trace.Tracer,
	logger log.Logger,
	validator *validator.Validate,
	service service.Service,
	blocker blocker.Blocker,
) *API {
	return &API{repository: repository, tracer: tracer, logger: logger, validator: validator, service: service, blocker: blocker}
}

func (handler *API) Add(writer http.ResponseWriter, request *http.Request) {
	ctx, span := handler.tracer.Start(request.Context(), "Add")
	defer span.End()

	span.SetAttributes(attribute.String("handler", "api.v1"))

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		handler.logger.Error(err)
		return
	}

	password := Password{}
	if err := json.Unmarshal(body, &password); err != nil {
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		handler.logger.Error(err)
		return
	}

	if err := handler.validator.Struct(password); err != nil {
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		handler.logger.Error(err)
		return
	}

	err = handler.service.Add(ctx, &domain.Password{
		Login:      password.Login,
		Password:   password.Password,
		OneTime:    password.OneTime,
		ValidUntil: password.ValidUntil,
	})
	if err != nil && err != service.AlreadyExistError {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		handler.logger.Error(err)
		return
	}

	if err == service.AlreadyExistError {
		http.Error(writer, http.StatusText(http.StatusConflict), http.StatusConflict)
		return
	}

	writer.WriteHeader(http.StatusOK)
}

func (handler *API) Check(writer http.ResponseWriter, request *http.Request) {
	ctx, span := handler.tracer.Start(request.Context(), "Check")
	defer span.End()

	span.SetAttributes(attribute.String("handler", "api.v1"))

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		handler.logger.Error(err)
		return
	}

	password := Password{}
	if err := json.Unmarshal(body, &password); err != nil {
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		handler.logger.Error(err)
		return
	}

	if err := handler.validator.Struct(password); err != nil {
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		handler.logger.Error(err)
		return
	}

	ok, err := handler.service.Check(ctx, &domain.Password{
		Login:      password.Login,
		Password:   password.Password,
		OneTime:    password.OneTime,
		ValidUntil: password.ValidUntil,
	})
	if err != nil && err != db.RecordNotFoundError {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		handler.logger.Error(err)
		return
	}

	if err == db.RecordNotFoundError || !ok {
		http.Error(writer, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	writer.WriteHeader(http.StatusOK)
}

func (handler *API) Delete(writer http.ResponseWriter, request *http.Request) {
	ctx, span := handler.tracer.Start(request.Context(), "Delete")
	defer span.End()

	span.SetAttributes(attribute.String("handler", "api.v1"))

	handler.blocker.Add(ctx, ctx.Value(UuidFieldName).(uuid.UUID))

	writer.WriteHeader(http.StatusAccepted)
}

func (handler *API) Page(writer http.ResponseWriter, request *http.Request) {
	ctx, span := handler.tracer.Start(request.Context(), "Page")
	defer span.End()

	span.SetAttributes(
		attribute.String("interface", "http"),
		attribute.String("handler", "api.v1"),
	)

	page := ctx.Value(middlewares.PageFieldName).(uint)
	limit := ctx.Value(middlewares.LimitFieldName).(uint)
	login := ctx.Value(LoginFieldName).(uuid.UUID)

	var totalCount int64
	var models []*repository.Password

	wg := &errgroup.Group{}

	wg.Go(func() error {
		count, err := handler.repository.Count(ctx)
		totalCount = count

		return err
	})

	wg.Go(func() error {
		passwords, err := handler.repository.Page(ctx, page-1, limit, login)
		models = passwords

		return err
	})

	if err := wg.Wait(); err != nil && err != io.EOF {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		handler.logger.Error(err)
		return
	}

	passwords := make([]*PasswordForPage, len(models))
	for index, password := range models {
		passwords[index] = &PasswordForPage{
			Uuid:       password.Uuid,
			Login:      password.Login,
			OneTime:    password.OneTime,
			ValidUntil: password.ValidUntil,
			Disabled:   password.Disabled,
		}
	}

	content, err := json.Marshal(&Page{
		Meta: &Meta{
			Count: totalCount,
			Page:  page,
			Limit: limit,
		},
		Records: passwords,
	})
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		handler.logger.Error(err)
		return
	}

	writer.Header().Set(headers.ContentType, mimetype.ApplicationJSON)
	writer.Header().Set(middlewares.CountHeaderName, strconv.FormatInt(totalCount, 10))
	writer.Header().Set(middlewares.PageHeaderName, strconv.FormatUint(uint64(page), 10))
	writer.Header().Set(middlewares.LimitHeaderName, strconv.FormatUint(uint64(limit), 10))
	writer.WriteHeader(http.StatusOK)

	if _, err := writer.Write(content); err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		handler.logger.Error(err)
	}
}
