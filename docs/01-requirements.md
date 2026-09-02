# Travel RPC — Current Requirements

**Version:** 0.1.2

## Goal

Provide the tourism domain core for the global Dubai tourism platform, serving customer Web/H5/App and integrating with the existing merchant/tenant platform.

## Core flow

Product → Package/SKU → Date → Time Slot → Inventory/Pricing → Traveler → Order → Payment → Voucher → Redemption

## Requirements

- Multi-tenant and merchant isolation
- Tenant and merchant codes are unique within their required scope
- Product catalog and package/SKU
- Date/time-slot inventory
- Real-time availability and pricing
- Transaction-safe inventory reservation; no overselling
- Idempotent reservation retries
- Booking/order lifecycle
- Traveler information
- Voucher generation and redemption
- Payment abstraction; PayPal is the first provider
- REST API gateway consumes RPC; domain rules stay in RPC
- Client-provided tenant/merchant identity and total amount are not authoritative
- Order totals are calculated from RPC-owned product/package/inventory pricing
- AI Planner may recommend only products/prices/availability returned by travel-rpc

## Development sequence

1. travel-rpc + travel-api
2. Payment / PayPal
3. Web/H5
4. App
5. AI Planner and operations

## Documentation rule

Every implementation step updates the requirements/architecture/function checklist/development log as applicable.

## Current milestone

M1 executable RPC data layer: schema hardening is in progress. Runtime repositories and services are not yet claimed as complete until generated-code CI and build verification succeed.
