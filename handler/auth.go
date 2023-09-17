package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

// HTTPInterceptor http请求拦截器
func HTTPInterceptor() gin.HandlerFunc {
	return func(c *gin.Context) {
		username := c.Request.FormValue("username") // 用户名
		token := c.Request.FormValue("token")       // 访问令牌
		fmt.Println(username)
		fmt.Println(len(token))

		if len(username) < 3 || !isTokenValid(token) {
			// 验证不通过，不再调用后续的函数处理
			c.Abort()
			c.JSON(http.StatusUnauthorized, gin.H{"message": "访问未授权"})
			// return可省略, 只要前面执行Abort()就可以让后面的handler函数不再执行
			return
		}
		c.Next()
	}
}
