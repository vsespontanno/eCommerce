package pg

import (
	"context"
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/vsespontanno/eCommerce/user-service/internal/models"
)

type UserStore struct {
	db      *sql.DB
	builder sq.StatementBuilderType
}

func NewUserStore(db *sql.DB) *UserStore {
	return &UserStore{
		db:      db,
		builder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (s *UserStore) SaveUser(ctx context.Context, user *models.User) error {
	query := s.builder.Insert("users").
		Columns("userid", "firstname", "lastname", "email", "password").
		Values(user.ID, user.FirstName, user.LastName, user.Email, user.Password)
	sqlStr, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("building insert query: %w", err)
	}

	_, err = s.db.ExecContext(ctx, sqlStr, args...)
	if err != nil {
		return fmt.Errorf("executing insert: %w", err)
	}

	return nil
}

func (s *UserStore) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := s.builder.Select("userid", "firstname", "lastname", "email", "password").
		From("users").
		Where(sq.Eq{"email": email})
	sqlStr, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("building select query: %w", err)
	}
	var user models.User
	err = s.db.QueryRowContext(ctx, sqlStr, args...).Scan(
		&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.Password,
	)
	if err != nil {
		return nil, fmt.Errorf("executing select: %w", err)
	}
	return &user, nil
}
