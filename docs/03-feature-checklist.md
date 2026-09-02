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
- [ ] repositories
- [ ] catalog service implementation

## Inventory
- [x] Inventory schema + indexes
- [x] availability RPC contract
- [x] reservation RPC contract
- [ ] repository
- [ ] availability implementation
- [ ] atomic reservation
- [ ] idempotent reservation key storage
- [ ] release/confirm

## Order
- [x] Order schema + indexes
- [x] OrderItem schema
- [x] CreateOrder RPC contract
- [ ] repository
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

Latest M1 work: tenant/merchant isolation indexes, traveler query index, authoritative order pricing boundary, transaction-safe reservation contract, data-model/security documentation, and CI generation/build/test workflow.

## Verification status

The latest CI run after the protobuf-path correction is still not green; runtime generation and build must be verified before marking generated code or M1 execution complete.
