package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"httpgo/pkg/fingerprint"
	"httpgo/pkg/httpgo"
	"httpgo/pkg/utils"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
)

const (
	reset = "\033[0m"
	green = "\033[32m"
	red   = "\033[31m"
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
	output := flag.String("output", "output", "输出结果文件夹名称,不用加后缀(包含csv,json,html文件)")
	//outputhtml := flag.String("outputhtml", "report.html", "输出文件")
	server := flag.String("server", "", "指定output路径，启动web服务，自带随机密码，增加安全性")

	//取当前路径
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	// 解析命令行标志
	flag.Parse()
	if *hash != "" {
		hashx, err := httpgo.GetResponse(*hash, *proxyFlag, *timeoutInt)
		if err != nil {
			fmt.Println("Error getting response:", err)
			return
		}
		ahash := utils.IconHash(hashx.Body)
		icohash := utils.Mmh3Hash32(ahash)
		fmt.Printf("icon_hash=\"%s\"\n", icohash)
		return
	}
	// 如果指定了server，则启动web服务
	if *server != "" {
		newdir := dir + "/" + *server + "/"
		// 取随机字符串作为密码
		Spasswd := utils.GenerateRandomString(10)
		fmt.Printf("Serving：http://127.0.0.1:%d/%s.html\n", 6231, *server)
		fmt.Printf("Serving：http://0.0.0.1:%d/%s.html\n", 6231, *server)
		fmt.Printf("UserInfo: admin/%s\n", Spasswd)
		err := httpgo.ServeDirectoryWithAuth(newdir, "admin", Spasswd, 6231)
		if err != nil {
			log.Fatal(err)
		}
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
		fmt.Printf("%-20s %-10s %-20s %-10s %-10s\n", "URL", "Status", "Title", "CMS List", "Other List")
		fmt.Printf("%-20s %-10d %-20s %s%-10s%s %s%-10s%s\n", a.Url, a.StatusCode, a.Title, green, utils.FormatCmsList(a.CmsList), reset, red, utils.FormatCmsList(a.OtherList), reset)
		return
	}

	//获取target.txt文件中的url
	targetlist, err := utils.ReadFileToSlice(*fileFlag)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	// 创建一个通道来控制并发数量
	sem := make(chan struct{}, *thead)

	// 使用WaitGroup和goroutines并发处理URL
	var wg sync.WaitGroup

	outdir := dir + "/" + *output
	outdircsv := outdir + "/" + *output + ".csv"
	outdirhtml := outdir + "/" + *output + ".html"
	outdirjson := outdir + "/" + *output + ".json"
	outhtmljson := *output + ".json"

	// 创建目录（如果不存在）
	err = os.MkdirAll(outdir, os.ModePerm)
	if err != nil {
		fmt.Println("创建目录出错:", err)
		return
	}

	// 替换.html为.json
	reportJson := outdirjson

	// 创建HTML报告
	htmlfile, err := utils.InitializeHTMLReport(outdirhtml, outhtmljson)
	if err != nil {
		fmt.Println("Error creating HTML report:", err)
		return
	}
	defer htmlfile.Close()

	// 创建CSV文件
	file, err := os.Create(outdircsv)
	if err != nil {
		fmt.Println("创建CSV文件出错:", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 写入CSV表头
	header := []string{"Url", "StatusCode", "Title", "CmsList", "OtherList"}
	if err := writer.Write(header); err != nil {
		fmt.Println("写入CSV表头出错:", err)
		return
	}
	fmt.Printf("%-40s %-10s %-30s %-10s %-10s\n", "URL", "Status", "Title", "CMS List", "Other List")

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

			fmt.Printf("%-40s %-10d %-30s %s%-10s%s %s%-10s%s\n", a.Url, a.StatusCode, a.Title, green, utils.FormatCmsList(a.CmsList), reset, red, utils.FormatCmsList(a.OtherList), reset)

			// 将 CmsList 转换为单个字符串
			cmsListStr := strings.Join(a.CmsList, ";")
			otherListStr := strings.Join(a.OtherList, ";")
			// 将结果写入CSV文件
			record := []string{url, strconv.Itoa(a.StatusCode), a.Title, cmsListStr, otherListStr}
			if err := writer.Write(record); err != nil {
				fmt.Println("写入CSV文件出错:", err)
			}

			reports := utils.URLFingerprint{
				Url:        a.Url,
				StatusCode: a.StatusCode,
				Title:      a.Title,
				CmsList:    cmsListStr,
				OtherList:  otherListStr,
				Screenshot: a.Screenshot,
			}
			// 保存.json文件
			if err := utils.AppendJSONReport(reportJson, reports); err != nil {
				fmt.Println("写入JSON报告出错:", err)
			}

		}(target)
	}

	wg.Wait()
	fmt.Println("处理完毕")
}
