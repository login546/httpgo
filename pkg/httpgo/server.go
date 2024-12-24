package httpgo

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"
)

// ServeDirectoryWithAuth 启动一个带有基本身份验证的文件服务器
func ServeDirectoryWithAuth(dir, username, password string, port int) error {
	// 检查目录是否存在
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return err
	}

	// 创建一个文件服务器处理程序
	fs := http.FileServer(http.Dir(dir))

	// 使用 BasicAuth 和 Logging 中间件保护和记录文件服务器
	protectedFS := BasicAuth(LoggingMiddleware(fs), username, password)

	// 启动 Web 服务器
	addr := ":" + strconv.Itoa(port)
	//log.Printf("Serving %s on HTTP port %d\n", dir, port)
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

// LoggingMiddleware 是一个中间件函数，用于记录每个请求的详细信息
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// 使用自定义的 ResponseWriter 来捕获状态码
		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(lrw, r)

		// 打印日志
		log.Printf("%s %s %s %d %s",
			r.RemoteAddr,
			r.Method,
			r.URL.Path,
			lrw.statusCode,
			time.Since(start))
	})
}

// loggingResponseWriter 是一个包装 ResponseWriter 的结构体，用于捕获响应状态码
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader 捕获状态码
func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func GetLocalIP() string {
	// 获取本地机器的所有网络接口
	addrs, err := net.Interfaces()
	if err != nil {
		fmt.Println("获取网卡信息失败:", err)
		return "0.0.0.0"
	}

	// 遍历网络接口
	for _, iface := range addrs {
		// 获取该接口的所有IP地址
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		// 遍历每个地址，查找第一个非回环地址
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
				// 返回找到的局域网IP地址
				return ipnet.IP.String()
			}
		}
	}

	return "0.0.0.0" // 如果未找到有效的局域网IP地址，返回默认地址
}
