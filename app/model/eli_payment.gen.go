// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package model

import (
	"time"
)

const TableNameEliPayment = "eli_payment"

// EliPayment mapped from table <eli_payment>
type EliPayment struct {
	ID                int64     `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	OrderID           int64     `gorm:"column:order_id;not null" json:"order_id"`
	PaymentType       int64     `gorm:"column:payment_type;not null" json:"payment_type"`
	PaymentIdentifier string    `gorm:"column:payment_identifier;not null;comment:微信支付时为openid，支付宝支付时为buyer_id" json:"payment_identifier"` // 微信支付时为openid，支付宝支付时为buyer_id
	PaymentAmount     float64   `gorm:"column:payment_amount;not null" json:"payment_amount"`
	PaymentStatus     string    `gorm:"column:payment_status;not null" json:"payment_status"`
	CallbackInfo      string    `gorm:"column:callback_info;comment:支付信息表，存储支付详情，兼容微信支付和支付宝支付" json:"callback_info"` // 支付信息表，存储支付详情，兼容微信支付和支付宝支付
	RequestInfo       string    `gorm:"column:request_info;comment:请求支付信息" json:"request_info"`                      // 请求支付信息
	CreateAt          time.Time `gorm:"column:create_at" json:"create_at"`
	CallbackAt        time.Time `gorm:"column:callback_at" json:"callback_at"`
	UserID            int64     `gorm:"column:user_id" json:"user_id"`
}

// TableName EliPayment's table name
func (*EliPayment) TableName() string {
	return TableNameEliPayment
}