package utils

import (
	"encoding/json"
	"fmt"
	"github.com/gofrs/flock"
	"os"
	"path/filepath"
)

// URLFingerprint 结构体表示每个 URL 的指纹信息
type URLFingerprint struct {
	Url        string
	StatusCode int
	Title      string
	CmsList    string
	OtherList  string
	Screenshot string
}

// HTML 模板
var HtmlHeaderA = "<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n    <meta charset=\"UTF-8\">\n    <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n    <title>httpgo Fingerprint Report</title>\n    <style>\n        body {\n            font-family: Arial, sans-serif;\n            margin: 0;\n            padding: 0;\n            background-color: #f4f4f4;\n            color: #333;\n        }\n        h1 {\n            text-align: center;\n            margin: 20px 0;\n            color: #444;\n        }\n        table {\n            width: 90%;\n            margin: 20px auto;\n            border-collapse: collapse;\n            background: #fff;\n            box-shadow: 0 0 10px rgba(0, 0, 0, 0.1);\n        }\n        table, th, td {\n            border: 1px solid #ddd;\n        }\n        th, td {\n            padding: 12px;\n            text-align: left;\n        }\n        th {\n            background-color: #f8f8f8;\n            color: #555;\n        }\n        .container {\n            display: flex;\n            justify-content: space-between;\n            align-items: flex-start;\n            padding: 10px;\n        }\n        .left {\n            flex: 1;\n            margin-right: 20px;\n            background: #fafafa;\n            padding: 15px;\n            border-radius: 8px;\n            box-shadow: 0 2px 5px rgba(0, 0, 0, 0.1);\n            max-width: 50%;\n        }\n        .right {\n            flex: 1;\n            max-width: 50%;\n            text-align: center;\n        }\n        .right img {\n            width: 40%;\n            height: auto;\n            border-radius: 8px;\n            cursor: pointer;\n            transition: opacity 0.3s;\n        }\n        .right img:hover {\n            opacity: 0.8;\n        }\n        .modal {\n            display: none;\n            position: fixed;\n            top: 0;\n            left: 0;\n            width: 100%;\n            height: 100%;\n            background-color: rgba(0, 0, 0, 0.8);\n            align-items: center;\n            justify-content: center;\n            z-index: 1000;\n        }\n        .modal-content {\n            max-width: 90%;\n            max-height: 90%;\n            position: relative;\n        }\n        .modal-content img {\n            width: 100%;\n            height: auto;\n            border: 5px solid #fff;\n            border-radius: 8px;\n        }\n        .modal-close {\n            position: absolute;\n            top: 20px;\n            right: 20px;\n            font-size: 2rem;\n            color: #fff;\n            cursor: pointer;\n            transition: color 0.3s;\n        }\n        .modal-close:hover {\n            color: #ddd;\n        }\n        .cms-info {\n            color: red;\n        }\n        .other-info {\n            color: green;\n        }\n        .stats {\n            margin: 20px auto;\n            width: 90%;\n            padding: 15px;\n            background: #fafafa;\n            border-radius: 8px;\n            box-shadow: 0 2px 5px rgba(0, 0, 0, 0.1);\n        }\n        .stats h2 {\n            margin-top: 0;\n            font-size: 1.2rem; /* 调整大小 */\n        }\n        .stats ul {\n            list-style: none;\n            padding: 0;\n            margin: 0;\n        }\n        .stats ul li {\n            margin: 5px 0;\n            font-size: 1rem; /* 调整大小 */\n        }\n        .button-group {\n            display: flex;\n            flex-wrap: wrap;\n            /* justify-content: center; */\n            margin: 20px 0;\n        }\n        .button-group button {\n            background-color: #007bff;\n            color: white;\n            border: none;\n            padding: 6px 12px; /* 减少内边距 */\n            margin: 4px; /* 减少外边距 */\n            border-radius: 4px; /* 减小圆角 */\n            cursor: pointer;\n            transition: background-color 0.3s;\n            font-size: 0.875rem; /* 调整字体大小 */\n        }\n\n        .button-group button:hover {\n            background-color: #0056b3;\n        }\n\n        #scroll-to-top {\n            position: fixed;\n            bottom: 20px;\n            right: 20px;\n            background-color: #007bff;\n            color: white;\n            border: none;\n            border-radius: 50%;\n            width: 40px; /* 减少宽度 */\n            height: 40px; /* 减少高度 */\n            display: flex;\n            align-items: center;\n            justify-content: center;\n            cursor: pointer;\n            font-size: 18px; /* 调整字体大小 */\n            box-shadow: 0 4px 8px rgba(0, 0, 0, 0.2);\n            transition: background-color 0.3s, box-shadow 0.3s;\n        }\n        \n        #scroll-to-top:hover {\n            background-color: #0056b3;\n            box-shadow: 0 6px 12px rgba(0, 0, 0, 0.3);\n        }\n\n    </style>\n    <script>\n        document.addEventListener(\"DOMContentLoaded\", function() {\n        const scrollToTopButton = document.getElementById(\"scroll-to-top\");\n                \n        scrollToTopButton.addEventListener(\"click\", function() {\n            window.scrollTo({\n                top: 0,\n                behavior: \"smooth\"\n            });\n        });\n        \n        // Show or hide the button based on scroll position\n        window.addEventListener(\"scroll\", function() {\n            if (window.scrollY > 300) {\n                scrollToTopButton.style.display = \"flex\";\n            } else {\n                scrollToTopButton.style.display = \"none\";\n            }\n        });\n        });\n\n        document.addEventListener(\"DOMContentLoaded\", function() {\n            let originalData = [];\n\n            function openModal(src) {\n                var modal = document.getElementById(\"modal\");\n                var modalImg = document.getElementById(\"modal-img\");\n                modal.style.display = \"flex\";\n                modalImg.src = src;\n            }\n\n            function closeModal(event) {\n                if (event.target === document.getElementById(\"modal\")) {\n                    document.getElementById(\"modal\").style.display = \"none\";\n                }\n            }\n\n            function updateStats(data) {\n                const cmsCount = {};\n                const otherCount = {};\n\n                data.forEach(item => {\n                    item.CmsList.split(';').forEach(cms => {\n                        cms = cms.trim();\n                        if (cms) {\n                            cmsCount[cms] = (cmsCount[cms] || 0) + 1;\n                        }\n                    });\n\n                    item.OtherList.split(';').forEach(other => {\n                        other = other.trim();\n                        if (other) {\n                            otherCount[other] = (otherCount[other] || 0) + 1;\n                        }\n                    });\n                });\n\n                const cmsStats = Object.entries(cmsCount).sort((a, b) => b[1] - a[1])\n                    .map(([key, value]) => `<button class=\"cms-item\" data-type=\"cms\" data-value=\"${key}\">${key}: ${value}</button>`)\n                    .join('');\n                document.getElementById('cms-stats').innerHTML = `<h2>CMS Fingerprint Information</h2><div class=\"button-group\">${cmsStats}</div>`;\n\n                const otherStats = Object.entries(otherCount).sort((a, b) => b[1] - a[1])\n                    .map(([key, value]) => `<button class=\"other-item\" data-type=\"other\" data-value=\"${key}\">${key}: ${value}</button>`)\n                    .join('');\n                document.getElementById('other-stats').innerHTML = `<br><h2>OTHER Fingerprint Information</h2><div class=\"button-group\">${otherStats}</div>`;\n\n                document.getElementById('all-stats').innerHTML = `<br><h2>All Fingerprint Information</h2><div class=\"button-group\"><button id=\"btn-all\">ALL</button></div>`;\n            }\n\n            function filterData(data, type, value) {\n                return data.filter(item => {\n                    if (type === 'cms') {\n                        return item.CmsList.split(';').map(cms => cms.trim()).includes(value);\n                    } else if (type === 'other') {\n                        return item.OtherList.split(';').map(other => other.trim()).includes(value);\n                    }\n                    return false;\n                });\n            }\n\n            function updateTable(data) {\n                const tableBody = document.querySelector(\"tbody\");\n                tableBody.innerHTML = '';\n                data.forEach(item => {\n                    const row = document.createElement('tr');\n                    row.innerHTML = `\n                        <td class=\"container\">\n                            <div class=\"left\">\n                                <p><strong>目标:</strong> <a href=\"${item.Url}\" target=\"_blank\">${item.Url}</a></p>\n                                <p><strong>状态码:</strong> ${item.StatusCode}</p>\n                                <p><strong>标题:</strong> ${item.Title}</p>\n                                <p><strong>CMS指纹信息:</strong> <span class=\"cms-info\">${item.CmsList}</span></p>\n                                <p><strong>OTHER信息:</strong> <span class=\"other-info\">${item.OtherList}</span></p>\n                            </div>\n                            <div class=\"right\">\n                                ${item.Screenshot ? `<img src=\"${item.Screenshot}\" alt=\"Screenshot\" onclick=\"openModal('${item.Screenshot}')\" loading=\"lazy\">` : `<p>No Screenshot</p>`}\n                            </div>\n                        </td>\n                    `;\n                    tableBody.appendChild(row);\n                });\n            }\n\n            function updateAllButton(data) {\n                const allCount = data.length;\n                const allButton = document.getElementById('btn-all');\n                allButton.textContent = `ALL (${allCount})`;\n            }\n\n            document.addEventListener(\"click\", function(event) {\n                if (event.target.classList.contains('cms-item') || event.target.classList.contains('other-item')) {\n                    const type = event.target.getAttribute('data-type');\n                    const value = event.target.getAttribute('data-value');\n                    const filteredData = filterData(originalData, type, value);\n                    updateTable(filteredData);\n                } else if (event.target.id === 'btn-all') {\n                    updateTable(originalData);\n                }\n            });\n\n            fetch('"
var HtmlHeaderB = "')\n                .then(response => {\n                    if (!response.ok) {\n                        throw new Error('Network response was not ok');\n                    }\n                    return response.json();\n                })\n                .then(data => {\n                    originalData = data;\n                    updateStats(data);\n                    updateTable(data);\n                    updateAllButton(data);\n                })\n                .catch(error => console.error('Error loading JSON data:', error));\n        });\n    </script>\n</head>\n<body>\n    <h1>URL Fingerprint Report</h1>\n    <div class=\"stats\">\n        <div id=\"cms-stats\"></div>\n        <div id=\"other-stats\"></div>\n        <div id=\"all-stats\"></div>\n    </div>\n    <div id=\"modal\" class=\"modal\">\n        <div class=\"modal-content\">\n            <span class=\"modal-close\">&times;</span>\n            <img id=\"modal-img\" src=\"\" alt=\"Screenshot\">\n        </div>\n    </div>\n    <table>\n        <thead>\n            <tr>\n                <th>Details</th>\n            </tr>\n        </thead>\n        <tbody>\n            <!-- Data rows will be inserted here by JavaScript -->\n        </tbody>\n    </table>\n    <button id=\"scroll-to-top\" title=\"Go to Top\">&#8679;</button>\n</body>\n</html>\n"

// 创建 HTML 报告
func InitializeHTMLReport(filename string) (*os.File, error) {
	// 去除.html后缀
	reportJson := filename[:len(filename)-5] + ".json"
	var HtmlHeader = HtmlHeaderA + reportJson + HtmlHeaderB
	file, err := os.Create(filename)
	if err != nil {
		return nil, err
	}
	_, err = file.WriteString(HtmlHeader)
	if err != nil {
		file.Close()
		return nil, err
	}
	return file, nil
}

// ReplaceHTMLWithJSON 函数将 HTML 文件名的扩展名替换为 JSON 扩展名
func ReplaceHTMLWithJSON(filePath string) string {
	// 获取文件名和扩展名
	base := filepath.Base(filePath)
	ext := filepath.Ext(base)
	nameWithoutExt := base[:len(base)-len(ext)]

	// 生成新的 JSON 文件名
	return filepath.Join(filepath.Dir(filePath), nameWithoutExt+".json")
}

// AppendJSONReport 将 URLFingerprint 数据追加到指定的 JSON 文件中
func AppendJSONReport(filename string, data URLFingerprint) error {
	locker := flock.New(filename + ".lock")

	// 获取文件锁
	if err := locker.Lock(); err != nil {
		return fmt.Errorf("无法获取文件锁: %v", err)
	}
	defer func() {
		if err := locker.Unlock(); err != nil {
			fmt.Printf("解锁失败: %v\n", err)
		}
	}()

	var existingData []URLFingerprint

	// 读取现有文件内容
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	fileContent, err := os.ReadFile(filename)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// 解码现有内容
	if len(fileContent) > 0 {
		if err := json.Unmarshal(fileContent, &existingData); err != nil {
			return fmt.Errorf("无法解码现有 JSON 内容: %v", err)
		}
	}

	// 添加新的数据
	existingData = append(existingData, data)

	// 创建临时文件并写入数据
	tempFile, err := os.Create(filename + ".tmp")
	if err != nil {
		return err
	}
	defer tempFile.Close()

	encoder := json.NewEncoder(tempFile)
	encoder.SetIndent("", "    ")

	if err := encoder.Encode(existingData); err != nil {
		return err
	}

	// 用临时文件替换原文件
	if err := os.Rename(filename+".tmp", filename); err != nil {
		return err
	}

	return nil
}
