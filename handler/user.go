package handler

import (
	"filestore-server/db"
	"filestore-server/util"
	"fmt"
	"net/http"
	"os"
	"time"
)

const (
	pwd_salt = "#123" // 用于加密的盐值(自定义)
)

// SignupHandler 处理用户注册的请求
func SignupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		data, err := os.ReadFile("./static/view/signup.html")
		if err != nil {
			fmt.Printf("文件读取失败: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(data)
		return
	}
	r.ParseForm()
	username := r.Form.Get("username")
	passwd := r.Form.Get("password")

	if len(username) < 3 || len(passwd) < 5 {
		w.Write([]byte("Invalid parameter"))
		return
	}

	// 对密码进行加盐及取Sha1值加密
	encode_pwd := util.Sha1([]byte(passwd + pwd_salt))
	// 将用户信息注册到用户表中
	ok := db.UserSignup(username, encode_pwd)

	if ok {
		w.Write([]byte("SUCCESS"))
	} else {
		w.Write([]byte("FAILED"))
	}
}

// SignInHandler : 登录接口
func SigninHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		//data, err := ioutil.ReadFile("./static/view/signin.html")
		//if err != nil {
		//	w.WriteHeader(http.StatusInternalServerError)
		//	return
		//}
		//w.Write(data)
		http.Redirect(w, r, "/static/view/signin.html", http.StatusFound)
		return
	}

	r.ParseForm()
	username := r.Form.Get("username")
	passwd := r.Form.Get("password")

	encode_pwd := util.Sha1([]byte(passwd + pwd_salt))

	// 1. 校验用户名及密码
	pwdChecked := db.UserSignin(username, encode_pwd)
	if !pwdChecked {
		w.Write([]byte("FAILED"))
		return
	}

	// 2. 生成访问凭证(token)
	token := genToken(username)
	upRes := db.UpdateToken(username, token)
	if !upRes {
		w.Write([]byte("FAILED"))
		return
	}

	// 3. 登录成功后重定向到首页
	// w.Write([]byte("http://" + r.Host + "/static/view/home.html"))
	resp := util.RespMsg{
		Code: 0,
		Msg:  "OK",
		Data: struct {
			Location string
			Username string
			Token    string
		}{
			Location: "http://" + r.Host + "/static/view/home.html",
			Username: username,
			Token:    token,
		},
	}
	w.Write(resp.JSONBytes())
}

// UserInfoHandler 查询用户信息
func UserInfoHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 解析请求参数
	r.ParseForm()
	username := r.Form.Get("username")
	token := r.Form.Get("token") // 从浏览器传来

	// 2. 验证token是否有效
	//valid := isTokenValid(token)
	//if !valid {
	//	w.WriteHeader(http.StatusForbidden) // 403。拒绝访问
	//	return
	//}
	// 后续添加了拦截器，无需这步了
	_ = token

	// 3. 查询用户信息
	user, err := db.GetUserInfo(username)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// 4. 组装并且响应用户数据
	resp := util.RespMsg{
		Code: 0,
		Msg:  "OK",
		Data: user,
	}
	w.Write(resp.JSONBytes())

}

// GenToken : 生成token
func genToken(username string) string {
	// 40位字符:md5(username+timestamp+token_salt)+timestamp[:8]
	ts := fmt.Sprintf("%x", time.Now().Unix())
	tokenPrefix := util.MD5([]byte(username + ts + "_tokensalt"))
	return tokenPrefix + ts[:8]
}

// 判断token是否有效
func isTokenValid(token string) bool {
	// todo
	// 判断token时效性，是否过期
	// 从数据库查username对应的token信息
	// 对比token是否一致
	return true
}
