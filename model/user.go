package model

type User struct {
	UserName     string `bson:"user_name" json:"user_name"`
	StreamNumber int64  `bson:"stream_number" json:"stream_number"`
	CreatedAt    int64  `bson:"created_at" json:"created_at"`
}
