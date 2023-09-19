package handler

import (
	"context"
	proto "filestore-server/service/account/proto"
	"github.com/gin-gonic/gin"
	micro "go-micro.dev/v4"
	"log"
	"net/http"
)

var (
	userCli proto.UserService
)

func init() {
	service := micro.NewService(
		// micro.Name("go.micro.api.user"), // apigw的服务无需注册到服务中心，留空也行
	)
	// 初始化，如解析命令行参数等
	service.Init()

	// 初始化一个rpcClient
	userCli = proto.NewUserService("go.micro.service.user", service.Client())

}

// 将用户的http请求，转换成rpc请求，从上面的客户端发送到服务端，拿到结果后再响应给用户

// SignupHandler : 响应注册页面
func SignupHandler(c *gin.Context) {
	c.Redirect(http.StatusFound, "http://"+c.Request.Host+"/static/view/signup.html")
}

// DoSignupHandler : 处理用户注册请求
func DoSignupHandler(c *gin.Context) {
	username := c.Request.FormValue("username")
	passwd := c.Request.FormValue("password")
	resp, err := userCli.Signup(context.TODO(), &proto.ReqSignup{
		Username: username,
		Password: passwd,
	})
	if err != nil {
		log.Println(err)
		c.Status(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    resp.Code,
		"message": resp.Message,
	})
}
