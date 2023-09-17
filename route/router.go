package route

import "github.com/gin-gonic/gin"

func Router() *gin.Engine {
	//
	router := gin.Default()
	// 处理静态资源
	router.Static("/static/", "./ststic")

	// 不需要验证就能访问的接口

	return router
}
