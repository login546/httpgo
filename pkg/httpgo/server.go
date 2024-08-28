package httpgo

import (
	"log"
	"net/http"
	"os"
	"strconv"
)

// ServeDirectoryWithAuth 启动一个带有基本身份验证的文件服务器
func ServeDirectoryWithAuth(dir, username, password string, port int) error {
	// 检查目录是否存在
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return err
	}

	// 创建一个文件服务器处理程序
	fs := http.FileServer(http.Dir(dir))

	// 使用 BasicAuth 中间件保护文件服务器
	protectedFS := BasicAuth(fs, username, password)

	// 启动 Web 服务器
	addr := ":" + strconv.Itoa(port)
	log.Printf("Serving %s on HTTP port %d\n", dir, port)
	return http.ListenAndServe(addr, protectedFS)
}

// BasicAuth 是一个中间件函数，用于实现基本身份验证
func BasicAuth(next http.Handler, username, password string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != username || pass != password {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
