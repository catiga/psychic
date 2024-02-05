package model

type SysCatalog struct {
	ID     int64  `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	NameCn string `gorm:"column:name_cn" json:"name_cn"`

	NameEn uint64 `gorm:"column:name_en" json:"name_en"`
}

func (*SysCatalog) TableName() string {
	return "sys_catalog"
}
