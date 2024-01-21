package config

import (
	"fmt"
	"strings"
	"testing"
)

const s = `
财运 
运势 
情感 
事业 
生育 
风水 
其他
出行 
灾祸 
官司 
学业 
健康 
失物 
射覆 
健康 
健康
情感 
金融 
运势
事业
风水 
情感
运势 
事业
财运 
财运
出行 
讨债
`

func TestSort(t *testing.T) {
	ss := strings.Split(s, "\n")
	var mm []string
	for _, v := range ss {
		if len(strings.Trim(v, " ")) > 0 {
			add := true
			for _, m := range mm {
				if m == v {
					add = false
					break
				}
			}
			if add {
				mm = append(mm, v)
			}
		}
	}
	fmt.Println(mm)
}
