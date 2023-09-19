package main

import (
	"filestore-server/service/account/handler"
	"filestore-server/service/account/proto"
	micro "go-micro.dev/v4"
	"log"
	"time"
)

func main() {
	// 创建一个服务
	service := micro.NewService(
		micro.Name("go.micro.service.user"),
		micro.RegisterTTL(10*time.Second),
		micro.RegisterInterval(5*time.Second))
	service.Init()

	proto.RegisterUserServiceHandler(service.Server(), &handler.User{})

	// 启动服务
	if err := service.Run(); err != nil {
		log.Println(err)
	}
}
