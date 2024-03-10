package gpt

import (
	"context"
	"fmt"
	"testing"

	"github.com/baidubce/bce-qianfan-sdk/go/qianfan"
)

// https://cloud.baidu.com/doc/WENXINWORKSHOP/s/xlmokikxe
func TestBaidu(t *testing.T) {

	fmt.Println("------------")

	qianfan.GetConfig().AccessKey = "ALTAK5GweQGJvRHeFnqObjbQjB"
	qianfan.GetConfig().SecretKey = "ca83392464cf4bf8a93db6e12021f99f"

	// 调用默认模型，即 ERNIE-Bot-turbo
	chat := qianfan.NewChatCompletion()
	resp, _ := chat.Do(
		context.TODO(),
		&qianfan.ChatCompletionRequest{
			Messages: []qianfan.ChatCompletionMessage{
				qianfan.ChatCompletionUserMessage("你好！"),
			},
		},
	)
	fmt.Println(resp.Result)

}
