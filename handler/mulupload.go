package handler

import (
	redisPool "filestore-server/cache/redis"
	"filestore-server/db"
	"filestore-server/util"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"math"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

// MultipartUploadInfo 初始化信息
type MultipartUploadInfo struct {
	FileHash   string
	FileSize   int    // 总大小
	UploadID   string // 与Filehash不一样
	ChunkSize  int    // 分块大小
	ChunkCount int    // 分块数量
}

// InitialMultipartUploadHandler 初始化分块上传
func InitialMultipartUploadHandler(w http.ResponseWriter, r *http.Request) {
	//1.解析用户请求参数
	r.ParseForm()
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filesize, err := strconv.Atoi(r.Form.Get("filesize"))
	if err != nil {
		w.Write(util.NewRespMsg(-1, "params invalid", nil).JSONBytes())
		return
	}

	//2.获得redis的一个连接
	rConn := redisPool.RedisPool().Get()
	defer rConn.Close()

	//3.生成分块上传的初始化信息
	upInfo := MultipartUploadInfo{
		FileHash:   filehash,
		FileSize:   filesize,
		UploadID:   username + fmt.Sprintf("%x", time.Now().UnixNano()),
		ChunkSize:  5 * 1024 * 1024, // 5MB
		ChunkCount: int(math.Ceil(float64(filesize) / (5 * 1024 * 1024))),
	}

	//4.将初始化信息写入到redis缓存
	rConn.Do("HMSET", "MP_"+upInfo.UploadID,
		"chunkcount", upInfo.ChunkCount,
		"filehash", upInfo.FileHash,
		"filesize", upInfo.FileSize)

	//5.将响应初始化数据返回到客户端
	w.Write(util.NewRespMsg(0, "OK", upInfo).JSONBytes())
}

// 上传文件分块
func UploadPartHandler(w http.ResponseWriter, r *http.Request) {
	// 1.解析用户请求参数
	r.ParseForm()
	uploadID := r.Form.Get("uploadid")
	chunkindex := r.Form.Get("index")

	// 2.获得redis连接池中的一个连接
	rConn := redisPool.RedisPool().Get()
	defer rConn.Close()

	// 3.获得文件句柄，用于储存分块内容
	fpath := "./data/" + uploadID + "/" + chunkindex
	os.MkdirAll(path.Dir(fpath), 0744)
	fd, err := os.Create(fpath)
	if err != nil {
		w.Write(util.NewRespMsg(-1, "Upload part failed", nil).JSONBytes())
		return
	}
	defer fd.Close()

	//使用一个大小为1MB的缓冲区buf, 可以处理较大的数据块，平衡内存开销与读取次数
	buf := make([]byte, 1024*1024) // 1MB
	for {
		n, err := r.Body.Read(buf)
		fd.Write(buf[:n])
		if err != nil { // 读取完毕或遇到文件末尾
			break
		}
	}
	// 4.更新redis缓存状态
	rConn.Do("HSET", "MP_"+uploadID, "chunk_"+chunkindex, "1")
	// 5. 返回处理结果到客户端
	w.Write(util.NewRespMsg(0, "OK", nil).JSONBytes())
}

// 通知上传合并
func CompleteUploadHandler(w http.ResponseWriter, r *http.Request) {
	// 1.解析请求参数
	r.ParseForm()
	uploadID := r.Form.Get("uploadid")
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filesize := r.Form.Get("filesize")
	filename := r.Form.Get("filename")

	// 2.获得redis连接池中的一个连接
	rConn := redisPool.RedisPool().Get()
	defer rConn.Close()

	// 3.通过uploadid查询redis并判断是否所有分块上传完成
	data, err := redis.Values(rConn.Do("HGETALL", "MP_"+uploadID))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var total int = 0
	var realnum = 0
	//var filesize = 0
	for i := 0; i < len(data); i += 2 {
		k := string(data[i].([]byte))
		v := string(data[i+1].([]byte))
		if strings.HasPrefix(k, "chunk_") && v == "1" {
			realnum += 1
		} else if k == "chunkcount" {
			total, _ = strconv.Atoi(v)
		}
		//else if k == "filesize" {
		//	filesize, _ = strconv.Atoi(v)
		//}
	}
	if total != realnum {
		w.Write(util.NewRespMsg(0, "invalid request(合并数量不足)", nil).JSONBytes())
		return
	}

	// 4.
	// TODO:合并分块
	// 5.更新唯一文件表及用户文件表
	// TODO:安排地址
	fsize, _ := strconv.Atoi(filesize)
	db.OnFileUploadFinished(filehash, filename, int64(fsize), "")
	db.OnUserFileUploadFinished(username, filehash, filename, int64(fsize))

	// 6.响应处理结果
	w.Write(util.NewRespMsg(0, "OK", nil).JSONBytes())
}
