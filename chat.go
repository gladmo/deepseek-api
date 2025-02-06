package deepseek_api

import (
	"bufio"
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
)

var (
	ModelTypeChat     = ModelType("deepseek-chat")
	ModelTypeReasoner = ModelType("deepseek-reasoner")
)

type ModelType string

// ChatRequest
// https://api-docs.deepseek.com/zh-cn/api/create-chat-completion
type ChatRequest struct {
	Messages []Message `json:"messages"`
	Model    ModelType `json:"model"`
	// >= -2 and <= 2
	// 介于 -2.0 和 2.0 之间的数字。如果该值为正，那么新 token 会根据其在已有文本中的出现频率受到相应的惩罚，降低模型重复相同内容的可能性。
	FrequencyPenalty int `json:"frequency_penalty,omitempty"`
	// 介于 1 到 8192 间的整数，限制一次请求中模型生成 completion 的最大 token 数。输入 token 和输出 token 的总长度受模型的上下文长度的限制。
	//
	// 如未指定 max_tokens参数，默认使用 4096。
	MaxTokens int `json:"max_tokens,omitempty"`
	// >= -2 and <= 2
	// 介于 -2.0 和 2.0 之间的数字。如果该值为正，那么新 token 会根据其是否已在已有文本中出现受到相应的惩罚，从而增加模型谈论新主题的可能性。
	PresencePenalty int `json:"presence_penalty,omitempty"`
	// 一个 object，指定模型必须输出的格式。
	//
	// 设置为 { "type": "json_object" } 以启用 JSON 模式，该模式保证模型生成的消息是有效的 JSON。
	//
	// 注意: 使用 JSON 模式时，你还必须通过系统或用户消息指示模型生成 JSON。否则，模型可能会生成不断的空白字符，直到生成达到令牌限制，
	// 从而导致请求长时间运行并显得“卡住”。此外，如果 finish_reason="length"，这表示生成超过了 max_tokens 或对话超过了最大上下文长度，
	// 消息内容可能会被部分截断。
	ResponseFormat ResponseFormat `json:"response_format"`
	// 一个 string 或最多包含 16 个 string 的 list，在遇到这些词时，API 将停止生成更多的 token。
	Stop []string `json:"stop"`
	// 如果设置为 True，将会以 SSE（server-sent events）的形式以流式发送消息增量。消息流以 data: [DONE] 结尾。
	Stream bool `json:"stream"`
	// 流式输出相关选项。只有在 stream 参数为 true 时，才可设置此参数。
	// 如果设置为 true，在流式消息最后的 data: [DONE] 之前将会传输一个额外的块。此块上的 usage 字段显示整个请求的 token 使用统计信息，
	// 而 choices 字段将始终是一个空数组。所有其他块也将包含一个 usage 字段，但其值为 null。
	StreamOptions *StreamOptions `json:"stream_options"`
	// <= 2
	// 代码生成/数学解题	0.0
	// 数据抽取/分析	    1.0
	// 通用对话	        1.3
	// 翻译	            1.3
	// 创意类写作/诗歌创作	1.5
	//
	// 采样温度，介于 0 和 2 之间。更高的值，如 0.8，会使输出更随机，而更低的值，如 0.2，会使其更加集中和确定。
	// 我们通常建议可以更改这个值或者更改 top_p，但不建议同时对两者进行修改。
	// https://api-docs.deepseek.com/zh-cn/quick_start/parameter_settings
	Temperature float32 `json:"temperature,omitempty"`
	// <= 1
	// 作为调节采样温度的替代方案，模型会考虑前 top_p 概率的 token 的结果。所以 0.1 就意味着只有包括在最高 10% 概率中的 token 会被考虑。
	// 我们通常建议修改这个值或者更改 temperature，但不建议同时对两者进行修改。
	TopP       float32 `json:"top_p,omitempty"`
	Tools      []Tool  `json:"tools"`
	ToolChoice string  `json:"tool_choice,omitempty"`
	// 是否返回所输出 token 的对数概率。如果为 true，则在 message 的 content 中返回每个输出 token 的对数概率。
	Logprobs bool `json:"logprobs"`
	// <= 20
	// 一个介于 0 到 20 之间的整数 N，指定每个输出位置返回输出概率 top N 的 token，
	// 且返回这些 token 的对数概率。指定此参数时，logprobs 必须为 true。
	TopLogprobs int `json:"top_logprobs,omitempty"`
}

func NewChatRequest(modelType ModelType, message ...Message) (cr *ChatRequest) {
	cr = new(ChatRequest)

	cr.Model = modelType
	cr.Messages = message
	cr.ResponseFormat = TextResponseFormat()
	return
}

func (th *ChatRequest) StreamOutput() {
	th.Stream = true
}

func (th *Client) Chat(request *ChatRequest) (reply ChatReply, err error) {
	if len(request.Messages) == 0 {
		err = ErrMessagesEmpty
		return
	}

	if request.Stream {
		err = ErrStreamChatRequest
		return
	}

	payload, err := json.Marshal(request)
	if err != nil {
		return
	}

	req, err := th.NewRequest(http.MethodPost, ChatRoute, bytes.NewBuffer(payload))
	if err != nil {
		return
	}

	res, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		err = findReplyError(res.StatusCode)
		return
	}

	err = json.NewDecoder(res.Body).Decode(&reply)
	return
}

func (th *Client) StreamChat(request *ChatRequest, f func(ChatReply)) (err error) {
	if len(request.Messages) == 0 {
		err = ErrMessagesEmpty
		return
	}

	if !request.Stream {
		err = ErrNormalChatRequest
		return
	}

	payload, err := json.Marshal(request)
	if err != nil {
		return
	}

	req, err := th.NewRequest(http.MethodPost, ChatRoute, bytes.NewBuffer(payload))
	if err != nil {
		return
	}
	req.Header.Add("Accept", "text/event-stream")

	res, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		err = findReplyError(res.StatusCode)
		return
	}

	scanner := bufio.NewScanner(res.Body)
	for scanner.Scan() {
		line := scanner.Text()

		dataIdx := strings.Index(line, ":")
		if dataIdx == -1 {
			continue
		}

		data := strings.TrimSpace(line[dataIdx+1:])
		if data == "[DONE]" {
			break
		}

		switch data {
		case "keep-alive":
			continue
		default:
			var reply ChatReply
			err = json.Unmarshal([]byte(data), &reply)
			if err != nil {
				return
			}

			f(reply)
		}
	}
	return
}

type ChatCompletionToolChoice string

type ChatCompletionNamedToolChoice struct {
	Type     string `json:"type"`
	Function struct {
		// 要调用的函数名称。
		Name string `json:"name"`
	}
}

type Tool struct {
	Type     string `json:"type"`
	Function struct {
		Description string `json:"description"`
		Name        string `json:"name"`
		parameters  struct {
			Property string `json:"property"`
		}
	}
}

type StreamOptions struct {
	IncludeUsage bool `json:"include_usage"`
}

type Messages []Message

func (th *Messages) AddMessage(m Message) {
	*th = append(*th, m)
}

type Message struct {
	Content          string `json:"content"`                     // 消息的内容
	Role             string `json:"role"`                        // 该消息的发起角色 system/user/assistant/tool
	Name             string `json:"name,omitempty"`              // 可以选填的参与者的名称，为模型提供信息以区分相同角色的参与者。
	Prefix           bool   `json:"prefix,omitempty"`            // (Beta) 设置此参数为 true，来强制模型在其回答中以此 assistant 消息中提供的前缀内容开始。
	ReasoningContent string `json:"reasoning_content,omitempty"` // (Beta) 用于 deepseek-reasoner 模型在对话前缀续写功能下，作为最后一条 assistant 思维链内容的输入。使用此功能时，prefix 参数必须设置为 true。
	ToolCallId       string `json:"tool_call_id,omitempty"`      // 此消息所响应的 tool call 的 ID。
}

func SystemMessage(content string, name ...string) Message {
	n := ""
	if len(name) != 0 {
		n = name[0]
	}

	return Message{
		Content: content,
		Role:    "system",
		Name:    n,
	}
}

func UserMessage(content string, name ...string) Message {
	n := ""
	if len(name) != 0 {
		n = name[0]
	}

	return Message{
		Content: content,
		Role:    "user",
		Name:    n,
	}
}

func AssistantMessage(content, reasoningContent string, prefix bool, name ...string) Message {
	n := ""
	if len(name) != 0 {
		n = name[0]
	}

	return Message{
		Content:          content,
		Role:             "assistant",
		ReasoningContent: reasoningContent,
		Prefix:           prefix,
		Name:             n,
	}
}

func ToolMessage(content, toolCallId string) Message {
	return Message{
		Content:    content,
		Role:       "tool",
		ToolCallId: toolCallId,
	}
}

type ResponseFormat struct {
	Type string `json:"type"`
}

func TextResponseFormat() ResponseFormat {
	return ResponseFormat{Type: "text"}
}

func JsonObjectResponseFormat() ResponseFormat {
	return ResponseFormat{Type: "json_object"}
}

type ChatReply struct {
	Id                string     `json:"id"`
	Choices           Choices    `json:"choices"`
	Created           int        `json:"created"`
	Model             string     `json:"model"`
	SystemFingerprint string     `json:"system_fingerprint"`
	Object            string     `json:"object"`
	Usage             ReplyUsage `json:"usage"`
}

type Choices []Choice

type Choice struct {
	FinishReason string       `json:"finish_reason"`
	Index        int          `json:"index"`
	Delta        ReplyMessage `json:"delta"`
	Message      ReplyMessage `json:"message"`
	Logprobs     Logprobs     `json:"logprobs"`
}

type ReplyMessage struct {
	Content          string           `json:"content"`
	ReasoningContent string           `json:"reasoning_content"`
	ToolCalls        []ReplyToolCalls `json:"tool_calls"`
	Role             string           `json:"role"`
}

type ReplyToolCalls struct {
	Id       string        `json:"id"`
	Type     string        `json:"type"`
	Function ReplyFunction `json:"function"`
}

type ReplyFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type Logprobs struct {
	Content []ReplyContent `json:"content"`
}

type ReplyContent struct {
	Token       string             `json:"token"`
	Logprob     int                `json:"logprob"`
	Bytes       []int              `json:"bytes"`
	TopLogprobs []ReplyTopLogprobs `json:"top_logprobs"`
}

type ReplyTopLogprobs struct {
	Token   string `json:"token"`
	Logprob int    `json:"logprob"`
	Bytes   []int  `json:"bytes"`
}

type ReplyUsage struct {
	CompletionTokens        int                     `json:"completion_tokens"`
	PromptTokens            int                     `json:"prompt_tokens"`
	PromptCacheHitTokens    int                     `json:"prompt_cache_hit_tokens"`
	PromptCacheMissTokens   int                     `json:"prompt_cache_miss_tokens"`
	TotalTokens             int                     `json:"total_tokens"`
	CompletionTokensDetails CompletionTokensDetails `json:"completion_tokens_details"`
}

type CompletionTokensDetails struct {
	ReasoningTokens int `json:"reasoning_tokens"`
}
