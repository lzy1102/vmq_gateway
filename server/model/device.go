package model

type Device struct {
	DeviceID      string `bson:"device_id" json:"device_id" gorm:"column:device_id"`
	VmqKey        string `bson:"key" json:"key" gorm:"column:vmq_key"`
	Status        string `bson:"status" json:"status"`
	LastHeartbeat int64  `bson:"last_heartbeat" json:"last_heartbeat"`
	QRCode        string `bson:"qr_code" json:"qr_code" gorm:"column:qr_code;default:'/qr/alipay.png'"`
}

const (
	DeviceOnline  = "online"
	DeviceOffline = "offline"
)
