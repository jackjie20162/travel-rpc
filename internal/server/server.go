package server

import (
	"google.golang.org/grpc"

	"gitee.com/meinongyihe/travel-rpc/ent"
	"gitee.com/meinongyihe/travel-rpc/internal/repository"
	"gitee.com/meinongyihe/travel-rpc/internal/service"
	"gitee.com/meinongyihe/travel-rpc/travel"
)

// Register wires all generated tourism RPC services onto the underlying gRPC server.
func Register(grpcServer *grpc.Server, client *ent.Client) {
	products := repository.NewProductRepository(client)
	inventory := repository.NewInventoryRepository(client)
	orders := repository.NewOrderRepository(client)
	travel.RegisterCatalogServiceServer(grpcServer, service.NewCatalogService(products))
	travel.RegisterInventoryServiceServer(grpcServer, service.NewInventoryService(inventory))
	travel.RegisterOrderServiceServer(grpcServer, service.NewOrderService(orders, inventory))
}
