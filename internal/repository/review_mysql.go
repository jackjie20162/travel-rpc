package repository

import (
	"context"
	"time"

	"gitee.com/meinongyihe/travel-rpc/ent"
	"gitee.com/meinongyihe/travel-rpc/ent/predicate"
	"gitee.com/meinongyihe/travel-rpc/ent/review"
)

type mysqlReviewRepository struct{ client *ent.Client }

func NewReviewRepository(client *ent.Client) ReviewRepository {
	return &mysqlReviewRepository{client: client}
}

func (r *mysqlReviewRepository) Create(ctx context.Context, input CreateReviewInput) (*ent.Review, error) {
	builder := r.client.Review.Create().
		SetTenantID(input.TenantID).
		SetMerchantID(input.MerchantID).
		SetOrderID(input.OrderID).
		SetOrderNo(input.OrderNo).
		SetProductID(input.ProductID).
		SetUserID(input.UserID).
		SetRating(input.Rating).
		SetStatus("APPROVED")
	if input.ServiceRating > 0 {
		builder.SetServiceRating(input.ServiceRating)
	}
	if input.ValueRating > 0 {
		builder.SetValueRating(input.ValueRating)
	}
	if input.Content != "" {
		builder.SetContent(input.Content)
	}
	if input.Images != "" {
		builder.SetImages(input.Images)
	}
	return builder.Save(ctx)
}

func (r *mysqlReviewRepository) GetByOrderNo(ctx context.Context, orderNo string) (*ent.Review, error) {
	return r.client.Review.Query().Where(review.OrderNoEQ(orderNo)).Only(ctx)
}

func (r *mysqlReviewRepository) ListByProduct(ctx context.Context, tenantID int64, productID int64, status string, page, pageSize int32) ([]*ent.Review, int64, error) {
	preds := []predicate.Review{review.TenantIDEQ(tenantID), review.ProductIDEQ(productID)}
	if status != "" {
		preds = append(preds, review.StatusEQ(status))
	} else {
		// Default: show APPROVED and REPLIED reviews
		preds = append(preds, review.StatusIn("APPROVED", "REPLIED"))
	}
	q := r.client.Review.Query().Where(preds...)
	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	if page > 0 && pageSize > 0 {
		q = q.Offset(int((page - 1) * pageSize)).Limit(int(pageSize))
	}
	items, err := q.Order(ent.Desc(review.FieldID)).All(ctx)
	if err != nil {
		return nil, 0, err
	}
	return items, int64(total), nil
}

func (r *mysqlReviewRepository) AvgRatingByProduct(ctx context.Context, tenantID int64, productID int64) (float64, int64, error) {
	reviews, err := r.client.Review.Query().Where(
		review.TenantIDEQ(tenantID),
		review.ProductIDEQ(productID),
		review.StatusIn("APPROVED", "REPLIED"),
	).All(ctx)
	if err != nil {
		return 0, 0, err
	}
	if len(reviews) == 0 {
		return 0, 0, nil
	}
	var sum int
	for _, rv := range reviews {
		sum += rv.Rating
	}
	return float64(sum) / float64(len(reviews)), int64(len(reviews)), nil
}

func (r *mysqlReviewRepository) Reply(ctx context.Context, tenantID, merchantID int64, reviewID int64, replyContent string) (*ent.Review, error) {
	return r.client.Review.UpdateOneID(int(reviewID)).
		Where(review.TenantIDEQ(tenantID), review.MerchantIDEQ(merchantID)).
		SetReplyContent(replyContent).
		SetReplyTime(time.Now().Unix()).
		SetStatus("REPLIED").
		Save(ctx)
}
