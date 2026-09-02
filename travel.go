package main

import (
	"flag"
	"gitee.com/meinongyihe/travel-rpc/internal/config"
	"gitee.com/meinongyihe/travel-rpc/internal/server"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/zrpc"
)

var configFile = flag.String("f", "etc/travel-rpc.yaml", "config file")

func main() {
	flag.Parse()
	var c config.Config
	conf.MustLoad(*configFile, &c)
	server := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *zrpc.RpcServer) {
		server.Register(grpcServer)
	})
	defer server.Stop()
	server.Start()
}
