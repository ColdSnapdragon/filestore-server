package oss

import (
	"filestore-server/config"
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// 管理bucket实例
// 封装上传/下载等逻辑

var ossCli *oss.Client

// Client 获取SSClient对象
func Client() *oss.Client {
	if ossCli != nil {
		return ossCli
	}
	// 创建OSSClient实例。
	client, err := oss.New(config.OSSEndpoint, config.OSSAccessKeyID, config.OSSAccessKeySecret)
	if err != nil {
		fmt.Println("连接阿里云oss失败:", err)
		return nil
	}
	return client
}

// 获取bucket存储空间
func Bucket() *oss.Bucket {
	cli := Client()
	if cli != nil {
		bucket, err := cli.Bucket(config.OSSBucket) // 获取存储空间。
		if err != nil {
			fmt.Println("获取bucket失败:", err)
			return nil
		}
		return bucket
	}
	return nil
}

// DownloadUrl 生成临时授权下载的url
func DownloadUrl(objName string) string {
	signedUrl, err := Bucket().SignURL(objName, oss.HTTPGet, 3600) // 过期时间3600s
	if err != nil {
		fmt.Println("生成临时url失败:", err)
		return ""
	}
	return signedUrl
}
