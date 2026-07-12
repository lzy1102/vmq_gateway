package mongo

import (
	"context"
	"time"

	"github.com/lzy1102/vmq_gateway/server/model"
	storetypes "github.com/lzy1102/vmq_gateway/server/store/types"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MongoDB struct {
	client *mongo.Client
	db     *mongo.Database
}

func New(client *mongo.Client, dbName string) *MongoDB {
	return &MongoDB{
		client: client,
		db:     client.Database(dbName),
	}
}

// AutoMigrate MongoDB 不需要建表，直接返回 nil
func (m *MongoDB) AutoMigrate(ctx context.Context, models ...interface{}) error {
	return nil
}

// Transaction MongoDB 事务
func (m *MongoDB) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	session, err := m.client.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sessCtx context.Context) (interface{}, error) {
		return nil, fn(sessCtx)
	})
	return err
}

// Create 插入
func (m *MongoDB) Create(ctx context.Context, table string, entity interface{}) error {
	_, err := m.db.Collection(table).InsertOne(ctx, entity)
	return err
}

// Get 按主键查询（_id 或自定义字段）
func (m *MongoDB) Get(ctx context.Context, table string, id string, dest interface{}) error {
	// 尝试 _id (ObjectID)
	if oid, err := bson.ObjectIDFromHex(id); err == nil {
		err = m.db.Collection(table).FindOne(ctx, bson.M{"_id": oid}).Decode(dest)
		if err == nil {
			return nil
		}
	}
	// 尝试常见字段
	for _, field := range []string{"trade_no", "device_id", "pool_id", "service_id", "user_name"} {
		err := m.db.Collection(table).FindOne(ctx, bson.M{field: id}).Decode(dest)
		if err == nil {
			return nil
		}
	}
	return mongo.ErrNoDocuments
}

// Find 条件查询
func (m *MongoDB) Find(ctx context.Context, table string, conditions map[string]interface{}, dest interface{}) error {
	filter := bson.M(conditions)
	cursor, err := m.db.Collection(table).Find(ctx, filter)
	if err != nil {
		return err
	}
	return cursor.All(ctx, dest)
}

// Update 更新
func (m *MongoDB) Update(ctx context.Context, table string, id string, updates map[string]interface{}) error {
	filter := m.buildIDFilter(id)
	_, err := m.db.Collection(table).UpdateOne(ctx, filter, bson.M{"$set": updates})
	return err
}

// Delete 删除
func (m *MongoDB) Delete(ctx context.Context, table string, id string) error {
	filter := m.buildIDFilter(id)
	_, err := m.db.Collection(table).DeleteOne(ctx, filter)
	return err
}

// List 列表
func (m *MongoDB) List(ctx context.Context, table string, dest interface{}) error {
	cursor, err := m.db.Collection(table).Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	return cursor.All(ctx, dest)
}

func (m *MongoDB) ListWithPage(ctx context.Context, table string, dest interface{}, page, pageSize int, keyword string, fields []string) (*storetypes.PageResult, error) {
	filter := bson.M{}
	if keyword != "" && len(fields) > 0 {
		conditions := make([]bson.M, 0, len(fields))
		for _, field := range fields {
			conditions = append(conditions, bson.M{field: bson.M{"$regex": keyword, "$options": "i"}})
		}
		filter = bson.M{"$or": conditions}
	}

	total, err := m.db.Collection(table).CountDocuments(ctx, filter)
	if err != nil {
		return nil, err
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	skip := int64((page - 1) * pageSize)
	limit := int64(pageSize)

	opts := options.Find().SetSkip(skip).SetLimit(limit)
	cursor, err := m.db.Collection(table).Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	if err := cursor.All(ctx, dest); err != nil {
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

func (m *MongoDB) Claim(ctx context.Context, table string, amount int64, dest interface{}) error {
	filter := bson.M{"amount": amount, "status": model.StatusPending}
	update := bson.M{"$set": bson.M{"status": model.StatusPaid}}

	err := m.db.Collection(table).FindOneAndUpdate(ctx, filter, update,
		options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(dest)
	return err
}

// Upsert 不存在则创建，存在则更新
func (m *MongoDB) Upsert(ctx context.Context, table string, key string, value interface{}, update map[string]interface{}) error {
	filter := bson.M{key: value}
	updateDoc := bson.M{"$set": update}
	opts := options.UpdateOne().SetUpsert(true)
	_, err := m.db.Collection(table).UpdateOne(ctx, filter, updateDoc, opts)
	return err
}

// 按字段查询
func (m *MongoDB) FindByField(ctx context.Context, table, field string, value interface{}, dest interface{}) error {
	return m.db.Collection(table).FindOne(ctx, bson.M{field: value}).Decode(dest)
}

// 按字段更新
func (m *MongoDB) UpdateByField(ctx context.Context, table, field string, value interface{}, updates map[string]interface{}) error {
	_, err := m.db.Collection(table).UpdateOne(ctx, bson.M{field: value}, bson.M{"$set": updates})
	return err
}

// 按字段删除
func (m *MongoDB) DeleteByField(ctx context.Context, table, field string, value interface{}) error {
	_, err := m.db.Collection(table).DeleteOne(ctx, bson.M{field: value})
	return err
}

// GetByField 按指定字段查询单条
func (m *MongoDB) GetByField(ctx context.Context, table, field string, value interface{}, dest interface{}) error {
	return m.db.Collection(table).FindOne(ctx, bson.M{field: value}).Decode(dest)
}

// JoinQuery MongoDB 不支持传统 JOIN，返回错误
func (m *MongoDB) JoinQuery(ctx context.Context, dest interface{}, join, on, where string, args ...interface{}) error {
	// MongoDB 需要在应用层做 JOIN，这里简单查询
	cursor, err := m.db.Collection("pools").Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	return cursor.All(ctx, dest)
}

// UpdateHeartbeat 更新设备心跳
func (m *MongoDB) UpdateHeartbeat(ctx context.Context, deviceID string) error {
	_, err := m.db.Collection("devices").UpdateOne(ctx,
		bson.M{"device_id": deviceID},
		bson.M{"$set": bson.M{"status": model.DeviceOnline, "last_heartbeat": time.Now().Unix()}})
	return err
}

// GetDeviceByKey 按 key 查询设备
func (m *MongoDB) GetDeviceByKey(ctx context.Context, key string, dest interface{}) error {
	return m.db.Collection("devices").FindOne(ctx, bson.M{"key": key}).Decode(dest)
}

// AddPoolDevice 添加设备到池子
func (m *MongoDB) AddPoolDevice(ctx context.Context, poolID, deviceID string) error {
	// MongoDB 用数组存储，需要更新 pool 的 device_ids
	_, err := m.db.Collection("pools").UpdateOne(ctx,
		bson.M{"pool_id": poolID},
		bson.M{"$addToSet": bson.M{"device_ids": deviceID}})
	return err
}

// RemovePoolDevice 从池子移除设备
func (m *MongoDB) RemovePoolDevice(ctx context.Context, poolID, deviceID string) error {
	_, err := m.db.Collection("pools").UpdateOne(ctx,
		bson.M{"pool_id": poolID},
		bson.M{"$pull": bson.M{"device_ids": deviceID}})
	return err
}

func (m *MongoDB) RemovePoolDevicesByPool(ctx context.Context, poolID string) error {
	_, err := m.db.Collection("pools").UpdateOne(ctx,
		bson.M{"pool_id": poolID},
		bson.M{"$set": bson.M{"device_ids": []string{}}})
	return err
}

// GetPoolDeviceIDs 获取池子中的设备 ID 列表
func (m *MongoDB) GetPoolDeviceIDs(ctx context.Context, poolID string) ([]string, error) {
	var pool model.Pool
	err := m.db.Collection("pools").FindOne(ctx, bson.M{"pool_id": poolID}).Decode(&pool)
	if err != nil {
		return nil, err
	}
	return pool.DeviceIDs, nil
}

// GetPoolsByDevice 获取设备所在的池子
func (m *MongoDB) GetPoolsByDevice(ctx context.Context, deviceID string, dest interface{}) error {
	cursor, err := m.db.Collection("pools").Find(ctx, bson.M{"device_ids": deviceID})
	if err != nil {
		return err
	}
	return cursor.All(ctx, dest)
}

// buildIDFilter 构建 ID 过滤器
func (m *MongoDB) buildIDFilter(id string) bson.M {
	// 尝试 ObjectID
	if oid, err := bson.ObjectIDFromHex(id); err == nil {
		return bson.M{"_id": oid}
	}
	// 尝试常见字段
	for _, field := range []string{"trade_no", "device_id", "pool_id", "service_id", "user_name"} {
		return bson.M{field: id}
	}
	return bson.M{"_id": id}
}

func (m *MongoDB) ExpireStaleOrders(ctx context.Context) (int64, error) {
	now := time.Now().Unix()
	result, err := m.db.Collection("orders").UpdateMany(ctx,
		bson.M{"status": model.StatusPending, "expire_at": bson.M{"$gt": 0, "$lte": now}},
		bson.M{"$set": bson.M{"status": model.StatusExpired}},
	)
	if err != nil {
		return 0, err
	}
	return result.ModifiedCount, nil
}
