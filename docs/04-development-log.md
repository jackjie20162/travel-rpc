# Development Log

## 2026-09-02 — M1 foundation hardening

### Implemented
- Added product/package/inventory/order/voucher indexes for tenant-safe lookup and uniqueness.
- Changed protobuf `go_package` to match the repository's go-zero generation convention.
- Replaced client-authoritative order tenant/merchant/currency/total fields with a narrower order request.
- Added explicit inventory reservation RPC contract with `reservation_key` for retry-safe design.
- Added data-model and security-boundary documentation.
- Added GitHub Actions CI to generate protobuf + Ent code, run tests and build.
- Aligned the Makefile RPC generation target with the generation style used by the existing merchant-rpc repository.

### Not yet verified
- Generated protobuf/Ent files have not been committed by this workflow.
- CI first green run has not yet been observed/verified.
- Runtime RPC services and database repositories are not implemented yet.
- Inventory reservation transaction has not yet been implemented.
- travel-api generated handlers and RPC client wiring are not yet complete.

### Next
M1 implementation: generate code, configure database/Ent client, implement repositories, implement Catalog/Inventory/Order services, add transactional/idempotent reservation, register services, then finish travel-api wiring and tests.
