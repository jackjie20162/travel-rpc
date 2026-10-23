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

//go:embed attraction_i18n_extra_seed.json
var attractionI18nExtraSeedRaw []byte

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

// attractionI18nExtraSeed 附加种子：开放时间/建议时长/门票信息的 en/ar 译文
//（后续扩展字段，与主种子按中文基础名合并）。
type attractionI18nExtraSeed struct {
	Items []attractionI18nExtraItem `json:"items"`
}

// attractionI18nExtraItem 单个景点的扩展字段译文。
type attractionI18nExtraItem struct {
	Name string `json:"name"` // 中文基础名（匹配主表 attractions.name）
	Hoen string `json:"hoen"` // 英语开放时间
	Hoar string `json:"hoar"` // 阿语开放时间
	Duen string `json:"duen"` // 英语建议游玩时长
	Duar string `json:"duar"` // 阿语建议游玩时长
	Tken string `json:"tken"` // 英语门票信息
	Tkar string `json:"tkar"` // 阿语门票信息
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
// （名称、分类、标签、地址、简介、开放时间、建议时长、门票信息）。
// 按中文基础名匹配主表记录；已存在的 (attraction_id, locale) 译文行
// 仅回填当前为空的字段（不覆盖运营在后台做过的编辑），新行直接创建，
// 因此可安全重复执行。返回本次新写入或回填的译文行数。
func SeedAttractionI18n(ctx context.Context, client *ent.Client) (int, error) {
	var seed attractionI18nSeed
	if err := json.Unmarshal(attractionI18nSeedRaw, &seed); err != nil {
		return 0, fmt.Errorf("parse attraction i18n seed: %w", err)
	}
	var extra attractionI18nExtraSeed
	if err := json.Unmarshal(attractionI18nExtraSeedRaw, &extra); err != nil {
		return 0, fmt.Errorf("parse attraction i18n extra seed: %w", err)
	}
	extraByName := make(map[string]attractionI18nExtraItem, len(extra.Items))
	for _, e := range extra.Items {
		extraByName[e.Name] = e
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
		ex := extraByName[item.Name]

		// en-US 译文：名称取主表 name_en，地址缺省沿用主表拉丁文。
		enName := a.NameEn
		enAddr := item.Aen
		if enAddr == "" {
			enAddr = a.Address
		}
		changed, err := seedAttractionTranslation(ctx, client, attrID, "en-US",
			enName, cat.En, marshalTags(item.Ten), enAddr, item.Den,
			ex.Hoen, ex.Duen, ex.Tken)
		if err != nil {
			return written, err
		}
		if changed {
			written++
		}

		// ar 译文：地址缺省沿用主表拉丁文（拉丁地址在阿联酋通用）。
		arAddr := item.Aar
		if arAddr == "" {
			arAddr = a.Address
		}
		changed, err = seedAttractionTranslation(ctx, client, attrID, "ar",
			item.Nar, cat.Ar, marshalTags(item.Tar), arAddr, item.Dar,
			ex.Hoar, ex.Duar, ex.Tkar)
		if err != nil {
			return written, err
		}
		if changed {
			written++
		}
	}

	if matched == 0 {
		return written, fmt.Errorf("attraction i18n seed matched 0 attractions (expected %d)", len(seed.Items))
	}
	return written, nil
}

// seedAttractionTranslation 幂等写入单条译文：不存在则创建（全空跳过）；
// 已存在则仅回填当前为空的字段，非空值（运营编辑过的）保持不变。
// 返回本次是否发生了写入或回填。
func seedAttractionTranslation(ctx context.Context, client *ent.Client, attrID int64,
	locale, name, category, tags, address, description, openingHours, duration, ticketInfo string) (bool, error) {
	existing, err := client.AttractionTranslation.Query().
		Where(attractiontranslation.AttractionIDEQ(attrID), attractiontranslation.LocaleEQ(locale)).
		First(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return false, err
	}
	if ent.IsNotFound(err) {
		if name == "" && category == "" && tags == "" && address == "" && description == "" &&
			openingHours == "" && duration == "" && ticketInfo == "" {
			return false, nil
		}
		_, err = client.AttractionTranslation.Create().
			SetAttractionID(attrID).
			SetLocale(locale).
			SetName(name).
			SetCategory(category).
			SetTags(tags).
			SetAddress(address).
			SetDescription(description).
			SetOpeningHours(openingHours).
			SetSuggestedDuration(duration).
			SetTicketInfo(ticketInfo).
			SetSource("MANUAL").
			SetStatus("DONE").
			SetUpdatedAt(time.Now().Unix()).
			Save(ctx)
		return err == nil, err
	}

	// 已存在：只补空字段，不覆盖已有内容
	upd := client.AttractionTranslation.UpdateOne(existing)
	changed := false
	if existing.Name == "" && name != "" {
		upd.SetName(name)
		changed = true
	}
	if existing.Category == "" && category != "" {
		upd.SetCategory(category)
		changed = true
	}
	if existing.Tags == "" && tags != "" {
		upd.SetTags(tags)
		changed = true
	}
	if existing.Address == "" && address != "" {
		upd.SetAddress(address)
		changed = true
	}
	if existing.Description == "" && description != "" {
		upd.SetDescription(description)
		changed = true
	}
	if existing.OpeningHours == "" && openingHours != "" {
		upd.SetOpeningHours(openingHours)
		changed = true
	}
	if existing.SuggestedDuration == "" && duration != "" {
		upd.SetSuggestedDuration(duration)
		changed = true
	}
	if existing.TicketInfo == "" && ticketInfo != "" {
		upd.SetTicketInfo(ticketInfo)
		changed = true
	}
	if !changed {
		return false, nil
	}
	_, err = upd.SetUpdatedAt(time.Now().Unix()).Save(ctx)
	return err == nil, err
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

