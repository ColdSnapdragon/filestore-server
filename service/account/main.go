package main

import (
	"github.com/go-micro/plugins/v4/registry/consul"
	"log"
	"time"

	micro "go-micro.dev/v4"

	"filestore-server/service/account/handler"
	proto "filestore-server/service/account/proto"
)

func main() {
	// 创建一个service
	service := micro.NewService(
		micro.Name("go.micro.service.user"),
		micro.RegisterTTL(time.Second*10),
		micro.RegisterInterval(time.Second*5),
		micro.Registry(consul.NewRegistry()),
	)
	service.Init()

	proto.RegisterUserServiceHandler(service.Server(), new(handler.User))
	if err := service.Run(); err != nil {
		log.Println(err)
	}
}
