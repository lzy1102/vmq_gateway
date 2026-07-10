package service

import (
	"context"
	"crypto/md5"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/lzy1102/vmq_gateway/server/model"
	"github.com/lzy1102/vmq_gateway/server/store"
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
	var orders []model.Order
	err := store.DBInstance.Find(ctx, "orders", map[string]interface{}{
		"amount": amount, "status": model.StatusPending,
	}, &orders)
	if err != nil {
		return false, err
	}
	return len(orders) > 0, nil
}

func CreateOrder(ctx context.Context, amount int64, serviceID, callbackURL string) (*model.Order, *model.Device, error) {
	mu.Lock()
	defer mu.Unlock()

	floatAmount := nextFloatAmount(ctx, amount)
	now := time.Now().Unix()

	devices, err := ListDevices(ctx)
	if err != nil || len(devices) == 0 {
		return nil, nil, fmt.Errorf("无可用设备")
	}

	device := devices[0]

	order := &model.Order{
		TradeNo:     fmt.Sprintf("V%d_%d", now, floatAmount),
		ServiceID:   serviceID,
		CallbackURL: callbackURL,
		Amount:      floatAmount,
		Status:      model.StatusPending,
		DeviceID:    device.DeviceID,
		CreatedAt:   now,
	}

	if err := store.DBInstance.Create(ctx, "orders", order); err != nil {
		return nil, nil, err
	}
	return order, &device, nil
}

func CancelOrder(ctx context.Context, orderID string) error {
	var order model.Order
	if err := store.DBInstance.Get(ctx, "orders", orderID, &order); err != nil {
		return fmt.Errorf("订单不存在")
	}

	if order.Status != model.StatusPending {
		return fmt.Errorf("订单状态不允许取消")
	}

	return store.DBInstance.UpdateByField(ctx, "orders", "trade_no", orderID,
		map[string]interface{}{"status": model.StatusCancelled})
}

func HandleCallback(ctx context.Context, device *model.Device, priceYuan float64) (*model.Order, string, string, error) {
	amountCents := int64(priceYuan*100 + 0.5)

	var order model.Order
	if err := store.DBInstance.Claim(ctx, "orders", amountCents, &order); err != nil {
		return nil, "", "", fmt.Errorf("未找到匹配订单: amount=%d", amountCents)
	}

	now := time.Now().Unix()
	if err := store.DBInstance.UpdateByField(ctx, "orders", "trade_no", order.TradeNo,
		map[string]interface{}{"status": model.StatusPaid, "paid_at": now}); err != nil {
		return nil, "", "", fmt.Errorf("更新订单状态失败: %w", err)
	}
	order.Status = model.StatusPaid
	order.PaidAt = now

	serviceID, callbackURL := order.ServiceID, order.CallbackURL
	return &order, serviceID, callbackURL, nil
}

func NotifyCallback(order *model.Order, serviceID, callbackURL string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	payload := fmt.Sprintf(`{"order_id":"%s","amount":%d,"service_id":"%s","status":"paid","paid_at":%d}`,
		order.TradeNo, order.Amount, serviceID, order.PaidAt)

	req, err := http.NewRequestWithContext(ctx, "POST", callbackURL, strings.NewReader(payload))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
}

func GetOrder(ctx context.Context, tradeNo string) (*model.Order, error) {
	var order model.Order
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
