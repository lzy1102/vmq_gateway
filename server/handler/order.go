package handler

import (
	"context"
	"fmt"
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
	PayType     string `json:"pay_type"`
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

	if err := service.CheckIPWhitelist(ctx, req.ServiceID, c.ClientIP()); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"code": 0, "msg": err.Error()})
		return
	}

	order, device, err := service.CreateOrder(ctx, req.Amount, req.ServiceID, req.CallbackURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "msg": "创建订单失败"})
		return
	}

	qrURL := device.AlipayQR
	if req.PayType == "wechat" {
		qrURL = device.WechatQR
	}

	amountYuan := float64(order.Amount) / 100.0
	requestedYuan := float64(req.Amount) / 100.0
	now := time.Now().Unix()
	remaining := order.ExpireAt - now
	if remaining < 0 {
		remaining = 0
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 1,
		"data": gin.H{
			"order_id":         order.TradeNo,
			"request_amount":   req.Amount,
			"request_str":      fmt.Sprintf("%.2f", requestedYuan),
			"pay_amount":       order.Amount,
			"pay_str":          fmt.Sprintf("%.2f", amountYuan),
			"device_id":        device.DeviceID,
			"pool_id":          "",
			"qr_url":           qrURL,
			"expire_at":        order.ExpireAt,
			"remaining_seconds": remaining,
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

	now := time.Now().Unix()
	remaining := order.ExpireAt - now
	if remaining < 0 {
		remaining = 0
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 1,
		"data": gin.H{
			"order_id":          order.TradeNo,
			"amount":            order.Amount,
			"status":            order.Status,
			"paid_at":           order.PaidAt,
			"created_at":        order.CreatedAt,
			"expire_at":         order.ExpireAt,
			"remaining_seconds": remaining,
		},
	})
}

func ListOrders(c *gin.Context) {
	keyword := c.Query("keyword")
	status := c.Query("status")
	serviceID := c.Query("service_id")
	page := 1
	pageSize := 10
	var startTime, endTime int64
	if p := c.Query("page"); p != "" {
		fmt.Sscanf(p, "%d", &page)
	}
	if ps := c.Query("page_size"); ps != "" {
		fmt.Sscanf(ps, "%d", &pageSize)
	}
	if st := c.Query("start_time"); st != "" {
		fmt.Sscanf(st, "%d", &startTime)
	}
	if et := c.Query("end_time"); et != "" {
		fmt.Sscanf(et, "%d", &endTime)
	}
	result, err := service.ListOrdersWithPage(c.Request.Context(), keyword, page, pageSize, status, serviceID, startTime, endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "msg": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1, "data": gin.H{"items": result.Items, "total": result.Total, "page": result.Page, "page_size": result.PageSize, "total_pages": result.TotalPages}})
}
