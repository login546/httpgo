package utils

import (
	"bytes"
	"fmt"
	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
	"regexp"
	"strings"
)

// 提取body中的title内容
func ExtractTitle(htmlContent []byte) (string, error) {
	// 创建一个字符集读取器
	reader, err := charset.NewReader(bytes.NewReader(htmlContent), "")
	if err != nil {
		return "", fmt.Errorf("创建字符集读取器时出错: %v", err)
	}

	// 解析HTML内容
	doc, err := html.Parse(reader)
	if err != nil {
		return "", fmt.Errorf("解析HTML时出错: %v", err)
	}

	// 遍历节点树，查找<title>标签
	var title string
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "title" {
			if n.FirstChild != nil {
				title = n.FirstChild.Data
			}
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	if title == "" {
		return "", fmt.Errorf("未找到<title>元素")
	}

	return TrimTitle(title), nil
}

// 去除title中首尾的空格、制表符和换行符
func TrimTitle(title string) string {
	title = strings.TrimSpace(title)
	return title
}

// // 提取body中的非/favicon.ico的内容
// // 1、link标签的 rel="icon" 属性，
// // 2、link标签的 rel="shortcut icon" 属性，
// // 3、link标签的 rel="apple-touch-icon" 属性，
// // 4、link标签的 rel="apple-touch-icon-precomposed" 属性，
// // 5、link标签的 rel="apple-touch-startup-image" 属性，
// // 6、link标签的 rel="mask-icon" 属性，
// // 7、link标签的 rel="fluid-icon" 属性，
// // 拼接获取的所有favicon的url，传入Geticonhash函数，获取所有favicon的hash值并去重复。
// // ExtractSpareFavicon 提取 body 中 link 标签指定 favicon 的 href 属性，返回去重后的 URL 列表
func ExtractSpareFavicon(htmlContent string) ([]string, error) {
	var fav []string
	unique := make(map[string]struct{})

	// 单次匹配 rel 属性中包含 favicon 相关内容的 <link> 标签
	re := regexp.MustCompile(`<link[^>]*rel=["'](icon|shortcut icon|apple-touch-icon|apple-touch-icon-precomposed|apple-touch-startup-image|mask-icon|fluid-icon)["'][^>]*href=["'](.*?)["'][^>]*>`)
	matches := re.FindAllStringSubmatch(htmlContent, -1)

	// 提取 href 并去重
	for _, match := range matches {
		if len(match) > 2 {
			url := match[2]
			if _, exists := unique[url]; !exists {
				unique[url] = struct{}{}
				fav = append(fav, url)
			}
		}
	}

	return fav, nil
}

// 去除换行符
func RemoveNewline(str string) string {
	// 统一替换所有换行符为单一换行符
	str = strings.ReplaceAll(str, "\r\n", "\n")
	str = strings.ReplaceAll(str, "\n", "")
	str = strings.ReplaceAll(str, "\r", "")
	return str
}

// 格式化 CMS List
func FormatCmsList(cmsList []string) string {
	return fmt.Sprintf("[%s]", JoinStrings(cmsList, ", "))
}

// 将字符串切片连接成单个字符串
func JoinStrings(slice []string, separator string) string {
	result := ""
	for i, s := range slice {
		if i > 0 {
			result += separator
		}
		result += s
	}
	return result
}
