package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lzy1102/vmq_gateway/model"
	"github.com/lzy1102/vmq_gateway/service"
)

// Heartbeat 处理 APP 心跳 GET /appHeart?t={timestamp_ms}&sign={md5}
func Heartbeat(c *gin.Context) {
	t := c.Query("t")
	sign := c.Query("sign")

	if t == "" || sign == "" {
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "参数缺失"})
		return
	}

	tsMs, err := strconv.ParseInt(t, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "时间戳格式错误"})
		return
	}
	if !service.ValidateTimestamp(tsMs / 1000) {
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "时间戳过期"})
		return
	}

	device, err := findDeviceByHeartbeat(c.Request.Context(), t, sign)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "签名错误"})
		return
	}

	service.UpdateHeartbeat(c.Request.Context(), device.DeviceID)
	c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "成功"})
}

// AppPush 处理 APP 收款回调 GET /appPush?t={timestamp_ms}&type={1|2}&price={金额元}&sign={md5}
func AppPush(c *gin.Context) {
	t := c.Query("t")
	payType := c.Query("type")
	price := c.Query("price")
	sign := c.Query("sign")

	if t == "" || payType == "" || price == "" || sign == "" {
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "参数缺失"})
		return
	}

	tsMs, err := strconv.ParseInt(t, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "时间戳格式错误"})
		return
	}
	if !service.ValidateTimestamp(tsMs / 1000) {
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "时间戳过期"})
		return
	}

	device, err := findDeviceByPush(c.Request.Context(), payType, price, t, sign)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "签名错误"})
		return
	}

	priceYuan, err := strconv.ParseFloat(price, 64)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "金额格式错误"})
		return
	}

	order, serviceID, callbackURL, err := service.HandleCallback(c.Request.Context(), device, priceYuan)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "未找到匹配订单"})
		return
	}

	if callbackURL != "" {
		go service.NotifyCallback(order, serviceID, callbackURL)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 1,
		"msg":  "成功",
		"data": gin.H{"trade_no": order.TradeNo},
	})
}

// findDeviceByHeartbeat 遍历所有设备 key，找到匹配心跳签名的设备
func findDeviceByHeartbeat(ctx context.Context, t, sign string) (*model.Device, error) {
	devices, err := service.ListDevices(ctx)
	if err != nil {
		return nil, err
	}
	for i := range devices {
		if service.VerifyHeartbeatSign(t, sign, devices[i].VmqKey) {
			return &devices[i], nil
		}
	}
	return nil, fmt.Errorf("签名不匹配任何设备")
}

// findDeviceByPush 遍历所有设备 key，找到匹配回调签名的设备
func findDeviceByPush(ctx context.Context, payType, price, t, sign string) (*model.Device, error) {
	devices, err := service.ListDevices(ctx)
	if err != nil {
		return nil, err
	}
	for i := range devices {
		if service.VerifyPushSignWithKey(payType, price, t, sign, devices[i].VmqKey) {
			return &devices[i], nil
		}
	}
	return nil, fmt.Errorf("签名不匹配任何设备")
}
