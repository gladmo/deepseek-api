package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	deepseekapi "github.com/gladmo/deepseek-api"
)

var apiKey string

func init() {
	flag.StringVar(&apiKey, "api-key", "", "deepseek api key")
}

func main() {
	flag.Parse()

	if len(os.Args) < 3 {
		flag.PrintDefaults()
		return
	}

	client := deepseekapi.NewClient(apiKey)

	var err error
	switch flag.Arg(0) {
	case "balance":
		err = balance(client)
	case "list-models":
		err = listModels(client)
	case "chat":
		err = testChat(client, flag.Arg(1))
	case "stream-chat":
		err = streamChat(client, flag.Arg(1))
	case "function-call":
		err = functionCall(client, flag.Arg(1))
	default:
		flag.PrintDefaults()
		return
	}
	if err != nil {
		panic(err)
	}
}

func testChat(client *deepseekapi.Client, message string) (err error) {
	chatRequest := deepseekapi.NewChatRequest(deepseekapi.ModelTypeReasoner, deepseekapi.UserMessage(message))

	b, err := json.Marshal(chatRequest)
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	// return

	chat, err := client.Chat(chatRequest)
	if err != nil {
		return err
	}

	b, err = json.Marshal(chat)
	if err != nil {
		return err
	}

	fmt.Println(string(b))
	return
}

func streamChat(client *deepseekapi.Client, message string) (err error) {
	chatRequest := deepseekapi.NewChatRequest(deepseekapi.ModelTypeReasoner,
		deepseekapi.UserMessage(message))
	chatRequest.StreamOutput()

	isThink := true
	fmt.Println("<Think>")
	err = client.StreamChat(chatRequest, func(reply deepseekapi.ChatReply) {
		if isThink {
			if reply.Choices[0].Delta.Content != "" {
				isThink = false
				fmt.Println("\n<Think>")
				fmt.Println()
			} else {
				fmt.Print(reply.Choices[0].Delta.ReasoningContent)
				return
			}
		}
		fmt.Print(reply.Choices[0].Delta.Content)
	})
	fmt.Println()
	if err != nil {
		return err
	}
	return
}

func listModels(client *deepseekapi.Client) (err error) {
	models, err := client.ListModels()
	if err != nil {
		return
	}

	b, err := json.Marshal(models)
	if err != nil {
		return
	}

	fmt.Println(string(b))
	return
}

func balance(client *deepseekapi.Client) (err error) {
	balanceData, err := client.UserBalance()
	if err != nil {
		return err
	}

	b, err := json.Marshal(balanceData)
	if err != nil {
		return err
	}

	fmt.Println(string(b))
	return
}

var getWeatherFunc = []deepseekapi.Tool{
	{
		Type: "function",
		Function: deepseekapi.ToolFunction{
			Description: "Get weather of an location, the user shoud supply a location first",
			Name:        "get_weather",
			Parameters: deepseekapi.ToolFunctionParameters{
				Type: "object",
				Properties: map[string]any{
					"location": map[string]any{
						"type":        "string",
						"description": "The city and state, e.g. San Francisco, CA",
					},
				},
				Required: []string{"location"},
			},
		},
	},
}

func functionCall(client *deepseekapi.Client, message string) (err error) {
	messages := deepseekapi.Messages{deepseekapi.UserMessage(message)}
	chatRequest := deepseekapi.NewChatRequest(deepseekapi.ModelTypeChat, messages...)
	chatRequest.Tools = getWeatherFunc

	b, err := json.Marshal(chatRequest)
	if err != nil {
		return err
	}
	fmt.Println("req1", string(b))
	// return

	chat, err := client.Chat(chatRequest)
	if err != nil {
		return err
	}

	b, err = json.Marshal(chat)
	if err != nil {
		return err
	}
	fmt.Println("stage1", string(b))

	// call function
	replyMessage := chat.Choices[0].Message
	tool := replyMessage.ToolCalls[0]
	messages.AddMessage(deepseekapi.Message{
		Content:   replyMessage.Content,
		Role:      replyMessage.Role,
		ToolCalls: replyMessage.ToolCalls,
	})

	toolMsg, err := testFunctionCallFn(tool.Function)
	if err != nil {
		return
	}
	messages.AddMessage(deepseekapi.ToolMessage(string(toolMsg), tool.Id))

	chatRequest = deepseekapi.NewChatRequest(deepseekapi.ModelTypeChat, messages...)
	b, err = json.Marshal(chatRequest)
	if err != nil {
		return err
	}
	fmt.Println("req2", string(b))

	chat, err = client.Chat(chatRequest)
	if err != nil {
		return err
	}

	b, err = json.Marshal(chat)
	if err != nil {
		return err
	}
	fmt.Println("stage2", string(b))

	return
}

type GetWeatherCallReply struct {
	Location    string `json:"location"`
	Temperature int32  `json:"temperature"`
	Unit        string `json:"unit"`
}

func testFunctionCallFn(params deepseekapi.ReplyFunction) (reply []byte, err error) {
	switch params.Name {
	case "get_weather":
		var r GetWeatherCallReply
		err = json.Unmarshal([]byte(params.Arguments), &r)
		if err != nil {
			return
		}
		r.Temperature = 110
		r.Unit = "Celsius"
		return json.Marshal(r)
	}

	return
}
