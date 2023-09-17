package main

// 上传服务

import (
	"filestore-server/config"
	"filestore-server/route"
	"log"
)

func main() {
	router := route.Router()
	log.Fatal(router.Run(config.UploadServiceHost))
}
