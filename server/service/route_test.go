package service

import (
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/lzy1102/vmq_gateway/server/model"
	"github.com/lzy1102/vmq_gateway/server/store"
	gormstore "github.com/lzy1102/vmq_gateway/server/store/gorm"
	gormdriver "gorm.io/gorm"
)

func setupTestDBForRoute(t *testing.T) {
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

func TestIdentifyDevice(t *testing.T) {
	setupTestDBForRoute(t)
	ctx := context.Background()

	device := &model.Device{
		DeviceID: "device1",
		VmqKey:   "testkey123",
		Status:   model.DeviceOffline,
	}
	if err := store.DBInstance.Create(ctx, "devices", device); err != nil {
		t.Fatal(err)
	}

	found, err := IdentifyDevice(ctx, "testkey123")
	if err != nil {
		t.Fatalf("IdentifyDevice failed: %v", err)
	}

	if found.DeviceID != "device1" {
		t.Errorf("DeviceID = %s, want device1", found.DeviceID)
	}
}

func TestIdentifyDevice_NotFound(t *testing.T) {
	setupTestDBForRoute(t)
	ctx := context.Background()

	_, err := IdentifyDevice(ctx, "nonexistent")
	if err == nil {
		t.Error("Should return error for nonexistent device")
	}
}

func TestListDevices(t *testing.T) {
	setupTestDBForRoute(t)
	ctx := context.Background()

	devices := []model.Device{
		{DeviceID: "device1", VmqKey: "key1", Status: model.DeviceOnline},
		{DeviceID: "device2", VmqKey: "key2", Status: model.DeviceOffline},
	}
	for _, d := range devices {
		if err := store.DBInstance.Create(ctx, "devices", &d); err != nil {
			t.Fatal(err)
		}
	}

	list, err := ListDevices(ctx)
	if err != nil {
		t.Fatalf("ListDevices failed: %v", err)
	}

	if len(list) != 2 {
		t.Errorf("Device count = %d, want 2", len(list))
	}
}

func TestAddDevice(t *testing.T) {
	setupTestDBForRoute(t)
	ctx := context.Background()

	device := &model.Device{
		DeviceID: "device1",
		VmqKey:   "key1",
		Status:   model.DeviceOffline,
	}
	if err := AddDevice(ctx, device); err != nil {
		t.Fatalf("AddDevice failed: %v", err)
	}

	list, err := ListDevices(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Errorf("Device count = %d, want 1", len(list))
	}
}

func TestAddPool(t *testing.T) {
	setupTestDBForRoute(t)
	ctx := context.Background()

	pool := &model.Pool{
		PoolID:    "pool1",
		Name:      "测试池",
		DeviceIDs: []string{},
	}
	if err := AddPool(ctx, pool); err != nil {
		t.Fatalf("AddPool failed: %v", err)
	}

	list, err := ListPools(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Errorf("Pool count = %d, want 1", len(list))
	}
}

func TestAddDeviceToPool(t *testing.T) {
	setupTestDBForRoute(t)
	ctx := context.Background()

	pool := &model.Pool{
		PoolID:    "pool1",
		Name:      "测试池",
		DeviceIDs: []string{},
	}
	if err := AddPool(ctx, pool); err != nil {
		t.Fatal(err)
	}

	device := &model.Device{
		DeviceID: "device1",
		VmqKey:   "key1",
		Status:   model.DeviceOnline,
	}
	if err := AddDevice(ctx, device); err != nil {
		t.Fatal(err)
	}

	if err := AddDeviceToPool(ctx, "pool1", "device1"); err != nil {
		t.Fatalf("AddDeviceToPool failed: %v", err)
	}

	var pools []model.Pool
	if err := store.DBInstance.GetPoolsByDevice(ctx, "device1", &pools); err != nil {
		t.Fatal(err)
	}
	if len(pools) != 1 {
		t.Errorf("Pool count for device = %d, want 1", len(pools))
	}
}

func TestAddBinding(t *testing.T) {
	setupTestDBForRoute(t)
	ctx := context.Background()

	binding := &model.Binding{
		ServiceID:   "service1",
		CallbackURL: "http://callback.test",
		DeviceID:    "device1",
		PoolID:      "",
	}
	if err := AddBinding(ctx, binding); err != nil {
		t.Fatalf("AddBinding failed: %v", err)
	}

	list, err := ListBindings(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Errorf("Binding count = %d, want 1", len(list))
	}
}

func TestRouteCallback_DirectBinding(t *testing.T) {
	setupTestDBForRoute(t)
	ctx := context.Background()

	device := &model.Device{
		DeviceID: "device1",
		VmqKey:   "key1",
		Status:   model.DeviceOnline,
	}
	if err := AddDevice(ctx, device); err != nil {
		t.Fatal(err)
	}

	binding := &model.Binding{
		ServiceID:   "service1",
		CallbackURL: "http://callback.test",
		DeviceID:    "device1",
		PoolID:      "",
	}
	if err := AddBinding(ctx, binding); err != nil {
		t.Fatal(err)
	}

	order := &model.Order{
		ServiceID:   "",
		CallbackURL: "",
	}

	serviceID, callbackURL := RouteCallback(ctx, device, order)
	if serviceID != "service1" {
		t.Errorf("ServiceID = %s, want service1", serviceID)
	}
	if callbackURL != "http://callback.test" {
		t.Errorf("CallbackURL = %s, want http://callback.test", callbackURL)
	}
}

func TestRouteCallback_PoolBinding(t *testing.T) {
	setupTestDBForRoute(t)
	ctx := context.Background()

	device := &model.Device{
		DeviceID: "device1",
		VmqKey:   "key1",
		Status:   model.DeviceOnline,
	}
	if err := AddDevice(ctx, device); err != nil {
		t.Fatal(err)
	}

	pool := &model.Pool{
		PoolID:    "pool1",
		Name:      "测试池",
		DeviceIDs: []string{},
	}
	if err := AddPool(ctx, pool); err != nil {
		t.Fatal(err)
	}

	if err := AddDeviceToPool(ctx, "pool1", "device1"); err != nil {
		t.Fatal(err)
	}

	binding := &model.Binding{
		ServiceID:   "service2",
		CallbackURL: "http://callback2.test",
		DeviceID:    "",
		PoolID:      "pool1",
	}
	if err := AddBinding(ctx, binding); err != nil {
		t.Fatal(err)
	}

	order := &model.Order{
		ServiceID:   "",
		CallbackURL: "",
	}

	serviceID, callbackURL := RouteCallback(ctx, device, order)
	if serviceID != "service2" {
		t.Errorf("ServiceID = %s, want service2", serviceID)
	}
	if callbackURL != "http://callback2.test" {
		t.Errorf("CallbackURL = %s, want http://callback2.test", callbackURL)
	}
}

func TestRouteCallback_OrderDirect(t *testing.T) {
	setupTestDBForRoute(t)
	ctx := context.Background()

	device := &model.Device{
		DeviceID: "device1",
		VmqKey:   "key1",
		Status:   model.DeviceOnline,
	}
	if err := AddDevice(ctx, device); err != nil {
		t.Fatal(err)
	}

	order := &model.Order{
		ServiceID:   "service3",
		CallbackURL: "http://callback3.test",
	}

	serviceID, callbackURL := RouteCallback(ctx, device, order)
	if serviceID != "service3" {
		t.Errorf("ServiceID = %s, want service3", serviceID)
	}
	if callbackURL != "http://callback3.test" {
		t.Errorf("CallbackURL = %s, want http://callback3.test", callbackURL)
	}
}
