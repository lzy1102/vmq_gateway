package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lzy1102/vmq_gateway/server/security"
	"github.com/lzy1102/vmq_gateway/server/service"
)

type createOrderReq struct {
	Amount      int64  `json:"amount" binding:"required"`
	ServiceID   string `json:"service_id" binding:"required"`
	CallbackURL string `json:"callback_url"`
	APIKey      string `json:"api_key" binding:"required"`
}

func CreateOrder(c *gin.Context) {
	var req createOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "msg": "参数错误"})
		return
	}

	if req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "msg": "金额必须大于 0"})
		return
	}

	if req.CallbackURL != "" {
		if err := security.ValidateCallbackURL(req.CallbackURL); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 0, "msg": err.Error()})
			return
		}
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if err := service.VerifyAPIKey(ctx, req.ServiceID, req.APIKey); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 0, "msg": "API Key 错误"})
		return
	}

	order, device, err := service.CreateOrder(ctx, req.Amount, req.ServiceID, req.CallbackURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "msg": "创建订单失败"})
		return
	}

	qrURL := device.QRCode
	if qrURL == "" {
		qrURL = "/qr/alipay.png"
	}

	amountYuan := float64(order.Amount) / 100.0
	c.JSON(http.StatusOK, gin.H{
		"code": 1,
		"data": gin.H{
			"order_id":   order.TradeNo,
			"amount":     order.Amount,
			"amount_str": amountYuan,
			"device_id":  device.DeviceID,
			"qr_url":     qrURL,
		},
	})
}

type cancelOrderReq struct {
	OrderID string `json:"order_id" binding:"required"`
}

func CancelOrder(c *gin.Context) {
	var req cancelOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "msg": "参数错误"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	if err := service.CancelOrder(ctx, req.OrderID); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "订单已取消"})
}

func QueryOrderStatus(c *gin.Context) {
	orderID := c.Query("order_id")
	if orderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "msg": "缺少 order_id"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	order, err := service.GetOrder(ctx, orderID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "data": gin.H{"status": "pending"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 1,
		"data": gin.H{
			"order_id":   order.TradeNo,
			"amount":     order.Amount,
			"status":     order.Status,
			"paid_at":    order.PaidAt,
			"created_at": order.CreatedAt,
		},
	})
}
