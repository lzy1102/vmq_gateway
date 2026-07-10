package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const SessionCookie = "vmq_admin_session"

func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Cookie(SessionCookie)
		if err != nil || cookie == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 0, "msg": "未登录"})
			c.Abort()
			return
		}
		c.Set("admin_id", cookie)
		c.Next()
	}
}
