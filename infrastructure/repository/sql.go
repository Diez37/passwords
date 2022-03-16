package repository

import (
	"context"
	"github.com/Diez37/passwords/infrastructure/time"
	"github.com/diez37/go-packages/clients/db"
	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const (
	sqlTableName = "passwords"
)

type sql struct {
	db     goqu.SQLDatabase
	tracer trace.Tracer
}

func NewSql(db goqu.SQLDatabase, tracer trace.Tracer) Repository {
	return &sql{db: db, tracer: tracer}
}

func (repository *sql) Count(ctx context.Context) (int64, error) {
	ctx, span := repository.tracer.Start(ctx, "Count")
	defer span.End()

	span.SetAttributes(
		attribute.String("repository", "sql"),
	)

	sql, _, err := goqu.From(sqlTableName).Select(goqu.COUNT("uuid")).ToSQL()
	if err != nil {
		return 0, err
	}

	rows, err := repository.db.QueryContext(ctx, sql)
	if err != nil {
		return 0, err
	}

	for rows.Next() {
		count := int64(0)

		if err := rows.Scan(&count); err != nil {
			return 0, err
		}

		return count, nil
	}

	return 0, nil
}

func (repository *sql) Page(ctx context.Context, page uint, limit uint, login uuid.UUID) ([]*Password, error) {
	ctx, span := repository.tracer.Start(ctx, "Page")
	defer span.End()

	span.SetAttributes(
		attribute.Int("page", int(page)),
		attribute.Int("limit", int(limit)),
		attribute.String("repository", "sql"),
	)

	sql, args, err := goqu.From(sqlTableName).
		Where(goqu.Ex{"login": login.String()}).
		Offset(page * limit).
		Limit(limit).
		ToSQL()
	if err != nil {
		return nil, err
	}

	rows, err := repository.db.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var passwords []*Password

	for rows.Next() {
		password := &Password{}

		err := rows.Scan(
			&password.Id,
			&password.Uuid,
			&password.Login,
			&password.Password,
			&password.Disabled,
			&password.OneTime,
			&password.CreatedAt,
			&password.UpdateAt,
			&password.ValidUntil,
		)
		if err != nil {
			return nil, err
		}

		passwords = append(passwords, password)
	}

	if len(passwords) == 0 {
		return nil, db.RecordNotFoundError
	}

	return passwords, nil
}

func (repository *sql) FindByLogin(ctx context.Context, login uuid.UUID) ([]*Password, error) {
	ctx, span := repository.tracer.Start(ctx, "FindByLogin")
	defer span.End()

	span.SetAttributes(
		attribute.String("login", login.String()),
		attribute.String("repository", "sql"),
	)

	sql, args, err := goqu.From(sqlTableName).Where(goqu.Ex{"login": login}).ToSQL()
	if err != nil {
		return nil, err
	}

	return repository.find(ctx, sql, args...)
}

func (repository *sql) FindActiveByLogin(ctx context.Context, login uuid.UUID) ([]*Password, error) {
	ctx, span := repository.tracer.Start(ctx, "FindActiveByLogin")
	defer span.End()

	span.SetAttributes(
		attribute.String("login", login.String()),
		attribute.String("repository", "sql"),
	)

	sql, args, err := goqu.From(sqlTableName).Where(goqu.Ex{"login": login}, goqu.Ex{"disabled": false}).ToSQL()
	if err != nil {
		return nil, err
	}

	return repository.find(ctx, sql, args...)
}

func (repository *sql) find(ctx context.Context, sql string, args ...interface{}) ([]*Password, error) {
	ctx, span := repository.tracer.Start(ctx, "find")
	defer span.End()

	span.SetAttributes(attribute.String("repository", "sql"))

	rows, err := repository.db.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var passwords []*Password

	for rows.Next() {
		password := &Password{}

		err := rows.Scan(
			&password.Id,
			&password.Uuid,
			&password.Login,
			&password.Password,
			&password.Disabled,
			&password.OneTime,
			&password.CreatedAt,
			&password.UpdateAt,
			&password.ValidUntil,
		)
		if err != nil {
			return nil, err
		}

		passwords = append(passwords, password)
	}

	if len(passwords) == 0 {
		return nil, db.RecordNotFoundError
	}

	return passwords, nil
}

func (repository *sql) Insert(ctx context.Context, password *Password) (*Password, error) {
	ctx, span := repository.tracer.Start(ctx, "Insert")
	defer span.End()

	span.SetAttributes(attribute.String("repository", "sql"))

	password.Uuid = uuid.New()

	now := time.NowUTC()
	password.CreatedAt = &now

	sql, args, err := goqu.Insert(sqlTableName).Rows(password).ToSQL()

	if err != nil {
		return nil, err
	}

	_, err = repository.db.ExecContext(ctx, sql, args...)

	return password, err
}

func (repository *sql) DisableByUuids(ctx context.Context, uuids ...uuid.UUID) (bool, error) {
	ctx, span := repository.tracer.Start(ctx, "DisableByUuids")
	defer span.End()

	span.SetAttributes(attribute.String("repository", "sql"))

	sql, args, err := goqu.Update(sqlTableName).Set(
		goqu.Record{"disabled": true, "update_at": time.NowUTC()},
	).Where(goqu.Ex{"uuid": uuids}).ToSQL()

	if err != nil {
		return false, err
	}

	result, err := repository.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return false, err
	}

	if countUpdate, err := result.RowsAffected(); err != nil {
		return false, err
	} else if countUpdate == 0 {
		return false, db.RecordNotFoundError
	}

	return true, nil
}
