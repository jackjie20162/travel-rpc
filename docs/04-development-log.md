# Development Log

## 2026-09-09 — C 端库存批量查询窗口调整与 RPC 客户端超时

### 背景
- C 端产品详情页库存批量查询原按「当月 1 号 → 当月最后一天」，业务要求改为「当天 → 当天+30 天」滚动窗口。
- 商户端产品列表/编辑偶发报错：RPC slowcall 显示 GetProduct/ListProducts 间歇 ~1.8s（同 SQL 热调用仅 12.7ms），为本地 Docker MySQL 间歇停顿；zrpc 客户端默认超时 2000ms，停顿超 2s 时 API 层 DeadlineExceeded，前端表现为接口异常。

### 变更
- **travel-app** `ProductDetail.vue`：`loadMonthInventory` 改为 `loadInventoryWindow`（startDate=当天、endDate=当天+30）；翻月导航不再重复请求（窗口与展示月份无关），selectPkg/openCalendar 缓存为空时仍触发。
- **travel-api** `etc/travel-api.yaml`：`TravelRpc` 显式 `Timeout: 10000`，容忍 DB 间歇停顿。

### 验证
- ✅ travel-app `vite build` 通过（507ms）

## 2026-09-09 — C 端上门接接送范围展示

### 背景
产品详情页行程时间线已上线，但集合节点为「上门接」模式时，商户在后台绘制的接送范围（type_params 中 pickupPolygon/pickupRangeMode/pickupRangeOptions）未面向消费者展示。

### 实现内容（仅 travel-app）
- 新增 `src/utils/amap.js`：高德 JS API 2.0 动态加载（Key 读 `VITE_AMAP_JS_API_KEY`/`VITE_AMAP_SECURITY_KEY`，.env 已 gitignore）。
- 新增 `src/utils/coordTransform.js`：WGS84 → GCJ02 转换（后端存 WGS84，高德国内版需 GCJ02）。
- 新增 `src/components/PickupRangeMap.vue`：只读地图组件，将 WGS84 多边形转 GCJ02 后用 AMap.Polygon 绘制并 setFitView；加载失败降级为文案提示。
- `ProductDetail.vue`：集合节点上门接模式下新增「接送范围」块——范围模式标签（自定义接送范围/仅列表部分酒店/地点）+ 覆盖选项标签（所有区域/所有酒店/机场火车站/超范围付费接送）+ 范围地图（顶点≥3 时）。
- `ProductDetail.vue`：返程「提供送回服务」每个送回行同样展示范围标签 + 范围地图 + 补充说明（note）；范围辅助函数兼容两套字段名（集合 pickupPolygon/pickupRangeMode/pickupRangeOptions，送回行 polygon/rangeMode/rangeOptions）。

### 验证
- ✅ travel-app `vite build` 通过（779ms / 333ms）

## 2026-09-08 — C 端产品详情页行程展示

### 背景
travel-app 产品详情页（/product/:id）此前无行程信息，商户后台录入的行程节点需要面向消费者展示。

### 实现内容
- **travel-rpc**: `CatalogService` 新增 `ListItineraryStops(ItineraryStopListRequest)`，租户可选、按 sequence/id 排序，复用 `itineraryStopMessage`；`make gen-rpc` 重新生成。
- **travel-api**: 新增公开路由 `GET /api/travel/products/:id/itinerary-stops`（`ListProductItineraryStops`），logic 走 `CatalogClient`；handler 补 `SetPathParams`（goctls 生成件默认缺失）。
- **travel-app**: `api.js` 新增 `getProductItineraryStops`；`ProductDetail.vue` 在套餐与详情之间新增行程时间线：分类型图标/标签/标题/元信息（时间、城市、时长、入内）、特色描述、集合图片、返程送回行/解散点/自由解散列表；typeParams JSON 解析展示。

### 验证
- ✅ travel-rpc / travel-api `go build ./...` 通过
- ✅ travel-app `vite build` 通过（752ms）

## 2026-09-08 — 行程选点回显与坐标持久化修复

### 背景
商户后台反馈：行中「地点和活动」地图选点后点确定，表单不更新。截图证据：弹窗名称框仍为旧值，而已选坐标/逆地理地址已是新点。

### 根因
- `PointPickMap.vue` 逆地理/搜索回填名称、城市时使用「仅空时填充」策略：已有旧名称时重新选点不会更新名称/城市，确定后旧名称回写表单，表现为「确定不会更改」。
- `ProductEdit.vue` 的 `stopPayload` 与 onMounted 映射遗漏节点自身 `latitude/longitude`（API/RPC 链路本已支持），选点坐标保存后丢失、重进无法回显。

### 修复
- `PointPickMap.vue`：点击地图/搜索命中后同步更新名称、城市、地址为所选点结果（之后仍可手动改），并加提示文案。
- `ProductEdit.vue`：stopPayload 与加载映射补齐 latitude/longitude。

### 验证
- ✅ merchant-frontend `vite build` 通过

## 2026-09-08 — 已发布产品允许修改行程

### 变更
- 移除 `ManagementService` 行程节点 Create/Update/Delete 中的 PUBLISHED 状态拦截（原报错 `published product cannot modify itinerary`）：已发布产品可直接修改行程，无需先下架。
- 保留产品存在性与租户/商户归属校验。
- 商户后台编辑页同步移除「已发布不可改行程」提示条与一键下架按钮。
- 注：套餐新增（`published product cannot add packages`）的限制未变。

## 2026-09-08 — 行程编辑器结构重构与 type_params 通道

### 背景
商户后台行程编辑旧实现允许任意增删/改类型节点，且集合/返程与中间节点参数混用同一表单，不符合业务规则：集合与返程各唯一且必须存在，中间仅可插入地点和活动/行中交通/行中餐食；各类型参数差异大（集合分上门接/集合点，返程分送回服务/解散点/自由解散）。

### 实现内容
- **Ent schema**: `itinerary_stop` 新增 `type_params` text 列，存储节点类型专属参数 JSON；`make gen-ent` 重新生成。
- **Proto**: `ItineraryStop` / `CreateItineraryStopRequest` / `UpdateItineraryStopRequest` 新增 `type_params` 字段；`make gen-rpc` 重新生成。
- **RPC service**: `itineraryStopMessage` / Create / Update 增加 TypeParams 映射（Update 无条件覆盖以支持清空）。
- **travel-api**: `.api` 类型新增 `typeParams`，`goctls api go` 重新生成 types.go；create/update logic 与 `toItineraryStop` 增加透传。
- **商户后台**:
  - `ProductEdit.vue` 步骤 2 重构：首集合/尾返程固定不可删，中间插入按钮（地点和活动/行中交通/行中餐食）；分类型表单（集合双模式、交通选项+时长、活动 POI+入内+时长、餐型单选+时长、返程三解散方案+送回行/解散点列表）；保存前结构+必填校验；标题自动生成；typeParams 序列化存取。
  - 新增 `PointPickMap.vue` 地图定位选点组件（高德国内/世界双供应商，点击选点+逆地理，统一输出 WGS84），服务于集合点/活动地点/解散点定位。

### 验证
- ✅ travel-rpc / travel-api `go build ./...` 通过
- ✅ merchant-frontend `vite build` 通过（903ms）

### 文档
- 更新 `05-data-model.md`：ItineraryStop 结构约束与 type_params JSON 结构说明

## 2026-09-06 — Merchant order list workflow refinement

### Implemented
- Reworked `merchant-frontend/src/plugin/travel/orders/Index.vue` around the merchant order-operating workflow.
- Added the screenshot-aligned status tabs, verification counters, dense search area, quick filters, color markers, sorting, and a horizontally rich order table.
- The table now prioritizes order number, lifecycle status, product/package context, booking/use date, contact data, quantity progress, amount summary, markers, and detail/remark actions.
- Kept the existing `travel-api` merchant order-list integration. Fields not yet supplied by the API render as placeholders rather than fabricated values.

## 2026-09-04 — 产品发布与行程信息优化

### 背景
参考携程产品发布页面，对 travel 项目的产品发布和商品管理进行全面优化。用户提供 10 张携程截图作为参考，展示 5 步产品创建流程：产品信息→行程信息→套餐&报价→预订&须知→高级设置。

### 实现内容

#### 1. Ent Schema 扩展
- **product.go**: 新增 6 个字段（highlights, cover_image, images, video_url, rich_content, booking_notice）
- **package.go**: 新增 7 个字段（pricing_mode, inventory_mode, min_order_qty, sell_currency, cost_currency, group_prices, tier_prices）
- **inventory.go**: 新增 3 个字段（inventory_mode, total_capacity, is_open）
- **itinerary_stop.go**: 新建 schema，17 个字段，支持 5 种节点类型

#### 2. Protobuf 合同扩展
- 新增 `ItineraryStop` message 及 CRUD RPC 方法
- `CreateItineraryStopRequest` / `UpdateItineraryStopRequest` 各新增 14 个字段
- 字段涵盖：POI 信息、入内标志、时长模式、活动特色、接送地点坐标

#### 3. RPC Service 实现
- `ManagementService` 新增行程节点 CRUD 方法
- 更新 `productMessage()`, `packageMessage()`, `inventoryMessage()`, `itineraryStopMessage()` 转换器
- 所有方法均包含租户/商户隔离验证

#### 4. API Handler 实现
- 新增 4 个行程节点 REST 端点
- 更新 `types.go` 中的 `ItineraryStopReq` 和 `ItineraryStop` 类型
- 更新 `routes.go` 注册新路由

#### 5. 前端页面重构
- **ProductEdit.vue**: 从简单表单重构为 5 步分步表单（413 行 → 558 行）
  - Step 1: 产品信息（标题、亮点、图片、视频）
  - Step 2: 行程信息（节点类型、POI 搜索、入内选择、体验时长、活动特色、接送地点）
  - Step 3: 套餐&报价（报价模式、库存模式、套餐列表）
  - Step 4: 预订&须知（退改规则、违约金、图文介绍）
  - Step 5: 高级设置（产品描述、预订须知文本、发布操作）
- **packages/Index.vue**: 重写为日历库存管理界面（325+ 行）
  - 月历网格视图
  - 每日显示：开/关班开关、卖价、底价、库存
  - 批量设置功能（按周几选择）
  - 单日编辑对话框

#### 6. API 层更新
- `api/travel.js`: 新增 4 个行程节点 API 方法
- `plugin/travel/api.js`: 新增 4 个 API 导出

### 技术细节

#### 字段命名规范
- Protobuf: `poi_id` → Go: `PoiId`（大驼峰）
- Ent: `video_url` → Go: `VideoURL`（URL 全大写）
- Ent: `product_id` (int64) → 直接使用，无需 int 转换

#### 代码生成流程
1. Ent schema → `go generate ./ent` → 生成 CRUD 代码
2. Proto 文件 → `protoc` → 生成 `travel.pb.go` 和 `travel_grpc.pb.go`
3. 使用 bat 文件解决 Windows PATH 问题

#### 编译验证
- ✅ `travel-rpc`: `go build ./internal/...` 通过
- ✅ `travel-api`: `go build ./...` 通过

### 待办事项
- [ ] 接入高德地图 POI 搜索 API
- [ ] 执行数据库迁移
- [ ] 前端页面实际运行测试
- [ ] 接送地点地图选点功能

### 文档
- 新增 `10-product-publishing-optimization.md` 详细记录本次优化

## 2026-09-04 — User module (travel-app integration)

### Implemented
- Added User Ent schema mirroring the legacy PHP user table (username, password+salt, email, mobile, avatar, level, gender, birthday, bio, money, score, token, status, login tracking fields).
- Added `UserService` gRPC contract: Register, Login, LoginByMobile, GetProfile, UpdateProfile, ChangePassword.
- Added UserRepository interface and MySQL implementation with token-based lookup, password update, login info tracking.
- Added UserService RPC implementation with sha256+salt password hashing, random token generation, authenticated user context extraction.
- Registered UserService in gRPC server.
- Added travel-api REST endpoints: POST register, POST login, POST login-mobile, GET profile, PUT profile, PUT password.
- Added Bearer token authentication middleware in travel-api (ExtractToken, user lookup by token, gRPC x-user-id propagation).
- Added travel-app user API layer (register, login, loginByMobile, getProfile, updateProfile, changePassword, logout).
- Added travel-app useUser composable for global auth state management.
- Added travel-app Login/Register page with form validation and redirect support.
- Updated travel-app Profile page: shows login/register when unauthenticated, shows user stats (score, money, level) and edit profile when authenticated.
- Added route guards for auth-required pages (booking, payment, orders, order detail).
- Updated AppHeader to show login status icon.

### Security
- Passwords are never stored in plaintext; sha256(sha256(password) + salt) with random 16-byte salt.
- Auth tokens are random 32-byte hex strings stored in the user record.
- Profile update and password change require valid Bearer token.
- Token lookup verifies user status is "normal".
- ChangePassword invalidates the existing token to force re-login.

### Architecture decision
- User authentication is token-based (not JWT) to match the existing PHP user table schema.
- The API gateway resolves the user from the token via the shared Ent/MySQL database, then propagates x-user-id via gRPC metadata to travel-rpc.
- travel-rpc remains stateless for auth; it trusts the x-user-id injected by the authenticated gateway.

## 2026-09-02 — M1 foundation hardening

### Implemented
- Added product/package/inventory/order/voucher indexes for tenant-safe lookup and uniqueness.
- Changed protobuf `go_package` to the repository module package `gitee.com/meinongyihe/travel-rpc/travel`.
- Corrected GitHub Actions protobuf generation to use Go module-aware output paths.
- Replaced client-authoritative order tenant/merchant/currency/total fields with a narrower order request.
- Added explicit inventory reservation RPC contract with `reservation_key` for retry-safe design.
- Added tenant schema with unique tenant code and merchant tenant-scoped uniqueness.
- Added data-model and security-boundary documentation.
- Added GitHub Actions CI to generate protobuf + Ent code, run tests and build.
- Aligned the Makefile RPC generation target with the existing merchant-rpc generation style.

### Verification status
- Real GitHub Actions verified protobuf generation succeeds.
- Real GitHub Actions verified Ent generation succeeds after renaming the Ent `Package` schema to `ProductPackage`.
- The subsequent test stage required module metadata updates; CI was changed to run `go mod tidy` before tests.
- Generated protobuf/Ent files remain CI-generated and are not manually substituted.

## 2026-09-02 — M1 executable RPC data layer

### Implemented
- Added authenticated tenant/merchant/customer scope extraction from gRPC metadata.
- Added CatalogService with tenant-scoped product reads/listing.
- Added InventoryService with availability checking and retry-safe reservation.
- Added transactional inventory reservation persistence with reservation-key idempotency, release, confirmation and expiration support.
- Added OrderService with server-derived inventory pricing/currency.
- Added MySQL order persistence and order-item creation.
- Registered generated Catalog/Inventory/Order services with zRPC.
- Wired the MySQL Ent client into RPC startup.

### Security boundary
- Tenant ID and merchant ID are not accepted as authoritative business-request fields.
- They are extracted from authenticated gateway metadata and enforced at the RPC boundary.
- Client-supplied total amount/currency are not trusted.

### Known limitation
- Order creation and reservation confirmation currently span separate repository transactions. The next hardening step is to make the reservation-to-order transition durable/idempotent so a process failure cannot leave an order and inventory hold inconsistent.
- Runtime code depends on actual protobuf/Ent generation and therefore remains unverified until CI passes.

## 2026-09-02 — Ent ID type compatibility fix

### Implemented
- Fixed the `int64` RPC/domain ID to `int` Ent primary-key boundary in product, inventory-reservation and order-item repository operations.
- Kept the public repository/service interfaces on `int64`; conversions are isolated to generated Ent calls.
- Preserved tenant-scoped predicates and server-side pricing/inventory rules.

### Verification status
- The previous CI failure was caused by generated Ent primary keys using Go `int` while repository inputs used `int64`.
- The fixes are committed to `main`; a new CI run is expected from the push-triggered workflow.
- CI must still pass both `go test ./...` and `go build ./...` before this milestone is marked green.

## 2026-09-02 — Merchant Travel plugin foundation

### Implemented
- Added the Travel plugin directory to `b2b2c-vben5-admin-ui/apps/simple-admin-core/src/plugin/travel`.
- Added merchant-side menu SQL for 旅游产品、库存与价格、旅游订单.
- Added initial Travel product/inventory/order API contracts and page skeletons.
- Explicitly deferred the total/platform-admin Travel module as requested.

### Architecture decision
- The merchant frontend must not bypass merchant authentication and tenant/merchant scope by directly calling the tourism service.
- The intended production path is `merchant-frontend -> merchant-api2 -> travel-rpc`.
- The current plugin API files remain contract/skeleton code until the merchant API gateway and travel RPC client are wired and verified.

### Next
1. Verify the new travel-rpc CI run.
2. Harden reservation/order consistency and add database concurrency tests.
3. Finish travel-api REST gateway and authentication/context propagation.
4. Add merchant-api2 Travel gateway endpoints backed by travel-rpc.
5. Replace Travel plugin skeleton pages with real CRUD/calendar/order workflows.
6. Keep documentation updated at every implementation step.
