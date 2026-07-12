package model

type Device struct {
	DeviceID      string `bson:"device_id" json:"device_id" gorm:"column:device_id"`
	VmqKey        string `bson:"key" json:"key" gorm:"column:vmq_key"`
	Status        string `bson:"status" json:"status"`
	LastHeartbeat int64  `bson:"last_heartbeat" json:"last_heartbeat"`
	WechatQR      string `bson:"wechat_qr" json:"wechat_qr" gorm:"column:wechat_qr;default:'/qr/wechat.png'"`
	AlipayQR      string `bson:"alipay_qr" json:"alipay_qr" gorm:"column:alipay_qr;default:'/qr/alipay.png'"`
}

const (
	DeviceOnline  = "online"
	DeviceOffline = "offline"
)
