# 商户旅游商品闭环

## 目标

商户端只通过 `travel-api` REST 管理旅游商品；`travel-api` 再通过 gRPC 调用本服务。Travel RPC 独立持有 Ent/MySQL 数据，不依赖 merchant-api2 或 merchant-rpc。

## 已定义 RPC

`TravelManagementService`：

- `CreateProduct`：创建 DRAFT 商品
- `UpdateProduct`：修改未发布商品
- `CreatePackage`：创建套餐/SKU
- `ListPackages`：查询套餐
- `UpsertInventory`：按套餐、日期、时段设置库存与价格
- `ListInventory`：查询库存
- `PublishProduct`：发布/下架商品

## 发布约束

商品发布前至少需要一个 ACTIVE 套餐。已发布商品不能直接编辑或新增套餐，先下架再修改。

库存更新不会修改 `reserved`，且禁止把 capacity 调低到已预占数量以下。

## 闭环

`merchant-frontend → travel-api → travel-rpc → Ent → MySQL`

发布后消费者路径为：

`travel-app/travel-pc → travel-api → CatalogService → 已发布商品 → InventoryService.Reserve → OrderService.Create`

## 验证状态

本次提交已完成 proto、RPC service、API gateway route 与文档代码提交。GitHub Actions 会负责 protobuf/Ent 生成、go test 和 go build；在 CI 通过前不声明为已测试完成。
