# Travel RPC — Architecture

```text
Customer Web/H5 / App
          │ REST
          ▼
     travel-api
          │ gRPC
          ▼
     travel-rpc
   ┌──────┼────────┐
   ▼      ▼        ▼
Catalog Inventory Order
   │      │        │
   └──────┼────────┘
          ▼
        Ent/DB
```

## Domain ownership

- CatalogService: products
- InventoryService: availability, price and reservation consistency
- OrderService: booking/order lifecycle
- Traveler/Voucher: order-associated domain data
- Tenant/Merchant: isolation context

## Data consistency

Inventory reservation must use a transaction/row lock or equivalent atomic conditional update. Order creation and inventory reservation must not use an unsafe read-then-write sequence.

## Service boundary

`travel-rpc` owns tourism business rules and persistence. `travel-api` owns REST presentation, validation, authentication/context forwarding and RPC orchestration.

## Contract

`desc/travel.proto` is the RPC source of truth. Generated protobuf code must come from the actual protoc/toolchain and must never be replaced with hand-written generated files.
