package repository

import (
	"context"
	"time"

	"gitee.com/meinongyihe/travel-rpc/ent"
	"gitee.com/meinongyihe/travel-rpc/ent/packagetranslation"
	"gitee.com/meinongyihe/travel-rpc/ent/producttranslation"
)

type translationRepository struct {
	client *ent.Client
}

// NewTranslationRepository builds the ent-backed translation repository.
func NewTranslationRepository(client *ent.Client) TranslationRepository {
	return &translationRepository{client: client}
}

func (r *translationRepository) UpsertProduct(ctx context.Context, productID int64, locale, source, status string, f ProductTranslationFields) error {
	now := time.Now().Unix()
	existing, err := r.client.ProductTranslation.Query().
		Where(producttranslation.ProductIDEQ(productID), producttranslation.LocaleEQ(locale)).
		Only(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return err
	}
	if existing != nil {
		_, err = existing.Update().
			SetTitle(f.Title).SetDescription(f.Description).SetHighlights(f.Highlights).
			SetRichContent(f.RichContent).SetBookingNotice(f.BookingNotice).
			SetSource(source).SetStatus(status).SetUpdatedAt(now).
			Save(ctx)
		return err
	}
	_, err = r.client.ProductTranslation.Create().
		SetProductID(productID).SetLocale(locale).
		SetTitle(f.Title).SetDescription(f.Description).SetHighlights(f.Highlights).
		SetRichContent(f.RichContent).SetBookingNotice(f.BookingNotice).
		SetSource(source).SetStatus(status).SetUpdatedAt(now).
		Save(ctx)
	return err
}

func (r *translationRepository) MapProducts(ctx context.Context, productIDs []int64, locale string) (map[int64]*ent.ProductTranslation, error) {
	out := make(map[int64]*ent.ProductTranslation, len(productIDs))
	if len(productIDs) == 0 || locale == "" {
		return out, nil
	}
	rows, err := r.client.ProductTranslation.Query().
		Where(producttranslation.ProductIDIn(productIDs...), producttranslation.LocaleEQ(locale), producttranslation.StatusEQ("DONE")).
		All(ctx)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		out[row.ProductID] = row
	}
	return out, nil
}

func (r *translationRepository) UpsertPackage(ctx context.Context, packageID int64, locale, name, source, status string) error {
	now := time.Now().Unix()
	existing, err := r.client.PackageTranslation.Query().
		Where(packagetranslation.PackageIDEQ(packageID), packagetranslation.LocaleEQ(locale)).
		Only(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return err
	}
	if existing != nil {
		_, err = existing.Update().
			SetName(name).SetSource(source).SetStatus(status).SetUpdatedAt(now).
			Save(ctx)
		return err
	}
	_, err = r.client.PackageTranslation.Create().
		SetPackageID(packageID).SetLocale(locale).
		SetName(name).SetSource(source).SetStatus(status).SetUpdatedAt(now).
		Save(ctx)
	return err
}

func (r *translationRepository) MapPackages(ctx context.Context, packageIDs []int64, locale string) (map[int64]*ent.PackageTranslation, error) {
	out := make(map[int64]*ent.PackageTranslation, len(packageIDs))
	if len(packageIDs) == 0 || locale == "" {
		return out, nil
	}
	rows, err := r.client.PackageTranslation.Query().
		Where(packagetranslation.PackageIDIn(packageIDs...), packagetranslation.LocaleEQ(locale), packagetranslation.StatusEQ("DONE")).
		All(ctx)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		out[row.PackageID] = row
	}
	return out, nil
}
