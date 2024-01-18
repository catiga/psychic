package util

import (
	"math"
	"strconv"
)

// Page 结构体表示分页对象
type Page struct {
	PageSize  int64       `json:"pageSize"`  // 每页记录数
	PageNo    int64       `json:"pageNo"`    // 当前页码
	Total     int64       `json:"total"`     // 总记录数
	TotalPage int64       `json:"totalPage"` // 总页数
	Result    interface{} `json:"result"`    // 查询结果
	Version   string      `json:"version"`   // 版本信息（如果需要）
}

// NewPage 初始化分页对象
func NewPage(pageNo, pageSize int64) Page {
	if pageNo < 1 {
		pageNo = 1
	}

	if pageSize <= 0 {
		pageSize = 10
	}

	if pageSize > 150 {
		pageSize = 150
	}

	return Page{
		PageSize: pageSize,
		PageNo:   pageNo,
	}
}

func NewPageWithStr(pageNoStr, pageSizeStr string) Page {
	pageNo, err := strconv.ParseInt(pageNoStr, 10, 64)
	if err != nil || pageNo < 1 {
		pageNo = 1
	}

	pageSize, err := strconv.ParseInt(pageSizeStr, 10, 64)
	if err != nil || pageSize <= 0 {
		pageSize = 10
	}

	if pageSize > 150 {
		pageSize = 150
	}

	return Page{
		PageSize: pageSize,
		PageNo:   pageNo,
	}
}

// SetTotal 设置总记录数并计算总页数
func (p Page) SetTotal(total int64) Page {
	if p.PageSize == 0 {
		p.PageSize = 10
	}

	p.Total = total
	p.TotalPage = int64(math.Ceil(float64(p.Total) / float64(p.PageSize)))

	return p
}

// Offset 计算偏移量
func (p Page) Offset() int {
	offset := (p.PageNo - 1) * p.PageSize
	if offset < 0 {
		return 0
	}
	return int(offset)
}
