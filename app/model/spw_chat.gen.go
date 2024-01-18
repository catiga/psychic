package model

import (
	"time"
)

const TableNameSpwChat = "spw_chat"

// SpwChat mapped from table <spw_chat>
type SpwChat struct {
	ID       int64      `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	DevID    string     `gorm:"column:dev_id" json:"dev_id"`
	UserID   int64      `gorm:"column:user_id;not null" json:"user_id"`
	CharID   int64      `gorm:"column:char_id;not null" json:"char_id"`
	Question string     `gorm:"column:question" json:"question"`
	AddTime  *time.Time `gorm:"column:add_time" json:"add_time"`
	Flag     int32      `gorm:"column:flag;not null" json:"flag"`
	CharCode string     `gorm:"column:char_code" json:"char_code"`
	Reply    string     `gorm:"column:reply" json:"reply"`
	CalId    string     `gorm:"column:cal_id" json:"calId"`
}

// TableName SpwChat's table name
func (*SpwChat) TableName() string {
	return TableNameSpwChat
}
