package handler

import (
	"fmt"
	"net/http"
)

// HTTPInterceptor http请求拦截器
func HTTPInterceptor(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc( // 函数类型转换
		func(w http.ResponseWriter, r *http.Request) {
			r.ParseForm()
			username := r.Form.Get("username")
			token := r.Form.Get("token")

			if len(username) < 3 || !isTokenValid(token) {
				fmt.Println("访问拒绝", username, token, "检查错误")
				w.WriteHeader(http.StatusForbidden)
				return
			}
			h(w, r)
		},
	)
}
