package meta

import (
	"filestore-server/db"
)

// FileMeta 文件元信息结构
type FileMeta struct {
	FileSha1 string // 文件唯一标志
	FileName string
	FileSize int64
	Location string // 路径
	UploadAt string // 时间戳(格式化后的字符串)
}

// 保存所有元信息(根据唯一标志索引)(初期在内存中，后续会放进数据库里)
// 后续再考虑map的并发安全问题
var fileMetas map[string]FileMeta

// 初始化
func init() {
	fileMetas = make(map[string]FileMeta)
}

// 新增/更新文件源信息
func UpdateFileMeta(fmeta FileMeta) {
	fileMetas[fmeta.FileSha1] = fmeta
}

// 新增/更新元信息到数据库中
func UpdateFileMetaDB(fmeta FileMeta) bool {
	return db.OnFileUploadFinished(fmeta.FileSha1, fmeta.FileName, fmeta.FileSize, fmeta.Location)
}

// 获取文件元信息
func GetFileMeta(fileSha1 string) FileMeta {
	return fileMetas[fileSha1]
}

// 从数据库获取文件元信息
func GetFileMetaDB(fileSha1 string) (FileMeta, error) {
	tfile, err := db.GetFileMeta(fileSha1)
	if err != nil {
		return FileMeta{}, err
	}
	fmeta := FileMeta{
		FileSha1: tfile.FileHash,
		FileName: tfile.FileName.String,
		FileSize: tfile.FileSize.Int64,
		Location: tfile.FileAddr.String,
	}
	return fmeta, nil
}

// 删除文件元信息
func RemoveFileMeta(filesha1 string) {
	delete(fileMetas, filesha1)
}
