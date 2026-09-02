package main

import (
	"context"
	"flag"
	"fmt"

	"gitee.com/meinongyihe/travel-rpc/internal/config"
	"gitee.com/meinongyihe/travel-rpc/internal/db"
	"gitee.com/meinongyihe/travel-rpc/internal/server"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

var configFile = flag.String("f", "etc/travel-rpc.yaml", "config file")

func main() {
	flag.Parse()
	var c config.Config
	conf.MustLoad(*configFile, &c)

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true&loc=UTC", c.DatabaseConf.Username, c.DatabaseConf.Password, c.DatabaseConf.Host, c.DatabaseConf.Port, c.DatabaseConf.Dbname)
	client, err := db.NewClient(context.Background(), "mysql", dsn)
	if err != nil { panic(err) }
	defer client.Close()

	rpcServer := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		server.Register(grpcServer, client)
	})
	defer rpcServer.Stop()
	rpcServer.Start()
}
