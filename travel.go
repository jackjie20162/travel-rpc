package main

import (
	"context"
	"flag"
	"fmt"

	"entgo.io/ent/dialect"
	"gitee.com/meinongyihe/travel-rpc/ent"
	"gitee.com/meinongyihe/travel-rpc/internal/config"
	"gitee.com/meinongyihe/travel-rpc/internal/server"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/travel-rpc.yaml", "config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	// build ent client from DatabaseConf
	var drv dialect.Driver = c.DatabaseConf.NewNoCacheDriver()
	if c.DatabaseConf.Debug {
		drv = dialect.Debug(drv, logx.Info)
	}
	client := ent.NewClient(
		ent.Log(logx.Info),
		ent.Driver(drv),
	)
	defer client.Close()

	// auto-create database tables
	if err := client.Schema.Create(context.Background()); err != nil {
		logx.Severef("failed to create schema: %v", err)
	}

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		server.Register(grpcServer, client, c)

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
