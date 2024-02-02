package model

type CoinList struct {
	ID     int64  `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	Symbol string `gorm:"column:symbol" json:"symbol"`

	ChainId   uint64 `gorm:"column:chain_id" json:"chain_id"`
	ChainName string `gorm:"column:chain_name" json:"chain_name"`
}

func (*CoinList) TableName() string {
	return "web3_coin_list"
}
