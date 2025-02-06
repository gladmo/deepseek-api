# deepseek-api

## Installation

```bash
go get github.com/edznux/deepseek-api
```

## Usage

```bash
package main

import (
	"fmt"

	deepseekapi "github.com/gladmo/deepseek-api"
)

func main() {
	client := deepseekapi.NewClient( /* apiKey */ )

	// chat
	chatRequest := deepseekapi.NewChatRequest(deepseekapi.ModelTypeReasoner, deepseekapi.UserMessage("你好"))
	chatReply, err := client.Chat(chatRequest)
	if err != nil {
		panic(err)
	}
	fmt.Println(chatReply)

	// stream-chat
	streamChatRequest := deepseekapi.NewChatRequest(deepseekapi.ModelTypeReasoner, deepseekapi.UserMessage("你好"))
	streamChatRequest.StreamOutput()

	isThink := true
	fmt.Println("<Think>")
	err = client.StreamChat(streamChatRequest, func(chatReply deepseekapi.ChatReply) {
		if isThink {
			if chatReply.Choices[0].Delta.Content != "" {
				isThink = false
				fmt.Println("\n<Think>")
				fmt.Println()
			} else {
				fmt.Print(chatReply.Choices[0].Delta.ReasoningContent)
				return
			}
		}
		fmt.Print(chatReply.Choices[0].Delta.Content)
	})
	if err != nil {
		panic(err)
	}
	fmt.Println()

	// list-models
	models, err := client.ListModels()
	if err != nil {
		panic(err)
	}
	fmt.Println(models)

	// balance
	balanceReply, err := client.UserBalance()
	if err != nil {
		panic(err)
	}
	fmt.Println(balanceReply)
}
```

## Command Line Interface
```bash
git clone github.com/gladmo/deepseek-api
cd deepseek-api
```

### Stream Chat
```bash
go run cmd/main.go -api-key /-- you deepseek API key --/ stream-chat "你好"
```

### Chat
```bash
go run cmd/main.go -api-key /-- you deepseek API key --/ chat "你好"
```

### User Balance
```bash
go run cmd/main.go -api-key /-- you deepseek API key --/ balance
```

### List Models
```bash
go run cmd/main.go -api-key /-- you deepseek API key --/ list-models
```