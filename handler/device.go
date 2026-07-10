package handler

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lzy1102/vmq_gateway/model"
	"github.com/lzy1102/vmq_gateway/service"
)

type addDeviceReq struct {
	DeviceID string `json:"device_id" binding:"required"`
}

func generateKey() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func AddDevice(c *gin.Context) {
	var req addDeviceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "msg": "参数错误"})
		return
	}

	key := generateKey()
	device := &model.Device{
		DeviceID: req.DeviceID,
		VmqKey:   key,
		Status:   model.DeviceOffline,
	}
	if err := service.AddDevice(c.Request.Context(), device); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "msg": "添加失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "成功", "data": gin.H{"key": key}})
}

func ListDevices(c *gin.Context) {
	devices, err := service.ListDevices(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "msg": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1, "data": devices})
}

type addPoolReq struct {
	PoolID string `json:"pool_id" binding:"required"`
	Name   string `json:"name" binding:"required"`
}

func AddPool(c *gin.Context) {
	var req addPoolReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "msg": "参数错误"})
		return
	}

	pool := &model.Pool{
		PoolID:    req.PoolID,
		Name:      req.Name,
		DeviceIDs: []string{},
	}
	if err := service.AddPool(c.Request.Context(), pool); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "msg": "添加失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "成功"})
}

type addDeviceToPoolReq struct {
	PoolID   string `json:"pool_id" binding:"required"`
	DeviceID string `json:"device_id" binding:"required"`
}

func AddDeviceToPool(c *gin.Context) {
	var req addDeviceToPoolReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "msg": "参数错误"})
		return
	}
	if err := service.AddDeviceToPool(c.Request.Context(), req.PoolID, req.DeviceID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "msg": "操作失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "成功"})
}

func ListPools(c *gin.Context) {
	pools, err := service.ListPools(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "msg": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1, "data": pools})
}

type addBindingReq struct {
	ServiceID   string `json:"service_id" binding:"required"`
	CallbackURL string `json:"callback_url" binding:"required"`
	DeviceID    string `json:"device_id"`
	PoolID      string `json:"pool_id"`
}

func AddBinding(c *gin.Context) {
	var req addBindingReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "msg": "参数错误"})
		return
	}

	binding := &model.Binding{
		ServiceID:   req.ServiceID,
		CallbackURL: req.CallbackURL,
		DeviceID:    req.DeviceID,
		PoolID:      req.PoolID,
	}
	if err := service.AddBinding(c.Request.Context(), binding); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "msg": "添加失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "成功"})
}

func ListBindings(c *gin.Context) {
	bindings, err := service.ListBindings(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "msg": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1, "data": bindings})
}
