package auth

import (
	"context"
	"fmt"
	"strconv"

	"google.golang.org/grpc/metadata"
)

const (
	TenantIDKey  = "x-tenant-id"
	MerchantIDKey = "x-merchant-id"
	CustomerIDKey = "x-customer-id"
)

// TenantID extracts the authenticated tenant scope propagated by the API gateway.
// The RPC layer deliberately does not accept tenant_id from business requests.
func TenantID(ctx context.Context) (int64, error) {
	return metadataID(ctx, TenantIDKey)
}

// MerchantID extracts the authenticated merchant scope propagated by the API gateway.
func MerchantID(ctx context.Context) (int64, error) {
	return metadataID(ctx, MerchantIDKey)
}

// CustomerID is optional and is used when an authenticated customer is available.
func CustomerID(ctx context.Context) (*int64, error) {
	id, err := metadataID(ctx, CustomerIDKey)
	if err != nil {
		if err == errMissingMetadata {
			return nil, nil
		}
		return nil, err
	}
	return &id, nil
}

var errMissingMetadata = fmt.Errorf("missing authenticated scope metadata")

func metadataID(ctx context.Context, key string) (int64, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, errMissingMetadata
	}
	values := md.Get(key)
	if len(values) == 0 || values[0] == "" {
		return 0, errMissingMetadata
	}
	id, err := strconv.ParseInt(values[0], 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid %s", key)
	}
	return id, nil
}
