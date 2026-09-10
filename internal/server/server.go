package server

import (
    "context"

    "google.golang.org/grpc"

    "gitee.com/meinongyihe/travel-rpc/ent"
    "gitee.com/meinongyihe/travel-rpc/internal/repository"
    "gitee.com/meinongyihe/travel-rpc/internal/service"
    "gitee.com/meinongyihe/travel-rpc/travel"
    "github.com/zeromicro/go-zero/core/logx"
)

// Register wires the standalone Travel domain services. No merchant-api or
// merchant-rpc dependency is required in the Travel runtime path.
func Register(grpcServer *grpc.Server, client *ent.Client) {
    products := repository.NewProductRepository(client)
    inventory := repository.NewInventoryRepository(client)
    orders := repository.NewOrderRepository(client)
    booking := repository.NewBookingRepository(client)
    payments := repository.NewPaymentRepository(client)
    users := repository.NewUserRepository(client)
    reviews := repository.NewReviewRepository(client)
    currencies := repository.NewCurrencyRepository(client)

    // 启动时幂等写入内置币种(AED/USD/CNY)与默认汇率；失败不阻断服务启动。
    if err := currencies.SeedDefaults(context.Background()); err != nil {
        logx.Errorf("failed to seed default currencies: %v", err)
    }

    travel.RegisterCatalogServiceServer(grpcServer, service.NewCatalogService(products, client))
    travel.RegisterInventoryServiceServer(grpcServer, service.NewInventoryService(inventory))
    travel.RegisterOrderServiceServer(grpcServer, service.NewOrderService(orders, inventory, booking, currencies))
    travel.RegisterPaymentServiceServer(grpcServer, service.NewPaymentService(payments))
    travel.RegisterTravelManagementServiceServer(grpcServer, service.NewManagementService(client))
    travel.RegisterUserServiceServer(grpcServer, service.NewUserService(users))
    travel.RegisterReviewServiceServer(grpcServer, service.NewReviewService(reviews, orders))
    travel.RegisterCurrencyServiceServer(grpcServer, service.NewCurrencyService(currencies))
}
