# Travel RPC — Feature Checklist

Legend: `[x]` implemented, `[~]` contract/schema started, `[ ]` pending. A contract/schema checkbox does not mean runtime verification is complete.

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
- [ ] repositories
- [ ] tenant authorization enforcement
- [ ] merchant scope enforcement

## Catalog
- [x] Product schema + indexes
- [x] Package/SKU schema + indexes
- [x] product RPC contract
- [x] product repository foundation
- [ ] catalog service implementation

## Inventory
- [x] Inventory schema + indexes
- [x] InventoryReservation schema + idempotency index
- [x] availability RPC contract
- [x] reservation RPC contract
- [x] inventory repository contract
- [ ] repository implementation
- [ ] availability implementation
- [ ] atomic reservation
- [ ] idempotent reservation key storage implementation
- [ ] release/confirm implementation

## Order
- [x] Order schema + indexes
- [x] OrderItem schema
- [x] CreateOrder RPC contract
- [x] order repository contract
- [ ] repository implementation
- [ ] create/query implementation
- [ ] authoritative price calculation
- [ ] order state machine
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

## Current milestone

**M0 — foundation/contracts: completed.**

**M1 — executable RPC data layer: in progress.**

Latest M1 work: database configuration, MySQL driver dependency, Ent client bootstrap, product repository implementation foundation, inventory reservation repository contract, and order repository contract.

## Verification status

CI generation/test/build still needs a new green run. Repository code that imports generated Ent packages is intentionally dependent on the real Ent generation step; no hand-written generated code is accepted.
