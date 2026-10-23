# Travel RPC — Data Model

## Scope

The first executable data layer covers tenant, merchant, product, package/SKU, inventory, order, traveler, voucher, and itinerary stops.

## Isolation

All business records carry `tenant_id` and `merchant_id` where applicable. Tenant and merchant scope must come from authenticated service context, not from public request fields.

## Uniqueness

- Tenant: `code`
- Merchant: `(tenant_id, code)`
- Product: `(tenant_id, code)`
- Package: `(tenant_id, product_id, code)`
- Inventory: `(package_id, service_date, time_slot)`
- Order: `order_no`
- Voucher: `voucher_no`
- ItineraryStop: `(product_id, sequence)` — ordered within a product

## Inventory rule

`available = status OPEN AND capacity - reserved >= requested quantity`.

Reservation must be atomic inside a database transaction. The service must never read a remaining quantity and then perform an independent update that can oversell under concurrency.

## Money

Amounts are integer minor units. Currency is stored explicitly. Client-supplied total amounts are not authoritative; order totals are calculated from the inventory/package price held by the RPC data layer.

## Time

`service_date` is an ISO date and `time_slot` is a normalized business slot string. A later migration may replace these strings with stronger temporal types after timezone policy is finalized.

## Current implementation status

The Ent schema layer is being hardened before repositories/services are added. Schema source is authoritative; generated Ent code is produced by the CI toolchain and must not be hand-written.

## Itinerary Stop model

The `ItineraryStop` schema represents a segment/stop in a product's itinerary. Each product can have multiple stops ordered by `sequence`.

### Stop types

- `MEETING` — 集合点（唯一，必须为第一个节点；分上门接/集合点两种模式）
- `ACTIVITY` — 活动景点（地点和活动）
- `TRANSPORT` — 交通路段（行中交通）
- `MEAL` — 餐饮安排（行中餐食）
- `RETURN` — 返程（唯一，必须为最后一个节点；送回服务/解散点/自由解散）

结构约束由商户后台编辑器保证：首节点 MEETING、尾节点 RETURN 各只有一个，中间节点仅允许 ACTIVITY/TRANSPORT/MEAL 插入。

### Fields

| Field | Type | Description |
|-------|------|-------------|
| tenant_id | int64 | Tenant isolation |
| merchant_id | int64 | Merchant isolation |
| product_id | int64 | Parent product |
| stop_type | string | MEETING/ACTIVITY/TRANSPORT/MEAL/RETURN |
| title | string | Stop title |
| description | text? | Stop description |
| sequence | int | Ordering within product |
| location_name | string? | Location name |
| address | string? | Address |
| latitude | float? | Latitude |
| longitude | float? | Longitude |
| poi_id | string? | Amap POI ID |
| poi_name | string? | POI name |
| is_entering | bool | Whether entering the venue |
| duration_mode | string? | FIXED/PER_PACKAGE/UNLIMITED |
| duration_hours | int? | Duration hours |
| duration_minutes | int? | Duration minutes |
| activity_features | text? | Activity features description |
| transport_type | string? | Vehicle type (car/boat/walk) |
| start_time | string? | Start time (HH:mm) |
| end_time | string? | End time (HH:mm) |
| pickup_location | string? | Pickup location name |
| pickup_address | string? | Pickup address |
| pickup_latitude | float? | Pickup latitude |
| pickup_longitude | float? | Pickup longitude |
| dropoff_location | string? | Dropoff location name |
| dropoff_address | string? | Dropoff address |
| dropoff_latitude | float? | Dropoff latitude |
| dropoff_longitude | float? | Dropoff longitude |
| is_pickup | bool | Has pickup service |
| is_dropoff | bool | Has dropoff service |
| agreement_no_shopping | bool | No shopping commitment |
| agreement_adjustable | bool | Itinerary adjustable |
| type_params | text? | 节点类型专属参数 JSON（集合方式/集合点/餐型/送回行/解散点等） |

### type_params JSON 结构

- MEETING: `{meetingMode: PICKUP|POINT, meetingTime, meetingCity, meetingPoint, meetingDesc, meetingImages, meetingLat, meetingLng, pickupTime, pickupCity, pickupDistrict, pickupRangeMode, pickupPolygon[], pickupRangeOptions{}, pickupNote}`
- MEAL: `{mealType: BREAKFAST|LUNCH|DINNER|AFTERNOON_TEA|LATE_NIGHT|MORNING_TEA}`
- RETURN: `{dropoffService, dropoffPoints[{city,time,rangeMode,polygon[],rangeOptions{},note}], dispersalService, dispersalPoints[{time,city,pointName,lat,lng,images,desc}], freeDispersal}`
- ACTIVITY/TRANSPORT: 复用公共列（poi/duration/transport_type/activity_features），type_params 为空对象

### Indexes

- `(tenant_id, product_id)` — tenant-scoped lookup
- `(product_id, sequence)` — ordered stops within product

## User model

The `User` schema stores end-user (customer) accounts. It mirrors the legacy PHP user table for data compatibility.

### Uniqueness

- User: `username` (unique)

### Authentication

- Password is stored as `sha256(sha256(password) + salt)` with a random 16-byte salt.
- Token is a random 32-byte hex string used for Bearer authentication.
- Status: `normal` (active) or `hidden` (disabled).

### Fields

| Field | Type | Description |
|-------|------|-------------|
| username | string | Unique login name |
| nickname | string | Display name |
| password | string | Hashed password |
| salt | string | Random salt for password hashing |
| email | string | Email address |
| mobile | string | Phone number |
| avatar | string | Avatar URL |
| level | uint8 | User level/tier |
| gender | int8 | 0=unknown, 1=male, 2=female |
| birthday | string? | Date of birth |
| bio | string | Short biography |
| money | float | Account balance |
| score | int | Loyalty score |
| token | string | Auth bearer token |
| status | string | Account status |
| jointime | int64? | Registration timestamp |
| logintime | int64 | Last login timestamp |
| loginip | string | Last login IP |

## Region model（国家 → 城市 字典）

`Region` 是一张单表自关联的区域字典，用于「国家 → 城市」两级结构，供景点（`attractions.city_id`）绑定城市并统一多语言展示。区域为全局字典，不带 `tenant_id`/`merchant_id`。

### 层级与自关联

- `level=COUNTRY`：国家，`parent_id=0`
- `level=CITY`：城市，`parent_id` 指向所属国家行

### 多语言

基础名称（默认 `zh-CN`）存放在 `regions.name`；其余语言存放在 `region_translation`（按 `locale` 区分，`(region_id, locale)` 唯一）。C 端按请求 `locale` 解析名称，命中译文则覆盖，否则回退基础名。三语口径与 `product_translation`/`package_translation` 一致：`zh-CN` / `en-US` / `ar`。

### Region 字段

| Field | Type | Description |
|-------|------|-------------|
| parent_id | int64 | 父级 ID：0=国家，城市指向所属国家 |
| level | string | COUNTRY / CITY |
| code | string? | ISO 代码（如国家 AE/CN），城市可空 |
| name | string | 基础名称（默认 zh-CN） |
| poi_id | string? | 代表 POI ID（高德/Google） |
| poi_name | string? | 代表 POI 名称 |
| latitude | float? | 中心纬度 WGS84 |
| longitude | float? | 中心经度 WGS84 |
| sort | int | 排序，越小越靠前 |
| status | string | ACTIVE / INACTIVE |

### RegionTranslation 字段

| Field | Type | Description |
|-------|------|-------------|
| region_id | int64 | 关联区域 ID |
| locale | string | zh-CN / en-US / ar |
| name | string? | 区域名称译文 |
| source | string | MACHINE / MANUAL |
| status | string | PENDING / DONE / FAILED |
| updated_at | int64 | 最近翻译时间戳(unix秒) |

### 唯一性与索引

- Region：`(parent_id, name)` 唯一（避免同国家内重复城市名）；`(level, status)` 辅助查询
- RegionTranslation：`(region_id, locale)` 唯一

### 景点绑定

`attractions` 新增 `city_id`（int64，默认 0）指向 `regions.id`（level=CITY）。保留原有 `city` 字符串作为冗余展示名与历史数据兜底。删除城市前会校验是否仍被景点绑定，删除国家前会校验其下是否仍有城市。

### 接口（travel-api，直用 Ent client，不经 gRPC）

- `GET /api/travel/regions`：公开字典，支持 `level`/`parentId`/`locale`/`keyword`/`status`，名称按 locale 解析
- `GET /api/travel/merchant/regions`：管理端列表，返回基础名 + 全部译文
- `POST/PUT/DELETE /api/travel/merchant/regions[/:id]`：区域增删改，写入时覆盖式重建译文

## Attraction 多语言（名称/分类/标签/地址/简介）

景点的可翻译文本采用与区域一致的「基础表存中文 + 译文表按 locale 存各语言」模式。

### 多语言

- `attractions` 主表的 `name`/`category`/`tags`/`address`/`description` 为基准语言（zh-CN）。
- `attraction_translation` 表按 `locale`（en-US/ar）存放上述 5 个字段的对应译文。
- 分类 `category` 主表存枚举 code（如 `THEME_PARK`），译文里存该语言的本地化显示名文本；C 端按 locale 解析后直接展示。
- 标签 `tags` 为 JSON 数组字符串，各语言各自维护一份。

### AttractionTranslation 字段

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| attraction_id | int64 | 关联景点 ID |
| locale | string | 目标语言：zh-CN/en-US/ar |
| name | string | 名称译文 |
| category | string | 分类本地化显示名译文 |
| tags | text | 标签译文（JSON 数组） |
| address | string | 地址译文 |
| description | text | 简介译文 |
| source | string | MACHINE/MANUAL，默认 MANUAL |
| status | string | PENDING/DONE/FAILED，默认 DONE |
| updated_at | int64 | 最近翻译时间戳 |

### 唯一性与索引

- `(attraction_id, locale)` 唯一；`(locale, status)` 辅助查询。

### 接口（travel-api，直用 Ent client，不经 gRPC）

- 管理端 `GET /api/travel/merchant/attractions[/:id]`：返回基础字段 + 全部语种译文明细（供编辑）。
- `POST/PUT /api/travel/merchant/attractions[/:id]`：随请求体 `translations` 覆盖式重建该景点的 en-US/ar 译文。
- 公开端 `GET /api/travel/attractions`、`GET /api/travel/products/:id/attractions`：按 `locale`（或请求头语言）解析，命中译文覆盖对应字段，否则回退中文基准。
