package selflogic

import (
	"eli/app/model"
	"eli/database"
	"fmt"
	"strings"
	"testing"
)

func TestSort(t *testing.T) {
	eles := []string{"亥", "丑", "辰", "子", "午"}

	ts := Transform{}

	sh := ts.OutputSeq(2, eles)

	fmt.Println(sh)
}

func TestAdjust(t *testing.T) {
	db := database.GetDb()
	var data []model.EliDzgx

	db.Model(model.EliDzgx{}).Find(&data)

	ts := Transform{}

	for i, _ := range data {
		d := data[i]
		d.Dz = strings.Join(ts.OutputSeq(2, strings.Split(d.Dz, "")), "")
		db.Updates(&d)
	}

	fmt.Println("success")
}

func TestComb(t *testing.T) {
	dizhi := []string{"亥", "丑", "辰", "子"}
	ts := Transform{}

	_1 := ts.UniqueCombination(2, dizhi, true)
	fmt.Println(_1)
}
