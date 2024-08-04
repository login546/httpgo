package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Fingerprint 结构体表示一个指纹规则
type Fingerprint struct {
	Name    string `json:"name"`
	Keyword string `json:"keyword"`
}

// unescape 去除转义字符
func unescape(s string) string {
	re := regexp.MustCompile(`\\(.)`)
	return re.ReplaceAllString(s, "$1")
}

// evaluateCondition 检查单个条件是否匹配
func evaluateCondition(condition, respBody, respHeader, respTitle string, iconHashes []string) bool {
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
		return strings.Contains(respTitle, value)
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

// evaluatePostfix 评估后缀表达式
func evaluatePostfix(postfix []string, respBody, respHeader, respTitle string, iconHashes []string) bool {
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
			stack = append(stack, evaluateCondition(token, respBody, respHeader, respTitle, iconHashes))
		}
	}

	if len(stack) != 1 {
		fmt.Println("Error: Invalid expression")
		return false
	}
	return stack[0]
}

// CheckFingerprint 检查响应内容是否匹配指纹规则
func CheckFingerprint(respBody, respHeader, respTitle string, iconHashes []string, fp Fingerprint) bool {
	expression := fp.Keyword

	postfix, err := shuntingYard(expression)
	if err != nil {
		fmt.Println("Error:", err)
		return false
	}
	return evaluatePostfix(postfix, respBody, respHeader, respTitle, iconHashes)
}

// readFingerprints 从JSON文件中读取指纹规则
func readFingerprints(filename string) ([]Fingerprint, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var fingerprints []Fingerprint
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&fingerprints); err != nil {
		return nil, err
	}

	return fingerprints, nil
}

func main() {
	// 从文件中读取指纹规则
	fingerprints, err := readFingerprints("fingers.json")
	if err != nil {
		fmt.Println("Error reading fingerprints:", err)
		return
	}

	// 示例响应数据
	respBody := `<link acRedseaPlatformtio123456n="/manager/loginC<strong>We're sorry but iKuai Cloud Platform doesn't ontroller.htm?act=login&id=1`
	respHeader := `realm="huawei x6781-z37`
	respTitle := "233"
	iconHashes := []string{"-3454", "444444"}

	// 检查每个指纹规则
	for _, fp := range fingerprints {
		if CheckFingerprint(respBody, respHeader, respTitle, iconHashes, fp) {
			fmt.Printf("匹配到指纹: %s\n", fp.Name)
		} else {
			//fmt.Printf("未匹配到指纹: %s\n", fp.Name)
		}
	}
}
