# DB Layer

The DB layer is intentionally small: it opens the SQL connection and creates the Ent client. Domain repositories and transaction boundaries belong in service/repository packages.

Current target database dialect: MySQL, matching the existing simple-admin ecosystem.

Before runtime integration, the module dependency and production connection pool settings must be verified against the actual `merchant-rpc` database configuration and go.mod.
