package mq

import "filestore-server/common"

// TransferData 转移队列中消息载体的结构格式
type TransferData struct {
	FileHash      string
	CurLocation   string // 临时储存地址
	DestLocation  string // 目标转移地址(比如oss objectKey)
	DestStoreType common.StoreType
}
