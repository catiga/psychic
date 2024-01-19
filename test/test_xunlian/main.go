package main

import (
	"bufio"
	"bytes"
	"eli/config"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
)

func main() {
	// 先把上个程序生成出来的文件转成模型格式
	// zhuanhuaData()

	// delFile("file-lKexmQOzhLVogDhdKQdNpdry")
	//上传文件
	// UploadFile("./mydata_real.jsonl")

	// //创建微调任务
	// response, err := createFineTuningJob("file-9il1XLGyJ47LQVrLfzp02Cvw", openai.GPT3Dot5Turbo1106)

	// if err != nil {
	// 	fmt.Println("Error:", err)
	// 	return
	// }

	// fmt.Println("Response:", response)

	//查询指定任务
	// chaxunFinetuning("ftjob-pfcubpjwugAsPiWSDPTWs8cT")

	// jiansuo("file-ZpHroikB3XI228ilekZ0QYpF", "./xunlian.jsonl")
}

// 上传训练文件
func UploadFile(filePath string) {
	OpenAIURL := "https://api.openai.com/v1/files"
	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	// 创建一个缓冲区来存储multipart的body
	var requestBody bytes.Buffer
	multipartWriter := multipart.NewWriter(&requestBody)

	// 添加文件
	filePart, err := multipartWriter.CreateFormFile("file", filePath)
	if err != nil {
		fmt.Println("Error creating form file:", err)
		return
	}
	_, err = io.Copy(filePart, file)
	if err != nil {
		fmt.Println("Error copying file:", err)
		return
	}

	// 添加其他字段
	_ = multipartWriter.WriteField("purpose", "fine-tune")

	// 关闭multipart writer
	multipartWriter.Close()

	// 创建请求
	req, err := http.NewRequest("POST", OpenAIURL, &requestBody)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	// 设置请求头部
	req.Header.Set("Authorization", "Bearer "+config.Get().Openai.Apikey)
	req.Header.Set("Content-Type", multipartWriter.FormDataContentType())

	// 创建HTTP客户端并发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	// 读取响应
	response, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return
	}

	// 打印响应内容
	fmt.Println("Response:", string(response))
}

type FineTuningRequest struct {
	TrainingFile string `json:"training_file"`
	Model        string `json:"model"`
}

func createFineTuningJob(trainingFile, model string) (string, error) {
	OpenAIURL := "https://api.openai.com/v1/fine_tuning/jobs"
	// 构建请求数据
	data := FineTuningRequest{
		TrainingFile: trainingFile,
		Model:        model,
	}

	// 将数据转换为JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("error marshaling data: %v", err)
	}

	// 创建请求
	req, err := http.NewRequest("POST", OpenAIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}

	// 设置请求头部
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.Get().Openai.Apikey)

	// 创建HTTP客户端并发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	response, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response: %v", err)
	}

	return string(response), nil
}

func delFile(fileID string) {
	OpenAIURL := "https://api.openai.com/v1/files/"
	client := &http.Client{}

	req, err := http.NewRequest("DELETE", OpenAIURL+fileID, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	req.Header.Add("Authorization", "Bearer "+config.Get().Openai.Apikey)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Println("File deleted successfully")
	} else {
		fmt.Println("Failed to delete file, status code:", resp.StatusCode)
	}
}

type OriginalFormat struct {
	Prompt     string `json:"prompt"`
	Completion string `json:"completion"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatExample struct {
	Messages []ChatMessage `json:"messages"`
}

func zhuanhuaData() {
	inputFilePath := "mydata.jsonl"       // 替换为你的输入文件路径
	outputFilePath := "mydata_real.jsonl" // 替换为你的输出文件路径

	inputFile, err := os.Open(inputFilePath)
	if err != nil {
		fmt.Println("Error opening input file:", err)
		return
	}
	defer inputFile.Close()

	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		fmt.Println("Error creating output file:", err)
		return
	}
	defer outputFile.Close()

	scanner := bufio.NewScanner(inputFile)
	writer := bufio.NewWriter(outputFile)
	for scanner.Scan() {
		var original OriginalFormat
		err := json.Unmarshal(scanner.Bytes(), &original)
		if err != nil {
			fmt.Println("Error unmarshaling JSON:", err)
			continue
		}

		chatExample := ChatExample{
			Messages: []ChatMessage{
				{Role: "user", Content: original.Prompt},
				{Role: "assistant", Content: original.Completion},
			},
		}

		chatJson, err := json.Marshal(chatExample)
		if err != nil {
			fmt.Println("Error marshaling chat JSON:", err)
			continue
		}

		_, err = writer.WriteString(string(chatJson) + "\n")
		if err != nil {
			fmt.Println("Error writing to output file:", err)
			return
		}
		writer.Flush()
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading input file:", err)
	}
}

func chaxunFinetuning(fineTuningJobID string) {
	OpenAIURL := "https://api.openai.com/v1/fine_tuning/jobs/"
	req, err := http.NewRequest("GET", OpenAIURL+fineTuningJobID, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	req.Header.Add("Authorization", "Bearer "+config.Get().Openai.Apikey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}

	fmt.Println("Response:", string(body))
}

func jiansuo(fileID string, outputFile string) {
	OpenAIURL := "https://api.openai.com/v1/files/"

	req, err := http.NewRequest("GET", OpenAIURL+fileID+"/content", nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	req.Header.Add("Authorization", "Bearer "+config.Get().Openai.Apikey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	outFile, err := os.Create(outputFile)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}

	fmt.Println("File downloaded successfully:", outputFile)
}
