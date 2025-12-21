package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/vsespontanno/eCommerce/services/products-service/internal/domain/apperrors"
	"github.com/vsespontanno/eCommerce/services/products-service/internal/domain/products/entity"

	sq "github.com/Masterminds/squirrel"
)

type ProductStore struct {
	db      *sqlx.DB
	builder sq.StatementBuilderType
}

func NewProductStore(db *sqlx.DB) *ProductStore {
	return &ProductStore{
		db:      db,
		builder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (s *ProductStore) SaveProduct(ctx context.Context, product *entity.Product) error {
	query := s.builder.Insert("products").
		Columns("productID", "productName", "productDescription", "productPrice", "productQuantity", "created_at").
		Values(product.ID, product.Name, product.Description, product.Price, product.CountInStock, time.Now().Format(time.RFC1123Z))

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx, sqlStr, args...)
	return err
}

func (s *ProductStore) GetProducts(ctx context.Context) ([]*entity.Product, error) {
	query := s.builder.Select("productID", "productName", "productDescription", "productPrice", "created_at").
		From("products").
		RunWith(s.db)

	rows, err := query.QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*entity.Product
	for rows.Next() {
		var p entity.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.CreatedAt); err != nil {
			return nil, err
		}
		products = append(products, &p)
	}

	return products, nil
}

func (s *ProductStore) GetProductByID(ctx context.Context, id int64) (*entity.Product, error) {
	query := s.builder.Select("productID", "productName", "productDescription", "productPrice", "created_at").
		From("products").
		Where(sq.Eq{"productID": id}).
		RunWith(s.db)

	var p entity.Product
	err := query.QueryRowContext(ctx).Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.ErrNoProductFound // No product found
		}
		return nil, err
	}

	return &p, nil
}

func (s *ProductStore) GetProductsByID(ctx context.Context, ids []int64) ([]*entity.Product, error) {
	if len(ids) == 0 {
		return []*entity.Product{}, nil
	}

	query := s.builder.Select("productID", "productName", "productDescription", "productPrice", "created_at").
		From("products").
		Where(sq.Eq{"productID": ids}).
		RunWith(s.db)

	rows, err := query.QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := make([]*entity.Product, 0, len(ids))
	for rows.Next() {
		var p entity.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.CreatedAt); err != nil {
			return nil, err
		}
		products = append(products, &p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return products, nil
}
