# Travel RPC — Reservation Data Layer

## Status

M1.2 started.

## Reservation model

`InventoryReservation` is the idempotency and lifecycle record for inventory holds.

Fields:

- `tenant_id`: tenant isolation boundary
- `merchant_id`: merchant ownership boundary
- `inventory_id`: concrete inventory slot
- `reservation_key`: client/request idempotency key; unique within tenant
- `quantity`: held quantity
- `order_id`: optional order association
- `status`: `RESERVED`, `CONFIRMED`, `RELEASED`, `EXPIRED`
- `expires_at`: automatic hold expiry deadline
- `created_at`, `updated_at`: lifecycle timestamps

## Consistency rules

1. Reservation creation and inventory increment must occur in one database transaction.
2. A duplicate `(tenant_id, reservation_key)` must return the existing reservation rather than consume inventory twice.
3. Inventory availability is `capacity - reserved >= requested quantity` and must be enforced atomically at write time.
4. Only an active `RESERVED` record may be confirmed or released.
5. Expired reservations must not remain counted as active inventory holds.
6. Tenant and merchant scope are derived from server-side context, never trusted from the public request.

## Ent layer

The schema is the source of truth. Generated Ent code must be produced by the real Ent toolchain and must not be hand-written.

## Verification

The reservation schema has been committed. Ent client/repository implementation is the next M1.2 task. CI must generate Ent code, run tests, and build before this milestone is marked complete.
