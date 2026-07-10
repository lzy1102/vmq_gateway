package config

import "os"

const (
	DefaultMongoURI = "mongodb://localhost:27017"
	DBName          = "vmq_gateway"
	ListenAddr      = ":8080"
)

func MongoURI() string {
	if v := os.Getenv("MONGO_URI"); v != "" {
		return v
	}
	return DefaultMongoURI
}

func AdminUser() string {
	if v := os.Getenv("ADMIN_USER"); v != "" {
		return v
	}
	return "admin"
}

func AdminPass() string {
	if v := os.Getenv("ADMIN_PASS"); v != "" {
		return v
	}
	return "vmq_gateway"
}

func DBDriver() string {
	if v := os.Getenv("DB_DRIVER"); v != "" {
		return v
	}
	return "sqlite"
}

func DBPath() string {
	if v := os.Getenv("DB_PATH"); v != "" {
		return v
	}
	return "vmq_gateway.db"
}

func MySQLDSN() string {
	if v := os.Getenv("MYSQL_DSN"); v != "" {
		return v
	}
	return "root:123456@tcp(127.0.0.1:3306)/vmq_gateway?charset=utf8mb4&parseTime=True&loc=Local"
}

func PostgresDSN() string {
	if v := os.Getenv("POSTGRES_DSN"); v != "" {
		return v
	}
	return "postgres://postgres:123456@localhost:5432/vmq_gateway?sslmode=disable"
}

// Package 套餐定义
type Package struct {
	Name         string
	Amount       int64 // 基础金额（分）
	StreamNumber int64 // 流量数
}

// Packages 预定义套餐（MVP 硬编码）
var Packages = map[string]Package{
	"small":  {Name: "小套餐", Amount: 1000, StreamNumber: 100}, // 10.00元
	"medium": {Name: "中套餐", Amount: 3000, StreamNumber: 300}, // 30.00元
	"big":    {Name: "大套餐", Amount: 5000, StreamNumber: 500}, // 50.00元
}
