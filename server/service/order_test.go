package service

import (
	"context"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
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

	addTestDevice(t, ctx)

	order, _, err := CreateOrder(ctx, 1000, "service1", "http://callback.test")
	if err != nil {
		t.Fatalf("CreateOrder failed: %v", err)
	}

	if order.TradeNo == "" {
		t.Error("TradeNo should not be empty")
	}
	if order.Amount < 1001 || order.Amount > 1019 {
		t.Errorf("Amount = %d, want 1001-1019", order.Amount)
	}
	if order.Status != model.StatusPending {
		t.Errorf("Status = %s, want pending", order.Status)
	}
	if order.ServiceID != "service1" {
		t.Errorf("ServiceID = %s, want service1", order.ServiceID)
	}
}

func addTestDevice(t *testing.T, ctx context.Context) {
	t.Helper()
	device := &model.Device{
		DeviceID: "device1",
		VmqKey:   "testkey",
		Status:   model.DeviceOffline,
	}
	if err := store.DBInstance.Create(ctx, "devices", device); err != nil {
		t.Fatal(err)
	}
}

func TestHasPendingAmount(t *testing.T) {
	setupTestDB(t)
	ctx := context.Background()

	addTestDevice(t, ctx)

	exists, err := HasPendingAmount(ctx, 1001)
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Error("Should not have pending amount yet")
	}

	_, _, err = CreateOrder(ctx, 1000, "service1", "http://callback.test")
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

	addTestDevice(t, ctx)

	created, _, err := CreateOrder(ctx, 1000, "service1", "http://callback.test")
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

	addTestDevice(t, ctx)

	created, _, err := CreateOrder(ctx, 1000, "service1", "http://callback.test")
	if err != nil {
		t.Fatal(err)
	}

	device := &model.Device{
		DeviceID: "device1",
		VmqKey:   "testkey",
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
