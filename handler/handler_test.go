package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/gin-gonic/gin"
	"github.com/lzy1102/vmq_gateway/config"
	"github.com/lzy1102/vmq_gateway/model"
	"github.com/lzy1102/vmq_gateway/service"
	"github.com/lzy1102/vmq_gateway/store"
	"github.com/lzy1102/vmq_gateway/store/gorm"
	gormdriver "gorm.io/gorm"
)

func setupTestEnv(t *testing.T) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, err := gormdriver.Open(sqlite.Open(":memory:"), &gormdriver.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(
		&gorm.GormOrder{},
		&gorm.GormUser{},
		&gorm.GormDevice{},
		&gorm.GormPool{},
		&gorm.GormPoolDevice{},
		&gorm.GormBinding{},
	); err != nil {
		t.Fatal(err)
	}
	store.DBInstance = gorm.New(db)

	config.Packages = map[string]config.Package{
		"small": {Name: "小套餐", Amount: 1000, StreamNumber: 100},
	}
}

func TestCreateRechargeOrder(t *testing.T) {
	setupTestEnv(t)

	router := gin.New()
	router.POST("/api/recharge/vmpay", CreateRechargeOrder)

	body := createOrderReq{
		UserName:    "test_user",
		Package:     "small",
		ServiceID:   "service1",
		CallbackURL: "http://callback.test",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequestWithContext(context.Background(), "POST", "/api/recharge/vmpay", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}

	if resp["code"] != float64(1) {
		t.Errorf("Code = %v, want 1", resp["code"])
	}
}

func TestCreateRechargeOrder_InvalidPackage(t *testing.T) {
	setupTestEnv(t)

	router := gin.New()
	router.POST("/api/recharge/vmpay", CreateRechargeOrder)

	body := createOrderReq{
		UserName:    "test_user",
		Package:     "invalid",
		ServiceID:   "service1",
		CallbackURL: "http://callback.test",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequestWithContext(context.Background(), "POST", "/api/recharge/vmpay", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestQueryOrderStatus(t *testing.T) {
	setupTestEnv(t)

	config.Packages = map[string]config.Package{
		"small": {Name: "小套餐", Amount: 1000, StreamNumber: 100},
	}

	order, err := service.CreateOrder(context.Background(), "test_user", config.Packages["small"], "service1", "http://callback.test")
	if err != nil {
		t.Fatal(err)
	}

	router := gin.New()
	router.GET("/api/recharge/vmpay-status", QueryOrderStatus)

	req, _ := http.NewRequestWithContext(context.Background(), "GET", "/api/recharge/vmpay-status?trade_no="+order.TradeNo, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}

	if resp["code"] != float64(1) {
		t.Errorf("Code = %v, want 1", resp["code"])
	}
}

func TestQueryOrderStatus_NotFound(t *testing.T) {
	setupTestEnv(t)

	router := gin.New()
	router.GET("/api/recharge/vmpay-status", QueryOrderStatus)

	req, _ := http.NewRequestWithContext(context.Background(), "GET", "/api/recharge/vmpay-status?trade_no=nonexistent", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestHeartbeat(t *testing.T) {
	setupTestEnv(t)

	device := &model.Device{
		DeviceID: "device1",
		VmqKey:      "testkey",
		Status:   model.DeviceOffline,
	}
	if err := store.DBInstance.Create(context.Background(), "devices", device); err != nil {
		t.Fatal(err)
	}

	tsMs := time.Now().UnixMilli()
	tsStr := strconv.FormatInt(tsMs, 10)
	sign := service.SignHeartbeat(tsStr, "testkey")

	router := gin.New()
	router.GET("/appHeart", Heartbeat)

	req, _ := http.NewRequestWithContext(context.Background(), "GET", "/appHeart?t="+tsStr+"&sign="+sign, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}

	if resp["code"] != float64(1) {
		t.Errorf("Code = %v, want 1", resp["code"])
	}
}

func TestAddDevice(t *testing.T) {
	setupTestEnv(t)

	router := gin.New()
	router.POST("/admin/device", AddDevice)

	body := addDeviceReq{
		DeviceID: "device1",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequestWithContext(context.Background(), "POST", "/admin/device", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}

	if resp["code"] != float64(1) {
		t.Errorf("Code = %v, want 1", resp["code"])
	}
}

func TestListDevices(t *testing.T) {
	setupTestEnv(t)

	device := &model.Device{
		DeviceID: "device1",
		VmqKey:      "key1",
		Status:   model.DeviceOnline,
	}
	if err := store.DBInstance.Create(context.Background(), "devices", device); err != nil {
		t.Fatal(err)
	}

	router := gin.New()
	router.GET("/admin/devices", ListDevices)

	req, _ := http.NewRequestWithContext(context.Background(), "GET", "/admin/devices", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}

	if resp["code"] != float64(1) {
		t.Errorf("Code = %v, want 1", resp["code"])
	}
}

func TestAddPool(t *testing.T) {
	setupTestEnv(t)

	router := gin.New()
	router.POST("/admin/pool", AddPool)

	body := addPoolReq{
		PoolID: "pool1",
		Name:   "测试池",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequestWithContext(context.Background(), "POST", "/admin/pool", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}

	if resp["code"] != float64(1) {
		t.Errorf("Code = %v, want 1", resp["code"])
	}
}

func TestAddBinding(t *testing.T) {
	setupTestEnv(t)

	router := gin.New()
	router.POST("/admin/binding", AddBinding)

	body := addBindingReq{
		ServiceID:   "service1",
		CallbackURL: "http://callback.test",
		DeviceID:    "device1",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequestWithContext(context.Background(), "POST", "/admin/binding", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}

	if resp["code"] != float64(1) {
		t.Errorf("Code = %v, want 1", resp["code"])
	}
}
