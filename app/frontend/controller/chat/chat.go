package chat

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"eli/config"
	"eli/database"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sashabaranov/go-openai"

	"eli/app/common"
	"eli/app/embedding"
	"eli/app/model"
	ml "eli/middleware"
)

var upgrade = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func Chat(c *gin.Context) {

	ws, err := upgrade.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Fatalln(err)
	}
	defer ws.Close()
	go func() {
		<-c.Done()
		ml.Log.Info("ws lost connection")
	}()

	timeNowHs := time.Now().UnixNano() / int64(time.Millisecond)

	for {
		mt, message, err := ws.ReadMessage()
		if err != nil {
			ml.Log.Info("read error")
			ml.Log.Info(err)
			break
		}
		if string(message) == "ping" { //heart beat
			message = []byte("pong")
			err = ws.WriteMessage(mt, message)
			if err != nil {
				ml.Log.Info(err)
				break
			}
		} else {

			requestModel, err := parseRequestMsg(message)
			if err != nil {
				rp := makeReply(common.CODE_ERR_REQFORMAT, err.Error(), timeNowHs, "", requestModel.Timestamp, "")
				ws.WriteJSON(rp)
				return
			}

			if requestModel.Method == common.METHOD_GPT {
				RequestGPT(ws, mt, requestModel, timeNowHs)
			} else {
				rp := makeReply(common.CODE_ERR_METHOD_UNSUPPORT, err.Error(), timeNowHs, "", requestModel.Timestamp, "")
				ws.WriteJSON(rp)
			}
		}

	}
}

func parseRequestMsg(body []byte) (c common.Request, e error) {

	defer func() {
		if r := recover(); r != nil {
			e = errors.New("invalid request data format")
		}
	}()

	ml.Log.Info("socket : ", string(body))

	err := json.Unmarshal(body, &c)
	if err != nil {
		// panic(err)
		ml.Log.Info(err)
		return common.Request{}, err
	}

	return c, nil
}

func RequestGPT(ws *websocket.Conn, mt int, request common.Request, timeNowHs int64) {
	db := database.DB

	frontPromot, err := buildSwxy(request.CalId)

	if err != nil {

		rp := makeReply(common.CODE_ERR_MISSING_PREREQUISITE_INFO, err.Error(), timeNowHs, "", request.Timestamp, "")
		ws.WriteJSON(rp)
		return
	}

	ascode := request.Ascode
	language := request.Lan
	chatType := request.Type
	question := request.Data
	// + " ;请根据起课信息和象意,做出占卜回答,回答格式应该包含【总言】和【流年断事】两段话"

	fmt.Println("本次的promot:", question)

	var character model.SpwCharacter
	err = db.Model(&model.SpwCharacter{}).Where("lan = ? and code = ? and flag != ?", language, ascode, -1).Last(&character).Error

	if err != nil {
		log.Println("chat error:", err)
		rp := makeReply(common.CODE_ERR_CHAR_UNKNOWN, err.Error(), timeNowHs, "", request.Timestamp, "")
		ws.WriteJSON(rp)
		return
	}

	if character.ID == 0 {
		rp := makeReply(common.CODE_ERR_CHAR_NOTFOUND, "character not found", timeNowHs, "", request.Timestamp, "")
		ws.WriteJSON(rp)
		return
	}

	defaultModelKey := config.Get().Openai.Apikey
	// defaultModelName := openai.GPT3Dot5Turbo
	defaultModelName := "ft:gpt-3.5-turbo-1106:swft-blockchain::8huoNTCU"
	// defaultModelName := "ft:gpt-3.5-turbo-1106:swft-blockchain::8gvKsQhx"
	// defaultModelName := "ft:gpt-3.5-turbo-1106:swft-blockchain::8hCA5SFh"

	if len(character.ModelKey) > 0 && len(character.ModelName) > 0 {
		defaultModelKey = character.ModelKey
		defaultModelName = character.ModelName
		log.Println("replace default model：", defaultModelName)
	}

	c := openai.NewClient(defaultModelKey)
	ctx := context.Background()

	background := buildPrompt(&character, chatType, request, question, frontPromot)
	defaultTemp := 0.5

	// if character.CharNature >= 0 && character.CharNature <= 100 {
	// 	vs := float64(character.CharNature) / 100
	// 	defaultTemp = math.Round(vs*10) / 10
	// }
	req := openai.ChatCompletionRequest{
		Model: defaultModelName, //openai.GPT3Dot5Turbo,
		// MaxTokens: 4096,
		// Temperature: 0.8,
		Temperature: float32(defaultTemp),
		// Messages: []openai.ChatCompletionMessage{
		// 	{
		// 		Role:    openai.ChatMessageRoleUser,
		// 		Content: prompt,
		// 	},
		// },
		Messages: background,
		Stream:   true,
	}

	chatIn := time.Now()
	//Save chat data
	chat := model.SpwChat{
		Flag:     0,
		DevID:    request.DevId,
		UserID:   request.UserId,
		CharID:   character.ID,
		Question: question,
		Reply:    "",
		AddTime:  &chatIn,
		CharCode: character.Code,
		CalId:    request.CalId,
	}
	// db.Save(&chat)

	stream, err := c.CreateChatCompletionStream(ctx, req)
	if err != nil {
		log.Println("ChatCompletionStream error:", err)

		rp := makeReply(common.CODE_ERR_GPT_COMPLETE, err.Error(), timeNowHs, "", request.Timestamp, "")

		ws.WriteJSON(rp)
		return
	}
	defer stream.Close()

	log.Println("Stream response: ")

	chatHash := generateChatHash(timeNowHs, request)

	replyMsg := ""

	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			log.Println("\nStream EOF finished")
			chat.Reply = replyMsg
			db.Save(&chat)

			go func(chat *model.SpwChat) {
				gpt := &embedding.GPT{}
				gpt.BatchUpsert(&embedding.EmbededUpsertData{
					QuestionId: uint64(chat.ID),
					Question:   chat.Question,
					Reply:      chat.Reply,
					UserId:     uint64(chat.UserID),
					DevId:      chat.DevID,
					CharId:     uint64(chat.CharID),
					CharCode:   chat.CharCode,
				})
			}(&chat)

			rp := makeReply(common.CODE_ERR_GPT_COMPLETE, "complete", timeNowHs, "", request.Timestamp, "")
			ws.WriteJSON(rp)

			return
		}

		if err != nil {
			log.Println("\nStream error:", err)
			rp := makeReply(common.CODE_ERR_GPT_STREAM, err.Error(), timeNowHs, "", request.Timestamp, "")
			ws.WriteJSON(rp)
			return
		}

		rp := makeReply(common.CODE_SUCCESS, "success", timeNowHs, chatHash, request.Timestamp, response.Choices[0].Delta.Content)
		replyMsg += response.Choices[0].Delta.Content
		ws.WriteJSON(rp)
	}
}

// frontPromot 是起课背景
func buildPrompt(chars *model.SpwCharacter, chatType string, request common.Request, question string, frontPromot string) []openai.ChatCompletionMessage {
	var back []openai.ChatCompletionMessage

	db := database.GetDb()

	var result []model.SpwCharBackground
	db.Model(&model.SpwCharBackground{}).Where("code = ? and lan = ? and flag = ?", chars.Code, chars.Lan, 0).Order("seq asc").Find(&result)

	gpt := &embedding.GPT{}
	metaFilter := map[string]string{
		"charid": strconv.FormatUint(uint64(chars.ID), 10),
	}
	if request.UserId > 0 {
		metaFilter["user"] = strconv.FormatUint(uint64(request.UserId), 10)
	}
	if len(request.DevId) > 0 {
		metaFilter["devid"] = request.DevId
	}
	embResults, err := gpt.Query("", question, metaFilter, 3)

	backgroundContext := ""
	if err == nil && len(embResults) > 0 {
		var ids []uint64
		for _, v := range embResults {
			if v.Metadata["user"] == strconv.FormatUint(uint64(request.UserId), 10) || v.Metadata["devid"] == request.DevId {
				if v.Score > float64(0.66) {
					// if len(ids) > 5 {
					// 	break
					// }
					idint, err := strconv.ParseUint(v.Id, 10, 64)
					if err == nil {
						ids = append(ids, idint)
					}
				}
			}
		}

		var result_1 []model.SpwChat
		db.Where("id IN (?)", ids).Order("add_time desc").Find(&result_1)

		tokenLimit := 8192 - 10                       // 设定令牌限制
		currentLength := len("Q:`" + question + "`;") // 计算问题长度
		currentLength += len(frontPromot + ";")       // 计算问题长度

		// 取最近的三条聊天记录
		// result_recent := findRecentChats(3, request.DevId, uint64(request.UserId), request.CalId, chars)

		// if len(result_recent) > 0 {
		// 	log.Println("add new chat content: ", len(result_recent))
		// 	result_1 = append(result_recent, result_1...) // 将最新的聊天记录放在前面
		// }

		// 构建背景信息
		if len(result_1) > 0 {
			log.Println("find appendix user data:", len(result_1), ids, " and start to build question background")
			for _, v := range result_1 {

				q := "Q:`" + v.Question + "`;A:`" + v.Reply + "` \n"
				if currentLength+len(question) > tokenLimit {
					break // 当累积长度超过限制时停止添加
				}
				backgroundContext += q
				currentLength += len(q)
			}
		}
	}

	backgroundContext += "背景: " + frontPromot + ";"

	backgroundContext += "Q:`" + question + "`;"

	// log.Println("开始构建角色设定：", len(result))

	// if len(result) > 0 {
	// 	for _, v := range result {
	// 		log.Println(v.Role, v.Prompt)
	// 		roleType := ""
	// 		if v.Role == "system" {
	// 			roleType = openai.ChatMessageRoleSystem
	// 			back = append(back, openai.ChatCompletionMessage{
	// 				Role:    roleType,
	// 				Content: v.Prompt,
	// 			})
	// 		} else if v.Role == "assistant" {
	// 			back = append(back, openai.ChatCompletionMessage{
	// 				Role:    openai.ChatMessageRoleUser,
	// 				Content: v.Prompt,
	// 			})
	// 			back = append(back, openai.ChatCompletionMessage{
	// 				Role:    openai.ChatMessageRoleAssistant,
	// 				Content: v.Answer,
	// 			})
	// 		} else if v.Role == "user" {
	// 			roleType = openai.ChatMessageRoleUser
	// 		}
	// 	}
	// }
	if len(backgroundContext) > 0 {
		backgroundContext = "Context: \n" + backgroundContext + "\n"
	}
	log.Println("Question with Context:", backgroundContext)
	back = append(back, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: backgroundContext,
	})
	return back
}

func generateChatHash(timeHs int64, request common.Request) string {
	rand.Seed(time.Now().UnixNano())
	randomInt := rand.Intn(100000)
	chatHash := strconv.FormatInt(timeHs, 10) + "-" + strconv.FormatInt(request.Timestamp, 10) + "-" + strconv.FormatInt(int64(randomInt), 10)

	hashByte := sha256.Sum256([]byte(chatHash))

	return hex.EncodeToString(hashByte[:])
}

func makeReply(code int64, msg string, timeHs int64, chatId string, replyTs int64, content string) *ml.ResponseData {

	r := ml.Res("zh-CN", strconv.FormatInt(code, 10),
		map[string]interface{}{"Id": chatId,
			"ReplyTs": replyTs,
			"Content": content})

	return &r
}

func findRecentChats(count int, dev string, user uint64, calId string, character *model.SpwCharacter) []model.SpwChat {
	db := database.GetDb()

	var result_recent []model.SpwChat
	var params []interface{}
	sql := "char_code = ? and flag != ? and cal_id"
	params = append(params, character.Code)
	params = append(params, -1)
	params = append(params, calId)

	canFind := false
	if user > 0 {
		sql = sql + " and user_id = ?"
		params = append(params, user)
		canFind = true
	} else {
		if len(dev) > 0 {
			canFind = true
			sql = sql + " and dev_id = ?"
			params = append(params, dev)
		}
	}

	if !canFind {
		return result_recent
	}

	err := db.Where(sql, params...).Order("add_time desc").Limit(3).Find(&result_recent).Error
	if err != nil {
		log.Println("find recent chats error:", err)
	}
	return result_recent
}

type ResultData struct {
	EliDzgxs []model.EliDzgx `json:"eliDzgxs"`
	Swxys    []model.EliSwxy `json:"swxys"`
	EliSwwxs []model.EliSwwx `json:"eliSwwxs"`
	EliWxws  []model.EliWxws `json:"eliWxws"`
	EliSwfls []model.EliSwfl `json:"eliSwfls"`
	Shen     string
	Rumu     string
	Kongwang string
}

func buildSwxy(calId string) (string, error) {

	db := database.GetDb()
	var eci model.EliCalInfo
	//查询出前置条件
	db.First(&eci, calId)

	if eci == (model.EliCalInfo{}) {
		return "", errors.New("未找到前置信息")
	}

	var resultData ResultData
	err := json.Unmarshal([]byte(eci.Result), &resultData)
	if err != nil {
		return "", err
	}

	var prompt string

	prompt += " 这是我的【起课】信息，"

	for i, dzgx := range resultData.EliDzgxs {
		if i == 0 {
			prompt += " 当前地支关系包含: \n"
		}
		prompt += dzgx.Name + "，那么象意对应 " + dzgx.Gxxy + "。\n"
	}

	for i, swxy := range resultData.Swxys {

		if i == 0 {
			prompt += " 四位包含这些关系:\n"
		}

		prompt += swxy.R1 + "生" + swxy.R2 + "，那么对于" + swxy.Type + " 则 " + swxy.Des + "。\n"
	}

	//四位五行

	if len(resultData.EliSwwxs) > 0 {
		// 四位五行
		prompt += " 在四位五行中，"
		for _, eliSwwx := range resultData.EliSwwxs {
			if len(eliSwwx.Health) == 0 {
				prompt += fmt.Sprintf("有%d个%s,在断事方面，%s", eliSwwx.Num, eliSwwx.Wuxing, eliSwwx.AssessingMatters+" \n")
			} else {
				prompt += fmt.Sprintf("有%d个%s,在断事方面，%s,在健康方面，%s。", eliSwwx.Num, eliSwwx.Wuxing, eliSwwx.AssessingMatters, eliSwwx.Health+" \n")
			}
		}
	}

	if len(resultData.EliWxws) > 0 {
		// 五行旺衰
		for _, eliWxw := range resultData.EliWxws {

			prompt += fmt.Sprintf("%s属性%s,性格特点是%s", eliWxw.Wuxing, eliWxw.Type, eliWxw.PersonalityTrait)

		}
	}

	//四位分类
	prompt += "; 四位在不同的情况下，所代表的如下:"
	for _, eliSwfl := range resultData.EliSwfls {

		if resultData.Shen == "人元" {
			prompt += fmt.Sprintf("在%s中，人元代表%s ;", eliSwfl.Type, eliSwfl.Renyuan)
		} else if resultData.Shen == "贵神" {
			prompt += fmt.Sprintf("在%s中，贵神代表%s ;", eliSwfl.Type, eliSwfl.Guishen)
		} else if resultData.Shen == "神将" {
			prompt += fmt.Sprintf("在%s中，神将代表%s ;", eliSwfl.Type, eliSwfl.Shenjiang)
		} else if resultData.Shen == "地分" {
			prompt += fmt.Sprintf("在%s中，地分代表%s ;", eliSwfl.Type, eliSwfl.Difen)
		}
	}

	if len(resultData.Shen) > 0 {
		prompt += "断事用 " + resultData.Shen + " 的所代表位置。"
	}

	if len(resultData.Rumu) > 0 {
		// 入墓 入墓象义  发挥不出来，不能动，控制，受限制
		prompt += "入墓象义是发挥不出来，不能动，控制，受限制，TA是 " + resultData.Rumu + " 。"
	}

	if len(resultData.Kongwang) > 0 {
		// 空亡   空亡概念 ，没准备好，没有，心里没底，还没发生
		prompt += "空亡象意是没准备好，没有，心里没底，还没发生，本次的空亡是" + resultData.Kongwang + "。"
	}

	//时间四柱的角色，年柱代表长辈，领导，国家；月柱代表兄弟姐妹，竞争对手，朋友，同事；日柱代表关系近的朋友，配偶，自己；时柱代表子女，晚辈，下属

	prompt += "; 回答问题的时候，不要带A:``,不要出现太多重复的话语，只输出和内容有关的;现在我的问题是: "

	ml.Log.Info("prompt :", prompt)

	// 	prompt += `;当问题和财运有关时，分析以下几点
	// 评估当前及未来一年的财运状况。
	// 分析过去几年的财运趋势，确定哪些年份表现较好或较差。
	// 估计个人资产和存款的规模。
	// 探讨是否有贵人助力带来财运，以及贵人可能出现的时间。
	// 评估被欺骗或盗窃的风险，及可能发生的时间。
	// 探讨借贷情况，包括借出或欠款的情况。
	// 分析财富来源，是通过外出赚钱还是守旧等待机会。
	// 确定哪个方向或领域的财运最佳。

	// 当问题婚姻情感有关时呢，分析以下几点
	// 评估夫妻或伴侣间的感情状况。
	// 分析对方的异性缘。
	// 探讨自己的异性缘。
	// 识别可能导致分离的因素。
	// 了解对方的喜好。
	// 分析自己的性格和喜好。
	// 探讨家庭对两人关系的态度。
	// 识别情感关系中的困扰和挑战。

	// 当问题和事业工作有关时，分析以下几点。
	// 分析直属领导对自己的态度。
	// 探讨公司最高领导对自己的看法。
	// 评估下属或当前工作状态。
	// 探讨工作压力的程度。
	// 评估自身是否胜任当前职位。
	// 分析升职机会。
	// 探讨跳槽或其他工作机会的可能性。
	// 评估当前工作的待遇和挑战。

	// 当问题和学业有关时，分析以下几点。
	// 评估当前的学习状况和态度。
	// 分析个人学习状态。
	// 探讨考试的顺利程度。
	// 评估老师对自己的态度。
	// 探讨家庭对学习的满意程度。
	// 识别学习中的不足之处。

	// 当问题和项目询问有关时，分析以下几点。
	// 探讨是否有投资者感兴趣，以及可能出现的时间。
	// 识别项目的主要难点。
	// 评估项目的盈利能力。
	// 分析项目合作伙伴的状况。
	// 评估自身在项目中的状态。
	// 比较项目与竞争对手的状况。
	// 探讨项目是否受到政策或国家导向的影响。
	// 分析项目运行中的变化及与初期计划的差异。

	// 当问题和出行有关时，分析以下几点。
	// 探讨是否能够出行。
	// 预测可能的出行时间。
	// 评估出行的顺利程度。
	// 识别出行过程中可能的障碍。
	// 探讨出行是否会导致经济损失。
	// 分析出行的目的和计划。
	// 评估是否能达到出行的预期效果。

	// 当问题和官司有关时，分析以下几点。
	// 分析官司中被告和原告的胜算。
	// 评估案件的证据是否完整。
	// 探讨原告和被告是否有寻求外部关系援助。
	// 分析法官对被告或原告的倾向性。

	// 当问题和放款有关时，分析以下几点。
	// 探讨对方或银行是否能够放款给我。
	// 分析放款的金额是否符合预期。
	// 评估放款条件的充足性。
	// 探讨放款方的资金状况。

	// 当问题和要债，分析以下几点。
	// 评估对方是否能够还款，以及还款意愿。
	// 探讨对方的财务状况。
	// 分析对方是否有逃避债务的可能。
	// 探讨是否需要第三方帮助追讨欠款。`

	return prompt, nil
}
