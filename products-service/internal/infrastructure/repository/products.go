package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/vsespontanno/eCommerce/products-service/internal/domain/apperrors"
	"github.com/vsespontanno/eCommerce/products-service/internal/domain/products/entity"

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
		Values(product.ID, product.Name, product.Description, product.Price, product.CountInStock, time.Now().Format(time.RFC1123Z)).
		RunWith(s.db)

	return query.QueryRowContext(ctx).Scan(&product.ID)
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
	products := make([]*entity.Product, 0, len(ids))
	for _, id := range ids {
		product, err := s.GetProductByID(ctx, id)
		if err != nil {
			if errors.Is(err, apperrors.ErrNoProductFound) {
				continue
			} else {
				return nil, err
			}
		}
		products = append(products, product)
	}
	return products, nil
}
