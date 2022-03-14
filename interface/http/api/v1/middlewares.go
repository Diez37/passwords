package v1

import (
	"context"
	"github.com/diez37/go-packages/log"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"net/http"
)

type Uuid struct {
	logger log.Logger
}

func NewUuid(logger log.Logger) *Uuid {
	return &Uuid{logger: logger}
}

func (middleware *Uuid) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		uuid, err := uuid.Parse(chi.URLParam(request, UuidFieldName))

		if err != nil {
			http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			middleware.logger.Error(err)
			return
		}

		ctx := context.WithValue(request.Context(), UuidFieldName, uuid)
		next.ServeHTTP(writer, request.WithContext(ctx))
	})
}
