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

// 提取body中的非/favicon.ico的内容
// 1、link标签的 rel="icon" 属性，
// 2、link标签的 rel="shortcut icon" 属性，
// 3、link标签的 rel="apple-touch-icon" 属性，
// 4、link标签的 rel="apple-touch-icon-precomposed" 属性，
// 5、link标签的 rel="apple-touch-startup-image" 属性，
// 6、link标签的 rel="mask-icon" 属性，
// 7、link标签的 rel="fluid-icon" 属性，
// 拼接获取的所有favicon的url，传入Geticonhash函数，获取所有favicon的hash值并去重复。
func ExtractSpareFavicon(htmlContent string) ([]string, error) {
	var fav []string

	//匹配link标签的rel属性为icon的内容
	re := regexp.MustCompile(`<link.*?rel=["']icon["'].*?>`)
	matches := re.FindAllString(htmlContent, -1)
	for i := range matches {
		re = regexp.MustCompile(`href=["'](.*?)["']`)
		m := re.FindStringSubmatch(matches[i])
		if len(m) > 1 {
			fav = append(fav, m[1])
		}
	}

	//匹配link标签的rel属性为shortcut icon的内容
	re = regexp.MustCompile(`<link.*?rel=["']shortcut icon["'].*?>`)
	matches = re.FindAllString(htmlContent, -1)
	for i := range matches {
		re = regexp.MustCompile(`href=["'](.*?)["']`)
		m := re.FindStringSubmatch(matches[i])
		if len(m) > 1 {
			fav = append(fav, m[1])
		}
	}

	//匹配link标签的rel属性为apple-touch-icon的内容
	re = regexp.MustCompile(`<link.*?rel=["']apple-touch-icon["'].*?>`)
	matches = re.FindAllString(htmlContent, -1)
	for i := range matches {
		re = regexp.MustCompile(`href=["'](.*?)["']`)
		m := re.FindStringSubmatch(matches[i])
		if len(m) > 1 {
			fav = append(fav, m[1])
		}
	}

	//匹配link标签的rel属性为apple-touch-icon-precomposed的内容
	re = regexp.MustCompile(`<link.*?rel=["']apple-touch-icon-precomposed["'].*?>`)
	matches = re.FindAllString(htmlContent, -1)
	for i := range matches {
		re = regexp.MustCompile(`href=["'](.*?)["']`)
		m := re.FindStringSubmatch(matches[i])
		if len(m) > 1 {
			fav = append(fav, m[1])
		}
	}

	//匹配link标签的rel属性为apple-touch-startup-image的内容
	re = regexp.MustCompile(`<link.*?rel=["']apple-touch-startup-image["'].*?>`)
	matches = re.FindAllString(htmlContent, -1)
	for i := range matches {
		re = regexp.MustCompile(`href=["'](.*?)["']`)
		m := re.FindStringSubmatch(matches[i])
		if len(m) > 1 {
			fav = append(fav, m[1])
		}
	}

	//匹配link标签的rel属性为mask-icon的内容
	re = regexp.MustCompile(`<link.*?rel=["']mask-icon["'].*?>`)
	matches = re.FindAllString(htmlContent, -1)
	for i := range matches {
		re = regexp.MustCompile(`href=["'](.*?)["']`)
		m := re.FindStringSubmatch(matches[i])
		if len(m) > 1 {
			fav = append(fav, m[1])
		}
	}

	//匹配link标签的rel属性为fluid-icon的内容
	re = regexp.MustCompile(`<link.*?rel=["']fluid-icon["'].*?>`)
	matches = re.FindAllString(htmlContent, -1)
	for i := range matches {
		re = regexp.MustCompile(`href=["'](.*?)["']`)
		m := re.FindStringSubmatch(matches[i])
		if len(m) > 1 {
			fav = append(fav, m[1])
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
