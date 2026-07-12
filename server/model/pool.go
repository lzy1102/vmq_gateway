package model

type Pool struct {
	PoolID    string   `bson:"pool_id" json:"pool_id"`
	Name      string   `bson:"name" json:"name"`
	DeviceIDs []string `bson:"device_ids" json:"device_ids" gorm:"-"`
}
