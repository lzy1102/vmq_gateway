package model

type Binding struct {
	ServiceID   string `bson:"service_id" json:"service_id" gorm:"size:128"`
	CallbackURL string `bson:"callback_url" json:"callback_url" gorm:"size:512"`
	DeviceID    string `bson:"device_id" json:"device_id" gorm:"size:128"`
	PoolID      string `bson:"pool_id" json:"pool_id" gorm:"size:128"`
	APIKey      string `bson:"api_key" json:"api_key" gorm:"size:128"`
	IPWhitelist string `bson:"ip_whitelist" json:"ip_whitelist" gorm:"size:1024;default:''"`
}
