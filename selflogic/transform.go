package selflogic

import (
	"eli/util"
	"fmt"
	"strings"
)

var tiangan = strings.Split("甲乙丙丁戊己庚辛壬癸", "")
var dizhi = strings.Split("子丑寅卯辰巳午未申酉戌亥", "")

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

var relations = map[string]map[string]string{
	"金": {"金": "比和", "木": "我克", "火": "克我", "水": "我生", "土": "生我"},
	"木": {"木": "比和", "土": "我克", "金": "克我", "火": "我生", "水": "生我"},
	"水": {"水": "比和", "火": "我克", "土": "克我", "木": "我生", "金": "生我"},
	"火": {"火": "比和", "金": "我克", "水": "克我", "土": "我生", "木": "生我"},
	"土": {"土": "比和", "水": "我克", "木": "克我", "金": "我生", "火": "生我"},
}

// 五行生克关系规则
var relationshipRules = map[string][]string{
	"木": {"土", "克火"},
	"火": {"金", "克水"},
	"土": {"水", "克木"},
	"金": {"木", "克土"},
	"水": {"火", "克金"},
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

type Transform struct {
}

func (Transform) CalElements(branches []string) []string {
	var result []string
	if len(branches) == 0 {
		return result
	}
	for _, v := range branches {
		result = append(result, wuxingMap[v])
	}
	return result
}

func (Transform) CalProsDec(elements []string) []string {
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

func (Transform) OutputSeq(types int, elements []string) []string {
	var seqs []string
	if len(elements) == 0 || (types != 1 && types != 2) {
		return seqs
	}
	var seqKey []int

	for _, i := range elements {
		if types == 1 { //天干表
			for index, j := range tiangan {
				if i == j {
					seqKey = append(seqKey, index)
					break
				}
			}
		} else { //地支表
			for index, j := range dizhi {
				if i == j {
					seqKey = append(seqKey, index)
					break
				}
			}
		}
	}
	n := len(seqKey)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if seqKey[j] > seqKey[j+1] {
				seqKey[j], seqKey[j+1] = seqKey[j+1], seqKey[j]
				elements[j], elements[j+1] = elements[j+1], elements[j]
			}
		}
	}
	return elements
}

// CalculateWuxingRelationship 计算两个五行元素之间的生克关系
func (Transform) CalculateWuxingRelationship(element1, element2 string) (string, int) {
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

func (t Transform) UniqueCombination(count int, eles []string, sort bool) []string {
	var helper func(int, []string, int)
	res := []string{}

	helper = func(start int, prev []string, left int) {
		if left == 0 {
			// 如果达到所需长度的组合，添加到结果中
			combo := make([]string, len(prev))
			copy(combo, prev)
			if sort {
				res = append(res, strings.Join(t.OutputSeq(2, combo), ""))
			} else {
				res = append(res, strings.Join(combo, ""))
			}
			// res = append(res, combo)
			return
		}

		for i := start; i <= len(eles)-left; i++ {
			// 递归地选择下一个元素
			helper(i+1, append(prev, eles[i]), left-1)
		}
	}

	helper(0, []string{}, count)
	return res
}
