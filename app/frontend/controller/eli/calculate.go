package eli

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"eli/app/model"
	"eli/database"
	ml "eli/middleware"
	"eli/util"
	"eli/util/chronos"

	"github.com/gin-gonic/gin"
)

// 计算四柱
func CalculateFourPillars(c *gin.Context) {
	lang := c.GetHeader("I18n-Language")
	dateTime := c.Query("date_time")

	if len(dateTime) == 0 {
		dateTime = time.Now().Format("2006-01-02 15:04:05")
	}

	// 解析时间字符串
	t, err := time.Parse("2006-01-02 15:04:05", dateTime)
	if err != nil {
		c.JSON(http.StatusOK, ml.Fail(lang, "100011"))
		return
	}

	sizhu := chronos.New(t).Lunar().EightCharacter()

	// 计算四柱八字
	yearGanZhi := sizhu[0] + sizhu[1]
	monthGanZhi := sizhu[2] + sizhu[3]
	dayGanZhi := sizhu[4] + sizhu[5]
	hourGanZhi := sizhu[6] + sizhu[7]

	// 返回结果
	// result := fmt.Sprintf("%s%s%s%s", yearGanZhi, monthGanZhi, dayGanZhi, hourGanZhi)
	// c.JSON(http.StatusOK, gin.H{"fourPillars": result})
	c.JSON(http.StatusOK, ml.Succ(lang, map[string]interface{}{"year_pillar": yearGanZhi, "month_pillar": monthGanZhi,
		"day_pillar": dayGanZhi, "hour_pillar": hourGanZhi}))
}

// 下面的地支和四柱的地支进行判断刑冲合害，第一个天干不需要

// 计算生克
func CalculateShenKe(c *gin.Context) {
	lang := c.GetHeader("I18n-Language")
	dizhi := c.PostForm("dizhi")
	kongwang := c.PostForm("kongwang") //贵神、人元
	rumu := c.PostForm("rumu")         //神将
	token := c.GetHeader("token")
	session, _ := util.ParseToken(token)

	//新增参数 空亡   空亡概念 ，没准备好，没有，心里没底，还没发生
	//新增参数 入墓 入墓象义  发挥不出来，不能动，控制，受限制
	//新增参数时间四柱 calculateFourPillars 的角色 对应 年柱代表长辈，领导，国家；月柱代表兄弟姐妹，竞争对手，朋友，同事；日柱代表关系近的朋友，配偶，自己；时柱代表子女，晚辈，下属

	//训练数据要调整 把数字转成 支 再转成对应的 干

	//拿地支算五行
	dizhis := strings.Split(dizhi, "")

	var wuxings []string

	for i := 0; i < len(dizhis); i++ {
		wx, b := getWuxing(dizhis[i])

		if !b {
			c.JSON(http.StatusOK, ml.Fail(lang, "100012"))
			return
		}

		wuxings = append(wuxings, wx)
	}

	fmt.Println("五行:", wuxings)
	wangshuais := wangxiang(wuxings)

	for i, _ := range wangshuais {
		if wangshuais[i] == "旺" || wangshuais[i] == "相" {
			wangshuais[i] = "旺"
		} else {
			wangshuais[i] = "衰"
		}
	}

	db := database.DB
	var eliSwxys []model.EliSwxy

	// 0 人元 1 贵神 2 神将 3 地分
	//计算出对应的五行 再计算出来对应的生克 再去表里查询出对应的关系记录
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
		c.JSON(http.StatusOK, ml.Fail(lang, "100013"))
		return
	}

	shen := yongshens[yongshenIndex]

	//eliSwxys 这就是对应的关系
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

	fmt.Println(eliSwxys, eliDzgxs, eliSwwxs, eliWxws)

	result := map[string]interface{}{"swxys": eliSwxys, "eliDzgxs": eliDzgxs,
		"eliSwwxs": eliSwwxs, "eliWxws": eliWxws, "kongwang": kongwang, "shen": shen, "rumu": rumu}

	jsonData, err := json.Marshal(result)
	if err != nil {
		log.Fatalf("Error occurred during marshaling. Error: %s", err.Error())
	}

	//保存结果 在之后询问的时候用到
	eliCalInfo := new(model.EliCalInfo)

	// eliCalInfo.ID, _ = util.Sf.GenerateID()
	eliCalInfo.Param = dizhi
	eliCalInfo.Result = string(jsonData)
	eliCalInfo.Type = 1
	eliCalInfo.UserID = session.Id
	eliCalInfo.CreateAt = time.Now()

	db.Create(&eliCalInfo)

	// "eliSwxys": eliSwxys, "eliDzgxs": eliDzgxs, "eliSwwxs": eliSwwxs, "eliWxws": eliWxws, "eliSwfls": eliSwfls
	c.JSON(http.StatusOK, ml.Succ(lang, map[string]interface{}{"cal_id": eliCalInfo.ID}))
}

// GetWuxing 根据前面的五行元素获取后面的五行元素
func getWuxing(element string) (string, bool) {
	fmt.Println("element: ", element)
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
		if _, exists := elements[element]; exists {
			elements[element]++
		}
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

	tiangan := []string{"", "甲", "已", "丙", "丁", "戊", "已", "亥", "辛", "壬", "癸"}
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
