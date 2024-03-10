// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package model

import (
	"time"
)

const TableNameEliOrder = "eli_order"

// EliOrder mapped from table <eli_order>
type EliOrder struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	OrderNo   string    `gorm:"column:order_no" json:"order_no"`
	OrderType int64     `gorm:"column:order_type;comment:1.算命订单" json:"order_type"` // 1.算命订单
	QType     string    `gorm:"column:q_type" json:"q_type"`
	QDetail   string    `gorm:"column:q_detail" json:"q_detail"`
	QNum      int64     `gorm:"column:q_num" json:"q_num"`
	CreateAt  time.Time `gorm:"column:create_at" json:"create_at"`
	UserID    int64     `gorm:"column:user_id" json:"user_id"`
	Ref       string    `gorm:"column:ref;comment:订单来源" json:"ref"` // 订单来源
	Sizhu     string    `gorm:"column:sizhu" json:"sizhu"`
}

// TableName EliOrder's table name
func (*EliOrder) TableName() string {
	return TableNameEliOrder
}
