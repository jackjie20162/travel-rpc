package repository

import (
	"context"

	"gitee.com/meinongyihe/travel-rpc/ent"
)

type ReviewRepository interface {
	Create(ctx context.Context, input CreateReviewInput) (*ent.Review, error)
	GetByOrderNo(ctx context.Context, orderNo string) (*ent.Review, error)
	ListByProduct(ctx context.Context, tenantID int64, productID int64, status string, page, pageSize int32) ([]*ent.Review, int64, error)
	AvgRatingByProduct(ctx context.Context, tenantID int64, productID int64) (float64, int64, error)
	Reply(ctx context.Context, tenantID, merchantID int64, reviewID int64, replyContent string) (*ent.Review, error)
}

type CreateReviewInput struct {
	TenantID      int64
	MerchantID    int64
	OrderID       int64
	OrderNo       string
	ProductID     int64
	UserID        int64
	Rating        int
	ServiceRating int
	ValueRating   int
	Content       string
	Images        string
}
