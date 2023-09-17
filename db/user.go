package db

import (
	mydb "filestore-server/db/mysql"
	"fmt"
)

// User : 用户表model
type User struct { // 与数据库表对应
	Username     string
	Email        string
	Phone        string
	SignupAt     string // 注册时间
	LastActiveAt string // 最后活跃时间
	Status       int
}

// UserSignup 通过用户名和密码完成user表的注册操作
func UserSignup(username string, passwd string) bool {
	stmt, err := mydb.DBConn().Prepare(
		"insert ignore into tbl_user (`user_name`,`user_pwd`) values (?,?)")
	if err != nil {
		fmt.Printf("预编译sql失败: %v", err)
		return false
	}
	defer stmt.Close()

	ret, err1 := stmt.Exec(username, passwd)
	if err1 != nil {
		fmt.Printf("插入数据失败: %v", err)
		return false
	}

	if affects, err := ret.RowsAffected(); nil == err && affects > 0 {
		return true
	}
	// affects==0表示账户已存在，不能重复注册
	return false
}

// UserSignin : 判断密码是否一致
func UserSignin(username string, encpwd string) bool {
	stmt, err := mydb.DBConn().Prepare("select * from tbl_user where user_name=? limit 1")
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	defer stmt.Close()

	rows, err := stmt.Query(username)
	if err != nil {
		fmt.Println(err.Error())
		return false
	} else if rows == nil {
		fmt.Println("用户不存在: " + username)
		return false
	}

	pRows := mydb.ParseRows(rows) //把多行查询结果(由于上面有limit 1，实际只有一行)转为[]map[string]any
	if len(pRows) > 0 && string(pRows[0]["user_pwd"].([]byte)) == encpwd {
		return true
	}
	return false
}

// UpdateToken : 刷新用户登录的token
func UpdateToken(username string, token string) bool {
	stmt, err := mydb.DBConn().Prepare(
		// REPLACE INTO: 插入或替换数据(根据提供的唯一索引(user_name为UNIQUE)来判断是否存在)
		// 插入数据的表必须有主键或者是唯一索引，否则replace into会直接插入数据
		"replace into tbl_user_token (`user_name`,`user_token`) values (?,?)")
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	defer stmt.Close()

	_, err = stmt.Exec(username, token)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	return true
}

// GetUserInfo : 查询用户信息
func GetUserInfo(username string) (User, error) {
	user := User{}

	stmt, err := mydb.DBConn().Prepare(
		"select `user_name`,`signup_at` from tbl_user where `user_name`=? limit 1")
	if err != nil {
		fmt.Println(err.Error())
		return user, err
	}
	defer stmt.Close()

	// 执行查询的操作
	err = stmt.QueryRow(username).Scan(&user.Username, &user.SignupAt)
	if err != nil {
		return user, err
	}
	return user, nil
}
