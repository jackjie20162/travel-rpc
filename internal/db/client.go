package db

import (
	"context"
	"database/sql"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"gitee.com/meinongyihe/travel-rpc/ent"
	_ "github.com/go-sql-driver/mysql"
)

// NewClient opens an Ent client from an existing SQL connection.
// Connection configuration remains owned by the RPC application config.
func NewClient(ctx context.Context, driverName, dataSourceName string) (*ent.Client, error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	driver := entsql.OpenDB(dialect.MySQL, db)
	return ent.NewClient(ent.Driver(driver)), nil
}
