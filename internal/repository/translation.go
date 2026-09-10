package repository

import (
	"context"

	"gitee.com/meinongyihe/travel-rpc/ent"
)

// ProductTranslationFields carries the translatable text of a product.
type ProductTranslationFields struct {
	Title         string
	Description   string
	Highlights    string
	RichContent   string
	BookingNotice string
}

// TranslationRepository persists locale-specific translations of products and
// packages. The base entities remain the source of truth; translation rows are
// derived content keyed by (entityID, locale). Reads only surface DONE rows so
// in-progress/failed machine translations never leak to the catalog.
type TranslationRepository interface {
	// UpsertProduct writes or updates the translation of a product for a locale.
	UpsertProduct(ctx context.Context, productID int64, locale, source, status string, f ProductTranslationFields) error
	// MapProducts returns DONE translations for many products in one locale, keyed by product ID.
	MapProducts(ctx context.Context, productIDs []int64, locale string) (map[int64]*ent.ProductTranslation, error)
	// UpsertPackage writes or updates the translation of a package for a locale.
	UpsertPackage(ctx context.Context, packageID int64, locale, name, source, status string) error
	// MapPackages returns DONE translations for many packages in one locale, keyed by package ID.
	MapPackages(ctx context.Context, packageIDs []int64, locale string) (map[int64]*ent.PackageTranslation, error)
}
