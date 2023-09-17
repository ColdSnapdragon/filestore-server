package handler

import (
	"encoding/json"
	"filestore-server/common"
	"filestore-server/config"
	"filestore-server/db"
	"filestore-server/meta"
	"filestore-server/mq"
	"filestore-server/store/oss"
	"filestore-server/util"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

// 简单的文件上传服务

// UploadHandler 处理文件上传
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// 返回上传页面
		data, err := os.ReadFile("./static/view/index.html")
		if err != nil {
			fmt.Printf("文件读取失败: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(data)
	} else if r.Method == "POST" {
		// 接受文件以及储存到本地目录
		file, fhead, err := r.FormFile("file") // 我们用form形式上传文件
		// file:文件句柄 fhead:文件头(可获取各种文件信息)
		if err != nil {
			fmt.Printf("获取上传数据失败: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer file.Close() // 关闭句柄

		newFile, err := os.Create("/tmp/" + fhead.Filename)
		if err != nil {
			log.Fatal(err)
		}
		defer newFile.Close()

		fileMeta := meta.FileMeta{
			FileName: fhead.Filename,
			Location: "/tmp/" + fhead.Filename,
			UploadAt: time.Now().Format("2006-05-04 15:02:01"),
		}
		fileMeta.FileSha1 = util.FileSha1(newFile) // 从文件内容中计算sha1哈希值

		// 临时存储到本地
		fileMeta.FileSize, err = io.Copy(newFile, file)
		if err != nil {
			fmt.Printf("复制到本地失败: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if config.CurrentStoreType == common.StoreCeph {
			// TODO: 同时将文件写入ceph存储
			newFile.Seek(0, 0)
		} else if config.CurrentStoreType == common.StoreOSS {
			// 同时将文件写入oss
			newFile.Seek(0, 0)
			ossPath := "oss/" + fileMeta.FileSha1
			//err = oss.Bucket().PutObject(ossPath, newFile)
			//if err != nil {
			//	fmt.Printf("上传到oss失败: %v", err)
			//	w.WriteHeader(http.StatusInternalServerError)
			//	return
			//}
			//fileMeta.Location = ossPath // 修改地址为oss类

			// 引入消息队列
			task := mq.TransferData{
				FileHash:      fileMeta.FileSha1,
				CurLocation:   fileMeta.Location,
				DestLocation:  ossPath,
				DestStoreType: common.StoreOSS,
			}
			msg, _ := json.Marshal(task)
			ok := mq.Publish(config.TransExchangeName, config.TransOSSRoutingKey, msg)
			if !ok {
				// TODO: 加入重发消息逻辑
			}
		}

		// 存储完毕，将元信息写入本地数据库
		// meta.UpdateFileMeta(fileMeta)
		_ = meta.UpdateFileMetaDB(fileMeta)
		fmt.Printf("%#v\n", fileMeta)

		// 更新用户文件表
		r.ParseForm()
		username := r.Form.Get("username")
		db.OnUserFileUploadFinished(username, fileMeta.FileSha1, fileMeta.FileName, fileMeta.FileSize)

		http.Redirect(w, r, "/file/upload/success", http.StatusFound)
	}
}

// 成功上传
func UploadSucHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("上传成功！"))
}

// GetFileMetaHandler 查询文件的元信息
func GetFileMetaHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()                      // 将表单数据解析为键值对的形式
	filehash := r.Form.Get("filehash") // 访问解析后的表单数据(string)
	// fMeta := meta.GetFileMeta(filehash) // 查找元信息
	fMeta, err := meta.GetFileMetaDB(filehash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	data, err := json.Marshal(fMeta) // 序列化为JSON格式的字节切片(slice)
	if err != nil {
		fmt.Printf("序列化失败: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

// FileQueryHandler : 查询批量的文件元信息
func FileQueryHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	limitCnt, _ := strconv.Atoi(r.Form.Get("limit"))
	username := r.Form.Get("username")
	//fileMetas, _ := meta.GetLastFileMetasDB(limitCnt)
	userFiles, err := db.QueryUserFileMetas(username, limitCnt)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(userFiles)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

// DownloadHandler 下载文件接口
func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fsha1 := r.Form.Get("filehash")
	fm := meta.GetFileMeta(fsha1)

	file, err := os.Open(fm.Location) // 简单打开文件(只读)
	if err != nil {
		fmt.Printf(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	// 设置响应头字段
	w.Header().Set("Content-Type", "application/octect-stream") // 表示响应的数据类型是二进制流
	// 设置了HTTP响应头的Content-Disposition字段为attachment;filename="[FileName]"，其中[FileName]是一个变量，代表要下载的文件名。
	// 这个设置指示浏览器将响应内容作为附件下载，并指定了下载时的文件名
	w.Header().Set("Content-Disposition", "attachment;filename=\""+fm.FileName+"\"")
	w.Write(data)
}

// FileMetaUpdateHandler 更新元信息(重命名)接口
func FileMetaUpdateHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	opType := r.Form.Get("op")
	fileSha1 := r.Form.Get("filehash")
	newFileName := r.Form.Get("filename")

	if opType != "0" {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	curFileMeta := meta.GetFileMeta(fileSha1)
	curFileMeta.FileName = newFileName
	meta.UpdateFileMeta(curFileMeta)

	w.WriteHeader(http.StatusOK)
	data, err := json.Marshal(curFileMeta) // 序列化为JSON格式的字节切片(slice)
	if err != nil {
		fmt.Printf("序列化失败: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

// FileDeleteHandler 删除文件
func FileDeleteHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fileSha1 := r.Form.Get("filehash")

	// fm := meta.GetFileMeta(fileSha1)
	fm, _ := meta.GetFileMetaDB(fileSha1)
	err := os.Remove(fm.Location)
	if err != nil {
		fmt.Printf("文件" + fm.Location + "删除失败: " + err.Error())
		return
	}

	meta.RemoveFileMeta(fileSha1)
	w.Write(util.NewRespMsg(0, "删除成功", nil).JSONBytes())
	w.WriteHeader(http.StatusOK)

}

// TryFastUploadHandler '尝试秒传'接口
func TryFastUploadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	// 1.解析请求参数
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filename := r.Form.Get("filename")
	filesize, _ := strconv.Atoi(r.Form.Get("filesize"))

	// 2.从文件表中查询相同hash的文件记录
	fileMeta, err := meta.GetFileMetaDB(filehash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 3.查不到记录则返回秒传失败
	if fileMeta.FileSha1 == "" {
		resp := util.RespMsg{
			Code: 1,
			Msg:  "秒传失败，请使用普通上传",
		}
		w.Write(resp.JSONBytes())
		return
	}

	// 4.上传过则将文件信息写入用户文件表，返回成功
	ok := db.OnUserFileUploadFinished(username, filehash, filename, int64(filesize))
	if ok {
		resp := util.RespMsg{
			Code: 0,
			Msg:  "秒传成功",
		}
		w.Write(resp.JSONBytes())
		return
	}
	resp := util.RespMsg{
		Code: -2,
		Msg:  "秒传失败，请稍后重试",
	}
	w.Write(resp.JSONBytes())
	return
}

// DownloadURLHandler 生成文件的下载地址
func DownloadURLHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	filehash := r.Form.Get("filehash")
	// TODO:判断地址是本地，ceph，还是oss？
	// 从文件表查找记录
	fMeta, err := meta.GetFileMetaDB(filehash)
	if err != nil {
		fmt.Println("文件查询失败: ", err)
		w.Write([]byte("文件查询失败"))
		return
	}
	signedUrl := oss.DownloadUrl(fMeta.Location)
	w.Write([]byte(signedUrl))
}
