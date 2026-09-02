package repository

import (
	"context"

	"gitee.com/meinongyihe/travel-rpc/ent"
)

// ProductRepository owns tenant-scoped product catalog persistence.
type ProductRepository interface {
	GetByID(ctx context.Context, id int64) (*ent.Product, error)
	List(ctx context.Context, tenantID int64, keyword, destination string, offset, limit int) ([]*ent.Product, int, error)
}

type productRepository struct {
	client *ent.Client
}

func NewProductRepository(client *ent.Client) ProductRepository {
	return &productRepository{client: client}
}

func (r *productRepository) GetByID(ctx context.Context, id int64) (*ent.Product, error) {
	return r.client.Product.Get(ctx, id)
}

func (r *productRepository) List(ctx context.Context, tenantID int64, keyword, destination string, offset, limit int) ([]*ent.Product, int, error) {
	q := r.client.Product.Query().Where(
		productTenantIDEQ(tenantID),
	)
	if keyword != "" {
		q = q.Where(productTitleContains(keyword))
	}
	if destination != "" {
		q = q.Where(productDestinationEQ(destination))
	}
	items, err := q.Offset(offset).Limit(limit).All(ctx)
	if err != nil {
		return nil, 0, err
	}
	count, err := q.Clone().Offset(0).Limit(0).Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	return items, count, nil
}
