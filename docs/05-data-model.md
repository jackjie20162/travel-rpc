# Travel RPC — Data Model

## Scope

The first executable data layer covers tenant, merchant, product, package/SKU, inventory, order, traveler and voucher.

## Isolation

All business records carry `tenant_id` where applicable. Tenant and merchant scope must come from authenticated service context, not from public request fields.

## Uniqueness

- Tenant: `code`
- Merchant: `(tenant_id, code)`
- Product: `(tenant_id, code)`
- Package: `(tenant_id, product_id, code)`
- Inventory: `(package_id, service_date, time_slot)`
- Order: `order_no`
- Voucher: `voucher_no`

## Inventory rule

`available = status OPEN AND capacity - reserved >= requested quantity`.

Reservation must be atomic inside a database transaction. The service must never read a remaining quantity and then perform an independent update that can oversell under concurrency.

## Money

Amounts are integer minor units. Currency is stored explicitly. Client-supplied total amounts are not authoritative; order totals are calculated from the inventory/package price held by the RPC data layer.

## Time

`service_date` is an ISO date and `time_slot` is a normalized business slot string. A later migration may replace these strings with stronger temporal types after timezone policy is finalized.

## Current implementation status

The Ent schema layer is being hardened before repositories/services are added. Schema source is authoritative; generated Ent code is produced by the CI toolchain and must not be hand-written.
