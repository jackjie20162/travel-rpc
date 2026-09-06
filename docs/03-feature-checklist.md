# Travel RPC — Feature Checklist

Legend: `[x]` implemented, `[~]` contract/schema started, `[ ]` pending. Implementation does not mean CI/runtime verification is complete.

## Foundation
- [x] Go module
- [x] RPC server bootstrap
- [x] protobuf contract
- [x] Ent generation entrypoint
- [x] code-generation Makefile targets
- [x] CI generation/test/build workflow
- [ ] generated protobuf Go/grpc code committed
- [ ] generated Ent code committed
- [ ] CI first green run verified

## Tenant / Merchant
- [x] Tenant schema + unique code index
- [x] Merchant schema + tenant-scoped unique code index
- [ ] tenant/merchant repositories
- [x] RPC tenant scope extraction
- [x] tenant scope enforcement in Catalog/Inventory/Order service entry points
- [ ] full merchant authorization policy

## Catalog
- [x] Product schema + indexes
- [x] Package/SKU schema + indexes
- [x] product RPC contract
- [x] product repository implementation foundation
- [x] CatalogService implementation

## Inventory
- [x] Inventory schema + indexes
- [x] InventoryReservation schema + idempotency index
- [x] availability RPC contract
- [x] reservation RPC contract
- [x] inventory repository implementation
- [x] availability implementation
- [x] atomic reservation update with optimistic reserved-count predicate
- [x] idempotent reservation key storage
- [x] release/confirm/expire implementation
- [ ] database-level concurrency integration tests

## Order
- [x] Order schema + indexes
- [x] OrderItem schema
- [x] CreateOrder RPC contract
- [x] order repository implementation foundation
- [x] create/query implementation
- [x] server-authoritative price/currency calculation from inventory result
- [~] reservation-to-order confirmation workflow
- [ ] durable order state machine
- [ ] cancellation/refund

## Traveler / Voucher
- [x] Traveler schema + tenant/order index
- [x] Voucher schema + indexes
- [ ] repositories
- [ ] voucher generation
- [ ] redemption

## User / Auth
- [x] User Ent schema (mirrors legacy PHP user table)
- [x] User RPC contract (Register, Login, LoginByMobile, GetProfile, UpdateProfile, ChangePassword)
- [x] User repository + MySQL implementation
- [x] UserService with password hashing (sha256+salt) and token generation
- [x] gRPC server registration
- [x] API gateway REST endpoints (register, login, profile, password)
- [x] Token-based authentication middleware (Bearer token)
- [x] travel-app Login/Register page
- [x] travel-app user state composable (useUser)
- [x] travel-app Profile page with real user data
- [x] travel-app route guards for auth-required pages
- [ ] password reset / verification code flow
- [ ] OAuth / social login integration

## API Gateway
- [x] REST contract
- [ ] generated API handlers/service context
- [ ] RPC client wiring
- [ ] request validation
- [ ] authentication/tenant context propagation

## Merchant Backend Plugin
- [x] Travel plugin directory
- [x] merchant menu SQL skeleton
- [x] product/inventory/order page skeletons
- [ ] merchant-api2 Travel gateway endpoints
- [x] 产品发布 5 步分步表单（ProductEdit.vue）
- [x] 行程节点 CRUD（活动/集合/交通/餐饮/返程）
- [x] 活动节点增强（POI、入内、时长、特色、接送地点）
- [x] 日历库存管理界面（packages/Index.vue）
- [ ] 订单操作
- [ ] 权限/角色映射

## 产品发布与行程信息（2026-09-04 新增）
- [x] 产品亮点（多条文本，JSON 存储）
- [x] 封面图 + 多图上传
- [x] 视频 URL
- [x] 图文介绍（富文本 HTML）
- [x] 预订须知文本
- [x] 行程节点 5 种类型（MEETING/ACTIVITY/TRANSPORT/MEAL/RETURN）
- [x] 活动节点 POI 信息（poi_id, poi_name）
- [x] 入内/不入内选择
- [x] 体验时长模式（FIXED/PER_PACKAGE/UNLIMITED）
- [x] 活动体验特色描述
- [x] 接送地点（接站/送站的位置和坐标）
- [x] 协议勾选（无购物承诺、行程可调整）
- [x] 报价模式（SAME_PRICE/GROUP_PRICE/TIER_PRICE）
- [x] 库存模式（UNLIMITED/DAILY/TOTAL）
- [x] 退改规则配置
- [ ] 高德地图 POI 搜索 API 接入
- [ ] 接送地点地图选点功能
- [ ] 行程节点拖拽排序

## Current milestone

**M0 — foundation/contracts: completed.**

**M1 — executable RPC data layer: completed.**

**M2 — 产品发布与行程信息优化：completed.**

Current M2 implementation includes:
- 5-step product creation form (ProductEdit.vue)
- Itinerary stop CRUD with 5 node types
- Activity node enhancements (POI, entering, duration, features, pickup/dropoff)
- Calendar inventory management (packages/Index.vue)
- Booking rules configuration
- All backend RPC/API handlers updated and compiled

## Verification status

The latest completed CI run verified protobuf and Ent generation but failed at `go test` because module metadata needed tidying. A new CI run has been triggered after the module-tidy fix. The newly added runtime services/repositories are not marked verified until that run completes successfully.
