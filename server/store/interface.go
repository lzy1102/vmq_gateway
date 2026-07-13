package store

import (
	"context"

	"github.com/lzy1102/vmq_gateway/server/store/types"
)

type DB interface {
	AutoMigrate(ctx context.Context, models ...interface{}) error
	Transaction(ctx context.Context, fn func(ctx context.Context) error) error

	Create(ctx context.Context, table string, entity interface{}) error
	Get(ctx context.Context, table string, id string, dest interface{}) error
	Find(ctx context.Context, table string, conditions map[string]interface{}, dest interface{}) error
	Update(ctx context.Context, table string, id string, updates map[string]interface{}) error
	Delete(ctx context.Context, table string, id string) error
	List(ctx context.Context, table string, dest interface{}) error
	ListWithPage(ctx context.Context, table string, dest interface{}, page, pageSize int, keyword string, fields []string) (*types.PageResult, error)

	Claim(ctx context.Context, table string, amount int64, dest interface{}) error
	Upsert(ctx context.Context, table string, key string, value interface{}, update map[string]interface{}) error

	FindByField(ctx context.Context, table, field string, value interface{}, dest interface{}) error
	UpdateByField(ctx context.Context, table, field string, value interface{}, updates map[string]interface{}) error
	DeleteByField(ctx context.Context, table, field string, value interface{}) error
	GetByField(ctx context.Context, table, field string, value interface{}, dest interface{}) error

	GetDeviceByKey(ctx context.Context, key string, dest interface{}) error
	UpdateHeartbeat(ctx context.Context, deviceID string) error
	AddPoolDevice(ctx context.Context, poolID, deviceID string) error
	RemovePoolDevice(ctx context.Context, poolID, deviceID string) error
	RemovePoolDevicesByPool(ctx context.Context, poolID string) error
	GetPoolDeviceIDs(ctx context.Context, poolID string) ([]string, error)
	GetPoolsByDevice(ctx context.Context, deviceID string, dest interface{}) error
	ExpireStaleOrders(ctx context.Context) (int64, error)
	ExpireOfflineDevices(ctx context.Context, thresholdSec int64) (int64, error)
}

type PageResult = types.PageResult

var DBInstance DB
