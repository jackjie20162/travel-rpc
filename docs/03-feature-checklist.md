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
- [ ] real product CRUD
- [ ] package/SKU management
- [ ] inventory calendar/time-slot management
- [ ] order operations
- [ ] permissions/role mapping

## Current milestone

**M0 — foundation/contracts: completed.**

**M1 — executable RPC data layer: in progress.**

Current M1 implementation includes authenticated tenant/merchant metadata extraction, CatalogService, InventoryService, OrderService, MySQL inventory reservation persistence, and order persistence. These changes require real protobuf/Ent generation before compilation.

## Verification status

The latest completed CI run verified protobuf and Ent generation but failed at `go test` because module metadata needed tidying. A new CI run has been triggered after the module-tidy fix. The newly added runtime services/repositories are not marked verified until that run completes successfully.
