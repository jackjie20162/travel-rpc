# 订单事件 IM 推送（11）

更新时间：2026-09-23

## 目标

订单关键状态变化时，通过 IM（imGateway）把「订单卡片 + 文案」推送给交易对方，让客户在客服会话里收到接单/退款/核销通知，让商户在客服工作台收到新支付订单提醒。同时支撑前端「基于订单的咨询」（客户端主动把订单卡片发给商户）。

## 边界与约束

- travel-rpc 不直接连 IM 数据库，只调用 imGateway 的 HTTP 接口 `POST /im/system-msg`。
- 推送是**旁路能力**：`Im.GatewayUrl` 为空即整体禁用；发送在独立 goroutine 中执行（3s 超时、panic recover、失败仅记日志），任何情况下都不阻塞或影响订单主流程。
- 不修改 im-common 远程模块：`content_type` 在网关侧按 int32 透传，新增「订单卡片=4」只是约定值。
- 金额取订单存储原值（单位：元，币种为订单 `currency`），不做展示币种换算——卡片语义是「订单事实」而非「当前浏览币种」。

## 配置

`etc/travel-rpc.yaml`：

```yaml
Im:
  GatewayUrl: http://localhost:9281   # 空则禁用推送
  Token:                              # 随 X-Internal-Token 发送，需与 imGateway 的 SystemMsg.InternalToken 一致
```

对应 `internal/config/config.go` 的 `ImConf`（均为 optional，不影响既有配置加载）。

## 实现

### `internal/imnotify`

- `Init(config.ImConf)`：由 `internal/server/server.go` 在启动时注入。
- `NotifyOrder(ctx, order *ent.Order, fromMerchant bool, text string)`：
  - `fromMerchant=false` → 客户(3, `order.UserID`) 发给 商户(2, `order.MerchantID`)
  - `fromMerchant=true` → 商户发给客户
  - 身份为 0 时跳过并记日志（避免脏数据打扰）
  - 先发订单卡片（`content_type=4`），再发文本说明（`content_type=1`），两次都是独立异步投递
- 订单卡片 content（`OrderCard`，JSON）：

  ```json
  { "orderNo": "", "title": "", "package": "", "date": "", "amount": 0, "currency": "", "status": "" }
  ```

  `title`/`package`/`date` 取订单上的产品快照字段（`product_name`/`package_name`/`service_date`），产品后续改名或下架不影响历史卡片。
- 请求体（`/im/system-msg`）：`{from_type, from_biz_uid, to_type, to_biz_uid, content, content_type}`。

### 注入点

| 事件 | 位置 | 方向 | 文案 |
| --- | --- | --- | --- |
| 支付成功（转待接单） | `internal/service/payment.go` MarkPaid 调用点 | 客户 → 商户 | 新订单已支付，请及时接单 |
| 接单/拒单 | `internal/service/order.go` AcceptOrder | 商户 → 客户 | 商家已接单，请等待出行 / 商家拒绝了订单。原因：xxx |
| 退款申请 | `internal/service/order.go` RequestRefund | 客户 → 商户 | 客户申请退款。原因：xxx |
| 退款审批 | `internal/service/order.go` HandleRefund | 商户 → 客户 | 退款申请已通过… / 被驳回。原因：xxx |
| 核销/核销驳回 | `internal/service/order.go` VerifyOrder | 商户 → 客户 | 订单已核销，祝您旅途愉快 / 核销未通过，已退款。原因：xxx |

均在状态变更成功之后调用，使用变更后的 `updated` 订单，保证卡片里的 `status` 是新状态。

## imGateway 侧

`POST /im/system-msg`（见 imGateway 仓库）：`X-Internal-Token` 校验 → 两端 `RegisterOrGetImUser` 懒注册解析 `im_uid` → 组装 `model.ImMsg`（单聊）→ 投递 `im_msg_topic`，完整复用既有「入库 + 实时推送 + 离线补偿」管线，因此离线用户上线后仍能收到订单通知。

## 前端消费

- travel-pc：`src/plugin/im/`（IM 抽屉，`contentType=4` 渲染为订单卡片，点击跳 `#/orders/:orderNo`）。
- travel-app：`src/plugin/im/Support.vue` 同样渲染订单卡片，订单详情「咨询此订单」带 `?orderNo=` 进入并自动发送卡片。
- merchant-frontend：IM 工作台 `ChatPanel.vue` 渲染订单卡片，点击进入 travel 商户订单详情。

## 验证

- `go build ./...` 通过（travel-rpc、imGateway）。
- 端到端（需先启动 Kafka、imWsRpc、imGateway(9281)、travel-rpc(9205)、travel-api(9206)）：PC 登录 → 产品详情咨询 → 下单支付 → 商户工作台收到订单卡片与「新订单已支付」→ 商户接单 → PC 客服会话收到「商家已接单」。
