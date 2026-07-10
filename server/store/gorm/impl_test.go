package gorm

import (
	"context"
	"testing"

	"github.com/lzy1102/vmq_gateway/server/model"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *GormDB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(
		&GormOrder{},
		&GormDevice{},
		&GormPool{},
		&GormPoolDevice{},
		&GormBinding{},
	); err != nil {
		t.Fatal(err)
	}
	return New(db)
}

func TestCreate(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	order := &GormOrder{
		TradeNo:     "V123_1001",
		ServiceID:   "service1",
		CallbackURL: "http://callback.test",
		Amount:      1001,
		Status:      model.StatusPending,
		DeviceID:    "device1",
		CreatedAt:   1234567890,
	}

	if err := db.Create(ctx, "orders", order); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	var found GormOrder
	if err := db.Get(ctx, "orders", "V123_1001", &found); err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if found.TradeNo != "V123_1001" {
		t.Errorf("TradeNo = %s, want V123_1001", found.TradeNo)
	}
	if found.ServiceID != "service1" {
		t.Errorf("ServiceID = %s, want service1", found.ServiceID)
	}
}

func TestFind(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	orders := []GormOrder{
		{TradeNo: "V1_1001", Amount: 1001, Status: model.StatusPending, CreatedAt: 1234567890},
		{TradeNo: "V2_1002", Amount: 1002, Status: model.StatusPaid, CreatedAt: 1234567891},
	}
	for _, o := range orders {
		if err := db.Create(ctx, "orders", &o); err != nil {
			t.Fatal(err)
		}
	}

	var found []GormOrder
	if err := db.Find(ctx, "orders", map[string]interface{}{"status": model.StatusPending}, &found); err != nil {
		t.Fatalf("Find failed: %v", err)
	}

	if len(found) != 1 {
		t.Errorf("Found count = %d, want 1", len(found))
	}
	if found[0].TradeNo != "V1_1001" {
		t.Errorf("TradeNo = %s, want V1_1001", found[0].TradeNo)
	}
}

func TestUpdate(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	order := &GormOrder{
		TradeNo:   "V1_1001",
		Amount:    1001,
		Status:    model.StatusPending,
		CreatedAt: 1234567890,
	}
	if err := db.Create(ctx, "orders", order); err != nil {
		t.Fatal(err)
	}

	if err := db.Update(ctx, "orders", "V1_1001", map[string]interface{}{
		"status":  model.StatusPaid,
		"paid_at": 1234567891,
	}); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	var found GormOrder
	if err := db.Get(ctx, "orders", "V1_1001", &found); err != nil {
		t.Fatal(err)
	}

	if found.Status != model.StatusPaid {
		t.Errorf("Status = %s, want paid", found.Status)
	}
	if found.PaidAt != 1234567891 {
		t.Errorf("PaidAt = %d, want 1234567891", found.PaidAt)
	}
}

func TestDelete(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	order := &GormOrder{
		TradeNo:   "V1_1001",
		Amount:    1001,
		Status:    model.StatusPending,
		CreatedAt: 1234567890,
	}
	if err := db.Create(ctx, "orders", order); err != nil {
		t.Fatal(err)
	}

	if err := db.Delete(ctx, "orders", "V1_1001"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	var found GormOrder
	err := db.Get(ctx, "orders", "V1_1001", &found)
	if err == nil {
		t.Error("Should return error after delete")
	}
}

func TestList(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	orders := []GormOrder{
		{TradeNo: "V1_1001", Amount: 1001, Status: model.StatusPending, CreatedAt: 1234567890},
		{TradeNo: "V2_1002", Amount: 1002, Status: model.StatusPaid, CreatedAt: 1234567891},
	}
	for _, o := range orders {
		if err := db.Create(ctx, "orders", &o); err != nil {
			t.Fatal(err)
		}
	}

	var list []GormOrder
	if err := db.List(ctx, "orders", &list); err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(list) != 2 {
		t.Errorf("List count = %d, want 2", len(list))
	}
}

func TestClaim(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	order := &GormOrder{
		TradeNo:   "V1_1001",
		Amount:    1001,
		Status:    model.StatusPending,
		CreatedAt: 1234567890,
	}
	if err := db.Create(ctx, "orders", order); err != nil {
		t.Fatal(err)
	}

	var claimed GormOrder
	if err := db.Claim(ctx, "orders", 1001, &claimed); err != nil {
		t.Fatalf("Claim failed: %v", err)
	}

	if claimed.TradeNo != "V1_1001" {
		t.Errorf("TradeNo = %s, want V1_1001", claimed.TradeNo)
	}

	var found GormOrder
	if err := db.Get(ctx, "orders", "V1_1001", &found); err != nil {
		t.Fatal(err)
	}
	if found.Status != model.StatusPaid {
		t.Errorf("Status after claim = %s, want paid", found.Status)
	}
}

func TestTransaction(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	err := db.Transaction(ctx, func(ctx context.Context) error {
		return nil
	})
	if err != nil {
		t.Fatalf("Transaction failed: %v", err)
	}
}

func TestDeviceOperations(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	device := &GormDevice{
		DeviceID:      "device1",
		VmqKey:        "key123",
		Status:        model.DeviceOffline,
		LastHeartbeat: 0,
	}
	if err := db.Create(ctx, "devices", device); err != nil {
		t.Fatal(err)
	}

	var found GormDevice
	if err := db.GetDeviceByKey(ctx, "key123", &found); err != nil {
		t.Fatal(err)
	}
	if found.DeviceID != "device1" {
		t.Errorf("DeviceID = %s, want device1", found.DeviceID)
	}

	if err := db.UpdateHeartbeat(ctx, "device1"); err != nil {
		t.Fatal(err)
	}

	var updated GormDevice
	if err := db.GetByField(ctx, "devices", "device_id", "device1", &updated); err != nil {
		t.Fatal(err)
	}
	if updated.Status != model.DeviceOnline {
		t.Errorf("Status = %s, want online", updated.Status)
	}
}

func TestPoolDeviceOperations(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	pool := &GormPool{
		PoolID: "pool1",
		Name:   "测试池",
	}
	if err := db.Create(ctx, "pools", pool); err != nil {
		t.Fatal(err)
	}

	if err := db.AddPoolDevice(ctx, "pool1", "device1"); err != nil {
		t.Fatal(err)
	}
	if err := db.AddPoolDevice(ctx, "pool1", "device2"); err != nil {
		t.Fatal(err)
	}

	ids, err := db.GetPoolDeviceIDs(ctx, "pool1")
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 2 {
		t.Errorf("Device count = %d, want 2", len(ids))
	}

	if err := db.RemovePoolDevice(ctx, "pool1", "device1"); err != nil {
		t.Fatal(err)
	}

	ids, err = db.GetPoolDeviceIDs(ctx, "pool1")
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 1 {
		t.Errorf("Device count = %d, want 1", len(ids))
	}
}
