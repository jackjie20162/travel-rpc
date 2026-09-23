// Package imnotify 将订单事件以系统消息形式推送到 IM（imGateway /im/system-msg）。
// 设计约束：绝不阻塞/影响订单主流程 —— 未配置 GatewayUrl 时静默禁用，
// 发送在独立 goroutine 中执行（3s 超时，panic recover，失败仅记日志）。
package imnotify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"gitee.com/meinongyihe/travel-rpc/ent"
	"gitee.com/meinongyihe/travel-rpc/internal/config"

	"github.com/zeromicro/go-zero/core/logx"
)

// IM 用户身份类型（与 im-common constant.UserType 一致）
const (
	UserTypeMerchant int32 = 2
	UserTypeClient   int32 = 3
)

// IM 消息内容类型（与 im-common/前端约定一致：1文本 2图片 3商品卡片 4订单卡片）
const (
	ContentTypeText   int32 = 1
	ContentTypeOrder  int32 = 4
)

// OrderCard 订单卡片消息 content 载荷（前端按 contentType=4 渲染）。
// Amount 为订单存储币种原值（单位：元），不做展示币种换算。
type OrderCard struct {
	OrderNo  string `json:"orderNo"`
	Title    string `json:"title"`
	Package  string `json:"package"`
	Date     string `json:"date"`
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
	Status   string `json:"status"`
}

var gatewayURL, internalToken string

// Init 从配置初始化推送目标；GatewayUrl 为空时保持禁用。
func Init(c config.ImConf) {
	gatewayURL = c.GatewayUrl
	internalToken = c.Token
	if gatewayURL == "" {
		logx.Info("imnotify: gateway url not configured, order IM notification disabled")
	}
}

// NotifyOrder 以订单归属双方身份推送一条订单卡片 + 一条文本说明。
// fromMerchant 为 true 时发送方是商户、接收方是客户；反之客户发给商户。
func NotifyOrder(ctx context.Context, order *ent.Order, fromMerchant bool, text string) {
	if order == nil || gatewayURL == "" {
		return
	}
	var (
		fromType, toType       int32
		fromBiz, toBiz         int64
	)
	if fromMerchant {
		fromType, fromBiz = UserTypeMerchant, order.MerchantID
		toType, toBiz = UserTypeClient, order.UserID
	} else {
		fromType, fromBiz = UserTypeClient, order.UserID
		toType, toBiz = UserTypeMerchant, order.MerchantID
	}
	if fromBiz <= 0 || toBiz <= 0 {
		logx.WithContext(ctx).Infof("imnotify: skip, invalid identity from=%d to=%d order=%s", fromBiz, toBiz, order.OrderNo)
		return
	}

	card := OrderCard{
		OrderNo:  order.OrderNo,
		Title:    order.ProductName,
		Package:  order.PackageName,
		Date:     order.ServiceDate,
		Amount:   order.TotalAmount,
		Currency: order.Currency,
		Status:   order.Status,
	}
	cardJSON, err := json.Marshal(card)
	if err == nil {
		send(ctx, fromType, fromBiz, toType, toBiz, ContentTypeOrder, string(cardJSON))
	}
	if text != "" {
		send(ctx, fromType, fromBiz, toType, toBiz, ContentTypeText, text)
	}
}

// send 异步投递系统消息，失败仅记日志。
func send(ctx context.Context, fromType int32, fromBiz int64, toType int32, toBiz int64, contentType int32, content string) {
	body, err := json.Marshal(map[string]any{
		"from_type":    fromType,
		"from_biz_uid": fromBiz,
		"to_type":      toType,
		"to_biz_uid":   toBiz,
		"content":      content,
		"content_type": contentType,
	})
	if err != nil {
		logx.WithContext(ctx).Errorf("imnotify: marshal failed: %v", err)
		return
	}
	url := fmt.Sprintf("%s/im/system-msg", gatewayURL)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logx.Errorf("imnotify: panic sending system msg: %v", r)
			}
		}()
		reqCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, url, bytes.NewReader(body))
		if err != nil {
			logx.Errorf("imnotify: build request failed: %v", err)
			return
		}
		req.Header.Set("Content-Type", "application/json")
		if internalToken != "" {
			req.Header.Set("X-Internal-Token", internalToken)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			logx.Errorf("imnotify: post %s failed: %v", url, err)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			logx.Errorf("imnotify: post %s unexpected status %d", url, resp.StatusCode)
		}
	}()
}
