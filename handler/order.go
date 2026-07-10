package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lzy1102/vmq_gateway/config"
	"github.com/lzy1102/vmq_gateway/model"
	"github.com/lzy1102/vmq_gateway/service"
)

// createOrderReq 创建订单请求体
type createOrderReq struct {
	UserName    string `json:"user_name" binding:"required"`
	Package     string `json:"package" binding:"required"`      // small / medium / big
	ServiceID   string `json:"service_id" binding:"required"`   // 调用方标识
	CallbackURL string `json:"callback_url"`                    // 支付成功回调地址（可选）
}

// CreateRechargeOrder POST /api/recharge/vmpay
func CreateRechargeOrder(c *gin.Context) {
	var req createOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "msg": "参数错误"})
		return
	}

	pkg, ok := config.Packages[req.Package]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "msg": "无效的套餐"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	order, err := service.CreateOrder(ctx, req.UserName, pkg, req.ServiceID, req.CallbackURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "msg": "创建订单失败"})
		return
	}

	// 返回元为单位（方便前端显示）
	amountYuan := float64(order.Amount) / 100.0
	c.JSON(http.StatusOK, gin.H{
		"code": 1,
		"data": gin.H{
			"trade_no":   order.TradeNo,
			"amount":     order.Amount,
			"amount_str": amountYuan,
			"qr_url":     "/qr/alipay.png",
			"pkg_name":   pkg.Name,
		},
	})
}

// QueryOrderStatus GET /api/recharge/vmpay-status?trade_no=xxx
func QueryOrderStatus(c *gin.Context) {
	tradeNo := c.Query("trade_no")
	if tradeNo == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "msg": "缺少 trade_no"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	order, err := service.GetOrder(ctx, tradeNo)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "data": gin.H{"status": model.StatusCancelled}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 1,
		"data": gin.H{
			"trade_no":   order.TradeNo,
			"amount":     order.Amount,
			"status":     order.Status,
			"paid_at":    order.PaidAt,
			"created_at": order.CreatedAt,
		},
	})
}
