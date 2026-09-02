# Travel RPC — Current Requirements

**Version:** 0.1.0

## Goal

Provide the tourism domain core for the global Dubai tourism platform, serving customer Web/H5/App and integrating with the existing merchant/tenant platform.

## Core flow

Product → Package/SKU → Date → Time Slot → Inventory/Pricing → Traveler → Order → Payment → Voucher → Redemption

## Requirements

- Multi-tenant and merchant isolation
- Product catalog and package/SKU
- Date/time-slot inventory
- Real-time availability and pricing
- Transaction-safe inventory reservation; no overselling
- Booking/order lifecycle
- Traveler information
- Voucher generation and redemption
- Payment abstraction; PayPal is the first provider
- REST API gateway consumes RPC; domain rules stay in RPC
- AI Planner may recommend only products/prices/availability returned by travel-rpc

## Development sequence

1. travel-rpc + travel-api
2. Payment / PayPal
3. Web/H5
4. App
5. AI Planner and operations

## Documentation rule

Every implementation step updates the requirements/architecture/function checklist/development log as applicable.
