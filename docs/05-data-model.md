# Travel RPC — Data Model

## Scope

The first executable data layer covers tenant, merchant, product, package/SKU, inventory, order, traveler, voucher, and itinerary stops.

## Isolation

All business records carry `tenant_id` and `merchant_id` where applicable. Tenant and merchant scope must come from authenticated service context, not from public request fields.

## Uniqueness

- Tenant: `code`
- Merchant: `(tenant_id, code)`
- Product: `(tenant_id, code)`
- Package: `(tenant_id, product_id, code)`
- Inventory: `(package_id, service_date, time_slot)`
- Order: `order_no`
- Voucher: `voucher_no`
- ItineraryStop: `(product_id, sequence)` — ordered within a product

## Inventory rule

`available = status OPEN AND capacity - reserved >= requested quantity`.

Reservation must be atomic inside a database transaction. The service must never read a remaining quantity and then perform an independent update that can oversell under concurrency.

## Money

Amounts are integer minor units. Currency is stored explicitly. Client-supplied total amounts are not authoritative; order totals are calculated from the inventory/package price held by the RPC data layer.

## Time

`service_date` is an ISO date and `time_slot` is a normalized business slot string. A later migration may replace these strings with stronger temporal types after timezone policy is finalized.

## Current implementation status

The Ent schema layer is being hardened before repositories/services are added. Schema source is authoritative; generated Ent code is produced by the CI toolchain and must not be hand-written.

## Itinerary Stop model

The `ItineraryStop` schema represents a segment/stop in a product's itinerary. Each product can have multiple stops ordered by `sequence`.

### Stop types

- `MEETING` — 集合点
- `ACTIVITY` — 活动景点
- `TRANSPORT` — 交通路段
- `MEAL` — 餐饮安排
- `RETURN` — 返程

### Fields

| Field | Type | Description |
|-------|------|-------------|
| tenant_id | int64 | Tenant isolation |
| merchant_id | int64 | Merchant isolation |
| product_id | int64 | Parent product |
| stop_type | string | MEETING/ACTIVITY/TRANSPORT/MEAL/RETURN |
| title | string | Stop title |
| description | text? | Stop description |
| sequence | int | Ordering within product |
| location_name | string? | Location name |
| address | string? | Address |
| latitude | float? | Latitude |
| longitude | float? | Longitude |
| poi_id | string? | Amap POI ID |
| poi_name | string? | POI name |
| is_entering | bool | Whether entering the venue |
| duration_mode | string? | FIXED/PER_PACKAGE/UNLIMITED |
| duration_hours | int? | Duration hours |
| duration_minutes | int? | Duration minutes |
| activity_features | text? | Activity features description |
| transport_type | string? | Vehicle type (car/boat/walk) |
| start_time | string? | Start time (HH:mm) |
| end_time | string? | End time (HH:mm) |
| pickup_location | string? | Pickup location name |
| pickup_address | string? | Pickup address |
| pickup_latitude | float? | Pickup latitude |
| pickup_longitude | float? | Pickup longitude |
| dropoff_location | string? | Dropoff location name |
| dropoff_address | string? | Dropoff address |
| dropoff_latitude | float? | Dropoff latitude |
| dropoff_longitude | float? | Dropoff longitude |
| is_pickup | bool | Has pickup service |
| is_dropoff | bool | Has dropoff service |
| agreement_no_shopping | bool | No shopping commitment |
| agreement_adjustable | bool | Itinerary adjustable |

### Indexes

- `(tenant_id, product_id)` — tenant-scoped lookup
- `(product_id, sequence)` — ordered stops within product

## User model

The `User` schema stores end-user (customer) accounts. It mirrors the legacy PHP user table for data compatibility.

### Uniqueness

- User: `username` (unique)

### Authentication

- Password is stored as `sha256(sha256(password) + salt)` with a random 16-byte salt.
- Token is a random 32-byte hex string used for Bearer authentication.
- Status: `normal` (active) or `hidden` (disabled).

### Fields

| Field | Type | Description |
|-------|------|-------------|
| username | string | Unique login name |
| nickname | string | Display name |
| password | string | Hashed password |
| salt | string | Random salt for password hashing |
| email | string | Email address |
| mobile | string | Phone number |
| avatar | string | Avatar URL |
| level | uint8 | User level/tier |
| gender | int8 | 0=unknown, 1=male, 2=female |
| birthday | string? | Date of birth |
| bio | string | Short biography |
| money | float | Account balance |
| score | int | Loyalty score |
| token | string | Auth bearer token |
| status | string | Account status |
| jointime | int64? | Registration timestamp |
| logintime | int64 | Last login timestamp |
| loginip | string | Last login IP |
