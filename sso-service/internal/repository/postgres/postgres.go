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

func (s *Storage) SaveUser(ctx context.Context, email, FirstName, LastName string, passHash []byte) (int64, error) {
	const op = "postgres.SaveUser"
	fmt.Println("Saving user:", email)
	query := s.builder.Insert("users").
		Columns("email", "first_name", "last_name", "pass_hash").
		Values(email, FirstName, LastName, passHash)
	sqlStr, args, err := query.ToSql()

	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	_, err = s.db.ExecContext(ctx, sqlStr, args...)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	var id int64
	err = s.db.QueryRowContext(ctx, "SELECT id FROM users WHERE email = $1", email).Scan(&id)
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
			return models.User{}, fmt.Errorf("%s: User does not exist", op)
		}
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}
