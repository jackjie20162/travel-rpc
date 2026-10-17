-- 迪拜景点 ↔ 线路 关联绑定（演示/修复用）
--
-- 背景：景点与线路（产品）的绑定关系记录在 itinerary_stops.type_params JSON 的 attractionId 字段，
-- 由商户后台「旅游插件 → 产品 → 行程编排」在活动类节点上选择景点写入。
-- 产品 1（Musandam 海豚湾一日游，PUBLISHED）唯一的活动节点 id=6 原先绑定 attractionId=1，
-- 而景点 1（早期测试数据）已被删除，属于悬空引用；这里改指向真实景点，
-- 使 C 端「景点详情 → 包含该景点的线路」有可见数据。
--
-- 可重复执行；如需更多景点关联同一线路，请在商户后台行程编排里为其它节点选择景点。

UPDATE itinerary_stops
SET type_params = '{"attractionId":6}'
WHERE id = 6 AND product_id = 1;

-- 校验：应返回 1 行（景点 6 朱美拉棕榈岛 ← 产品 1）
SELECT s.id AS stop_id, s.product_id, s.type_params, a.name, a.name_en
FROM itinerary_stops s
JOIN attractions a ON a.id = CAST(REPLACE(REPLACE(s.type_params, '{"attractionId":', ''), '}', '') AS UNSIGNED)
WHERE s.id = 6;
