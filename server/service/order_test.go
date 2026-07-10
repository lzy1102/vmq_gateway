package service

import (
	"context"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/lzy1102/vmq_gateway/server/config"
	"github.com/lzy1102/vmq_gateway/server/model"
	"github.com/lzy1102/vmq_gateway/server/store"
	gormstore "github.com/lzy1102/vmq_gateway/server/store/gorm"
	gormdriver "gorm.io/gorm"
)

func setupTestDB(t *testing.T) {
	t.Helper()
	db, err := gormdriver.Open(sqlite.Open(":memory:"), &gormdriver.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(
		&gormstore.GormOrder{},
		&gormstore.GormUser{},
		&gormstore.GormDevice{},
		&gormstore.GormPool{},
		&gormstore.GormPoolDevice{},
		&gormstore.GormBinding{},
	); err != nil {
		t.Fatal(err)
	}
	store.DBInstance = gormstore.New(db)
}

func TestCreateOrder(t *testing.T) {
	setupTestDB(t)
	ctx := context.Background()

	config.Packages = map[string]config.Package{
		"small": {Name: "小套餐", Amount: 1000, StreamNumber: 100},
	}

	order, err := CreateOrder(ctx, "test_user", config.Packages["small"], "service1", "http://callback.test")
	if err != nil {
		t.Fatalf("CreateOrder failed: %v", err)
	}

	if order.TradeNo == "" {
		t.Error("TradeNo should not be empty")
	}
	if order.UserName != "test_user" {
		t.Errorf("UserName = %s, want test_user", order.UserName)
	}
	if order.Amount < 1001 || order.Amount > 1019 {
		t.Errorf("Amount = %d, want 1001-1019", order.Amount)
	}
	if order.Status != model.StatusPending {
		t.Errorf("Status = %s, want pending", order.Status)
	}
}

func TestHasPendingAmount(t *testing.T) {
	setupTestDB(t)
	ctx := context.Background()

	config.Packages = map[string]config.Package{
		"small": {Name: "小套餐", Amount: 1000, StreamNumber: 100},
	}

	exists, err := HasPendingAmount(ctx, 1001)
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Error("Should not have pending amount yet")
	}

	_, err = CreateOrder(ctx, "test_user", config.Packages["small"], "service1", "http://callback.test")
	if err != nil {
		t.Fatal(err)
	}

	exists, err = HasPendingAmount(ctx, 1001)
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Error("Should have pending amount")
	}
}

func TestGetOrder(t *testing.T) {
	setupTestDB(t)
	ctx := context.Background()

	config.Packages = map[string]config.Package{
		"small": {Name: "小套餐", Amount: 1000, StreamNumber: 100},
	}

	created, err := CreateOrder(ctx, "test_user", config.Packages["small"], "service1", "http://callback.test")
	if err != nil {
		t.Fatal(err)
	}

	order, err := GetOrder(ctx, created.TradeNo)
	if err != nil {
		t.Fatalf("GetOrder failed: %v", err)
	}

	if order.TradeNo != created.TradeNo {
		t.Errorf("TradeNo = %s, want %s", order.TradeNo, created.TradeNo)
	}
}

func TestHandleCallback(t *testing.T) {
	setupTestDB(t)
	ctx := context.Background()

	config.Packages = map[string]config.Package{
		"small": {Name: "小套餐", Amount: 1000, StreamNumber: 100},
	}

	created, err := CreateOrder(ctx, "test_user", config.Packages["small"], "service1", "http://callback.test")
	if err != nil {
		t.Fatal(err)
	}

	device := &model.Device{
		DeviceID: "device1",
		VmqKey:      "testkey",
		Status:   model.DeviceOnline,
	}

	priceYuan := float64(created.Amount) / 100.0
	order, serviceID, callbackURL, err := HandleCallback(ctx, device, priceYuan)
	if err != nil {
		t.Fatalf("HandleCallback failed: %v", err)
	}

	if order.Status != model.StatusPaid {
		t.Errorf("Status = %s, want paid", order.Status)
	}
	if serviceID != "service1" {
		t.Errorf("ServiceID = %s, want service1", serviceID)
	}
	if callbackURL != "http://callback.test" {
		t.Errorf("CallbackURL = %s, want http://callback.test", callbackURL)
	}

	var users []model.User
	if err := store.DBInstance.Find(ctx, "users", map[string]interface{}{"user_name": "test_user"}, &users); err != nil {
		t.Fatal(err)
	}
	if len(users) == 0 {
		t.Error("User should be created")
	}
}

func TestValidateTimestamp(t *testing.T) {
	now := time.Now().Unix()

	if !ValidateTimestamp(now) {
		t.Error("Current timestamp should be valid")
	}

	if ValidateTimestamp(now - 400) {
		t.Error("Timestamp 400s ago should be invalid")
	}

	if ValidateTimestamp(now + 400) {
		t.Error("Timestamp 400s in future should be invalid")
	}
}

func TestSignAndVerify(t *testing.T) {
	ts := "1234567890"
	key := "testkey"

	sign := SignHeartbeat(ts, key)
	if !VerifyHeartbeatSign(ts, sign, key) {
		t.Error("Heartbeat sign verification failed")
	}

	payType := "1"
	price := "10.00"
	pushSign := SignPush(payType, price, ts, key)
	if !VerifyPushSignWithKey(payType, price, ts, pushSign, key) {
		t.Error("Push sign verification failed")
	}
}
