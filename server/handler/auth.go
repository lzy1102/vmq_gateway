package handler

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lzy1102/vmq_gateway/server/config"
	"github.com/lzy1102/vmq_gateway/server/middleware"
)

type loginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

var (
	loginAttempts = make(map[string][]time.Time)
	loginMu       sync.Mutex
	maxAttempts   = 5
	windowSize    = 60 * time.Second
)

func isRateLimited(ip string) bool {
	loginMu.Lock()
	defer loginMu.Unlock()

	now := time.Now()
	attempts := loginAttempts[ip]

	var valid []time.Time
	for _, t := range attempts {
		if now.Sub(t) < windowSize {
			valid = append(valid, t)
		}
	}
	loginAttempts[ip] = valid

	return len(valid) >= maxAttempts
}

func recordAttempt(ip string) {
	loginMu.Lock()
	defer loginMu.Unlock()
	loginAttempts[ip] = append(loginAttempts[ip], time.Now())
}

func generateSessionToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func Login(c *gin.Context) {
	ip := c.ClientIP()

	if isRateLimited(ip) {
		c.JSON(http.StatusTooManyRequests, gin.H{"code": 0, "msg": "登录尝试过多，请稍后再试"})
		return
	}

	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "msg": "参数错误"})
		return
	}

	if req.Username != config.AdminUser() || req.Password != config.AdminPass() {
		recordAttempt(ip)
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "用户名或密码错误"})
		return
	}

	token := generateSessionToken()

	// 内存存储，重启失效
	middleware.SetSession(token, req.Username)

	c.SetSameSite(http.SameSiteStrictMode)
	secure := config.IsProduction()
	c.SetCookie(middleware.SessionCookie, token, 86400, "/admin", "", secure, true)
	c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "登录成功"})
}

func Logout(c *gin.Context) {
	cookie, _ := c.Cookie(middleware.SessionCookie)
	if cookie != "" {
		middleware.DeleteSession(cookie)
	}
	c.SetCookie(middleware.SessionCookie, "", -1, "/admin", "", false, true)
	c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "已退出"})
}
