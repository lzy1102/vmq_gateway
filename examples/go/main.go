// Go 接入示例 - V免签支付网关
// 支持回调模式和轮询模式
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ========== 配置 ==========
const (
	GatewayURL = "http://186.241.107.44"
	ServiceID  = "my_go_service"   // 你的服务ID
	APIKey     = "xxxxxxxxxxxx"   // 绑定时生成的 API Key
)

// ========== 数据结构 ==========
type OrderReq struct {
	Amount      int    `json:"amount"`       // 金额，单位分
	ServiceID   string `json:"service_id"`
	APIKey      string `json:"api_key"`
	CallbackURL string `json:"callback_url,omitempty"`
	PayType     string `json:"pay_type"` // wechat / alipay
}

type OrderResp struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data *OrderData  `json:"data"`
}

type OrderData struct {
	OrderID          string `json:"order_id"`
	RequestAmount    int    `json:"request_amount"`
	PayAmount        int    `json:"pay_amount"`
	PayStr           string `json:"pay_str"`
	DeviceID         string `json:"device_id"`
	QRURL            string `json:"qr_url"`
	ExpireAt         int64  `json:"expire_at"`
	RemainingSeconds int    `json:"remaining_seconds"`
}

type QueryResp struct {
	Code int       `json:"code"`
	Data *QueryData `json:"data"`
}

type QueryData struct {
	OrderID  string `json:"order_id"`
	Amount   int    `json:"amount"`
	Status   string `json:"status"`
	PaidAt   int64  `json:"paid_at"`
}

type CallbackData struct {
	TradeNo   string `json:"trade_no"`
	Amount    int    `json:"amount"`
	PayType   string `json:"pay_type"`
	PaidAt    int64  `json:"paid_at"`
	ServiceID string `json:"service_id"`
}

// ========== 1. 创建订单 ==========
func CreateOrder(amount int, payType string, callbackURL string) (*OrderData, error) {
	req := OrderReq{
		Amount:      amount,
		ServiceID:   ServiceID,
		APIKey:      APIKey,
		PayType:     payType,
		CallbackURL: callbackURL,
	}

	body, _ := json.Marshal(req)
	resp, err := http.Post(GatewayURL+"/api/order", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result OrderResp
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.Code == 1 {
		fmt.Printf("订单创建成功: %s\n", result.Data.OrderID)
		fmt.Printf("  实付金额: %s 元\n", result.Data.PayStr)
		fmt.Printf("  二维码: %s\n", result.Data.QRURL)
		fmt.Printf("  有效期: %d 秒\n", result.Data.RemainingSeconds)
		return result.Data, nil
	}

	return nil, fmt.Errorf("创建失败: %s", result.Msg)
}

// ========== 2. 查询订单状态 ==========
func QueryOrder(orderID string) (*QueryData, error) {
	resp, err := http.Get(fmt.Sprintf("%s/api/order/status?order_id=%s", GatewayURL, orderID))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result QueryResp
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.Code == 1 {
		return result.Data, nil
	}
	return nil, fmt.Errorf("查询失败")
}

// ========== 3. 轮询等待支付 ==========
func PollOrder(orderID string, timeout time.Duration) string {
	start := time.Now()
	fmt.Printf("开始轮询订单 %s...\n", orderID)

	for time.Since(start) < timeout {
		result, err := QueryOrder(orderID)
		if err == nil {
			switch result.Status {
			case "paid":
				fmt.Printf("✅ 收款成功！金额: %.2f 元\n", float64(result.Amount)/100)
				return "paid"
			case "expired", "cancelled":
				fmt.Printf("❌ 订单已%s\n", result.Status)
				return result.Status
			}
		}
		time.Sleep(3 * time.Second)
	}

	fmt.Println("⏰ 轮询超时")
	return "timeout"
}

// ========== 4. 回调处理 ==========
func handleCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, _ := io.ReadAll(r.Body)
	var data CallbackData
	if err := json.Unmarshal(body, &data); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	fmt.Printf("收到回调:\n")
	fmt.Printf("  订单号: %s\n", data.TradeNo)
	fmt.Printf("  金额: %.2f 元\n", float64(data.Amount)/100)
	fmt.Printf("  支付方式: %s\n", data.PayType)
	fmt.Printf("  服务: %s\n", data.ServiceID)

	// TODO: 在这里处理业务逻辑
	// - 更新订单状态
	// - 发货/开通服务

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"code":1,"msg":"ok"}`))
}

// ========== 主函数 ==========
func main() {
	// 方式1: 轮询模式
	fmt.Println("========== 轮询模式 ==========")
	order, err := CreateOrder(100, "alipay", "")
	if err != nil {
		fmt.Println(err)
		return
	}
	PollOrder(order.OrderID, 15*time.Minute)

	// 方式2: 回调模式（取消注释启用）
	// fmt.Println("========== 回调模式 ==========")
	// order, err = CreateOrder(500, "wechat", "http://你的公网IP:8080/callback")
	// if err != nil {
	//     fmt.Println(err)
	//     return
	// }
	// http.HandleFunc("/callback", handleCallback)
	// fmt.Println("回调服务已启动: :8080/callback")
	// http.ListenAndServe(":8080", nil)
}
