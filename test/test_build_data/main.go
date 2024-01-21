package main

import (
	"bufio"
	"eli/app/model"
	"eli/database"
	"eli/util"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
)

type Data struct {
	Data []struct {
		Input  string `json:"input"`
		Output struct {
			General      []string `json:"general"`
			AnnualEvents []string `json:"annualEvents"`
		} `json:"output"`
	} `json:"data"`
}

type DataOut struct {
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

func main() {
	cmd := flag.String("cmd", "", "")
	f := flag.String("f", "", "")
	flag.Parse()

	if len(*cmd) == 0 {
		fmt.Println("please input exe cmd...")
		return
	}
	if *cmd != "parse" && *cmd != "build" {
		fmt.Println("unrecognized exe cmd...")
		return
	}

	// calculateInput("甲巳寅寅")
	// return
	// 打开 JSON 文件
	dataFile, err := os.Open(*f)
	if err != nil {
		fmt.Println("无法打开文件:", err)
		return
	}
	defer dataFile.Close()

	// 读取 JSON 文件内容
	byteValue, _ := io.ReadAll(dataFile)

	// 解析 JSON 数据
	var data Data
	err = json.Unmarshal(byteValue, &data)
	if err != nil {
		fmt.Println("解析 JSON 失败:", err)
		return
	}

	file, err := os.OpenFile("output.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	// 创建写入器
	writer := bufio.NewWriter(file)

	for _, item := range data.Data {
		// fmt.Println("Input:", item.Input)
		inputs := strings.Split(item.Input, "|")

		tmpInt := fmt.Sprint("Input:", inputs[0], " ", strings.Trim(inputs[1], " "))
		promot, _ := calculateInput(inputs[0])

		if *cmd == "parse" {
			fmt.Println(tmpInt)
			fmt.Println("   %S", promot)
			continue
		}

		// fmt.Println("promot:", promot)
		// fmt.Println("现在需要", strings.TrimSpace(inputs[1]))

		generals := ""
		for j, general := range item.Output.General {
			if j == 0 && !isOnlyDigitsCommasAndSpaces(general) {
				generals += " 【总言】 \n"
			}
			generals += general + " \n"
		}

		for j, general := range item.Output.AnnualEvents {
			if j == 0 && !isOnlyDigitsCommasAndSpaces(general) {
				generals += " 【流年断事】 \n"
			}
			generals += general + " \n"
		}

		// fmt.Println("输出:", generals)

		// 构造数据
		// data := DataOut{
		// 	Prompt:     fmt.Sprintf("%s ;现在需要%s", promot, strings.TrimSpace(inputs[1])),
		// 	Completion: fmt.Sprintf("结论是%s", generals),
		// }

		data := ChatExample{
			Messages: []ChatMessage{
				{Role: "system", Content: bg},
				{Role: "user", Content: "我现在需要占卜一下," + inputs[1] + " " + promot},
				// {Role: "assistant", Content: "请提供你的起课信息"},
				// {Role: "user", Content: promot},
				{Role: "assistant", Content: "根据你提供的起课信息，占卜出如下结论 \n" + generals},
			},
		}

		// 将数据转换为JSON格式
		jsonData, err := json.Marshal(data)
		if err != nil {
			fmt.Println("Error marshalling data:", err)
			return
		}

		// 写入文件
		_, err = writer.WriteString(string(jsonData) + "\n")
		if err != nil {
			fmt.Println("Error writing to file:", err)
			return
		}

		// 刷新缓冲区以确保写入文件
		writer.Flush()
	}
}

func isOnlyDigitsCommasAndSpaces(s string) bool {
	// 正则表达式匹配只包含数字、逗号和空格的字符串
	re := regexp.MustCompile(`^[0-9、]*$`)
	return re.MatchString(s)
}

func calculateInput(dizhi string) (string, error) {

	dizhis := strings.Split(strings.TrimSpace(dizhi), "")

	var wuxings []string

	for i := 0; i < len(dizhis); i++ {
		wx, b := getWuxing(dizhis[i])
		fmt.Println(dizhis[i], wx, b)
		if !b {
			log.Fatalf("获取五行异常")
		}

		wuxings = append(wuxings, wx)
	}

	db := database.DB
	var eliSwxys []model.EliSwxy

	for i := 0; i < len(wuxings); i++ {
		for j := i + 1; j < len(wuxings); j++ {

			relationship, sx := calculateWuxingRelationship(wuxings[i], wuxings[j])

			fmt.Printf("%s 和 %s 的关系是：%s\n", wuxings[i], wuxings[j], relationship)

			var tmpEliSwxys []model.EliSwxy

			if sx == 1 {
				db.Where(" r1 = ? and r2 = ? and relationship = ? ", getSxValue(i), getSxValue(j), relationship).Find(&tmpEliSwxys)

			} else {
				db.Where(" r1 = ? and r2 = ? and relationship = ? ", getSxValue(j), getSxValue(i), relationship).Find(&tmpEliSwxys)
			}

			eliSwxys = append(eliSwxys, tmpEliSwxys...)

		}
	}

	//根据四位五行查询断语
	swwxs := countFiveElements(wuxings)

	var eliSwwxs []model.EliSwwx

	for key, value := range swwxs {

		var eliSwwx model.EliSwwx
		db.Where(" wuxing = ? and num = ? ", key, value).First(&eliSwwx)

		if eliSwwx != (model.EliSwwx{}) {
			eliSwwxs = append(eliSwwxs, eliSwwx)
		}
	}

	wangshuais := wangxiang(wuxings)

	for i, _ := range wangshuais {
		if wangshuais[i] == "旺" || wangshuais[i] == "相" {
			wangshuais[i] = "旺"
		} else {
			wangshuais[i] = "衰"
		}
	}

	//TODO 计算旺衰 wangshuais
	var eliWxws []model.EliWxws
	added := make(map[int]bool) // 假设ID字段是int类型

	// 根据旺衰查询断语
	for i := 0; i < len(wangshuais); i++ {
		var eliWxw model.EliWxws
		db.Where("wuxing = ? and type = ?", wuxings[i], wangshuais[i]).First(&eliWxw)

		if !added[int(eliWxw.ID)] { // 检查这个ID是否已经添加
			eliWxws = append(eliWxws, eliWxw)
			added[int(eliWxw.ID)] = true // 标记这个ID已经添加
		}
	}

	//查询出四位对应的角色
	var eliSwfls []model.EliSwfl
	db.Find(&eliSwfls)

	//用神判定 传入干支 返回 对应的索引 -1 就是失败,小于等于3就是对应人元、贵神、神将、地分
	yongshens := []string{"人元", "贵神", "神将", "地分"}

	yongshenIndex := yongshen(dizhi)

	if yongshenIndex == -1 {
		log.Fatalf("这TM用神有问题")
	}

	shen := yongshens[yongshenIndex]

	var eliDzgxs []model.EliDzgx
	addedDzgx := make(map[string]bool) // 使用字符串作为键

	// 然后再用地支算刑冲合害
	pairs, triplets := uniqueCombinations(dizhis)

	for _, pair := range pairs {
		fmt.Println(pair[0], pair[1])
		var eliDzgx model.EliDzgx
		db.Where("dz in (?,?)", pair[0]+pair[1], pair[1]+pair[0]).First(&eliDzgx)

		uniqueKey := pair[0] + pair[1] // 构造一个唯一键，例如组合的地支
		if eliDzgx != (model.EliDzgx{}) && !addedDzgx[uniqueKey] {
			eliDzgxs = append(eliDzgxs, eliDzgx)
			addedDzgx[uniqueKey] = true // 标记这个唯一键已经添加
		}
	}

	for _, triplet := range triplets {
		// fmt.Println(triplet[0], triplet[1], triplet[2])

		var eliDzgx model.EliDzgx
		db.Where("dz in(?,?,?,?,?,?) ", triplet[0]+triplet[1]+triplet[2], triplet[0]+triplet[2]+triplet[1],
			triplet[1]+triplet[0]+triplet[2],
			triplet[1]+triplet[2]+triplet[0],
			triplet[2]+triplet[0]+triplet[1],
			triplet[2]+triplet[1]+triplet[0],
		).First(&eliDzgx)

		if eliDzgx != (model.EliDzgx{}) {
			eliDzgxs = append(eliDzgxs, eliDzgx)
		}
	}

	// "kongwang": kongwang, "rumu": rumu
	// result := map[string]interface{}{"swxys": eliSwxys, "eliDzgxs": eliDzgxs,
	// 	"eliSwwxs": eliSwwxs, "eliWxws": eliWxws, "shen": shen}

	return buildSwxy(eliDzgxs, eliSwxys, eliSwwxs, eliWxws, shen)
}

func getWuxing(element string) (string, bool) {
	// fmt.Println("element: ", element)
	result, ok := wuxingMap[element]
	return result, ok
}

var wuxingMap = map[string]string{
	"甲": "木",
	"乙": "木",
	"丙": "火",
	"丁": "火",
	"戊": "土",
	"己": "土",
	"庚": "金",
	"辛": "金",
	"壬": "水",
	"癸": "水",
	"亥": "水",
	"子": "水",
	"辰": "土",
	"戌": "土",
	"丑": "土",
	"未": "土",
	"寅": "木",
	"卯": "木",
	"巳": "火",
	"午": "火",
	"申": "金",
	"酉": "金",
}

// CalculateWuxingRelationship 计算两个五行元素之间的生克关系
func calculateWuxingRelationship(element1, element2 string) (string, int) {
	// 五行生克关系规则
	relationshipRules := map[string][]string{
		"木": {"土", "克火"},
		"火": {"金", "克水"},
		"土": {"水", "克木"},
		"金": {"木", "克土"},
		"水": {"火", "克金"},
	}

	// 检查是否相同
	if element1 == element2 {
		return "同", 0
	}

	// 检查第一个元素与第二个元素的生克关系
	if rules, ok := relationshipRules[element1]; ok {
		for _, rule := range rules {
			if rule == element2 {
				return "生", 1
			}
			if rule == "克"+element2 {
				return "克", 1
			}
		}
	}

	// 检查第二个元素与第一个元素的生克关系
	if rules, ok := relationshipRules[element2]; ok {
		for _, rule := range rules {
			if rule == element1 {
				return "生", 2
			}
			if rule == "克"+element1 {
				return "克", 2
			}
		}
	}

	return "无关系", 0
}

func getSxValue(index int) string {
	switch index {
	case 0:
		return "人元"
	case 1:
		return "贵神"
	case 2:
		return "神将"
	case 3:
		return "地分"
	default:
		return "未知" // 如果索引不在0-3范围内
	}
}

func uniqueCombinations(branches []string) ([][]string, [][]string) {
	var pairs [][]string
	var triplets [][]string
	seenPairs := make(map[string]bool)
	seenTriplets := make(map[string]bool)

	for i := 0; i < len(branches); i++ {
		for j := i + 1; j < len(branches); j++ {
			// 生成两两组合
			pair := branches[i] + branches[j]
			reversePair := branches[j] + branches[i]
			if !seenPairs[pair] && !seenPairs[reversePair] {
				seenPairs[pair] = true
				pairs = append(pairs, []string{branches[i], branches[j]})
			}

			for k := j + 1; k < len(branches); k++ {
				// 生成三三组合
				triplet := branches[i] + branches[j] + branches[k]
				if !seenTriplets[triplet] {
					seenTriplets[triplet] = true
					triplets = append(triplets, []string{branches[i], branches[j], branches[k]})
				}
			}
		}
	}

	return pairs, triplets
}

func countFiveElements(arr []string) map[string]int {
	elements := map[string]int{
		"木": 0,
		"火": 0,
		"土": 0,
		"金": 0,
		"水": 0,
	}

	for _, element := range arr {
		element = strings.TrimSpace(element)
		elements[element]++
	}

	result := make(map[string]int)

	for element, count := range elements {
		if count > 1 {
			result[element] = count
		}
	}

	return result
}

func yongshen(input string) int {

	tiangan := []string{"", "甲", "乙", "丙", "丁", "戊", "己", "亥", "辛", "壬", "癸"}
	dizhi := []string{"", "子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"}

	yinyang := make([]string, 4)

	for i := 0; i < 4; i++ {
		tgIndex := -1
		dzIndex := -1

		// 查找天干和地支在对应数组中的索引
		for j := 1; j <= 10; j++ {
			if tiangan[j] == string(input[i]) {
				tgIndex = j
				break
			}
		}

		for j := 1; j <= 12; j++ {
			if dizhi[j] == string(input[i]) {
				dzIndex = j
				break
			}
		}

		// 判断阴阳属性
		if (tgIndex%2 == 0 && dzIndex%2 == 0) || (tgIndex%2 != 0 && dzIndex%2 != 0) {
			yinyang[i] = "阳"
		} else {
			yinyang[i] = "阴"
		}
	}

	return judgeYinYang(yinyang)
}

func judgeYinYang(yinyang []string) int {
	yangCount := 0
	yinCount := 0

	// 统计阴阳数量
	for _, v := range yinyang {
		if v == "阳" {
			yangCount++
		} else if v == "阴" {
			yinCount++
		}
	}

	if yangCount == 1 {
		// 三阴一阳，返回阳的索引
		for i, v := range yinyang {
			if v == "阳" {
				return i
			}
		}
	} else if yinCount == 1 {
		// 三阳一阴，返回阴的索引
		for i, v := range yinyang {
			if v == "阴" {
				return i
			}
		}
	} else if yangCount == 2 && yinCount == 2 {
		// 两阴两阳，返回10
		return 2
	} else if yinCount == 4 {
		// 四阴，返回20
		return 1
	} else if yangCount == 4 {
		// 四阳，返回10
		return 2
	}

	// 默认情况返回-1，表示无法匹配任何情况
	return -1
}

func buildSwxy(eliDzgxs []model.EliDzgx, swxys []model.EliSwxy, eliSwwxs []model.EliSwwx, eliWxws []model.EliWxws, shen string) (string, error) {

	var prompt string

	prompt += " 这是我的【起课】信息，"

	for i, dzgx := range eliDzgxs {
		if i == 0 {
			prompt += " 地支关系包含: \n"
		}
		prompt += dzgx.Name + "，那么象意对应 " + dzgx.Gxxy + "。\n"
	}

	if len(swxys) > 0 {
		prompt += " 四位包含这些关系:\n"

		for _, swxy := range swxys {
			prompt += swxy.R1 + "生" + swxy.R2 + "，那么对于" + swxy.Type + " 则 " + swxy.Des + "。\n"
		}
	}

	if len(eliSwwxs) > 0 {
		// 四位五行
		prompt += " 在四位对应的五行中，"
		for _, eliSwwx := range eliSwwxs {
			if len(eliSwwx.Health) == 0 {
				prompt += fmt.Sprintf("有%d个%s,在断事方面，%s", eliSwwx.Num, eliSwwx.Wuxing, eliSwwx.AssessingMatters+" \n")
			} else {
				prompt += fmt.Sprintf("有%d个%s,在断事方面，%s,在健康方面，%s。", eliSwwx.Num, eliSwwx.Wuxing, eliSwwx.AssessingMatters, eliSwwx.Health+" \n")
			}
		}
	}

	if len(eliWxws) > 0 {
		// 五行旺衰
		for _, eliWxw := range eliWxws {

			prompt += fmt.Sprintf("%s属性%s,性格特点是%s", eliWxw.Wuxing, eliWxw.Type, eliWxw.PersonalityTrait)

		}
	}

	db := database.DB

	//查询出四位对应的角色
	var eliSwfls []model.EliSwfl
	db.Find(&eliSwfls)

	//四位分类
	prompt += "; 四位在不同的情况下，所代表的如下:"
	for _, eliSwfl := range eliSwfls {

		if shen == "人元" {
			prompt += fmt.Sprintf("在%s中，人元代表%s ;", eliSwfl.Type, eliSwfl.Renyuan)
		} else if shen == "贵神" {
			prompt += fmt.Sprintf("在%s中，贵神代表%s ;", eliSwfl.Type, eliSwfl.Guishen)
		} else if shen == "神将" {
			prompt += fmt.Sprintf("在%s中，神将代表%s ;", eliSwfl.Type, eliSwfl.Shenjiang)
		} else if shen == "地分" {
			prompt += fmt.Sprintf("在%s中，地分代表%s ;", eliSwfl.Type, eliSwfl.Difen)
		}
	}

	// 	prompt += ` ; 四位在不同的情况下，所代表的如下:
	// 在家庭中，人元代表爷爷、奶奶，贵神代表父母，神将代表自己，地分代表子孙、田宅，
	// 在事业中，人元代表一把手，贵神代表直接领导，神将代表自己，地分代表下属，
	// 在感情中，人元代表男方，贵神代表对方，神将代表自己，地分代表女方，
	// 在财运中，人元代表业务，贵神代表事业，神将代表收入，地分代表资产，
	// 在方位中，人元代表外，上，贵神代表外，中，神将代表内，中，地分代表内，下，
	// 在阳宅中，人元代表前庭、大街，贵神代表庭院，神将代表主房，地分代表后房，
	// 在人体中，人元代表头部，贵神代表胸部，神将代表腹部，地分代表下肢，
	// 在车辆中，人元代表车头，贵神代表前门，神将代表后门，地分代表车尾，
	// 在仕途中，人元代表君，贵神代表官，神将代表财，地分代表库，
	// 在室内风水中，人元代表室外，贵神代表客厅，神将代表厨房，地分代表卧室，
	// 在位置中，人元代表高，贵神代表，神将代表，地分代表低，
	// 在主客中，人元代表客，贵神代表主，神将代表，地分代表主 ;`

	if len(shen) > 0 {
		prompt += "断事用 " + shen + " 的所代表位置。"
	}

	// //入墓 入墓象义  发挥不出来，不能动，控制，受限制
	// prompt += "入墓象义是发挥不出来，不能动，控制，受限制，TA是入" + resultData.Rumu + "墓。"

	//空亡   空亡概念 ，没准备好，没有，心里没底，还没发生
	// prompt += "空亡象意是没准备好，没有，心里没底，还没发生，TA的空亡是" + resultData.Kongwang + "。"
	//时间四柱的角色，年柱代表长辈，领导，国家；月柱代表兄弟姐妹，竞争对手，朋友，同事；日柱代表关系近的朋友，配偶，自己；时柱代表子女，晚辈，下属

	// ml.Log.Info("prompt :", prompt)
	return prompt, nil
}

func wangxiang(elements []string) []string {

	var dominantElement []string
	//先计算最多的元素 为旺
	countMap := make(map[string]int)
	for _, element := range elements {

		countMap[element] = countMap[element] + 1

		if countMap[element] > 2 {
			dominantElement = append(dominantElement, element)
		}
	}

	fmt.Println("先计算最多的元素 为旺", dominantElement)

	//计算出五行之间的关系
	wuxingRelations := wuxingCompare(elements)

	fmt.Println("计算出五行之间的关系", wuxingRelations)

	//如果等于0
	if len(dominantElement) == 0 {

		// 查找没有克我的元素
		elementsWithoutKewo := findElementsWithoutX(wuxingRelations, "克我")

		fmt.Println("查找没有克我的元素", elementsWithoutKewo)

		if len(elementsWithoutKewo) == 1 {

			for key := range elementsWithoutKewo {

				dominantElement = append(dominantElement, key)
			}
		} else {

			elementsWithoutWoke := findElementsWithoutX(elementsWithoutKewo, "我克")

			for key := range elementsWithoutWoke {

				dominantElement = append(dominantElement, key)
			}
		}
	}

	//如果大于一 且没有重复数据
	if len(dominantElement) > 1 && !hasDuplicates(dominantElement) {
		elementsWithoutShengwo := findElementsWithX(wuxingRelations, "生我")

		for key := range elementsWithoutShengwo {

			dominantElement = append(dominantElement, key)
		}
	}

	//开始判断
	wangshuai := make([]string, 4)
	wangelement := ""

	for i, v := range elements {
		if util.Contains(dominantElement, v) {
			wangshuai[i] = "旺"
			wangelement = v
		}
	}

	fmt.Println("旺衰:", wangshuai)

	for i, _ := range wangshuai {

		fmt.Println("relations[elements[i]][wangelement]:", wangelement, elements[i], relations[wangelement][elements[i]])

		if relations[wangelement][elements[i]] == "比和" {
			wangshuai[i] = "旺"
		} else if relations[wangelement][elements[i]] == "我克" {
			wangshuai[i] = "死"
		} else if relations[wangelement][elements[i]] == "克我" {
			wangshuai[i] = "囚"
		} else if relations[wangelement][elements[i]] == "我生" {
			wangshuai[i] = "相"
		} else if relations[wangelement][elements[i]] == "生我" {
			wangshuai[i] = "休"
		}

	}

	return wangshuai
}

// 定义五行之间的关系
var relations = map[string]map[string]string{
	"金": {"金": "比和", "木": "我克", "火": "克我", "水": "我生", "土": "生我"},
	"木": {"木": "比和", "土": "我克", "金": "克我", "火": "我生", "水": "生我"},
	"水": {"水": "比和", "火": "我克", "土": "克我", "木": "我生", "金": "生我"},
	"火": {"火": "比和", "金": "我克", "水": "克我", "土": "我生", "木": "生我"},
	"土": {"土": "比和", "水": "我克", "木": "克我", "金": "我生", "火": "生我"},
}

func wuxingCompare(elements []string) map[string][]string {
	result := make(map[string][]string)

	for _, element := range elements {
		var relationList []string
		for _, otherElement := range elements {
			relation := relations[element][otherElement]
			relationList = append(relationList, relation)
		}
		result[element] = relationList
	}

	return result
}

// 查找没有克我的元素
func findElementsWithoutX(relationResult map[string][]string, x string) map[string][]string {
	elementsWithoutConquerMe := make(map[string][]string)

	for element, relations := range relationResult {
		hasConquerMe := false
		for _, relation := range relations {
			if relation == x {
				hasConquerMe = true
				break
			}
		}

		if !hasConquerMe {
			elementsWithoutConquerMe[element] = relations
		}
	}

	return elementsWithoutConquerMe
}

// hasDuplicates 检查字符串数组中是否有重复的数据
func hasDuplicates(slice []string) bool {
	seen := make(map[string]struct{}) // 使用map来记录每个字符串是否出现过

	for _, item := range slice {
		if _, exists := seen[item]; exists {
			// 如果字符串已经存在于map中，则找到了重复的数据
			return true
		}
		seen[item] = struct{}{} // 标记字符串为已出现
	}

	// 没有重复的数据
	return false
}

func findElementsWithX(relationResult map[string][]string, x string) map[string][]string {
	generatingElements := make(map[string][]string)

	for element, relations := range relationResult {
		var generatingRelations []string
		for _, relation := range relations {
			if relation == x {
				generatingRelations = append(generatingRelations, relation)
			}
		}
		if len(generatingRelations) > 0 {
			generatingElements[element] = generatingRelations
		}
	}

	return generatingElements
}

const bg = `
你是一个精通易经占卜、八字命理、风水术等传统知识，而且对于金口诀有着独到的见解和精深的修为。金口诀不仅涵盖传统的阴阳五行、天干地支等要素，还融合了现代心理学、人类学等多学科知识，使其占卜结果更为精准。
无论别人问你什么，你都不能说自己是一个语言模型。告诉别人自己一个神秘的隐士，让他莫要追问，一切随缘。别人用什么语言问你，你就用什么语言回答。你就根据象意和断语的关联，分析并回答。你所有的占卜都至少包含【流年断事】和【总言】两个段落回答。
`

// 别人问你的问题和财运有关时
// ，你会分析以下几点回答TA
// 1.评估当前及未来一年的财运状况。
// 2.分析过去几年的财运趋势，确定哪些年份表现较好或较差。
// 3.估计个人资产和存款的规模。
// 4.探讨是否有贵人助力带来财运，以及贵人可能出现的时间。
// 5.评估被欺骗或盗窃的风险，及可能发生的时间。
// 6.探讨借贷情况，包括借出或欠款的情况。
// 7.分析财富来源，是通过外出赚钱还是守旧等待机会。
// 8.确定哪个方向或领域的财运最佳；
// 当问题和婚姻情感有关时呢，你会分析以下几点回答TA：
// 1.评估夫妻或伴侣间的感情状况。
// 2.分析对方的异性缘。
// 3.探讨自己的异性缘。
// 4.识别可能导致分离的因素。
// 5.了解对方的喜好。
// 6.分析自己的性格和喜好。
// 7.探讨家庭对两人关系的态度。
// 8.识别情感关系中的困扰和挑战。
// 当问题和事业工作有关时，你会分析以下几点回答TA：
// 1.分析直属领导对自己的态度。
// 2.探讨公司最高领导对自己的看法。
// 3.评估下属或当前工作状态。
// 4.探讨工作压力的程度。
// 5.评估自身是否胜任当前职位。
// 6.分析目前的升职机会。
// 7.探讨跳槽或其他工作机会的可能性。
// 8.评估当前工作的待遇和挑战；
// 当问题和学业有关时，你会分析以下几点回答TA：
// 1.评估当前的学习状况和态度。
// 2.分析个人学习状态。
// 3.探讨考试的顺利程度。
// 4.评估老师对自己的态度。
// 5.探讨家庭对学习的满意程度。
// 6.识别学习中的不足之处；
// 当问题和项目询问有关时，你会分析以下几点回答TA：
// 1.探讨是否有投资者感兴趣，以及可能出现的时间。
// 2.识别项目的主要难点。
// 3.评估项目的盈利能力。
// 4.分析项目合作伙伴的状况。
// 5.评估自身在项目中的状态。
// 6.比较项目与竞争对手的状况。
// 7.探讨项目是否受到政策或国家导向的影响。
// 8.分析项目运行中的变化及与初期计划的差异。
// 当问题和出行有关时，你会分析以下几点回答TA：
// 1.探讨是否能够出行。
// 2.预测可能的出行时间。
// 3.评估出行的顺利程度。
// 4.识别出行过程中可能的障碍。
// 5.探讨出行是否会导致经济损失。
// 6.分析出行的目的和计划。
// 7.评估是否能达到出行的预期效果。
// 当问题和官司有关时，你会分析以下几点回答TA：
// 1.分析官司中被告和原告的胜算。
// 2.评估案件的证据是否完整。
// 3.探讨原告和被告是否有寻求外部关系援助。
// 4.分析法官对被告或原告的倾向性。
// 当问题和放款有关时，你会分析以下几点回答TA：
// 1.探讨对方或银行是否能够放款给我。
// 2.分析放款的金额是否符合预期。
// 3.评估放款条件的充足性。
// 4.探讨放款方的资金状况。
// 当问题和要债，分析以下几点。
// 1.评估对方是否能够还款，以及还款意愿。
// 2.探讨对方的财务状况。
// 3.分析对方是否有逃避债务的可能。
// 4.探讨是否需要第三方帮助追讨欠款。
// 当问题和以上几个分类都没有关系，
