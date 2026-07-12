package model

type Order struct {
	TradeNo     string `bson:"trade_no" json:"trade_no" gorm:"column:trade_no"`
	ServiceID   string `bson:"service_id" json:"service_id" gorm:"column:service_id"`
	CallbackURL string `bson:"callback_url" json:"callback_url" gorm:"column:callback_url"`
	Amount      int64  `bson:"amount" json:"amount" gorm:"column:amount"`
	Status      string `bson:"status" json:"status" gorm:"column:status"`
	DeviceID    string `bson:"device_id" json:"device_id" gorm:"column:device_id"`
	CreatedAt   int64  `bson:"created_at" json:"created_at" gorm:"column:created_at"`
	PaidAt      int64  `bson:"paid_at" json:"paid_at,omitempty" gorm:"column:paid_at"`
	ExpireAt    int64  `bson:"expire_at" json:"expire_at" gorm:"column:expire_at"`
}

func (Order) TableName() string {
	return "orders"
}

const (
	StatusPending   = "pending"
	StatusPaid      = "paid"
	StatusCancelled = "cancelled"
	StatusExpired   = "expired"

	DefaultExpireMinutes = 15
)
