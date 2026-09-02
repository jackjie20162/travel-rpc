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
- GitHub workflow was triggered by the CI-path fix and still requires a green verification run before generated/runtime work is marked verified.
- Generated protobuf/Ent files are not yet committed.
- Runtime RPC services and database repositories are not implemented yet.
- Inventory reservation transaction has not yet been implemented.
- travel-api generated handlers and RPC client wiring are not yet complete.

### Next
M1 implementation: verify generation, configure database/Ent client, implement repositories, implement Catalog/Inventory/Order services, add transactional/idempotent reservation, register services, then finish travel-api wiring and tests.
