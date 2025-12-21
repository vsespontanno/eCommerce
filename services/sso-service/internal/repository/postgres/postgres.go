package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/vsespontanno/eCommerce/services/sso-service/internal/domain/models"
	"github.com/vsespontanno/eCommerce/services/sso-service/internal/repository"
)

var (
	ErrUserExists = repository.ErrUserExists
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

func (s *Storage) SaveUser(ctx context.Context, email, firstName, lastName string, passHash []byte) (int64, error) {
	const op = "postgres.SaveUser"

	query := s.builder.Insert("users").
		Columns("email", "first_name", "last_name", "pass_hash").
		Values(email, firstName, lastName, passHash).
		Suffix("RETURNING id")

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return 0, fmt.Errorf("%s: failed to build query: %w", op, err)
	}

	var id int64
	err = s.db.QueryRowContext(ctx, sqlStr, args...).Scan(&id)
	if err != nil {
		// Check for unique constraint violation
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return 0, fmt.Errorf("%s: %w", op, ErrUserExists)
		}
		return 0, fmt.Errorf("%s: failed to insert user: %w", op, err)
	}

	return id, nil
}

func (s *Storage) User(ctx context.Context, email string) (models.User, error) {
	const op = "postgres.User"
	query := s.builder.Select("id", "email", "first_name", "last_name", "pass_hash").
		From("users").
		Where(sq.Eq{"email": email})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return models.User{}, fmt.Errorf("%s: failed to build query: %w", op, err)
	}

	row := s.db.QueryRowContext(ctx, sqlStr, args...)
	var user models.User
	if err := row.Scan(&user.ID, &user.Email, &user.FirstName, &user.LastName, &user.PassHash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, fmt.Errorf("%s: %w", op, repository.ErrUserNotFound)
		}
		return models.User{}, fmt.Errorf("%s: failed to scan user: %w", op, err)
	}

	return user, nil
}
