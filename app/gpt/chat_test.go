package gpt

import (
	"context"
	"eli/config"
	"errors"
	"fmt"
	"io"
	"testing"

	openai "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

func Hello(name string) string {
	return "Hello " + name
}
func TestMain(t *testing.T) {
	c := openai.NewClient(config.Get().Openai.Apikey)
	ctx := context.Background()

	// params := jsonschema.Definition{
	// 	Type: jsonschema.Object,
	// 	Properties: map[string]jsonschema.Definition{
	// 		"name": {
	// 			Type:        jsonschema.String,
	// 			Description: "someone's name",
	// 		},
	// 	},
	// 	Required: []string{"location"},
	// }

	// f := openai.FunctionDefinition{
	// 	Name:        "Hello",
	// 	Description: "say hello to someone",
	// 	Parameters:  params,
	// }
	ft := openai.FunctionCall{
		Name:      "Hello",
		Arguments: "name",
	}

	req := openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		// MaxTokens: 4096,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:         openai.ChatMessageRoleAssistant,
				Content:      "我是Jackie，生日：19990729，帮我算一下八字",
				FunctionCall: &ft,
			},
		},
		Stream: true,
	}
	stream, err := c.CreateChatCompletionStream(ctx, req)
	if err != nil {
		fmt.Printf("ChatCompletionStream error: %v\n", err)
		return
	}
	defer stream.Close()

	fmt.Printf("Stream response: ")

	var result string
	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			fmt.Println("\nStream finished")
			break
		}

		if err != nil {
			fmt.Printf("\nStream error: %v\n", err)
			return
		}

		// fmt.Printf(response.Choices[0].Delta.Content)
		result += response.Choices[0].Delta.Content
	}

	fmt.Println(result)
}

func TestWether(tv *testing.T) {
	ctx := context.Background()
	client := openai.NewClient(config.Get().Openai.Apikey)

	// describe the function & its inputs
	params := jsonschema.Definition{
		Type: jsonschema.Object,
		Properties: map[string]jsonschema.Definition{
			"location": {
				Type:        jsonschema.String,
				Description: "The city and state, e.g. San Francisco, CA",
			},
			"unit": {
				Type: jsonschema.String,
				Enum: []string{"celsius", "fahrenheit"},
			},
		},
		Required: []string{"location"},
	}
	f := openai.FunctionDefinition{
		Name:        "get_current_weather",
		Description: "Get the current weather in a given location",
		Parameters:  params,
	}
	t := openai.Tool{
		Type:     openai.ToolTypeFunction,
		Function: f,
	}

	// simulate user asking a question that requires the function
	dialogue := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleUser, Content: "What is the weather in Boston today?"},
	}
	fmt.Printf("Asking OpenAI '%v' and providing it a '%v()' function...\n",
		dialogue[0].Content, f.Name)
	resp, err := client.CreateChatCompletion(ctx,
		openai.ChatCompletionRequest{
			Model:    openai.GPT4TurboPreview,
			Messages: dialogue,
			Tools:    []openai.Tool{t},
		},
	)
	if err != nil || len(resp.Choices) != 1 {
		fmt.Printf("Completion error: err:%v len(choices):%v\n", err,
			len(resp.Choices))
		return
	}
	msg := resp.Choices[0].Message
	if len(msg.ToolCalls) != 1 {
		fmt.Printf("Completion error: len(toolcalls): %v\n", len(msg.ToolCalls))
		return
	}

	// simulate calling the function & responding to OpenAI
	dialogue = append(dialogue, msg)
	fmt.Printf("OpenAI called us back wanting to invoke our function '%v' with params '%v'\n",
		msg.ToolCalls[0].Function.Name, msg.ToolCalls[0].Function.Arguments)
	dialogue = append(dialogue, openai.ChatCompletionMessage{
		Role:       openai.ChatMessageRoleTool,
		Content:    "Sunny and 80 degrees.",
		Name:       msg.ToolCalls[0].Function.Name,
		ToolCallID: msg.ToolCalls[0].ID,
	})
	fmt.Printf("Sending OpenAI our '%v()' function's response and requesting the reply to the original question...\n",
		f.Name)
	resp, err = client.CreateChatCompletion(ctx,
		openai.ChatCompletionRequest{
			Model:    openai.GPT4TurboPreview,
			Messages: dialogue,
			Tools:    []openai.Tool{t},
		},
	)
	if err != nil || len(resp.Choices) != 1 {
		fmt.Printf("2nd completion error: err:%v len(choices):%v\n", err,
			len(resp.Choices))
		return
	}

	// display OpenAI's response to the original question utilizing our function
	msg = resp.Choices[0].Message
	fmt.Printf("OpenAI answered the original request with: %v\n",
		msg.Content)
}
