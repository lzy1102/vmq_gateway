package handler

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lzy1102/vmq_gateway/server/model"
	"github.com/lzy1102/vmq_gateway/server/service"
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
	keyword := c.Query("keyword")
	page := 1
	pageSize := 10
	if p := c.Query("page"); p != "" {
		fmt.Sscanf(p, "%d", &page)
	}
	if ps := c.Query("page_size"); ps != "" {
		fmt.Sscanf(ps, "%d", &pageSize)
	}
	result, err := service.ListDevicesWithPage(c.Request.Context(), keyword, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "msg": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1, "data": gin.H{"items": result.Items, "total": result.Total, "page": result.Page, "page_size": result.PageSize, "total_pages": result.TotalPages}})
}

type deleteDeviceReq struct {
	DeviceID string `json:"device_id" binding:"required"`
}

func DeleteDevice(c *gin.Context) {
	var req deleteDeviceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "msg": "参数错误"})
		return
	}
	if err := service.DeleteDevice(c.Request.Context(), req.DeviceID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "msg": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "删除成功"})
}

type updateDeviceReq struct {
	DeviceID string `json:"device_id" binding:"required"`
	Key      string `json:"key"`
}

func UpdateDevice(c *gin.Context) {
	var req updateDeviceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "msg": "参数错误"})
		return
	}
 updates := map[string]interface{}{}
	if req.Key != "" {
		updates["vmq_key"] = req.Key
	}
	if err := service.UpdateDevice(c.Request.Context(), req.DeviceID, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "msg": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "更新成功"})
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

type deletePoolReq struct {
	PoolID string `json:"pool_id" binding:"required"`
}

func DeletePool(c *gin.Context) {
	var req deletePoolReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "msg": "参数错误"})
		return
	}
	if err := service.DeletePool(c.Request.Context(), req.PoolID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "msg": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "删除成功"})
}

type removeDeviceFromPoolReq struct {
	PoolID   string `json:"pool_id" binding:"required"`
	DeviceID string `json:"device_id" binding:"required"`
}

func RemoveDeviceFromPool(c *gin.Context) {
	var req removeDeviceFromPoolReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "msg": "参数错误"})
		return
	}
	if err := service.RemoveDeviceFromPool(c.Request.Context(), req.PoolID, req.DeviceID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "msg": "移除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "删除成功"})
}

func ListPools(c *gin.Context) {
	keyword := c.Query("keyword")
	page := 1
	pageSize := 10
	if p := c.Query("page"); p != "" {
		fmt.Sscanf(p, "%d", &page)
	}
	if ps := c.Query("page_size"); ps != "" {
		fmt.Sscanf(ps, "%d", &pageSize)
	}
	result, err := service.ListPoolsWithPage(c.Request.Context(), keyword, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "msg": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1, "data": gin.H{"items": result.Items, "total": result.Total, "page": result.Page, "page_size": result.PageSize, "total_pages": result.TotalPages}})
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

	apiKey := generateKey()
	binding := &model.Binding{
		ServiceID:   req.ServiceID,
		CallbackURL: req.CallbackURL,
		DeviceID:    req.DeviceID,
		PoolID:      req.PoolID,
		APIKey:      apiKey,
	}
	if err := service.AddBinding(c.Request.Context(), binding); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "msg": "添加失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "成功", "data": gin.H{"api_key": apiKey}})
}

type updateBindingReq struct {
	ServiceID   string `json:"service_id" binding:"required"`
	CallbackURL string `json:"callback_url" binding:"required"`
	DeviceID    *string `json:"device_id"`
	PoolID      *string `json:"pool_id"`
}

func UpdateBinding(c *gin.Context) {
	var req updateBindingReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "msg": "参数错误"})
		return
	}
	updates := map[string]interface{}{
		"callback_url": req.CallbackURL,
	}
	if req.DeviceID != nil {
		updates["device_id"] = *req.DeviceID
	}
	if req.PoolID != nil {
		updates["pool_id"] = *req.PoolID
	}
	if err := service.UpdateBinding(c.Request.Context(), req.ServiceID, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "msg": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "更新成功"})
}

type deleteBindingReq struct {
	ServiceID string `json:"service_id" binding:"required"`
}

func DeleteBinding(c *gin.Context) {
	var req deleteBindingReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "msg": "参数错误"})
		return
	}
	if err := service.DeleteBinding(c.Request.Context(), req.ServiceID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "msg": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "删除成功"})
}

func ListBindings(c *gin.Context) {
	keyword := c.Query("keyword")
	page := 1
	pageSize := 10
	if p := c.Query("page"); p != "" {
		fmt.Sscanf(p, "%d", &page)
	}
	if ps := c.Query("page_size"); ps != "" {
		fmt.Sscanf(ps, "%d", &pageSize)
	}
	result, err := service.ListBindingsWithPage(c.Request.Context(), keyword, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "msg": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1, "data": gin.H{"items": result.Items, "total": result.Total, "page": result.Page, "page_size": result.PageSize, "total_pages": result.TotalPages}})
}
