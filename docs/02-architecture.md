# Travel RPC — Architecture

## 1. Standalone microservice boundary

Travel is a **pluggable standalone tourism microservice**. Its runtime boundary is intentionally limited to Travel API, Travel RPC and Travel-owned persistence.

```text
Web / H5 / App / external adapters
              │
              │ REST/HTTP
              ▼
          travel-api
              │
              │ gRPC
              ▼
          travel-rpc
              │
              ▼
           Ent / MySQL
```

The Travel service must be deployable, testable and upgradeable without requiring merchant, admin, payment or other business services to be present.

## 2. Responsibility split

### travel-api

- REST/HTTP presentation layer
- request validation
- authentication/context boundary
- REST response mapping
- RPC orchestration
- no tourism persistence
- no direct database access

### travel-rpc

- tourism domain rules
- catalog/product/package
- inventory and pricing
- inventory reservation
- order lifecycle
- traveler/voucher domain data
- tenant/merchant isolation rules
- Ent/MySQL persistence

## 3. Domain ownership

- CatalogService: products and packages
- InventoryService: availability, pricing and reservation consistency
- OrderService: booking/order lifecycle
- Traveler/Voucher: order-associated domain data
- Tenant/Merchant: isolation context within Travel domain

## 4. Dependency rule

`travel-api` may depend on `travel-rpc`.

`travel-rpc` must not depend on `travel-api`.

Neither component may require `merchant-api2`, `merchant-rpc`, simple-admin or the total/platform admin service at runtime.

External systems integrate through the published Travel API contract rather than importing Travel's internal repository or database implementation.

## 5. Data consistency

Inventory reservation must use a transaction/row lock or equivalent atomic conditional update. Order creation and inventory reservation must not use an unsafe read-then-write sequence.

Reservation/order recovery, idempotency and state transitions belong entirely inside `travel-rpc` so that consistency does not depend on an external service.

## 6. Contract

`desc/travel.proto` is the RPC source of truth. Generated protobuf code must come from the actual protoc/toolchain and must never be replaced with hand-written generated files.

The REST API contract in `travel-api` is the external integration contract; the gRPC contract in `travel-rpc` is the internal Travel service contract.
