package gorm

import (
	"context"
	"strings"
	"time"

	"github.com/lzy1102/vmq_gateway/server/model"
	storetypes "github.com/lzy1102/vmq_gateway/server/store/types"
	"gorm.io/gorm"
)

type GormOrder struct {
	gorm.Model
	TradeNo     string `gorm:"uniqueIndex;size:64"`
	ServiceID   string `gorm:"size:128"`
	CallbackURL string `gorm:"size:512"`
	Amount      int64
	Status      string `gorm:"size:16;default:pending;index:idx_status_amount"`
	DeviceID    string `gorm:"size:128"`
	CreatedAt   int64
	PaidAt      int64
	ExpireAt    int64 `gorm:"column:expire_at;index"`
}

func (GormOrder) TableName() string { return "orders" }

type GormDevice struct {
	gorm.Model
	DeviceID      string `gorm:"uniqueIndex;size:128"`
	VmqKey        string `gorm:"column:vmq_key;size:128"`
	Status        string `gorm:"size:16;default:offline"`
	LastHeartbeat int64
	WechatQR      string `gorm:"column:wechat_qr;size:256;default:'/qr/wechat.png'"`
	AlipayQR      string `gorm:"column:alipay_qr;size:256;default:'/qr/alipay.png'"`
}

func (GormDevice) TableName() string { return "devices" }

type GormPool struct {
	gorm.Model
	PoolID string `gorm:"uniqueIndex;size:128"`
	Name   string `gorm:"size:128"`
}

func (GormPool) TableName() string { return "pools" }

type GormPoolDevice struct {
	gorm.Model
	PoolID   string `gorm:"uniqueIndex:idx_pool_device;size:128"`
	DeviceID string `gorm:"uniqueIndex:idx_pool_device;size:128"`
}

func (GormPoolDevice) TableName() string { return "pool_devices" }

type GormBinding struct {
	gorm.Model
	ServiceID   string `gorm:"uniqueIndex;size:128"`
	CallbackURL string `gorm:"size:512"`
	DeviceID    string `gorm:"size:128"`
	PoolID      string `gorm:"size:128"`
	APIKey      string `gorm:"size:128"`
}

func (GormBinding) TableName() string { return "bindings" }

type GormDB struct {
	db *gorm.DB
}

func New(db *gorm.DB) *GormDB {
	return &GormDB{db: db}
}

// AutoMigrate 建表
func (g *GormDB) AutoMigrate(ctx context.Context, models ...interface{}) error {
	return g.db.WithContext(ctx).AutoMigrate(models...)
}

// Transaction 事务
func (g *GormDB) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return g.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(ctx)
	})
}

// Create 插入
func (g *GormDB) Create(ctx context.Context, table string, entity interface{}) error {
	return g.db.WithContext(ctx).Table(table).Create(entity).Error
}

func (g *GormDB) Get(ctx context.Context, table string, id string, dest interface{}) error {
	field := idField(table)
	return g.db.WithContext(ctx).Table(table).Where(field+" = ?", id).First(dest).Error
}

func (g *GormDB) Find(ctx context.Context, table string, conditions map[string]interface{}, dest interface{}) error {
	return g.db.WithContext(ctx).Table(table).Where(conditions).Find(dest).Error
}

func (g *GormDB) Update(ctx context.Context, table string, id string, updates map[string]interface{}) error {
	field := idField(table)
	return g.db.WithContext(ctx).Table(table).Where(field+" = ?", id).Updates(updates).Error
}

func (g *GormDB) Delete(ctx context.Context, table string, id string) error {
	field := idField(table)
	return g.db.WithContext(ctx).Table(table).Where(field+" = ?", id).Delete(nil).Error
}

func idField(table string) string {
	switch table {
	case "orders":
		return "trade_no"
	case "devices":
		return "device_id"
	case "pools":
		return "pool_id"
	case "bindings":
		return "service_id"
	case "users":
		return "user_name"
	default:
		return "id"
	}
}

// List 列表
func (g *GormDB) List(ctx context.Context, table string, dest interface{}) error {
	return g.db.WithContext(ctx).Table(table).Find(dest).Error
}

// ListWithPage 分页查询
func (g *GormDB) ListWithPage(ctx context.Context, table string, dest interface{}, page, pageSize int, keyword string, fields []string) (*storetypes.PageResult, error) {
	db := g.db.WithContext(ctx).Table(table)

	if keyword != "" && len(fields) > 0 {
		like := "%" + keyword + "%"
		conditions := make([]string, 0, len(fields))
		args := make([]interface{}, 0, len(fields))
		for _, field := range fields {
			conditions = append(conditions, field+" LIKE ?")
			args = append(args, like)
		}
		where := strings.Join(conditions, " OR ")
		db = db.Where(where, args...)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	if err := db.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(dest).Error; err != nil {
		return nil, err
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &storetypes.PageResult{
		Items:      dest,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (g *GormDB) Claim(ctx context.Context, table string, amount int64, dest interface{}) error {
	now := time.Now().Unix()
	return g.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Table(table).Where("amount = ? AND status = ? AND expire_at > ?", amount, model.StatusPending, now).First(dest).Error; err != nil {
			return err
		}
		return tx.Table(table).Where("amount = ? AND status = ? AND expire_at > ?", amount, model.StatusPending, now).
			Update("status", model.StatusPaid).Error
	})
}

// Upsert 不存在则创建，存在则更新
func (g *GormDB) Upsert(ctx context.Context, table string, key string, value interface{}, update map[string]interface{}) error {
	var count int64
	if err := g.db.WithContext(ctx).Table(table).Where(key+" = ?", value).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		entity := map[string]interface{}{key: value}
		for k, v := range update {
			entity[k] = v
		}
		return g.db.WithContext(ctx).Table(table).Create(entity).Error
	}
	return g.db.WithContext(ctx).Table(table).Where(key+" = ?", value).Updates(update).Error
}

// 按字段查询
func (g *GormDB) FindByField(ctx context.Context, table, field string, value interface{}, dest interface{}) error {
	return g.db.WithContext(ctx).Table(table).Where(field+" = ?", value).First(dest).Error
}

// 按字段更新
func (g *GormDB) UpdateByField(ctx context.Context, table, field string, value interface{}, updates map[string]interface{}) error {
	return g.db.WithContext(ctx).Table(table).Where(field+" = ?", value).Updates(updates).Error
}

// 按字段删除
func (g *GormDB) DeleteByField(ctx context.Context, table, field string, value interface{}) error {
	return g.db.WithContext(ctx).Table(table).Where(field+" = ?", value).Delete(nil).Error
}

// GetByField 按指定字段查询单条
func (g *GormDB) GetByField(ctx context.Context, table, field string, value interface{}, dest interface{}) error {
	return g.db.WithContext(ctx).Table(table).Where(field+" = ?", value).First(dest).Error
}

// JoinQuery 联表查询
func (g *GormDB) JoinQuery(ctx context.Context, dest interface{}, join, on, where string, args ...interface{}) error {
	return g.db.WithContext(ctx).Joins(join+" ON "+on, args...).Where(where).Find(dest).Error
}

// UpdateHeartbeat 更新设备心跳
func (g *GormDB) UpdateHeartbeat(ctx context.Context, deviceID string) error {
	return g.db.WithContext(ctx).Table("devices").Where("device_id = ?", deviceID).
		Updates(map[string]interface{}{"status": model.DeviceOnline, "last_heartbeat": time.Now().Unix()}).Error
}

// GetDeviceByKey 按 key 查询设备
func (g *GormDB) GetDeviceByKey(ctx context.Context, key string, dest interface{}) error {
	return g.db.WithContext(ctx).Table("devices").Where("vmq_key = ?", key).First(dest).Error
}

// AddPoolDevice 添加设备到池子
func (g *GormDB) AddPoolDevice(ctx context.Context, poolID, deviceID string) error {
	return g.db.WithContext(ctx).Table("pool_devices").Create(&map[string]interface{}{
		"pool_id":   poolID,
		"device_id": deviceID,
	}).Error
}

// RemovePoolDevice 从池子移除设备
func (g *GormDB) RemovePoolDevice(ctx context.Context, poolID, deviceID string) error {
	return g.db.WithContext(ctx).Table("pool_devices").
		Where("pool_id = ? AND device_id = ?", poolID, deviceID).Delete(nil).Error
}

func (g *GormDB) RemovePoolDevicesByPool(ctx context.Context, poolID string) error {
	return g.db.WithContext(ctx).Table("pool_devices").
		Where("pool_id = ?", poolID).Delete(nil).Error
}

// GetPoolDeviceIDs 获取池子中的设备 ID 列表
func (g *GormDB) GetPoolDeviceIDs(ctx context.Context, poolID string) ([]string, error) {
	var ids []string
	err := g.db.WithContext(ctx).Table("pool_devices").Where("pool_id = ?", poolID).
		Pluck("device_id", &ids).Error
	return ids, err
}

// GetPoolsByDevice 获取设备所在的池子
func (g *GormDB) GetPoolsByDevice(ctx context.Context, deviceID string, dest interface{}) error {
	return g.db.WithContext(ctx).Table("pools").
		Joins("JOIN pool_devices ON pools.pool_id = pool_devices.pool_id").
		Where("pool_devices.device_id = ?", deviceID).Find(dest).Error
}

func (g *GormDB) ExpireStaleOrders(ctx context.Context) (int64, error) {
	now := time.Now().Unix()
	result := g.db.WithContext(ctx).Table("orders").
		Where("status = ? AND expire_at > 0 AND expire_at <= ?", model.StatusPending, now).
		Update("status", model.StatusExpired)
	return result.RowsAffected, result.Error
}
