package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"gitee.com/meinongyihe/travel-rpc/ent"
	"gitee.com/meinongyihe/travel-rpc/internal/config"
	"gitee.com/meinongyihe/travel-rpc/internal/repository"

	"github.com/zeromicro/go-zero/core/logx"
)

// TranslationService fills the product/package translation tables by calling an
// external machine-translation provider (DeepL-compatible). It is invoked
// asynchronously after a merchant creates or updates a product/package.
//
// Graceful degradation: when no provider/api-key is configured the service is a
// no-op, leaving the catalog on base-language content. This keeps the system
// fully functional without external credentials (see 多语言多货币改造方案 P4).
type TranslationService struct {
	client *ent.Client
	repo   repository.TranslationRepository
	conf   config.TranslateConf
	http   *http.Client
}

// NewTranslationService builds the translation service. The api key may come
// from config or, preferably, the TRANSLATE_API_KEY environment variable so it
// is never hardcoded in yaml.
func NewTranslationService(client *ent.Client, repo repository.TranslationRepository, conf config.TranslateConf) *TranslationService {
	if strings.TrimSpace(conf.ApiKey) == "" {
		conf.ApiKey = strings.TrimSpace(os.Getenv("TRANSLATE_API_KEY"))
	}
	if len(conf.TargetLangs) == 0 {
		conf.TargetLangs = []string{"zh-CN", "en-US", "ar"}
	}
	if strings.TrimSpace(conf.SourceLang) == "" {
		conf.SourceLang = "auto"
	}
	return &TranslationService{
		client: client,
		repo:   repo,
		conf:   conf,
		http:   &http.Client{Timeout: 20 * time.Second},
	}
}

// Enabled reports whether an external provider is configured and usable.
func (s *TranslationService) Enabled() bool {
	return strings.TrimSpace(s.conf.Provider) != "" &&
		strings.TrimSpace(s.conf.Endpoint) != "" &&
		strings.TrimSpace(s.conf.ApiKey) != ""
}

// TranslateProduct translates the product's base-language text into every
// configured target locale and upserts the translation rows. A per-locale
// failure is recorded as status=FAILED and never aborts the whole run.
func (s *TranslationService) TranslateProduct(ctx context.Context, productID int64) error {
	if !s.Enabled() {
		logx.Infof("[translation] product %d skipped: provider not configured", productID)
		return nil
	}
	p, err := s.client.Product.Get(ctx, int(productID))
	if err != nil {
		return err
	}
	src := repository.ProductTranslationFields{
		Title:         p.Title,
		Description:   p.Description,
		Highlights:    p.Highlights,
		RichContent:   p.RichContent,
		BookingNotice: p.BookingNotice,
	}
	for _, locale := range s.conf.TargetLangs {
		fields, ok := s.translateProductFields(ctx, src, locale)
		status := "DONE"
		if !ok {
			status = "FAILED"
		}
		if err := s.repo.UpsertProduct(ctx, productID, locale, "MACHINE", status, fields); err != nil {
			logx.Errorf("[translation] upsert product %d locale %s failed: %v", productID, locale, err)
		}
	}
	return nil
}

// TranslatePackage translates a package name into every configured target locale.
func (s *TranslationService) TranslatePackage(ctx context.Context, packageID int64) error {
	if !s.Enabled() {
		logx.Infof("[translation] package %d skipped: provider not configured", packageID)
		return nil
	}
	pkg, err := s.client.ProductPackage.Get(ctx, int(packageID))
	if err != nil {
		return err
	}
	for _, locale := range s.conf.TargetLangs {
		name, err := s.translateText(ctx, pkg.Name, locale)
		status := "DONE"
		if err != nil {
			status = "FAILED"
			name = pkg.Name
		}
		if uerr := s.repo.UpsertPackage(ctx, packageID, locale, name, "MACHINE", status); uerr != nil {
			logx.Errorf("[translation] upsert package %d locale %s failed: %v", packageID, locale, uerr)
		}
	}
	return nil
}

func (s *TranslationService) translateProductFields(ctx context.Context, src repository.ProductTranslationFields, locale string) (repository.ProductTranslationFields, bool) {
	out := src
	ok := true
	translate := func(text string, html bool) string {
		if strings.TrimSpace(text) == "" {
			return text
		}
		var (
			v   string
			err error
		)
		if html {
			v, err = s.translateHTML(ctx, text, locale)
		} else {
			v, err = s.translateText(ctx, text, locale)
		}
		if err != nil {
			ok = false
			return text
		}
		return v
	}
	out.Title = translate(src.Title, false)
	out.Description = translate(src.Description, false)
	out.BookingNotice = translate(src.BookingNotice, false)
	out.RichContent = translate(src.RichContent, true)
	out.Highlights = s.translateHighlights(ctx, src.Highlights, locale, &ok)
	return out, ok
}

// translateHighlights parses the highlights JSON ([{"text":"..."}]), translates
// each text value and re-marshals. Unparseable content is translated as plain
// text so nothing is silently dropped.
func (s *TranslationService) translateHighlights(ctx context.Context, raw, locale string, ok *bool) string {
	if strings.TrimSpace(raw) == "" {
		return raw
	}
	var items []map[string]any
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		v, terr := s.translateText(ctx, raw, locale)
		if terr != nil {
			*ok = false
			return raw
		}
		return v
	}
	for _, item := range items {
		t, found := item["text"]
		if !found {
			continue
		}
		ts, isStr := t.(string)
		if !isStr || strings.TrimSpace(ts) == "" {
			continue
		}
		v, err := s.translateText(ctx, ts, locale)
		if err != nil {
			*ok = false
			continue
		}
		item["text"] = v
	}
	out, err := json.Marshal(items)
	if err != nil {
		*ok = false
		return raw
	}
	return string(out)
}

// translateText translates plain text via the configured provider.
func (s *TranslationService) translateText(ctx context.Context, text, locale string) (string, error) {
	if strings.TrimSpace(text) == "" {
		return text, nil
	}
	switch strings.ToLower(strings.TrimSpace(s.conf.Provider)) {
	case "deepl":
		return s.translateDeepL(ctx, text, locale, "")
	default:
		return "", fmt.Errorf("unsupported translate provider: %q", s.conf.Provider)
	}
}

// translateHTML translates markup, preserving tags where the provider supports it.
func (s *TranslationService) translateHTML(ctx context.Context, htmlText, locale string) (string, error) {
	if strings.TrimSpace(htmlText) == "" {
		return htmlText, nil
	}
	switch strings.ToLower(strings.TrimSpace(s.conf.Provider)) {
	case "deepl":
		return s.translateDeepL(ctx, htmlText, locale, "html")
	default:
		return "", fmt.Errorf("unsupported translate provider: %q", s.conf.Provider)
	}
}

// translateDeepL calls a DeepL-compatible /translate endpoint. tagHandling may be
// "" (plain) or "html"/"xml" to preserve markup.
func (s *TranslationService) translateDeepL(ctx context.Context, text, locale, tagHandling string) (string, error) {
	form := url.Values{}
	form.Set("text", text)
	form.Set("target_lang", deeplLang(locale))
	if src := strings.TrimSpace(s.conf.SourceLang); src != "" && strings.ToLower(src) != "auto" {
		form.Set("source_lang", deeplLang(src))
	}
	if tagHandling != "" {
		form.Set("tag_handling", tagHandling)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.conf.Endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "DeepL-Auth-Key "+s.conf.ApiKey)
	resp, err := s.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("deepl http %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var out struct {
		Translations []struct {
			DetectedSourceLanguage string `json:"detected_source_language"`
			Text                   string `json:"text"`
		} `json:"translations"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return "", err
	}
	if len(out.Translations) == 0 {
		return "", fmt.Errorf("deepl returned no translations")
	}
	return out.Translations[0].Text, nil
}

// deeplLang maps an app locale to a DeepL language code.
func deeplLang(locale string) string {
	l := strings.ToLower(strings.TrimSpace(locale))
	switch {
	case strings.HasPrefix(l, "zh"):
		return "ZH"
	case strings.HasPrefix(l, "ar"):
		return "AR"
	case strings.HasPrefix(l, "en"):
		return "EN"
	case strings.HasPrefix(l, "fr"):
		return "FR"
	case strings.HasPrefix(l, "de"):
		return "DE"
	case strings.HasPrefix(l, "es"):
		return "ES"
	default:
		return strings.ToUpper(l)
	}
}
