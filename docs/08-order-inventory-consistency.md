# 订单与库存一致性

## 目标

Travel 的订单创建必须在 `travel-rpc` 内完成，不依赖 merchant-api2 或 merchant-rpc。

## 状态边界

```text
InventoryReservation
  RESERVED
      |
      | single Ent transaction
      v
Order PENDING_PAYMENT
      + OrderItem
      |
      v
Reservation CONFIRMED + order_id
```

`Reserve` 负责建立短期库存 hold；`CreateFromReservation` 负责把 hold 原子地转换为订单并确认 reservation。

## 幂等

- `reservation_key` 在 tenant 内唯一。
- 已 `CONFIRMED` 的 reservation 重试时直接返回关联订单。
- `RELEASED`、`EXPIRED` 或业务参数不一致的 reservation 不得复用。
- 同一 reservation 的并发创建使用条件更新作为并发闸门；失败事务整体回滚，不允许提交第二个订单。

## 失败恢复

如果客户端在提交事务后超时，再次使用相同 `reservation_key`，服务会读取已确认 reservation 并返回原订单。

如果进程在 Reserve 后、订单事务前退出，reservation 保持短期 `RESERVED`，由过期任务释放库存。

## 权威数据

订单金额、币种、库存状态均由 `travel-rpc` 从数据库读取，客户端不能覆盖。
