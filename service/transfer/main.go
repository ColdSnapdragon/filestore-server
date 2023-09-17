package main

// 转移服务

import (
	"encoding/json"
	"filestore-server/config"
	"filestore-server/db"
	"filestore-server/mq"
	"filestore-server/store/oss"
	"fmt"
	"log"
	"os"
)

// 处理文件转移的真正逻辑
func ProcessTransfer(msg []byte) bool {
	// 1.解析msg
	task := mq.TransferData{}
	err := json.Unmarshal(msg, &task)
	if err != nil {
		log.Println(err)
		return false
	}
	fmt.Printf("消费: %#v\n", task)

	// 2.根据临时文件存储路径，创建文件句柄
	file, err := os.Open(task.CurLocation)
	if err != nil {
		log.Println(err)
		return false
	}

	// 3.通过文件句柄将文件内容读出来并上传到oss
	err = oss.Bucket().PutObject(task.DestLocation, file)
	if err != nil {
		log.Println(err)
		return false
	}

	// 4.更新文件的存储路径到文件表
	ok := db.UpdateFileLocation(task.FileHash, task.DestLocation)
	if !ok {
		log.Println("更新存储路径失败")
		return false
	}

	// 5.删除临时文件
	err = os.Remove(task.CurLocation)
	if err != nil {
		log.Println("临时文件删除失败:", err)
		return false
	}

	return true
}

func main() {
	log.Println("开始监听转移任务队列...")
	mq.StartConsume(
		config.TransOSSQueueName,
		"transfer_oss",
		ProcessTransfer,
	)
}
