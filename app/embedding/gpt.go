package embedding

import (
	"bytes"
	"eli/config"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
)

const uri = "https://api.openai.com"
const defaultModel = "text-embedding-ada-002"

const pinekey = "2fc2b6b8-375c-45d1-a05d-2b43c7c26195"
const pineuri = "https://eli-412cf08.svc.gcp-starter.pinecone.io"

var apikey = config.Get().Openai.Apikey

type RpError struct {
	Message string
	Type    string
	Param   interface{}
	Code    string
}

type GPT struct {
}

type EmbedResult struct {
	Error  *RpError
	Object string                 `json:"object"`
	Model  string                 `json:"model"`
	Usage  map[string]interface{} `json:"usage"`
	Data   []EmbedData            `json:"data"`
}

type EmbedData struct {
	Index     int       `json:"index"`
	Object    string    `json:"object"`
	Embedding []float64 `json:"embedding"`
}

type EmbedQueryResult struct {
	Id       string            `json:"id"`
	Score    float64           `json:"score"`
	Values   []float64         `json:"values"`
	Metadata map[string]string `json:"metadata"`
}

type EmbedQueryMatch struct {
	Results   []string           `json:"results"`
	Matches   []EmbedQueryResult `json:"matches"`
	Namespace string             `json:"namespace"`
}

type EmbededUpsertData struct {
	QuestionId uint64
	Question   string
	Reply      string
	UserId     uint64
	DevId      string
	CharId     uint64
	CharCode   string
}

func (*GPT) Embedding(content string, model string, key string) (*EmbedResult, error) {
	body := map[string]string{
		"input": content,
		"model": model,
	}
	dataBytes, _ := json.Marshal(body)
	request, err := http.NewRequest(http.MethodPost, uri+"/v1/embeddings", bytes.NewBuffer(dataBytes))

	if err != nil {
		log.Println("Failed to create request:", err)
		return nil, err
	}

	request.Header.Set("Content-Type", "application/json")
	if len(key) > 0 {
		log.Println("use the replaced key ", key)
		request.Header.Set("Authorization", "Bearer "+key)
	} else {
		request.Header.Set("Authorization", "Bearer "+apikey)
	}
	log.Println("embedding request: ", request)
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Println("Failed to send request:", err)
		return nil, err
	}
	defer response.Body.Close()

	data, _ := io.ReadAll(response.Body)

	log.Println("embedding response: ", string(data))

	var v EmbedResult
	err = json.Unmarshal(data, &v)

	if err != nil {
		return nil, err
	}

	return &v, nil
}

func (ins *GPT) BatchUpsert(data *EmbededUpsertData, replaceKey string) error {

	// var emb []EmbedResult

	content := "question:`" + data.Question + "`;\n reply: `" + data.Reply + "`"
	emb, err := ins.Embedding(content, defaultModel, replaceKey)
	log.Println("build gpt embedding:", content, err)
	if err != nil {
		return err
	}
	// emb = append(emb, *d)

	if emb == nil {
		return nil
	}
	return ins.SaveChatEmbeddings(emb, data)
}

func (*GPT) SaveChatEmbeddings(data *EmbedResult, richData *EmbededUpsertData) error {
	if len(data.Data) == 0 {
		return errors.New("empty embeddings")
	}
	var embReq []map[string]interface{}
	// for index, _ := range data {
	// 	if len(data[index].Data) > 0 {
	// 		embReq = append(embReq, map[string]interface{}{
	// 			"id":     strconv.FormatUint(richData[index].QuestionId, 10) + "-" + strconv.FormatUint(richData[index].ReplyId, 10),
	// 			"values": data[index].Data[0].Embedding,
	// 			"metadata": map[string]string{
	// 				"user":     strconv.FormatUint(richData[index].UserId, 10),
	// 				"devid":    richData[index].DevId,
	// 				"charid":   strconv.FormatUint(richData[index].CharId, 10),
	// 				"charcode": richData[index].CharCode,
	// 			},
	// 		})
	// 	} else {
	// 		log.Println("pinecone save why:::", data[index].Data)
	// 	}
	// }

	embDa := map[string]interface{}{
		"id":     strconv.FormatUint(richData.QuestionId, 10),
		"values": data.Data[0].Embedding,
		"metadata": map[string]string{
			"user":     strconv.FormatUint(richData.UserId, 10),
			"devid":    richData.DevId,
			"charid":   strconv.FormatUint(richData.CharId, 10),
			"charcode": richData.CharCode,
			"version":  "2.0",
		},
	}
	embReq = append(embReq, embDa)
	log.Println("build pinecone upsertdata:", len(embReq))

	bytesData, _ := json.Marshal(map[string]interface{}{
		"vectors":   embReq,
		"namespace": "spwvec",
	})

	payload := strings.NewReader(string(bytesData))

	req, _ := http.NewRequest("POST", pineuri+"/vectors/upsert", payload)

	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/json")
	req.Header.Add("Api-Key", pinekey)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)

	if err != nil {
		return err
	}

	log.Println("pinecone upsert:::", string(body))

	return nil
}

func (ins *GPT) Query(id string, question string, filter map[string]string, limitation int, replaceKey string) ([]EmbedQueryResult, error) {

	r, err := ins.Embedding(question, defaultModel, replaceKey)
	if err != nil {
		return nil, err
	}
	if r.Error != nil {
		return nil, fmt.Errorf("code:%s, error:%s", r.Error.Code, r.Error.Message)
	}

	filter["version"] = "2.0" //固定查询
	queryCond := map[string]interface{}{
		"filter":          filter,
		"topK":            limitation,
		"includeMetadata": true,
		"includeValues":   false,
		"namespace":       "spwvec",
	}
	if len(id) > 0 {
		queryCond["id"] = id
	}

	queryCond["vector"] = r.Data[0].Embedding

	jsonCond, _ := json.Marshal(queryCond)

	payload := strings.NewReader(string(jsonCond))

	req, _ := http.NewRequest("POST", pineuri+"/query", payload)

	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/json")
	req.Header.Add("Api-Key", pinekey)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	var m EmbedQueryMatch
	err = json.Unmarshal(body, &m)

	return m.Matches, err
}
