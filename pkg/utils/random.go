package utils

import (
	"math/rand"
	"time"
)

var Charset1 = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func GenerateRandomString(length int) string {
	rand.Seed(time.Now().UnixNano())

	// 创建一个字节数组来存储生成的随机字符
	b := make([]byte, length)

	// 随机选择字符集中的字符并填充到字节数组中
	for i := range b {
		b[i] = Charset1[rand.Intn(len(Charset1))]
	}

	return string(b)
}
