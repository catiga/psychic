// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package model

const TableNameEliWxws = "eli_wxws"

// EliWxw mapped from table <eli_wxws>
type EliWxws struct {
	ID               int64  `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	Wuxing           string `gorm:"column:wuxing" json:"wuxing"`
	Type             string `gorm:"column:type" json:"type"`
	PersonalityTrait string `gorm:"column:personality_trait" json:"personality_trait"`
}

// TableName EliWxw's table name
func (*EliWxws) TableName() string {
	return TableNameEliWxws
}
