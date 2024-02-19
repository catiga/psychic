package main

import (
	"gorm.io/driver/mysql"
	"gorm.io/gen"
	"gorm.io/gorm"
)

// Dynamic SQL
type Querier interface {
	// SELECT * FROM @@table WHERE name = @name{{if role !=""}} AND role = @role{{end}}
	FilterWithNameAndRole(name, role string) ([]gen.T, error)
}

func main() {
	g := gen.NewGenerator(gen.Config{
		OutPath: "../model",
		Mode:    gen.WithoutContext | gen.WithDefaultQuery | gen.WithQueryInterface, // generate mode
	})

	gormdb, _ := gorm.Open(mysql.Open("root:SDFD12312dddasdas1#@123123@(123.249.126.212:3306)/spirit_war?charset=utf8mb4&parseTime=True&loc=Local"))
	g.UseDB(gormdb)

	// Generate basic type-safe DAO API for struct `model.User` following conventions
	// g.GenerateModel("account_user_info")
	// g.GenerateModel("property_info")
	// g.GenerateModel("property_chain_info")
	// g.GenerateModel("property_info_media")
	// g.GenerateModel("eli_swxy")
	// g.GenerateModel("eli_dzgx")
	// g.GenerateModel("spw_character")
	// g.GenerateModel("spw_method")
	// g.GenerateModel("spw_char_background")
	// g.GenerateModel("spw_chat")
	// g.GenerateModel("spw_sample_chat")
	// g.GenerateModel("eli_cal_info")
	// g.GenerateModel("eli_swwx")
	// g.GenerateModel("eli_wxws")
	// g.GenerateModel("eli_swfl")
	// g.GenerateModel("dis_airdrop")
	// g.GenerateModel("sys_catalog")
	// g.GenerateModel("account_user_info")
	// g.ApplyBasic(g.GenerateModel("account_user_info"))

	// Generate Type Safe API with Dynamic SQL defined on Querier interface for `model.User` and `model.Company`
	// g.ApplyInterface(func(Querier) {}, model.User{}, model.Company{})

	// Generate the code
	g.Execute()
}
