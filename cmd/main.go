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
