package model

import (
	"time"
)

const TableNameEliCalInfo = "eli_cal_info"

// EliCalInfo mapped from table <eli_cal_info>
type EliCalInfo struct {
	ID       int64     `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	Param    string    `gorm:"column:param" json:"param"`
	Result   string    `gorm:"column:result" json:"result"`
	Type     int32     `gorm:"column:type;comment:1.生克" json:"type"` // 1.生克
	UserID   int64     `gorm:"column:user_id" json:"user_id"`
	CreateAt time.Time `gorm:"column:create_at" json:"create_at"`
}

// TableName EliCalInfo's table name
func (*EliCalInfo) TableName() string {
	return TableNameEliCalInfo
}
