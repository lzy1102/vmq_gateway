package middleware

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

const SessionCookie = "vmq_admin_session"

var (
	sessions = make(map[string]string)
	sessionsMu sync.RWMutex
)

func SetSession(token, username string) {
	sessionsMu.Lock()
	defer sessionsMu.Unlock()
	sessions[token] = username
}

func GetSession(token string) (string, bool) {
	sessionsMu.RLock()
	defer sessionsMu.RUnlock()
	username, ok := sessions[token]
	return username, ok
}

func DeleteSession(token string) {
	sessionsMu.Lock()
	defer sessionsMu.Unlock()
	delete(sessions, token)
}

func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Cookie(SessionCookie)
		if err != nil || cookie == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 0, "msg": "未登录"})
			c.Abort()
			return
		}

		username, ok := GetSession(cookie)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 0, "msg": "会话已过期"})
			c.Abort()
			return
		}

		c.Set("admin_id", username)
		c.Next()
	}
}
