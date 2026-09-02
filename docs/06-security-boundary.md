# Travel RPC — Security Boundary

1. Public API clients must not choose `tenant_id` or `merchant_id`.
2. Public API clients must not choose the authoritative order total or currency.
3. RPC services derive tenant/merchant scope from authenticated context or a trusted upstream gateway.
4. Product, package and inventory ownership must be checked before order creation.
5. Inventory reservation is the concurrency boundary and must be transaction-safe.
6. Reservation keys should be idempotent so retries cannot double-reserve inventory.
7. Payment callbacks must be authenticated, verified and idempotent before changing order state.
8. Logs must not expose payment credentials or unnecessary traveler personal data.
