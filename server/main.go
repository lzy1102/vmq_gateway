package main

import (
	"context"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lzy1102/vmq_gateway/server/config"
	"github.com/lzy1102/vmq_gateway/server/handler"
	"github.com/lzy1102/vmq_gateway/server/middleware"
	"github.com/lzy1102/vmq_gateway/server/store"
)

func main() {
	// 1. 连接 MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := store.Init(ctx); err != nil {
		log.Fatalf("MongoDB 连接失败: %v", err)
	}
	log.Println("MongoDB 连接成功")

	// 2. 创建路由
	r := gin.Default()

	// 3. APP 回调接口（GET，签名在 handler 内验证）
	r.GET("/appHeart", handler.Heartbeat)
	r.GET("/appPush", handler.AppPush)

	// 4. 用户 API
	api := r.Group("/api")
	{
		api.POST("/recharge/vmpay", handler.CreateRechargeOrder)
		api.GET("/recharge/vmpay-status", handler.QueryOrderStatus)
	}

	// 5. 设备管理 API（需登录）
	admin := r.Group("/admin")
	admin.POST("/login", handler.Login)
	admin.POST("/logout", handler.Logout)

	protected := admin.Group("")
	protected.Use(middleware.RequireAuth())
	{
		protected.POST("/device", handler.AddDevice)
		protected.GET("/devices", handler.ListDevices)
		protected.POST("/pool", handler.AddPool)
		protected.POST("/pool/device", handler.AddDeviceToPool)
		protected.GET("/pools", handler.ListPools)
		protected.POST("/binding", handler.AddBinding)
		protected.GET("/bindings", handler.ListBindings)
	}

	// 6. 启动
	log.Printf("V免签支付网关启动 → %s", config.ListenAddr)
	if err := r.Run(config.ListenAddr); err != nil {
		log.Fatalf("启动失败: %v", err)
	}
}
