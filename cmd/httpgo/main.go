package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"httpgo/pkg/fingerprint"
	"httpgo/pkg/httpgo"
	"httpgo/pkg/utils"
	"os"
	"strconv"
	"strings"
	"sync"
)

const (
	reset = "\033[0m"
	green = "\033[32m"
)

func main() {
	// 定义命令行标志
	urlFlag := flag.String("url", "", "请求的url")
	fileFlag := flag.String("file", "target.txt", "请求的文件")
	proxyFlag := flag.String("proxy", "", "添加代理")
	timeoutInt := flag.Duration("timeout", 15, "超时时间")
	thead := flag.Int("thead", 20, "并发数")
	fingers := flag.String("fingers", "fingers.json", "指纹文件")
	hash := flag.String("hash", "", "计算hash")
	output := flag.String("output", "output.csv", "输出文件")

	// 解析命令行标志
	flag.Parse()
	if *hash != "" {
		hash, err := httpgo.GetResponse(*hash, "", *timeoutInt)
		if err != nil {
			fmt.Println("Error getting response:", err)
			return
		}

		icohash := utils.Mmh3Hash32([]byte(hash.Body))
		fmt.Println("icon的hash为：", icohash)
		return
	}

	fingerlist, err := utils.LoadFingerprints(*fingers)
	if err != nil {
		fmt.Println("Error loading fingerprints:", err)
		return
	}

	// 如果指定了url，则只处理单个url
	if *urlFlag != "" {
		if err != nil {
			fmt.Println("Error loading fingerprints:", err)
			return
		}

		a, err := fingerprint.GetFinger(*urlFlag, *proxyFlag, fingerlist, *timeoutInt)
		if err != nil {
			fmt.Println("Error getting fingerprint:", err)
			return
		}
		fmt.Printf("%-40s %-10s %-30s %-s\n", "URL", "Status", "Title", "CMS List")
		fmt.Printf("%-40s %-10d %-30s %s%s%s\n", a.Url, a.StatusCdoe, a.Title, green, utils.FormatCmsList(a.CmsList), reset)
		return
	}

	//获取target.txt文件中的url
	targetlist, err := utils.ReadFileToSlice(*fileFlag)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	//提取指纹
	if err != nil {
		fmt.Println("Error loading fingerprints:", err)
		return
	}

	// 创建一个通道来控制并发数量
	sem := make(chan struct{}, *thead)

	// 使用WaitGroup和goroutines并发处理URL
	var wg sync.WaitGroup

	// 创建CSV文件
	file, err := os.Create(*output)
	if err != nil {
		fmt.Println("创建CSV文件出错:", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 写入CSV表头
	header := []string{"Url", "StatusCode", "Title", "CmsList"}
	if err := writer.Write(header); err != nil {
		fmt.Println("写入CSV表头出错:", err)
		return
	}

	for _, target := range targetlist {
		wg.Add(1)
		sem <- struct{}{} // 向通道发送数据，阻塞直到通道有可用空间
		go func(url string) {
			defer wg.Done()
			defer func() { <-sem }() // 从通道读取数据，以释放空间

			a, err := fingerprint.GetFinger(url, *proxyFlag, fingerlist, *timeoutInt)
			if err != nil {
				fmt.Println("获取指纹失败:", err)
				return
			}
			fmt.Printf("%-40s %-10d %-30s %s%s%s\n", a.Url, a.StatusCdoe, a.Title, green, utils.FormatCmsList(a.CmsList), reset)

			// 将 CmsList 转换为单个字符串
			cmsListStr := strings.Join(a.CmsList, ";")
			// 将结果写入CSV文件
			record := []string{url, strconv.Itoa(a.StatusCdoe), a.Title, cmsListStr}
			if err := writer.Write(record); err != nil {
				fmt.Println("写入CSV文件出错:", err)
			}
		}(target)
	}

	wg.Wait()
}
