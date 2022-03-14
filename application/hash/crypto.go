package hash

import (
	"context"
	"fmt"
	"github.com/Diez37/passwords/infrastructure/config"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/crypto/bcrypt"
)

type crypto struct {
	config *config.Hash
	tracer trace.Tracer
}

func NewCrypto(config *config.Hash, tracer trace.Tracer) Hasher {
	return &crypto{config: config, tracer: tracer}
}

func (hasher *crypto) Password(ctx context.Context, login uuid.UUID, password string) (string, error) {
	ctx, span := hasher.tracer.Start(ctx, "Password")
	defer span.End()

	span.SetAttributes(attribute.String("service", "crypto"))

	hash, err := bcrypt.GenerateFromPassword(hasher.makePassword(ctx, login, password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func (hasher *crypto) Check(ctx context.Context, login uuid.UUID, password string, hash string) bool {
	ctx, span := hasher.tracer.Start(ctx, "Check")
	defer span.End()

	span.SetAttributes(attribute.String("service", "crypto"))

	return bcrypt.CompareHashAndPassword([]byte(hash), hasher.makePassword(ctx, login, password)) == nil
}

func (hasher *crypto) makePassword(ctx context.Context, login uuid.UUID, password string) []byte {
	_, span := hasher.tracer.Start(ctx, "makePassword")
	defer span.End()

	return []byte(fmt.Sprintf(
		"%s%s%s",
		password,
		hasher.config.Salt,
		login.String(),
	))
}
