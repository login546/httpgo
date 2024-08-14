package fingerprint

import (
	"fmt"
	"httpgo/pkg/httpgo"
	"httpgo/pkg/utils"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type Fingers struct {
	Url        string
	StatusCode int
	Title      string
	CmsList    []string
	OtherList  []string
	Screenshot string
}

func GetFinger(urlStr string, proxyStr string, fingerlist []utils.FingerprintFile, timeoutInt time.Duration) (*Fingers, error) {
	// 截图
	ScreenShotPath := "https://s0.wp.com/mshots/v1/" + url.QueryEscape(urlStr)

	var cms []string
	var other []string
	a, err := httpgo.GetResponse(urlStr, proxyStr, timeoutInt)
	if err != nil {
		//fmt.Println("Error making HTTP request:", err)
		return &Fingers{
			Url:        urlStr,
			StatusCode: 000,
			Title:      "",
			CmsList:    nil,
			OtherList:  nil,
			Screenshot: ScreenShotPath,
		}, nil
	}

	// 获取faviconhash
	faviconhash, err := a.GetFaviconHash(proxyStr, timeoutInt)

	if err != nil {
		//fmt.Println("Error getting favicon hash:", err)
		return &Fingers{
			Url:        urlStr,
			StatusCode: a.StatusCode,
			Title:      a.Title,
			CmsList:    nil,
			OtherList:  nil,
			Screenshot: ScreenShotPath,
		}, nil
	}

	for _, fp := range fingerlist {
		if CheckFingerprint(a, fp.Keyword, faviconhash) {
			//fmt.Printf("Matched fingerprint: %s\n", fp.Name)
			if fp.Type == "cms" {
				cms = append(cms, fp.Name)
			} else {
				other = append(other, fp.Name)
			}
		}
	}

	cmslist := httpgo.RemoveDuplicates(cms)
	otherlist := httpgo.RemoveDuplicates(other)

	return &Fingers{
		Url:        urlStr,
		StatusCode: a.StatusCode,
		Title:      utils.RemoveNewline(a.Title),
		CmsList:    cmslist,
		OtherList:  otherlist,
		Screenshot: ScreenShotPath,
	}, nil
}

// unescape 去除转义字符
func unescape(s string) string {
	re := regexp.MustCompile(`\\(.)`)
	return re.ReplaceAllString(s, "$1")
}

// evaluateCondition 检查单个条件是否匹配
func evaluateCondition(condition, respBody, respHeader, respTitle string, respCert string, iconHashes []string) bool {
	condition = strings.TrimSpace(condition)

	switch {
	case strings.HasPrefix(condition, "body="):
		value := strings.Trim(strings.TrimPrefix(condition, "body="), "\"")
		value = unescape(value)
		return strings.Contains(respBody, value)
	case strings.HasPrefix(condition, "header="):
		value := strings.Trim(strings.TrimPrefix(condition, "header="), "\"")
		value = unescape(value)
		return strings.Contains(respHeader, value)
	case strings.HasPrefix(condition, "title="):
		value := strings.Trim(strings.TrimPrefix(condition, "title="), "\"")
		value = unescape(value)
		return strings.Contains(respTitle, value)
	case strings.HasPrefix(condition, "cert="):
		value := strings.Trim(strings.TrimPrefix(condition, "cert="), "\"")
		value = unescape(value)
		return strings.Contains(respCert, value)
	case strings.HasPrefix(condition, "icon_hash="):
		value := strings.Trim(strings.TrimPrefix(condition, "icon_hash="), "\"")
		value = unescape(value)
		for _, hash := range iconHashes {
			if hash == value {
				return true
			}
		}
	case strings.HasPrefix(condition, "body!="):
		value := strings.Trim(strings.TrimPrefix(condition, "body!="), "\"")
		value = unescape(value)
		return !strings.Contains(respBody, value)
	case strings.HasPrefix(condition, "header!="):
		value := strings.Trim(strings.TrimPrefix(condition, "header!="), "\"")
		value = unescape(value)
		return !strings.Contains(respHeader, value)
	case strings.HasPrefix(condition, "title!="):
		value := strings.Trim(strings.TrimPrefix(condition, "title!="), "\"")
		value = unescape(value)
		return !strings.Contains(respTitle, value)
	case strings.HasPrefix(respCert, "cert!="):
		value := strings.Trim(strings.TrimPrefix(condition, "cert!="), "\"")
		value = unescape(value)
		return !strings.Contains(respTitle, value)
	}
	return false
}

// tokenize 函数处理表达式，将其分割成token
func tokenize(expression string) []string {
	var tokens []string
	var token strings.Builder
	inQuotes := false
	escaped := false
	parens := 0

	for i := 0; i < len(expression); i++ {
		ch := expression[i]

		if escaped {
			token.WriteByte(ch)
			escaped = false
			continue
		}

		if ch == '\\' {
			if i+1 < len(expression) && expression[i+1] == '"' {
				escaped = true
			} else {
				token.WriteByte(ch)
			}
			continue
		}

		if ch == '"' {
			if !escaped {
				inQuotes = !inQuotes
			}
			token.WriteByte(ch)
			continue
		}

		if ch == '(' {
			if inQuotes {
				token.WriteByte(ch)
			} else {
				if token.Len() > 0 {
					tokens = append(tokens, token.String())
					token.Reset()
				}
				tokens = append(tokens, string(ch))
				parens++
			}
		} else if ch == ')' {
			if inQuotes {
				token.WriteByte(ch)
			} else {
				if token.Len() > 0 {
					tokens = append(tokens, token.String())
					token.Reset()
				}
				tokens = append(tokens, string(ch))
				parens--
			}
		} else if ch == ' ' && !inQuotes {
			if token.Len() > 0 {
				tokens = append(tokens, token.String())
				token.Reset()
			}
		} else if ch == '&' && i+1 < len(expression) && expression[i+1] == '&' {
			if token.Len() > 0 {
				tokens = append(tokens, token.String())
				token.Reset()
			}
			tokens = append(tokens, "&&")
			i++
		} else if ch == '|' && i+1 < len(expression) && expression[i+1] == '|' {
			if token.Len() > 0 {
				tokens = append(tokens, token.String())
				token.Reset()
			}
			tokens = append(tokens, "||")
			i++
		} else {
			token.WriteByte(ch)
		}
	}

	if token.Len() > 0 {
		tokens = append(tokens, token.String())
	}

	if parens != 0 {
		fmt.Println("Error: Mismatched parentheses")
	}

	return tokens
}

// Shunting Yard 算法实现
func shuntingYard(expression string) ([]string, error) {
	var output []string
	var operators []string

	precedence := map[string]int{
		"||": 1,
		"&&": 2,
	}
	//fmt.Println("Expression:", expression)

	tokens := tokenize(expression)

	//fmt.Println("Tokens:", tokens) // 调试输出 tokens

	for _, token := range tokens {
		//fmt.Println("Token:", token)
		switch token {
		case "&&", "||":
			for len(operators) > 0 {
				top := operators[len(operators)-1]
				if top == "(" || precedence[token] > precedence[top] {
					break
				}
				output = append(output, top)
				operators = operators[:len(operators)-1]
			}
			operators = append(operators, token)
		case "(":
			operators = append(operators, token)
		case ")":
			for len(operators) > 0 && operators[len(operators)-1] != "(" {
				output = append(output, operators[len(operators)-1])
				operators = operators[:len(operators)-1]
			}
			if len(operators) == 0 {
				return nil, fmt.Errorf("mismatched parentheses")
			}
			operators = operators[:len(operators)-1]
		default:
			output = append(output, token)
		}
	}

	for len(operators) > 0 {
		output = append(output, operators[len(operators)-1])
		operators = operators[:len(operators)-1]
	}

	return output, nil
}

// CheckFingerprint 检查响应内容是否匹配指纹规则
func CheckFingerprint(response *httpgo.Response, expression string, faviconhashs *httpgo.FaviconList) bool {
	//expression 为fp.Keyword
	postfix, err := shuntingYard(expression)
	if err != nil {
		fmt.Println("Error parsing expression:", err)
		return false
	}
	return evaluatePostfix(postfix, response.Body, response.HeadersStr, response.Title, response.Cert, faviconhashs.FaviconHash)
}

// evaluatePostfix 评估后缀表达式
func evaluatePostfix(postfix []string, respBody, respHeader, respTitle string, respCert string, iconHashes []string) bool {
	var stack []bool

	for _, token := range postfix {
		switch token {
		case "&&":
			if len(stack) < 2 {
				fmt.Println("Error: Insufficient values for AND operation")
				return false
			}
			v1 := stack[len(stack)-2]
			v2 := stack[len(stack)-1]
			stack = stack[:len(stack)-2]
			stack = append(stack, v1 && v2)
		case "||":
			if len(stack) < 2 {
				fmt.Println("Error: Insufficient values for OR operation")
				return false
			}
			v1 := stack[len(stack)-2]
			v2 := stack[len(stack)-1]
			stack = stack[:len(stack)-2]
			stack = append(stack, v1 || v2)
		default:
			stack = append(stack, evaluateCondition(token, respBody, respHeader, respTitle, respCert, iconHashes))
		}
	}

	if len(stack) != 1 {
		fmt.Println("Error: Invalid expression")
		return false
	}
	return stack[0]
}
