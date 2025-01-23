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
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	reset = "\033[0m"
	green = "\033[32m"
	red   = "\033[31m"
)

func main() {
	// 记录开始时间
	start := time.Now()

	// 定义命令行标志
	urlFlag := flag.String("url", "", "请求的url")
	fileFlag := flag.String("file", "", "请求的文件")
	proxyFlag := flag.String("proxy", "", "添加代理")
	timeoutInt := flag.Duration("timeout", 8, "超时时间")
	thead := flag.Int("thead", 20, "并发数")
	fingers := flag.String("fingers", "fingers.json", "指纹文件")
	hash := flag.String("hash", "", "计算hash")
	output := flag.String("output", "output", "输出结果文件夹名称,不用加后缀(包含csv,json,html文件)")
	//outputhtml := flag.String("outputhtml", "report.html", "输出文件")
	server := flag.String("server", "", "指定需要远程访问的output的文件夹名称，启动web服务，自带随机密码，增加安全性")
	checkf := flag.Bool("check", false, "检查新添加指纹规则的合规性")

	//取当前路径
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	//设置运行输出符号画httpgo
	fmt.Println(`
 _       _     _                           
| |__   | |_  | |_   _ __     __ _    ___  
| '_ \  | __| | __| | '_ \   / _' |  / _ \ 
| | | | | |_  | |_  | |_) | | (_| | | (_) |
|_| |_|  \__|  \__| | .__/   \__, |  \___/ 
                    |_|      |___/  
							Version: 1.2.3
	`)

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
		go func() {
			newdir := dir + "/" + *server + "/"
			// 取随机字符串作为密码
			Spasswd := utils.GenerateRandomString(10)
			ipadd := httpgo.GetLocalIP()
			//fmt.Println("ipadd:", ipadd)
			// 获取随机未占用端口
			port, err := utils.GetRandomPort()
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("----------------------------------------------------------------------------------\n")
			fmt.Printf("已启动web服务，可直接访问下面链接，进行实时查看结果\n")
			fmt.Printf("localhost：http://127.0.0.1:%d/%s.html\n", port, *server)
			fmt.Printf("Serving：http://%s:%d/%s.html\n", ipadd, port, *server)
			fmt.Printf("UserName: admin\n")
			fmt.Printf("Password: %s\n", Spasswd)
			fmt.Printf("一键访问：http://admin:%s@127.0.0.1:%d/%s.html\n", Spasswd, port, *server)
			fmt.Printf("一键访问：http://admin:%s@%s:%d/%s.html\n", Spasswd, ipadd, port, *server)
			fmt.Printf("----------------------------------------------------------------------------------\n")
			time.Sleep(3 * time.Second)
			err = httpgo.ServeDirectoryWithAuth(newdir, "admin", Spasswd, port)
			if err != nil {
				log.Fatal(err)
			}
		}()
	}

	fingerlist, err := utils.LoadFingerprints(*fingers)
	if err != nil {
		fmt.Println("Error loading fingerprints:", err)
		return
	}

	// 检查指纹文件
	if *checkf != false {
		// 检查指纹规则
		err := fingerprint.ValidateFingerprints(fingerlist)
		if err != nil {
			fmt.Println("Fingerprint validation error:", err)
		} else {
			fmt.Println("Fingerprint validation successful")
		}
		return
	}

	// 如果指定了url，则只处理单个url
	if *urlFlag != "" {
		if err != nil {
			fmt.Println("Error getting url:", err)
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

	targetlist, err := utils.ReadFileToSlice(*fileFlag)
	if err != nil {
		//fmt.Println("Error reading file:", err)
		//return
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
	if *fileFlag != "" {
		fmt.Printf("%-40s %-10s %-30s %-10s %-10s\n", "URL", "Status", "Title", "CMSList", "OtherList")
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

	// 记录结束时间并计算耗时
	elapsed := time.Since(start)
	fmt.Printf("处理完毕，共计耗时: %s\n", elapsed)

	//捕获系统信号，保持程序运行，防止web服务关闭
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	fmt.Println("指纹识别完成，按 Ctrl+C 停止，\n如开启了web服务，请不再浏览web结果时使用 Ctrl+C 关闭，否则无法正常访问结果展示页面。")
	<-c // 等待信号
	fmt.Println("程序已退出")
}
