package model

type Binding struct {
	ServiceID   string `bson:"service_id" json:"service_id"`
	CallbackURL string `bson:"callback_url" json:"callback_url"`
	DeviceID    string `bson:"device_id" json:"device_id"`
	PoolID      string `bson:"pool_id" json:"pool_id"`
	APIKey      string `bson:"api_key" json:"api_key"`
}
