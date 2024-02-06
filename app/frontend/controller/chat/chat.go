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
	fmt.Println("open ai key:", defaultModelKey)
	// defaultModelName := openai.GPT3Dot5Turbo
	defaultModelName := "ft:gpt-3.5-turbo-1106:swft-blockchain::8icpgPrw"
	// defaultModelName := openai.GPT4
	// defaultModelName := "ft:gpt-3.5-turbo-1106:swft-blockchain::8huoNTCU"
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

	if character.CharNature > 0 && character.CharNature <= 200 {
		defaultTemp = float64(character.CharNature) / 100
	}
	fmt.Println("当前角色的温度设定：", defaultTemp)

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
		log.Println("ChatCompletionStream error111:", err)

		rp := makeReply(common.CODE_ERR_GPT_COMPLETE, err.Error(), timeNowHs, "", request.Timestamp, "")

		ws.WriteJSON(rp)
		return
	}
	defer stream.Close()

	// rp := makeReply(common.CODE_ERR_GPT_COMPLETE, "err.Error()", timeNowHs, "", request.Timestamp, "")
	// ws.WriteJSON(rp)
	// return

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

		// tokenLimit := 8192 - 10                       // 设定令牌限制
		currentLength := len(question)          // 计算问题长度
		currentLength += len(frontPromot + ";") // 计算问题长度

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

				// q := "Q:`" + v.Question + "`;A:`" + v.Reply + "` \n"
				// if currentLength+len(question) > tokenLimit {
				// 	break // 当累积长度超过限制时停止添加
				// }
				// backgroundContext += q
				// currentLength += len(q)

				back = append(back, openai.ChatCompletionMessage{
					Role:    "user",
					Content: v.Question,
				})
				back = append(back, openai.ChatCompletionMessage{
					Role:    "assistant",
					Content: v.Reply,
				})
			}
		}
	}

	backgroundContext += "背景: " + frontPromot + ";"
	backgroundContext += "\n" + question
	log.Println("打印用户背景问题：", backgroundContext)

	if len(backgroundContext) > 0 {
		backgroundContext = "背景信息: \n" + backgroundContext + "\n"
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

	prompt += "本次【金口诀起课】信息如下：\n"

	for i, dzgx := range resultData.EliDzgxs {
		if i == 0 {
			prompt += "地支关系为:"
		}
		prompt += dzgx.Name + "，那么象意对应 " + dzgx.Gxxy + ";"
	}
	prompt += "\n"

	//四位分类
	prompt += "将人元、贵神、神将、地分定义为四位，四位在不同的情况下，所代表意义如下:"
	for _, eliSwfl := range resultData.EliSwfls {
		if resultData.Shen == "人元" {
			prompt += fmt.Sprintf("在%s中，人元代表%s;", eliSwfl.Type, eliSwfl.Renyuan)
		} else if resultData.Shen == "贵神" {
			prompt += fmt.Sprintf("在%s中，贵神代表%s;", eliSwfl.Type, eliSwfl.Guishen)
		} else if resultData.Shen == "神将" {
			prompt += fmt.Sprintf("在%s中，神将代表%s;", eliSwfl.Type, eliSwfl.Shenjiang)
		} else if resultData.Shen == "地分" {
			prompt += fmt.Sprintf("在%s中，地分代表%s;", eliSwfl.Type, eliSwfl.Difen)
		}
	}
	prompt += "\n"

	if len(resultData.Shen) > 0 {
		prompt += "当需要断事时以" + resultData.Shen + "位置为入手分析不同位置的关系。"
	}

	for i, swxy := range resultData.Swxys {
		if i == 0 {
			prompt += "四位关系为:"
		}
		prompt += swxy.R1 + "生" + swxy.R2 + "，那么对于" + swxy.Type + " 则 " + swxy.Des + ";"
	}
	prompt += "\n"

	prompt += "五行的象意参考为:\n"
	prompt += `
	金:	代表白色，西方，辛辣，秋天，呼吸系统，筋骨，肺部分忧伤，仁义，公正，原则、法律、霸道、阻碍、改革，精细，收敛，威严，强硬，矛盾，坚硬
	木:	绿色、青色，东方，酸的，春天，肝胆，神经，筋骨，生气，仁慈，尊重，同情，耿直，发展，延伸，突破，积极，有根基，纠缠，正义，高贵 ，直接
	水:	黑色，北方，咸味，冬天，肾脏，泌尿系统，膀胱，恐吓，惊吓，聪明，有智慧，聪明好动，随遇而安，没有形状，滚动，桃花，是非，胆子小，积蓄，储备，向下，湿润，向下，寒凉，困境，低谷，随遇而安
	火:	红色，男方，苦，夏天，喜悦，心脏，血液，炎症，懂礼数，有礼貌，快速，表现，表演，玩乐，娱乐，变化，想法多，热情，有脾气，有灵感，顶峰，向上，虚幻，虚荣，热的，证书、证件、文书、信息
	土:	黄色，中间，甜的，仲季，脾胃，消化系统，思念，诚信，包容，有责任，不灵活，保守，转化，可以克服的障碍，房屋，资产，收纳，承载，保障，缓慢
	`
	prompt += "\n"

	prompt += "地支的象意参考为:\n"
	prompt += `
	子:隐蔽、不为人知，保密，流动、没有主见、不清晰、偷盗	在方位中代表正北，在人物中代表孙子、儿子，在职业中代表小偷、盗贼、黑衣人、斜眼睛的人、秘书、机密人员、保密性工作、流动性质的工作者，在自然环境中代表水、河流、冰，在描述事物中代表流动、转动，在物品中代表车、船，在工作中代表技能，在人物性格中代表圆润、聪明，在人物关系中代表异性缘好，人体中代表肾脏、耳在朵、膀胱、泌尿系统、血液、腰、精子、喉咙，在动物中代表老鼠、蝙蝠、猫头鹰，在场景地点中代表洗手间、下水道、书店、图书馆，水产品市场，在空中代表云气，在地下代表水泽，在描述一个人中代表适应能力强、适合做推广交际，聪明、变化快，头脑清醒精于算计，桃花
	丑:金融相关、官人、贤士，冤仇、田宅	在方位中代表东北偏北，在人体中代表脾脏、肚腹、嘴唇、皮肤、脚、子宫，在人物中代表老妇、贵人、尊长、神佛、大肚子的人，在动物中代表牛、螃蟹、龟，在环境中代表阴湿的地方、桥梁、银行、土坡，在物品中代表柜子、珠宝、鞋子、首饰、钥匙，在描述一个人中代表忠厚老实、性情倔强，在行业中代表金融行业
	寅:文书、财帛、官贵、清高、财神、官吏	在方位中代表东北偏东，在物品中代表树木、花草、家居、木制品，在环境场所中代表会所、楼宇、政府机关、文化场所、高雅的场所、山林，在人体中代表受、肢体、胆、毛发、神经、筋脉，在钱财方面代表大的财富、财神
	卯:交易、门户、合作、多、组合、交通相关	在方位中代表东方，物品中代表树木、花草、灌木、纺织物、木材、网络、门窗，卯又被称为门户代表门、窗，在人体中代表肝胆、四肢、毛发、腰，在钱财中代表交易、贸易、合作、结合，在婚姻情感中代表婚姻、约会，在工作事业中代表合作、结合，在描述人的时候，做事情有多手准备，狡兔三窟，桃花
	辰:争斗、对抗、执法、旧事、不好的事	在方位中代表东南偏东，在人体中代表膀胱、内分泌、肩、胃、肌肤，在行为中代表争斗、阻止、阻力，征伐、争论，在四位中有两个辰就代表争执、阻力，有五行中土所表示的意思
	巳:文书信息、惊恐、多疑、变幻、吸引注意、争斗、口舌	在方位中代表东南偏南，在婚姻情感和工作事业中代表口舌、变化、文章、幻想，在环境中代表闹市、道路，在人体中代表心脏、三焦、咽喉、眼睛、肛门，在动物中代表蛇、蚯蚓，在人物中代表小女孩、年轻女性
	午:文书、荣誉、奖项、惊恐、光明、官事	在方位中代表南方，在事物中代表火器、光彩、电子、信息、广告、文学、语言、文章，在描述一个人特点中代表热情、激动、有文化，在环境场所中代表冶炼、战场、大厦、剧场、体育场、热闹的地方、学校，在人体中代表新章、小肠、血液、精力、血液、舌头，在人物中代表成熟年龄的女性、漂亮的人，桃花
	未:平常的、吃的东西，和厨房有关、祭祀相关的事、平常常用的	在方位中代表西南偏南方，在场景环境中代表田园、酒店、工厂，在人体中代表、脾胃、肌肤、口腔，人物代表老年男性，在事件中代表宴席，在钱财中代表财帛，在疾病和长辈中代表孝服
	申:移动、奔波、阻隔、杀伐	在方位中代表西南偏西，在环境场所中代表道路、高大的楼、医院，在事物中代表疾病、官灾、掌权，在行业中代表金融机构、银行，在人体中代表肺、大肠，在动物中代表猴子，在职业中代表司法部门工作的人、军人、警察，在事物中代表有阻隔、法律、规定
	酉:精致、细腻、阴私、解散、赏赐	在方位中代表西方，人物中代表少女，在动物中代表鸟、鸡，在物品中代表器皿、金属物品、钟表、银行，描述事物中代表暗中隐藏隐瞒，暗昧不明，事物有缺陷，有口舌，在描述人的心情情绪中代表喜悦、欢乐，在场所环境中代表门、窗，在人体中代表肺、小肠、耳朵、骨骼、精气，桃花
	戌:旧事重新之象，凡事都是聚众，不是个别现象，虚耗，印绶，欺诈	在方位中代表西北偏西，在地理环境场所中代表庙宇、加油站、电站、影院、市场、坟墓，在物品中代表炉子、古物，在刑罚官司中代表牢狱，在人体代表心包、后背、肌肉、鼻，在动物中代表狗、狼，在事物中代表是非毁败、虚伪、言约私契、文书、空话大话、开会
	亥:索取、拖延、赏赐、暗昧、妄想，什么也不想干，也不知道干什么	在方位中代表西北偏北，在地理环境中代表池塘、沟渠、小河流，在人体代表头、肾脏、膀胱、尿道，在动物中代表猪、熊，在物品中代表酒、汤药，在人物中代表小孩，有小孩性格的人，在事物中代表科技、理性思维、网络、思想
	`
	prompt += "\n"

	//四位五行
	if len(resultData.EliSwwxs) > 0 {
		// 四位五行
		prompt += "四位五行中:"
		for _, eliSwwx := range resultData.EliSwwxs {
			if len(eliSwwx.Health) == 0 {
				prompt += fmt.Sprintf("有%d个%s,在断事方面，%s", eliSwwx.Num, eliSwwx.Wuxing, eliSwwx.AssessingMatters+";")
			} else {
				prompt += fmt.Sprintf("有%d个%s,在断事方面，%s,在健康方面，%s。", eliSwwx.Num, eliSwwx.Wuxing, eliSwwx.AssessingMatters, eliSwwx.Health+";")
			}
		}
	}
	prompt += "\n"

	if len(resultData.EliWxws) > 0 {
		// 五行旺衰
		for _, eliWxw := range resultData.EliWxws {
			prompt += fmt.Sprintf("%s属性%s,性格特点是%s;", eliWxw.Wuxing, eliWxw.Type, eliWxw.PersonalityTrait)
		}
	}
	prompt += "\n"

	if len(resultData.Rumu) > 0 {
		// 入墓 入墓象义  发挥不出来，不能动，控制，受限制
		prompt += "入墓象义是发挥不出来，不能动，控制，受限制，TA是 " + resultData.Rumu + " 。"
	}

	if len(resultData.Kongwang) > 0 {
		// 空亡   空亡概念 ，没准备好，没有，心里没底，还没发生
		prompt += "空亡象意是没准备好，没有，心里没底，还没发生，本次的空亡是" + resultData.Kongwang + "。"
	}

	//时间四柱的角色，年柱代表长辈，领导，国家；月柱代表兄弟姐妹，竞争对手，朋友，同事；日柱代表关系近的朋友，配偶，自己；时柱代表子女，晚辈，下属
	prompt += "用中文回答问题，以四位的生克关系为主，并结合五行和天干地支象意，以及空亡入墓和刑冲合害信息，按分类输出分析结果，并且尽可能携带流年信息。\n"
	prompt += "现在我的问题是: "

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
