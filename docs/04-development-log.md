# Development Log

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
