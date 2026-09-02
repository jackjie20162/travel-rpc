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
- Aligned the Makefile RPC generation target with the generation style used by the existing merchant-rpc repository.

### Verification status
- Real GitHub Actions run confirmed protobuf generation succeeds.
- Real GitHub Actions run confirmed Ent generation succeeds after renaming the Ent `Package` schema to `ProductPackage`.
- The same run then failed at `go test ./...` because the generated workspace required module metadata updates; the workflow has now been changed to run `go mod tidy` before tests.
- Generated protobuf/Ent files are still generated during CI and are not yet committed to the repository.
- Runtime RPC services and database repositories are not implemented yet.
- Inventory reservation transaction has not yet been implemented.
- travel-api generated handlers and RPC client wiring are not yet complete.

## 2026-09-02 — Merchant Travel plugin foundation

### Implemented
- Added the Travel plugin directory to `b2b2c-vben5-admin-ui/apps/simple-admin-core/src/plugin/travel`.
- Added merchant-side menu SQL for 旅游产品、库存与价格、旅游订单.
- Added initial Travel product/inventory/order API contracts and page skeletons.
- Explicitly deferred the total/platform-admin Travel module as requested.

### Architecture decision
- The merchant frontend must not bypass merchant authentication and tenant/merchant scope by directly calling the tourism service.
- The intended production path is `merchant-frontend -> merchant-api2 -> travel-rpc`.
- The current plugin API files are therefore considered contract/skeleton code until the merchant API gateway and travel RPC client are wired and verified.

### Next
1. Make travel-rpc CI green after the module-tidy fix.
2. Implement Catalog/Inventory/Order RPC services and tenant/merchant scope enforcement.
3. Finish travel-api REST gateway and authentication/context propagation.
4. Add merchant-api2 Travel gateway endpoints backed by travel-rpc.
5. Replace Travel plugin skeleton pages with real CRUD/calendar/order workflows.
6. Keep documentation updated at every implementation step.
