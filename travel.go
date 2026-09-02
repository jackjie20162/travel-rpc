package main

import (
	"context"
	"flag"

	"gitee.com/meinongyihe/travel-rpc/internal/config"
	"gitee.com/meinongyihe/travel-rpc/internal/db"
	"gitee.com/meinongyihe/travel-rpc/internal/server"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/zrpc"
)

var configFile = flag.String("f", "etc/travel-rpc.yaml", "config file")

func main() {
	flag.Parse()
	var c config.Config
	conf.MustLoad(*configFile, &c)

	client, err := db.NewClient(context.Background(), "mysql", c.DatabaseConf.DataSource())
	if err != nil { panic(err) }
	defer client.Close()

	rpcServer := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *zrpc.RpcServer) {
		// zRPC invokes this callback with the concrete gRPC registrar in the generated server template.
		_ = grpcServer
	})
	defer rpcServer.Stop()
	rpcServer.Start()
}
