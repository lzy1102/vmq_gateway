package service

import (
	"context"
	"crypto/md5"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/lzy1102/vmq_gateway/config"
	"github.com/lzy1102/vmq_gateway/model"
	"github.com/lzy1102/vmq_gateway/store"
)

var mu sync.Mutex

func nextFloatAmount(ctx context.Context, baseAmount int64) int64 {
	for i := int64(1); i <= 19; i++ {
		candidate := baseAmount + i
		exists, err := HasPendingAmount(ctx, candidate)
		if err != nil || !exists {
			return candidate
		}
	}
	return baseAmount
}

func HasPendingAmount(ctx context.Context, amount int64) (bool, error) {
	var orders []model.RechargeOrder
	err := store.DBInstance.Find(ctx, "orders", map[string]interface{}{
		"amount": amount, "status": model.StatusPending,
	}, &orders)
	if err != nil {
		return false, err
	}
	return len(orders) > 0, nil
}

func CreateOrder(ctx context.Context, userName string, pkg config.Package, serviceID, callbackURL string) (*model.RechargeOrder, error) {
	mu.Lock()
	defer mu.Unlock()

	floatAmount := nextFloatAmount(ctx, pkg.Amount)
	now := time.Now().Unix()
	order := &model.RechargeOrder{
		TradeNo:      fmt.Sprintf("V%d_%d", now, floatAmount),
		UserName:     userName,
		ServiceID:    serviceID,
		CallbackURL:  callbackURL,
		Amount:       floatAmount,
		StreamNumber: pkg.StreamNumber,
		Status:       model.StatusPending,
		CreatedAt:    now,
	}

	if err := store.DBInstance.Create(ctx, "orders", order); err != nil {
		return nil, err
	}
	return order, nil
}

// HandleCallback 处理收款回调，返回订单、服务ID、回调地址
func HandleCallback(ctx context.Context, device *model.Device, priceYuan float64) (*model.RechargeOrder, string, string, error) {
	amountCents := int64(priceYuan*100 + 0.5)

	var order model.RechargeOrder
	if err := store.DBInstance.Claim(ctx, "orders", amountCents, &order); err != nil {
		return nil, "", "", fmt.Errorf("未找到匹配订单: amount=%d", amountCents)
	}

	// 给用户加流量
	var users []model.User
	if err := store.DBInstance.Find(ctx, "users", map[string]interface{}{"user_name": order.UserName}, &users); err != nil {
		return nil, "", "", fmt.Errorf("查询用户失败: %w", err)
	}
	if len(users) == 0 {
		// 创建用户
		user := &model.User{UserName: order.UserName, StreamNumber: order.StreamNumber, CreatedAt: time.Now().Unix()}
		if err := store.DBInstance.Create(ctx, "users", user); err != nil {
			return nil, "", "", fmt.Errorf("创建用户失败: %w", err)
		}
	} else {
		// 更新用户流量
		newStream := users[0].StreamNumber + order.StreamNumber
		if err := store.DBInstance.UpdateByField(ctx, "users", "user_name", order.UserName,
			map[string]interface{}{"stream_number": newStream}); err != nil {
			return nil, "", "", fmt.Errorf("充值失败: %w", err)
		}
	}

	// 更新订单状态
	now := time.Now().Unix()
	if err := store.DBInstance.UpdateByField(ctx, "orders", "trade_no", order.TradeNo,
		map[string]interface{}{"status": model.StatusPaid, "paid_at": now}); err != nil {
		return nil, "", "", fmt.Errorf("更新订单状态失败: %w", err)
	}
	order.Status = model.StatusPaid
	order.PaidAt = now

	serviceID, callbackURL := RouteCallback(ctx, device, &order)
	return &order, serviceID, callbackURL, nil
}

// NotifyCallback 异步通知调用方支付成功
func NotifyCallback(order *model.RechargeOrder, serviceID, callbackURL string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	payload := fmt.Sprintf(`{"trade_no":"%s","amount":%d,"service_id":"%s","user_name":"%s","status":"paid"}`,
		order.TradeNo, order.Amount, serviceID, order.UserName)

	req, err := http.NewRequestWithContext(ctx, "POST", callbackURL, strings.NewReader(payload))
	if err != nil {
		fmt.Printf("[callback] 构造请求失败: %v\n", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("[callback] 通知失败: service_id=%s url=%s err=%v\n", serviceID, callbackURL, err)
		return
	}
	defer resp.Body.Close()
	fmt.Printf("[callback] 通知成功: service_id=%s url=%s status=%d\n", serviceID, callbackURL, resp.StatusCode)
}

// GetOrder 查询订单
func GetOrder(ctx context.Context, tradeNo string) (*model.RechargeOrder, error) {
	var order model.RechargeOrder
	if err := store.DBInstance.Get(ctx, "orders", tradeNo, &order); err != nil {
		return nil, err
	}
	return &order, nil
}

func SignHeartbeat(ts, key string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(ts+key)))
}

func SignPush(payType, price, ts, key string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(payType+price+ts+key)))
}

func VerifyHeartbeatSign(ts, sign, key string) bool {
	return SignHeartbeat(ts, key) == sign
}

func VerifyPushSignWithKey(payType, price, ts, sign, key string) bool {
	return SignPush(payType, price, ts, key) == sign
}

func ValidateTimestamp(tsSec int64) bool {
	diff := time.Now().Unix() - tsSec
	if diff < 0 {
		diff = -diff
	}
	return diff <= 300
}
