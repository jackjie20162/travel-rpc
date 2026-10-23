package repository

import (
	"context"
	"time"

	"gitee.com/meinongyihe/travel-rpc/ent"
	"gitee.com/meinongyihe/travel-rpc/ent/attraction"
	"gitee.com/meinongyihe/travel-rpc/ent/region"
	"gitee.com/meinongyihe/travel-rpc/ent/regiontranslation"
)

// citySeed 城市种子：中/英/阿三语名 + 中心经纬度（WGS84）。
// poi_id/poi_name 不在此硬编码：不同地图供应商(高德/Google)的 POI ID 无法臆造，
// 由后台「区域管理 → 地图选点」补齐代表 POI。
type citySeed struct {
	zh, en, ar string
	lat, lng   float64
}

// countrySeed 国家种子：ISO 代码 + 中/英/阿三语名 + 若干常用城市。
type countrySeed struct {
	code       string
	zh, en, ar string
	cities     []citySeed
}

// defaultRegionSeeds 覆盖中东海湾国家(本业务主战场)与常用发达国家/客源地。
var defaultRegionSeeds = []countrySeed{
	// ── 海湾 / 中东 ──
	{code: "AE", zh: "阿联酋", en: "United Arab Emirates", ar: "الإمارات العربية المتحدة", cities: []citySeed{
		{"迪拜", "Dubai", "دبي", 25.2048, 55.2708},
		{"阿布扎比", "Abu Dhabi", "أبوظبي", 24.4539, 54.3773},
		{"沙迦", "Sharjah", "الشارقة", 25.3463, 55.4209},
		{"拉斯海玛", "Ras Al Khaimah", "رأس الخيمة", 25.7895, 55.9432},
	}},
	{code: "OM", zh: "阿曼", en: "Oman", ar: "عُمان", cities: []citySeed{
		{"马斯喀特", "Muscat", "مسقط", 23.5880, 58.3829},
		{"萨拉拉", "Salalah", "صلالة", 17.0151, 54.0924},
	}},
	{code: "QA", zh: "卡塔尔", en: "Qatar", ar: "قطر", cities: []citySeed{
		{"多哈", "Doha", "الدوحة", 25.2854, 51.5310},
	}},
	{code: "SA", zh: "沙特阿拉伯", en: "Saudi Arabia", ar: "المملكة العربية السعودية", cities: []citySeed{
		{"利雅得", "Riyadh", "الرياض", 24.7136, 46.6753},
		{"吉达", "Jeddah", "جدة", 21.4858, 39.1925},
		{"麦加", "Mecca", "مكة المكرمة", 21.3891, 39.8579},
		{"麦地那", "Medina", "المدينة المنورة", 24.5247, 39.5692},
	}},
	{code: "BH", zh: "巴林", en: "Bahrain", ar: "البحرين", cities: []citySeed{
		{"麦纳麦", "Manama", "المنامة", 26.2285, 50.5860},
	}},
	{code: "KW", zh: "科威特", en: "Kuwait", ar: "الكويت", cities: []citySeed{
		{"科威特城", "Kuwait City", "مدينة الكويت", 29.3759, 47.9774},
	}},

	// ── 亚洲客源地 / 发达国家 ──
	{code: "CN", zh: "中国", en: "China", ar: "الصين", cities: []citySeed{
		{"北京", "Beijing", "بكين", 39.9042, 116.4074},
		{"上海", "Shanghai", "شنغهاي", 31.2304, 121.4737},
		{"广州", "Guangzhou", "غوانغجو", 23.1291, 113.2644},
		{"成都", "Chengdu", "تشنغدو", 30.5728, 104.0668},
	}},
	{code: "JP", zh: "日本", en: "Japan", ar: "اليابان", cities: []citySeed{
		{"东京", "Tokyo", "طوكيو", 35.6762, 139.6503},
		{"大阪", "Osaka", "أوساكا", 34.6937, 135.5023},
		{"京都", "Kyoto", "كيوتو", 35.0116, 135.7681},
	}},
	{code: "KR", zh: "韩国", en: "South Korea", ar: "كوريا الجنوبية", cities: []citySeed{
		{"首尔", "Seoul", "سيول", 37.5665, 126.9780},
		{"釜山", "Busan", "بوسان", 35.1796, 129.0756},
	}},
	{code: "SG", zh: "新加坡", en: "Singapore", ar: "سنغافورة", cities: []citySeed{
		{"新加坡", "Singapore", "سنغافورة", 1.3521, 103.8198},
	}},
	{code: "TH", zh: "泰国", en: "Thailand", ar: "تايلاند", cities: []citySeed{
		{"曼谷", "Bangkok", "بانكوك", 13.7563, 100.5018},
		{"普吉岛", "Phuket", "فوكيت", 7.8804, 98.3923},
	}},
	{code: "MY", zh: "马来西亚", en: "Malaysia", ar: "ماليزيا", cities: []citySeed{
		{"吉隆坡", "Kuala Lumpur", "كوالالمبور", 3.1390, 101.6869},
	}},

	// ── 欧洲发达国家 ──
	{code: "GB", zh: "英国", en: "United Kingdom", ar: "المملكة المتحدة", cities: []citySeed{
		{"伦敦", "London", "لندن", 51.5074, -0.1278},
		{"曼彻斯特", "Manchester", "مانشستر", 53.4808, -2.2426},
	}},
	{code: "FR", zh: "法国", en: "France", ar: "فرنسا", cities: []citySeed{
		{"巴黎", "Paris", "باريس", 48.8566, 2.3522},
		{"尼斯", "Nice", "نيس", 43.7102, 7.2620},
	}},
	{code: "DE", zh: "德国", en: "Germany", ar: "ألمانيا", cities: []citySeed{
		{"柏林", "Berlin", "برلين", 52.5200, 13.4050},
		{"慕尼黑", "Munich", "ميونخ", 48.1351, 11.5820},
	}},
	{code: "IT", zh: "意大利", en: "Italy", ar: "إيطاليا", cities: []citySeed{
		{"罗马", "Rome", "روما", 41.9028, 12.4964},
		{"米兰", "Milan", "ميلانو", 45.4642, 9.1900},
		{"威尼斯", "Venice", "البندقية", 45.4408, 12.3155},
	}},
	{code: "ES", zh: "西班牙", en: "Spain", ar: "إسبانيا", cities: []citySeed{
		{"马德里", "Madrid", "مدريد", 40.4168, -3.7038},
		{"巴塞罗那", "Barcelona", "برشلونة", 41.3851, 2.1734},
	}},
	{code: "CH", zh: "瑞士", en: "Switzerland", ar: "سويسرا", cities: []citySeed{
		{"苏黎世", "Zurich", "زيورخ", 47.3769, 8.5417},
		{"日内瓦", "Geneva", "جنيف", 46.2044, 6.1432},
	}},
	{code: "NL", zh: "荷兰", en: "Netherlands", ar: "هولندا", cities: []citySeed{
		{"阿姆斯特丹", "Amsterdam", "أمستردام", 52.3676, 4.9041},
	}},
	{code: "RU", zh: "俄罗斯", en: "Russia", ar: "روسيا", cities: []citySeed{
		{"莫斯科", "Moscow", "موسكو", 55.7558, 37.6173},
		{"圣彼得堡", "Saint Petersburg", "سانت بطرسبرغ", 59.9311, 30.3609},
	}},
	{code: "TR", zh: "土耳其", en: "Turkey", ar: "تركيا", cities: []citySeed{
		{"伊斯坦布尔", "Istanbul", "إسطنبول", 41.0082, 28.9784},
		{"安塔利亚", "Antalya", "أنطاليا", 36.8969, 30.7133},
	}},

	// ── 美洲 / 大洋洲发达国家 ──
	{code: "US", zh: "美国", en: "United States", ar: "الولايات المتحدة", cities: []citySeed{
		{"纽约", "New York", "نيويورك", 40.7128, -74.0060},
		{"洛杉矶", "Los Angeles", "لوس أنجلوس", 34.0522, -118.2437},
		{"拉斯维加斯", "Las Vegas", "لاس فيغاس", 36.1699, -115.1398},
	}},
	{code: "CA", zh: "加拿大", en: "Canada", ar: "كندا", cities: []citySeed{
		{"多伦多", "Toronto", "تورونتو", 43.6532, -79.3832},
		{"温哥华", "Vancouver", "فانكوفر", 49.2827, -123.1207},
	}},
	{code: "AU", zh: "澳大利亚", en: "Australia", ar: "أستراليا", cities: []citySeed{
		{"悉尼", "Sydney", "سيدني", -33.8688, 151.2093},
		{"墨尔本", "Melbourne", "ملبورن", -37.8136, 144.9631},
	}},
}

// SeedDefaultRegions 启动时幂等写入常用国家与城市(含中/英/阿三语)。
// 已存在的区域按 (parent_id, name, level) 命中即跳过，译文按 (region_id, locale) 命中即跳过，
// 因此可安全重复执行，不覆盖运营在后台做过的编辑。
func SeedDefaultRegions(ctx context.Context, client *ent.Client) error {
	for _, c := range defaultRegionSeeds {
		countryID, err := seedRegion(ctx, client, 0, "COUNTRY", c.code, c.zh, 0, 0)
		if err != nil {
			return err
		}
		if err := seedRegionTranslations(ctx, client, countryID, c.en, c.ar); err != nil {
			return err
		}
		for _, city := range c.cities {
			cityID, err := seedRegion(ctx, client, countryID, "CITY", "", city.zh, city.lat, city.lng)
			if err != nil {
				return err
			}
			if err := seedRegionTranslations(ctx, client, cityID, city.en, city.ar); err != nil {
				return err
			}
		}
	}
	return nil
}

// seedRegion 按 (parent_id, name, level) 幂等写入一个区域，返回其 id。
func seedRegion(ctx context.Context, client *ent.Client, parentID int64, level, code, name string, lat, lng float64) (int64, error) {
	existing, err := client.Region.Query().
		Where(region.ParentIDEQ(parentID), region.NameEQ(name), region.LevelEQ(level)).
		Only(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return 0, err
	}
	if existing != nil {
		return int64(existing.ID), nil
	}

	create := client.Region.Create().
		SetParentID(parentID).
		SetLevel(level).
		SetName(name).
		SetStatus("ACTIVE")
	if code != "" {
		create.SetCode(code)
	}
	if lat != 0 || lng != 0 {
		create.SetLatitude(lat).SetLongitude(lng)
	}
	created, err := create.Save(ctx)
	if err != nil {
		return 0, err
	}
	return int64(created.ID), nil
}

// seedRegionTranslations 幂等写入 en-US / ar 两种译文（基础名 zh-CN 已在 regions.name）。
func seedRegionTranslations(ctx context.Context, client *ent.Client, regionID int64, enName, arName string) error {
	now := time.Now().Unix()
	pairs := []struct {
		locale string
		name   string
	}{{"en-US", enName}, {"ar", arName}}
	for _, p := range pairs {
		if p.name == "" {
			continue
		}
		exists, err := client.RegionTranslation.Query().
			Where(regiontranslation.RegionIDEQ(regionID), regiontranslation.LocaleEQ(p.locale)).
			Exist(ctx)
		if err != nil {
			return err
		}
		if exists {
			continue
		}
		if _, err := client.RegionTranslation.Create().
			SetRegionID(regionID).
			SetLocale(p.locale).
			SetName(p.name).
			SetSource("MANUAL").
			SetStatus("DONE").
			SetUpdatedAt(now).
			Save(ctx); err != nil {
			return err
		}
	}
	return nil
}

// BackfillAttractionCityIDs 幂等回填存量景点的 city_id：
// 按 attraction.city（中文展示名）匹配 regions 中 level=CITY 的城市名，命中即写入 city_id。
// 仅处理 city_id=0 且 city 非空的景点，已归属的不重复改动，因此可安全在每次启动执行。
// 返回本次成功归属的景点数。
func BackfillAttractionCityIDs(ctx context.Context, client *ent.Client) (int, error) {
	cities, err := client.Region.Query().
		Where(region.LevelEQ("CITY")).
		All(ctx)
	if err != nil {
		return 0, err
	}
	byName := make(map[string]int64, len(cities))
	for _, c := range cities {
		byName[c.Name] = int64(c.ID)
	}

	pending, err := client.Attraction.Query().
		Where(attraction.CityIDEQ(0), attraction.CityNEQ("")).
		All(ctx)
	if err != nil {
		return 0, err
	}

	updated := 0
	for _, a := range pending {
		regionID, ok := byName[a.City]
		if !ok {
			continue
		}
		if err := client.Attraction.UpdateOneID(a.ID).SetCityID(regionID).Exec(ctx); err != nil {
			return updated, err
		}
		updated++
	}
	return updated, nil
}
