// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package model
import (
    "github.com/shopspring/decimal"
)

const TableNameDisAirdrop = "dis_airdrop"

// DisAirdrop mapped from table <dis_airdrop>
type DisAirdrop struct {
	ID      int64   `gorm:"column:id;primaryKey" json:"id"`
	Address string  `gorm:"column:address" json:"address"`
	Status  int32   `gorm:"column:status" json:"status"`
	Amount  decimal.Decimal `gorm:"column:amount" json:"amount"`
}

// TableName DisAirdrop's table name
func (*DisAirdrop) TableName() string {
	return TableNameDisAirdrop
}
