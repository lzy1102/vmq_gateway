package service

import (
	"context"
	"fmt"

	"github.com/lzy1102/vmq_gateway/server/model"
	"github.com/lzy1102/vmq_gateway/server/store"
)

func IdentifyDevice(ctx context.Context, key string) (*model.Device, error) {
	var device model.Device
	if err := store.DBInstance.GetDeviceByKey(ctx, key, &device); err != nil {
		return nil, fmt.Errorf("未知设备: key=%s", key)
	}
	store.DBInstance.UpdateHeartbeat(ctx, device.DeviceID)
	return &device, nil
}

func UpdateHeartbeat(ctx context.Context, deviceID string) error {
	return store.DBInstance.UpdateHeartbeat(ctx, deviceID)
}

func RouteCallback(ctx context.Context, device *model.Device, order *model.RechargeOrder) (string, string) {
	if order.ServiceID != "" && order.CallbackURL != "" {
		return order.ServiceID, order.CallbackURL
	}

	// 查设备绑定
	var bindings []model.Binding
	if err := store.DBInstance.Find(ctx, "bindings", map[string]interface{}{"device_id": device.DeviceID}, &bindings); err == nil && len(bindings) > 0 {
		binding := bindings[0]
		if binding.ServiceID != "" {
			return binding.ServiceID, binding.CallbackURL
		}
	}

	// 查池子绑定
	var pools []model.Pool
	if err := store.DBInstance.GetPoolsByDevice(ctx, device.DeviceID, &pools); err == nil {
		for _, pool := range pools {
			var poolBindings []model.Binding
			if err := store.DBInstance.Find(ctx, "bindings", map[string]interface{}{"pool_id": pool.PoolID}, &poolBindings); err == nil && len(poolBindings) > 0 {
				binding := poolBindings[0]
				if binding.ServiceID != "" {
					return binding.ServiceID, binding.CallbackURL
				}
			}
		}
	}

	return "", ""
}

// ListDevices 列出所有设备
func ListDevices(ctx context.Context) ([]model.Device, error) {
	var devices []model.Device
	if err := store.DBInstance.List(ctx, "devices", &devices); err != nil {
		return nil, err
	}
	return devices, nil
}

// ListPools 列出所有池子
func ListPools(ctx context.Context) ([]model.Pool, error) {
	var pools []model.Pool
	if err := store.DBInstance.List(ctx, "pools", &pools); err != nil {
		return nil, err
	}
	return pools, nil
}

// ListBindings 列出所有绑定
func ListBindings(ctx context.Context) ([]model.Binding, error) {
	var bindings []model.Binding
	if err := store.DBInstance.List(ctx, "bindings", &bindings); err != nil {
		return nil, err
	}
	return bindings, nil
}

// AddDevice 添加设备
func AddDevice(ctx context.Context, device *model.Device) error {
	return store.DBInstance.Create(ctx, "devices", device)
}

// AddPool 添加池子
func AddPool(ctx context.Context, pool *model.Pool) error {
	return store.DBInstance.Create(ctx, "pools", pool)
}

// AddDeviceToPool 添加设备到池子
func AddDeviceToPool(ctx context.Context, poolID, deviceID string) error {
	return store.DBInstance.AddPoolDevice(ctx, poolID, deviceID)
}

// AddBinding 添加绑定
func AddBinding(ctx context.Context, binding *model.Binding) error {
	return store.DBInstance.Create(ctx, "bindings", binding)
}
