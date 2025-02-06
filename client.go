package deepseek_api

import (
	"errors"
	"io"
	"net/http"
)

const (
	ApiHost = "https://api.deepseek.com"

	ChatRoute    = "/chat/completions"
	ModelRoute   = "/models"
	BalanceRoute = "/user/balance"
)

var (
	ErrMessagesEmpty     = errors.New("messages is empty")
	ErrNormalChatRequest = errors.New("normal chat request must use Chat method")
	ErrStreamChatRequest = errors.New("stream chat request must use StreamChat method")
)

type Client struct {
	ApiKey string
}

func NewClient(apiKey string) *Client {
	return &Client{ApiKey: apiKey}
}

func (th *Client) NewRequest(method, route string, body io.Reader) (*http.Request, error) {
	url := ApiHost + route
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+th.ApiKey)
	return req, nil
}

// deepSeekError
// https://api-docs.deepseek.com/zh-cn/quick_start/error_codes
var deepSeekError = map[int]DeepSeekError{
	400: {
		Code:    400,
		Message: "格式错误",
		Cause:   "请求体格式错误",
		Fix:     "请根据错误信息提示修改请求体",
	},
	401: {
		Code:    401,
		Message: "认证失败",
		Cause:   "API key 错误，认证失败",
		Fix:     "请检查您的 API key 是否正确，如没有 API key，请先 创建 API key",
	},
	402: {
		Code:    402,
		Message: "余额不足",
		Cause:   "账号余额不足",
		Fix:     "请确认账户余额，并前往 充值 页面进行充值",
	},
	422: {
		Code:    422,
		Message: "参数错误",
		Cause:   "请求体参数错误",
		Fix:     "请根据错误信息提示修改相关参数",
	},
	429: {
		Code:    429,
		Message: "请求速率达到上限",
		Cause:   "请求速率（TPM 或 RPM）达到上限",
		Fix:     "请合理规划您的请求速率",
	},
	500: {
		Code:    500,
		Message: "服务器故障",
		Cause:   "服务器内部故障",
		Fix:     "请等待后重试。若问题一直存在，请联系我们解决",
	},
	503: {
		Code:    503,
		Message: "服务器繁忙",
		Cause:   "服务器负载过高",
		Fix:     "请稍后重试您的请求",
	},
}

type DeepSeekError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Cause   string `json:"cause"`
	Fix     string `json:"fix"`
}

func findReplyError(code int) DeepSeekError {
	if err, ok := deepSeekError[code]; ok {
		return err
	}
	return DeepSeekError{
		Code:    code,
		Message: "未知错误",
		Cause:   "未知错误",
		Fix:     "请联系我们解决",
	}
}
