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
	config.ValidateConfig()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := store.Init(ctx); err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	log.Println("数据库连接成功")

	r := gin.Default()

	r.GET("/appHeart", handler.Heartbeat)
	r.GET("/appPush", handler.AppPush)

	api := r.Group("/api")
	{
		api.POST("/order", handler.CreateOrder)
		api.POST("/order/cancel", handler.CancelOrder)
		api.GET("/order/status", handler.QueryOrderStatus)
	}

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

	log.Printf("V免签支付网关启动 → %s", config.ListenAddr)
	if err := r.Run(config.ListenAddr); err != nil {
		log.Fatalf("启动失败: %v", err)
	}
}
