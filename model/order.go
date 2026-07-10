package model

type RechargeOrder struct {
	TradeNo      string `bson:"trade_no" json:"trade_no"`
	UserName     string `bson:"user_name" json:"user_name"`
	ServiceID    string `bson:"service_id" json:"service_id"`
	CallbackURL  string `bson:"callback_url" json:"callback_url"`
	Amount       int64  `bson:"amount" json:"amount"`
	StreamNumber int64  `bson:"stream_number" json:"stream_number"`
	Status       string `bson:"status" json:"status"`
	CreatedAt    int64  `bson:"created_at" json:"created_at"`
	PaidAt       int64  `bson:"paid_at" json:"paid_at,omitempty"`
}

const (
	StatusPending    = "pending"
	StatusProcessing = "processing"
	StatusPaid       = "paid"
	StatusCancelled  = "cancelled"
)
