package postgres

import (
	"context"
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/vsespontanno/eCommerce/sso-service/internal/domain/models"
)

type Storage struct {
	db      *sql.DB
	builder sq.StatementBuilderType
}

func NewStorage(db *sql.DB) *Storage {
	return &Storage{
		db:      db,
		builder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (s *Storage) SaveUser(ctx context.Context, email string, passHash []byte) (int64, error) {
	const op = "postgres.SaveUser"
	query := s.builder.Insert("users").
		Columns("email", "pass_hash").
		Values(email, passHash)
	sqlStr, args, err := query.ToSql()

	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	res, err := s.db.ExecContext(ctx, sqlStr, args...)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	return id, nil
}

func (s *Storage) User(ctx context.Context, email string) (models.User, error) {
	const op = "postgres.User"
	query := s.builder.Select("id", "email", "pass_hash").
		From("users").
		Where(sq.Eq{"email": email})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	row := s.db.QueryRowContext(ctx, sqlStr, args...)
	var user models.User
	if err := row.Scan(&user.ID, &user.Email, &user.PassHash); err != nil {
		if err == sql.ErrNoRows {
			return models.User{}, fmt.Errorf("%s: user not found", op)
		}
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (s *Storage) App(ctx context.Context, appID int) (models.App, error) {
	const op = "postgres.App"
	query := s.builder.Select("id", "name", "secret").
		From("apps").
		Where(sq.Eq{"id": appID})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return models.App{}, fmt.Errorf("%s: %w", op, err)
	}

	row := s.db.QueryRowContext(ctx, sqlStr, args...)
	var app models.App
	if err := row.Scan(&app.ID, &app.Name, &app.Secret); err != nil {
		if err == sql.ErrNoRows {
			return models.App{}, fmt.Errorf("%s: app not found", op)
		}
		return models.App{}, fmt.Errorf("%s: %w", op, err)
	}

	return app, nil
}
