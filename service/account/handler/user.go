package handler

import (
	"context"
	"filestore-server/common"
	"filestore-server/config"
	"filestore-server/db"
	pb "filestore-server/service/account/proto"
	"filestore-server/util"
)

type User struct {
	// pb.UnimplementedUserServiceServer
}

// Signup 处理用户注册请求
func (*User) Signup(ctx context.Context, req *pb.ReqSignup, resp *pb.RespSignup) (err error) {
	username := req.Username
	passwd := req.Password

	// 简单参数校验
	if len(username) < 3 || len(passwd) < 5 {
		resp.Code = common.StatusParamInvalid
		resp.Message = "注册参数无效"
		return
	}

	// 对密码进行加盐及取Sha1值加密
	encPasswd := util.Sha1([]byte(passwd + config.PasswdSalt))
	// 将用户信息注册到用户表中
	suc := db.UserSignup(username, encPasswd)
	if suc {
		//c.JSON(http.StatusOK,
		//	gin.H{
		//		"code":    0,
		//		"msg":     "注册成功",
		//		"data":    nil,
		//		"forward": "/user/signin",
		//	})
		resp.Code = common.StatusOK
		resp.Message = "注册成功"
	} else {
		//c.JSON(http.StatusOK,
		//	gin.H{
		//		"code": 0,
		//		"msg":  "注册失败",
		//		"data": nil,
		//	})
		resp.Code = common.StatusOK
		resp.Message = "注册失败"
	}
	return nil
}
