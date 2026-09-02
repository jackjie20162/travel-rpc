package server

import (
    "google.golang.org/grpc"

    "gitee.com/meinongyihe/travel-rpc/ent"
    "gitee.com/meinongyihe/travel-rpc/internal/repository"
    "gitee.com/meinongyihe/travel-rpc/internal/service"
    "gitee.com/meinongyihe/travel-rpc/travel"
)

// Register wires the standalone Travel domain services. No merchant-api or
// merchant-rpc dependency is required in the Travel runtime path.
func Register(grpcServer *grpc.Server, client *ent.Client) {
    products := repository.NewProductRepository(client)
    inventory := repository.NewInventoryRepository(client)
    orders := repository.NewOrderRepository(client)
    booking := repository.NewBookingRepository(client)
    payments := repository.NewPaymentRepository(client)

    travel.RegisterCatalogServiceServer(grpcServer, service.NewCatalogService(products))
    travel.RegisterInventoryServiceServer(grpcServer, service.NewInventoryService(inventory))
    travel.RegisterOrderServiceServer(grpcServer, service.NewOrderService(orders, inventory, booking))
    travel.RegisterPaymentServiceServer(grpcServer, service.NewPaymentService(payments))
    travel.RegisterTravelManagementServiceServer(grpcServer, service.NewManagementService(client))
}
