# deepseek-api

## Installation

```bash
go get github.com/gladmo/deepseek-api
```

## Document
[DeepSeek API Docs](https://api-docs.deepseek.com/api/deepseek-api)

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
	chatRequest := deepseekapi.NewChatRequest(deepseekapi.ModelTypeReasoner, deepseekapi.UserMessage("Hi"))
	chatReply, err := client.Chat(chatRequest)
	if err != nil {
		panic(err)
	}
	fmt.Println(chatReply)

	// stream-chat
	streamChatRequest := deepseekapi.NewChatRequest(deepseekapi.ModelTypeReasoner, deepseekapi.UserMessage("Hi"))
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
go run cmd/main.go -api-key /-- you deepseek API key --/ stream-chat "Hi"

<Think>
Okay, the user just said "Hi". I should respond in a friendly and welcoming manner. Let me make sure to keep the tone positive and open. Maybe start with a greeting and offer assistance. Something like, "Hello! How can I assist you today?" That should work.
<Think>

Hello! How can I assist you today?
```

### Chat
```bash
go run cmd/main.go -api-key /-- you deepseek API key --/ chat "Hi"

{"messages":[{"content":"Hi","role":"user"}],"model":"deepseek-reasoner","response_format":{"type":"text"},"stop":null,"stream":false,"stream_options":null,"tools":null,"logprobs":false}
{"id":"a3d182fe-xxxx-xxxx-xxxx-6a8c85464d02","choices":[{"finish_reason":"stop","index":0,"delta":{"content":"","reasoning_content":"","tool_calls":null,"role":""},"message":{"content":"Hello! How can I assist you today?","reasoning_content":"Okay, the user just said \"Hi\". I should respond in a friendly and welcoming manner. Let me make sure to keep it natural and open-ended so they feel comfortable asking anything. Maybe start with a greeting and offer help. Something like, \"Hello! How can I assist you today?\" That should work.","tool_calls":null,"role":"assistant"},"logprobs":{"content":null}}],"created":1739117803,"model":"deepseek-reasoner","system_fingerprint":"fp_7e73fd9a08","object":"chat.completion","usage":{"completion_tokens":74,"prompt_tokens":6,"prompt_cache_hit_tokens":0,"prompt_cache_miss_tokens":6,"total_tokens":80,"completion_tokens_details":{"reasoning_tokens":63}}}
```

### User Balance
```bash
go run cmd/main.go -api-key /-- you deepseek API key --/ balance

{"is_available":true,"balance_infos":[{"currency":"CNY","total_balance":"291.86","granted_balance":"291.86","topped_up_balance":"0.00"}]}
```

### List Models
```bash
go run cmd/main.go -api-key /-- you deepseek API key --/ list-models

{"object":"list","data":[{"id":"deepseek-chat","object":"model","owned_by":"deepseek"},{"id":"deepseek-reasoner","object":"model","owned_by":"deepseek"}]}
```
