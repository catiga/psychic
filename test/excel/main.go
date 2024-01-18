package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/tealeg/xlsx"
)

func main() {
	// 打开Excel文件
	excelFileName := "/Users/wanggaowei/Downloads/4位关系.xlsx"
	xlFile, err := xlsx.OpenFile(excelFileName)
	if err != nil {
		log.Fatalf("无法打开Excel文件: %s\n", err)
	}

	var params []map[string]string

	// 遍历所有工作表
	for _, sheet := range xlFile.Sheets {
		// fmt.Printf("工作表名称: %s\n", sheet.Name)

		if sheet.Name == "四位分类" {
			// 遍历每一行
			for _, row := range sheet.Rows {
				param := make(map[string]string)

				param["r1"] = row.Cells[0].String()
				param["r2"] = row.Cells[1].String()
				param["r3"] = row.Cells[2].String()
				param["r4"] = row.Cells[3].String()
				// param["r5"] = row.Cells[4].String()

				fmt.Println(fmt.Sprintf(" insert into eli_swfl(renyun,guishen,shenjiang,difen,type) values ('%s','%s','%s','%s','%s');", param["r1"], param["r2"], param["r3"]))

				params = append(params, param)
			}

		}
	}

	jsonString, err := json.Marshal(params)
	if err != nil {
		fmt.Printf("JSON 编码失败: %s\n", err)
		return
	}

	// 打印 JSON 字符串
	fmt.Printf("JSON 字符串: %s\n", jsonString)
}

// func main() {
// 	// 打开Excel文件
// 	excelFileName := "/Users/wanggaowei/Downloads/4位关系.xlsx"
// 	xlFile, err := xlsx.OpenFile(excelFileName)
// 	if err != nil {
// 		log.Fatalf("无法打开Excel文件: %s\n", err)
// 	}

// 	var params []map[string]string

// 	// 遍历所有工作表
// 	for _, sheet := range xlFile.Sheets {
// 		// fmt.Printf("工作表名称: %s\n", sheet.Name)

// 		if sheet.Name == "四位象意" {
// 			// 遍历每一行
// 			for _, row := range sheet.Rows {
// 				param := make(map[string]string)

// 				param["r1"] = row.Cells[0].String()
// 				param["r2"] = row.Cells[1].String()
// 				param["r3"] = row.Cells[2].String()
// 				param["r4"] = row.Cells[3].String()
// 				param["r5"] = row.Cells[4].String()

// 				fmt.Println(fmt.Sprintf("insert into eli_swxy (r1,r2,relationship,des,`type`) values('%s','%s','%s','%s','%s');", param["r1"], param["r3"], param["r2"], param["r5"], param["r4"]))

// 				params = append(params, param)
// 			}

// 		}
// 	}

// 	jsonString, err := json.Marshal(params)
// 	if err != nil {
// 		fmt.Printf("JSON 编码失败: %s\n", err)
// 		return
// 	}

// 	// 打印 JSON 字符串
// 	fmt.Printf("JSON 字符串: %s\n", jsonString)
// }
