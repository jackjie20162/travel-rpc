# Travel RPC — Feature Checklist

Legend: `[x]` implemented, `[~]` contract/schema started, `[ ]` pending.

## Foundation
- [x] Go module
- [x] RPC server bootstrap
- [x] protobuf contract
- [x] Ent generation entrypoint
- [ ] generated protobuf Go/grpc code
- [ ] generated Ent code
- [ ] compile verification
- [ ] automated tests / CI

## Tenant / Merchant
- [~] Tenant schema placeholder
- [x] Merchant schema
- [ ] repositories
- [ ] tenant authorization enforcement
- [ ] merchant scope enforcement

## Catalog
- [x] Product schema
- [x] Package/SKU schema
- [x] product RPC contract
- [ ] repositories
- [ ] catalog service implementation

## Inventory
- [x] Inventory schema
- [x] availability RPC contract
- [ ] repository
- [ ] availability implementation
- [ ] atomic reservation
- [ ] release/confirm

## Order
- [x] Order schema
- [x] OrderItem schema
- [x] CreateOrder RPC contract
- [ ] repository
- [ ] create/query implementation
- [ ] order state machine
- [ ] cancellation/refund

## Traveler / Voucher
- [x] Traveler schema
- [x] Voucher schema
- [ ] repositories
- [ ] voucher generation
- [ ] redemption

## Current milestone

**M0 — foundation/contracts: completed.**

**M1 — executable RPC data layer: next.**

M1 target: finalize Ent indexes/edges, generate code, implement repositories/services, implement transaction-safe inventory reservation, then connect travel-api and add CI verification.
