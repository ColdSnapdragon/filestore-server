package db

import (
	"database/sql"
	mydb "filestore-server/db/mysql"
	"fmt"
)

// 文件上传完成，保存meta
func OnFileUploadFinished(filehash string, filename string,
	filesize int64, fileaddr string) bool {
	// 进行预编译，可尽量防止sql注入
	stmt, err := mydb.DBConn().Prepare(
		"insert ignore into tbl_file (`file_sha1`, `file_name`, `file_size`," +
			" `file_addr`, `status` ) values (?,?,?,?,1)",
	)
	// ignore关键字表示当发生唯一索引冲突时(file_sha1是UNIQUE)，忽略此处插入而不是报错
	if err != nil {
		fmt.Printf("预编译sql失败: %v", err)
		return false
	}
	defer stmt.Close() // 关闭资源

	ret, err := stmt.Exec(filehash, filename, filesize, fileaddr) // 替代"?"
	if err != nil {
		fmt.Printf("sql执行失败: %v", err)
		return false
	}
	// 影响的行(这里就是插入的行数)
	if rf, err := ret.RowsAffected(); err == nil {
		if rf <= 0 { // 没有插入
			fmt.Printf("File with hash:%s has been uploaded before", filehash)
		}
		return true
	}
	fmt.Printf("写入数据库失败: %v", err)
	return false
}

// TableFile 文件表结构体
type TableFile struct {
	FileHash string
	FileName sql.NullString // (很简单的结构体)可能为null的string
	FileSize sql.NullInt64
	FileAddr sql.NullString
}

// GetFileMeta 从mysql获取文件元信息
func GetFileMeta(filehash string) (*TableFile, error) {
	stmt, err := mydb.DBConn().Prepare(
		"select `file_sha1`, `file_name`, `file_size`," +
			" `file_addr` from tbl_file where `file_sha1`=? " +
			"and status=1 limit 1",
	)

	if err != nil {
		fmt.Printf("预编译sql失败: %v", err)
		return nil, err
	}
	defer stmt.Close()

	tfile := TableFile{}
	err = stmt.QueryRow(filehash).Scan( // 将查询结果扫描进结构体中
		&tfile.FileHash, &tfile.FileName, &tfile.FileSize, &tfile.FileAddr)
	if err != nil {
		if err == sql.ErrNoRows {
			// 查不到对应记录， 返回参数及错误均为nil
			return nil, nil
		} else {
			fmt.Printf("绑定结构体失败: %v", err)
			return nil, err
		}
	}

	return &tfile, nil
}

// GetFileMetaList : 从mysql批量获取文件元信息
func GetFileMetaList(limit int) ([]TableFile, error) {
	stmt, err := mydb.DBConn().Prepare(
		"select file_sha1,file_addr,file_name,file_size from tbl_file " +
			"where status=1 limit ?")
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(limit)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	cloumns, _ := rows.Columns()
	values := make([]sql.RawBytes, len(cloumns))
	var tfiles []TableFile
	for i := 0; i < len(values) && rows.Next(); i++ {
		tfile := TableFile{}
		err = rows.Scan(&tfile.FileHash, &tfile.FileAddr,
			&tfile.FileName, &tfile.FileSize)
		if err != nil {
			fmt.Println(err.Error())
			break
		}
		tfiles = append(tfiles, tfile)
	}
	fmt.Println(len(tfiles))
	return tfiles, nil
}

// UpdateFileLocation : 更新文件的存储地址(如文件被转移了)
func UpdateFileLocation(filehash string, fileaddr string) bool {
	stmt, err := mydb.DBConn().Prepare(
		"update tbl_file set`file_addr`=? where  `file_sha1`=? limit 1")
	if err != nil {
		fmt.Println("预编译sql失败, err:" + err.Error())
		return false
	}
	defer stmt.Close()

	ret, err := stmt.Exec(fileaddr, filehash)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	if rf, err := ret.RowsAffected(); nil == err {
		if rf <= 0 {
			fmt.Printf("更新文件location失败, filehash:%s", filehash)
		}
		return true
	}
	return false
}
