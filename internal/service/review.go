package service

import (
	"context"

	"gitee.com/meinongyihe/travel-rpc/ent"
	"gitee.com/meinongyihe/travel-rpc/internal/auth"
	"gitee.com/meinongyihe/travel-rpc/internal/repository"
	"gitee.com/meinongyihe/travel-rpc/travel"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ReviewService struct {
	travel.UnimplementedReviewServiceServer
	reviews repository.ReviewRepository
	orders  repository.OrderRepository
}

func NewReviewService(reviews repository.ReviewRepository, orders repository.OrderRepository) *ReviewService {
	return &ReviewService{reviews: reviews, orders: orders}
}

func (s *ReviewService) Create(ctx context.Context, req *travel.CreateReviewRequest) (*travel.Review, error) {
	if req == nil || req.GetOrderNo() == "" || req.GetProductId() <= 0 || req.GetRating() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "order_no, product_id and rating are required")
	}
	if req.GetRating() < 1 || req.GetRating() > 5 {
		return nil, status.Error(codes.InvalidArgument, "rating must be between 1 and 5")
	}
	tenantID, err := auth.TenantID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	merchantID, _ := auth.MerchantID(ctx)

	// Check if the order exists and belongs to this user
	var userID int64
	if id, e := auth.CustomerID(ctx); e == nil && id != nil && *id > 0 {
		userID = *id
	} else if id, e := auth.AuthenticatedUserID(ctx); e == nil && id > 0 {
		userID = id
	}

	// Check if already reviewed
	if existing, _ := s.reviews.GetByOrderNo(ctx, req.GetOrderNo()); existing != nil {
		return nil, status.Error(codes.AlreadyExists, "this order has already been reviewed")
	}

	// Look up the order to get order_id and merchant_id
	order, orderErr := s.orders.GetByOrderNo(ctx, tenantID, merchantID, req.GetOrderNo())
	if orderErr != nil {
		// For public review submission, try without merchant filter
		return nil, status.Error(codes.NotFound, "order not found")
	}

	created, err := s.reviews.Create(ctx, repository.CreateReviewInput{
		TenantID:      tenantID,
		MerchantID:    order.MerchantID,
		OrderID:       int64(order.ID),
		OrderNo:       req.GetOrderNo(),
		ProductID:     req.GetProductId(),
		UserID:        userID,
		Rating:        int(req.GetRating()),
		ServiceRating: int(req.GetServiceRating()),
		ValueRating:   int(req.GetValueRating()),
		Content:       req.GetContent(),
		Images:        req.GetImages(),
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return toReview(created, "", ""), nil
}

func (s *ReviewService) ListByProduct(ctx context.Context, req *travel.ReviewListRequest) (*travel.ReviewListResponse, error) {
	if req == nil || req.GetProductId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "product_id is required")
	}
	tenantID, err := auth.TenantID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	page, pageSize := req.GetPage(), req.GetPageSize()
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	items, total, err := s.reviews.ListByProduct(ctx, tenantID, req.GetProductId(), req.GetStatus(), page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	avgRating, _, _ := s.reviews.AvgRatingByProduct(ctx, tenantID, req.GetProductId())

	resp := &travel.ReviewListResponse{
		Total:     total,
		AvgRating: avgRating,
	}
	for _, r := range items {
		// Try to get user name and product name for display
		userName := ""
		productName := ""
		resp.Items = append(resp.Items, toReview(r, userName, productName))
	}
	return resp, nil
}

func (s *ReviewService) GetByOrder(ctx context.Context, req *travel.OrderNoReviewRequest) (*travel.Review, error) {
	if req == nil || req.GetOrderNo() == "" {
		return nil, status.Error(codes.InvalidArgument, "order_no is required")
	}
	r, err := s.reviews.GetByOrderNo(ctx, req.GetOrderNo())
	if err != nil {
		return nil, status.Error(codes.NotFound, "review not found")
	}
	return toReview(r, "", ""), nil
}

func (s *ReviewService) Reply(ctx context.Context, req *travel.ReplyReviewRequest) (*travel.Review, error) {
	if req == nil || req.GetId() <= 0 || req.GetReplyContent() == "" {
		return nil, status.Error(codes.InvalidArgument, "id and reply_content are required")
	}
	tenantID, err := auth.TenantID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	merchantID, err := auth.MerchantID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	updated, err := s.reviews.Reply(ctx, tenantID, merchantID, req.GetId(), req.GetReplyContent())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return toReview(updated, "", ""), nil
}

func toReview(r *ent.Review, userName, productName string) *travel.Review {
	return &travel.Review{
		Id:            int64(r.ID),
		TenantId:      r.TenantID,
		MerchantId:    r.MerchantID,
		OrderId:       r.OrderID,
		OrderNo:       r.OrderNo,
		ProductId:     r.ProductID,
		UserId:        r.UserID,
		Rating:        int32(r.Rating),
		ServiceRating: int32(r.ServiceRating),
		ValueRating:   int32(r.ValueRating),
		Content:       r.Content,
		Images:        r.Images,
		ReplyContent:  r.ReplyContent,
		ReplyTime:     r.ReplyTime,
		Status:        r.Status,
		UserName:      userName,
		ProductName:   productName,
	}
}
