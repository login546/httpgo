package httpgo

import (
	"crypto/tls"
	"fmt"
	"httpgo/pkg/utils"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Response struct {
	Url        string
	StatusCode int
	Title      string
	Body       []byte
	HeadersMap map[string][]string
	HeadersStr string
	Cert       string // 添加证书字段
}

func GetResponse(urlStr string, proxyStr string, timeoutInt time.Duration) (*Response, error) {

	var httpClient *http.Client
	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		MinVersion:         tls.VersionTLS10,
		MaxVersion:         tls.VersionTLS13,
		CipherSuites: []uint16{
			tls.TLS_RSA_WITH_RC4_128_SHA,
			tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA,
			tls.TLS_RSA_WITH_AES_128_CBC_SHA,
			tls.TLS_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_RSA_WITH_AES_128_CBC_SHA256,
			tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_RC4_128_SHA,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
			tls.TLS_ECDHE_RSA_WITH_RC4_128_SHA,
			tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA,
			tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
			tls.TLS_AES_128_GCM_SHA256,
			tls.TLS_AES_256_GCM_SHA384,
			tls.TLS_CHACHA20_POLY1305_SHA256,
			tls.TLS_FALLBACK_SCSV,
		},
	}

	if proxyStr != "" {
		proxyURL, err := url.Parse(proxyStr)
		if err != nil {
			log.Println("Error parsing proxy URL:", err)
			return nil, err
		}

		httpClient = &http.Client{
			Transport: &http.Transport{
				Proxy:           http.ProxyURL(proxyURL),
				TLSClientConfig: tlsconfig,
			},
			Timeout: timeoutInt * time.Second,
		}
	} else {
		httpClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: tlsconfig,
			},
			Timeout: timeoutInt * time.Second,
		}
	}

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		log.Println("Error creating HTTP request:", err)
		return nil, err
	}

	// 设置自定义header请求头
	req.Header.Set("User-Agent", getRandomUserAgent())
	req.Header.Set("Referer", urlStr)

	resp, err := httpClient.Do(req)
	if err != nil {
		tlsconfig.CipherSuites = []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,   // 常用且支持较好的前向保密
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,   // 较高安全性，性能稍低
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256, // 更快的ECDSA验证
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384, // 高安全性支持
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,    // 高效并适合低性能设备
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,  // 高效且兼容移动设备
		}

		// Retry the request
		resp, err = httpClient.Do(req)
		if err != nil {
			//log.Println("Error making HTTP request after retry:", err)
			return &Response{
				Url:        urlStr,
				StatusCode: -1,
				Title:      "",
				Body:       nil,
				HeadersMap: nil,
				HeadersStr: "",
				Cert:       "",
			}, nil
		}
	}
	defer resp.Body.Close()

	// 获取证书信息
	var certInfo strings.Builder
	if resp.TLS != nil && len(resp.TLS.PeerCertificates) > 0 {
		for _, cert := range resp.TLS.PeerCertificates {
			certInfo.WriteString("Certificate:\n")
			certInfo.WriteString(fmt.Sprintf("  Subject: %s\n", cert.Subject))
			certInfo.WriteString(fmt.Sprintf("  Issuer: %s\n", cert.Issuer))
			certInfo.WriteString(fmt.Sprintf("  Not Before: %s\n", cert.NotBefore))
			certInfo.WriteString(fmt.Sprintf("  Not After: %s\n", cert.NotAfter))
			certInfo.WriteString(fmt.Sprintf("  Serial Number: %s\n", cert.SerialNumber))
			certInfo.WriteString(fmt.Sprintf("  Public Key Algorithm: %s\n", cert.PublicKeyAlgorithm))
			certInfo.WriteString(fmt.Sprintf("  Public Key: %x\n", cert.PublicKey))
			certInfo.WriteString(fmt.Sprintf("  Signature Algorithm: %s\n", cert.SignatureAlgorithm))
			certInfo.WriteString(fmt.Sprintf("  Version: %d\n", cert.Version))
			certInfo.WriteString(fmt.Sprintf("  OCSP Server: %s\n", cert.OCSPServer))
			certInfo.WriteString(fmt.Sprintf("  DNS Names: %s\n", cert.DNSNames))
			certInfo.WriteString(fmt.Sprintf("  Email Addresses: %s\n", cert.EmailAddresses))
			certInfo.WriteString(fmt.Sprintf("  IP Addresses: %s\n", cert.IPAddresses))
			certInfo.WriteString(fmt.Sprintf("  URIs: %s\n", cert.URIs))
			certInfo.WriteString(fmt.Sprintf("  CRL Distribution Points: %s\n", cert.CRLDistributionPoints))
			certInfo.WriteString(fmt.Sprintf("  Issuing Certificate URL: %s\n", cert.IssuingCertificateURL))
			certInfo.WriteString("\n") // 添加一个换行符用于分隔每个证书的信息
		}
	}

	//获取body内容
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		//log.Println("read from resp.Body failed, err:", err)
		return nil, err
	}

	//获取状态码
	statusCode := resp.StatusCode

	//获取title
	title, _ := utils.ExtractTitle(body)

	//获取返回包headers
	headers := make(map[string][]string)
	for k, v := range resp.Header {
		headers[k] = v
	}

	headersstr := headerToString(resp.Header)
	//fmt.Println(headersstr)

	return &Response{
		Url:        urlStr,
		StatusCode: statusCode,
		Title:      title,
		Body:       body,
		HeadersMap: headers,
		HeadersStr: headersstr,
		Cert:       certInfo.String(),
	}, nil
}

// 随机更换user-agent
// List of User-Agent strings
var userAgentList = []string{
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 11_2_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.107 Safari/537.36",
	//"Mozilla/5.0 (Linux; Android 14; SM-T970) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.6353.215 Mobile Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
	//"Mozilla/5.0 (iPhone; CPU iPhone OS 9_1 like Mac OS X) AppleWebKit/601.1.46 (KHTML, like Gecko) Version/9.0 Mobile/13B143 Safari/601.1",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
}

// 随机获取user-agent
func getRandomUserAgent() string {
	rand.Seed(time.Now().UnixNano())
	return userAgentList[rand.Intn(len(userAgentList))]
}

// header to string
func headerToString(headers http.Header) string {
	var sb strings.Builder
	for key, values := range headers {
		for _, value := range values {
			sb.WriteString(fmt.Sprintf("%s: %s\n", key, value))
		}
	}
	return sb.String()
}
