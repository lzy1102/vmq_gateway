package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lzy1102/vmq_gateway/config"
	"github.com/lzy1102/vmq_gateway/middleware"
)

type loginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "msg": "参数错误"})
		return
	}

	if req.Username != config.AdminUser() || req.Password != config.AdminPass() {
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "用户名或密码错误"})
		return
	}

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(middleware.SessionCookie, req.Username, 86400, "/admin", "", false, false)
	c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "登录成功"})
}

func Logout(c *gin.Context) {
	c.SetCookie(middleware.SessionCookie, "", -1, "/admin", "", false, false)
	c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "已退出"})
}
