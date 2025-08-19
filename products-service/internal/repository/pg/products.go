package pg

import (
	"context"
	"database/sql"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/vsespontanno/eCommerce/products-service/internal/models"
)

type ProductStore struct {
	db      *sql.DB
	builder sq.StatementBuilderType
}

func NewProductStore(db *sql.DB) *ProductStore {
	return &ProductStore{
		db:      db,
		builder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (s *ProductStore) SaveProduct(ctx context.Context, product *models.Product) error {
	query := s.builder.Insert("products").
		Columns("productID", "productName", "productDescription", "productPrice", "created_at").
		Values(product.ID, product.Name, product.Description, product.Price, time.Now().Format(time.RFC1123Z)).
		RunWith(s.db)

	return query.QueryRowContext(ctx).Scan(&product.ID)
}

func (s *ProductStore) GetProducts(ctx context.Context) ([]*models.Product, error) {
	query := s.builder.Select("productID", "productName", "productDescription", "productPrice", "created_at").
		From("products").
		RunWith(s.db)

	rows, err := query.QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*models.Product
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.CreatedAt); err != nil {
			return nil, err
		}
		products = append(products, &p)
	}

	return products, nil
}
