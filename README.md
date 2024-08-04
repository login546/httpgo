# httpgo

## 关于
httpgo是一个web指纹识别工具，支持多线程、添加代理、批量识别、保存结果。可自行添加指纹。

## 使用
### 帮助
```
[shym]% ./httpgo -h

Usage of ./httpgo:
  -file string
        请求的文件 (default "target.txt")
  -fingers string
        指纹文件 (default "fingers.json")
  -hash string
        计算hash
  -output string
        输出文件 (default "output.csv")
  -proxy string
        添加代理
  -thead int
        并发数 (default 20)
  -timeout duration
        超时时间 (default 15)
  -url string
        请求的url
```
### 单个url识别
```
[shym]% ./httpgo -url https://exmail.qq.com/         
URL   Status     Title                CMS List
https://exmail.qq.com/  00  腾讯企业邮箱  [【DigiCert-Cert】, 【Baidu- webmaster platform】, 【腾讯企业邮箱】]
```

### 批量url识别
-file 指定批量url文件，每行一个url

-output 保存结果到文件，仅支持csv格式，未设置默认输出到output.csv

-thead 指定并发数，未设置默认20
```
[shym]% ./httpgo -file target.txt -output test.csv
https://email.163.com/                   200        网易免费邮箱 - 你的专业电子邮局              [【DigiCert-Cert】, 【Alibaba - Ali CDN】]
https://exmail.qq.com/                   200        腾讯企业邮箱                         [【DigiCert-Cert】, 【Baidu- webmaster platform】, 【腾讯企业邮箱】]
https://www.baidu.com/                   200        百度一下，你就知道                      []
```


## 指纹规则

~~~
title="xxxxx" 匹配title的内容
header="bbbb"	匹配响应标头的内容
icon_hash="1111111"	匹配favico.ico图标hash内容
body="cccc"	匹配body中的内容
cert="dddd"	匹配证书中内容
body="xxxx" && header!="ccc" 匹配body中包含xxxx并且header中不包含ccc的内容

=为包含关系，即包含关系即可匹配
!=为不包含关系，即不包含关系即可匹配

支持逻辑&& 以及 || 和 ()
比如
body=\"aaaa\" && (title=\"123\" || title=\"456\" )

双引号"记得转义，如果是搜索的具体内容里有"需要加使用\\\",如
body=\"<link href=\\\"/jcms/\" 匹配的为body中是否包含<link href="/jcms/

{
  "name": "jcms or fcms",
  "keyword": "body=\"<link href=\\\"/jcms/\" || body=\"<link href=\\\"/fcms/\" || body=\"jcms/Login.do\" || body=\"fcms/Login.do\""
}
~~~



