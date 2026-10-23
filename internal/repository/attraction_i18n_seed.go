package repository

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"time"

	"gitee.com/meinongyihe/travel-rpc/ent"
	"gitee.com/meinongyihe/travel-rpc/ent/attractiontranslation"
)

//go:embed attraction_i18n_seed.json
var attractionI18nSeedRaw []byte

// attractionI18nSeed 是景点 en-US/ar 译文种子文件反序列化结构。
// 基础语言 zh-CN 存在 attractions 主表（name/address/description/tags），
// 此处仅覆盖两种目标语言：名称(阿)、分类(本地化显示名)、标签、地址、简介。
type attractionI18nSeed struct {
	// Categories 按枚举 code 提供 en/ar 本地化分类显示名（主表 category 仍是 code 供筛选）。
	Categories map[string]struct {
		En string `json:"en"`
		Ar string `json:"ar"`
	} `json:"categories"`
	Items []attractionI18nItem `json:"items"`
}

// attractionI18nItem 单个景点的译文数据，以中文基础名 name 与主表匹配。
type attractionI18nItem struct {
	Name string   `json:"name"` // 中文基础名（匹配主表 attractions.name）
	Nar  string   `json:"nar"`  // 阿拉伯语名称
	Ten  []string `json:"ten"`  // 英语标签
	Tar  []string `json:"tar"`  // 阿拉伯语标签
	Den  string   `json:"den"`  // 英语简介
	Dar  string   `json:"dar"`  // 阿拉伯语简介
	Aen  string   `json:"aen"`  // 英语地址（选填，缺省沿用主表拉丁文地址）
	Aar  string   `json:"aar"`  // 阿拉伯语地址（选填，缺省沿用主表拉丁文地址）
}

// SeedAttractionI18n 启动时幂等为存量景点补齐 en-US/ar 两语译文
// （名称、分类、标签、地址、简介）。按中文基础名匹配主表记录，
// 已存在的 (attraction_id, locale) 译文跳过，不覆盖运营在后台做过的编辑，
// 因此可安全重复执行。返回本次新写入的译文行数。
func SeedAttractionI18n(ctx context.Context, client *ent.Client) (int, error) {
	var seed attractionI18nSeed
	if err := json.Unmarshal(attractionI18nSeedRaw, &seed); err != nil {
		return 0, fmt.Errorf("parse attraction i18n seed: %w", err)
	}

	// 主表景点按中文名建索引，供种子匹配。
	attrs, err := client.Attraction.Query().All(ctx)
	if err != nil {
		return 0, err
	}
	byName := make(map[string]*ent.Attraction, len(attrs))
	for _, a := range attrs {
		byName[a.Name] = a
	}

	written := 0
	matched := 0
	for _, item := range seed.Items {
		a, ok := byName[item.Name]
		if !ok {
			continue // 主表无此景点（可能已删除），跳过
		}
		matched++
		attrID := int64(a.ID)

		cat := seed.Categories[a.Category]

		// en-US 译文：名称取主表 name_en，地址缺省沿用主表拉丁文。
		enName := a.NameEn
		enAddr := item.Aen
		if enAddr == "" {
			enAddr = a.Address
		}
		if err := seedAttractionTranslation(ctx, client, attrID, "en-US",
			enName, cat.En, marshalTags(item.Ten), enAddr, item.Den); err != nil {
			return written, err
		}
		written++

		// ar 译文：地址缺省沿用主表拉丁文（拉丁地址在阿联酋通用）。
		arAddr := item.Aar
		if arAddr == "" {
			arAddr = a.Address
		}
		if err := seedAttractionTranslation(ctx, client, attrID, "ar",
			item.Nar, cat.Ar, marshalTags(item.Tar), arAddr, item.Dar); err != nil {
			return written, err
		}
		written++
	}

	if matched == 0 {
		return written, fmt.Errorf("attraction i18n seed matched 0 attractions (expected %d)", len(seed.Items))
	}
	return written, nil
}

// seedAttractionTranslation 幂等写入单条译文：已存在 (attractionID, locale) 则跳过；
// 五个字段全空则不写入空行。
func seedAttractionTranslation(ctx context.Context, client *ent.Client, attrID int64,
	locale, name, category, tags, address, description string) error {
	exists, err := client.AttractionTranslation.Query().
		Where(attractiontranslation.AttractionIDEQ(attrID), attractiontranslation.LocaleEQ(locale)).
		Exist(ctx)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	if name == "" && category == "" && tags == "" && address == "" && description == "" {
		return nil
	}
	_, err = client.AttractionTranslation.Create().
		SetAttractionID(attrID).
		SetLocale(locale).
		SetName(name).
		SetCategory(category).
		SetTags(tags).
		SetAddress(address).
		SetDescription(description).
		SetSource("MANUAL").
		SetStatus("DONE").
		SetUpdatedAt(time.Now().Unix()).
		Save(ctx)
	return err
}

// marshalTags 将标签数组序列化为 JSON 字符串；空数组返回空串（跳过写入）。
func marshalTags(tags []string) string {
	if len(tags) == 0 {
		return ""
	}
	b, err := json.Marshal(tags)
	if err != nil {
		return ""
	}
	return string(b)
}

