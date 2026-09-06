# Development Log

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
