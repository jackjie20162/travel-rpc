# Development Log

## 2026-09-02 — Repository correction and M0

- Corrected repository ownership: tourism RPC belongs to `jackjie20162/travel-rpc`, not `travel-app`.
- Migrated the initial RPC module, protobuf contract, Ent schemas, server bootstrap, config, Makefile and documentation into `travel-rpc`.
- `travel-app` is not the target repository for RPC implementation.
- Generated protobuf/Ent code has not yet been generated in the repository.
- Local compile/test verification has not yet been executed through this workflow.

### Next
M1: generated code + Ent data layer + repositories + catalog/inventory/order services + transaction-safe inventory reservation + API wiring + CI verification.
