package model

import (
	"errors"
	"reflect"
)

type PageDTO struct {
	// 当前页数
	PageIndex int64 `json:"page_index"`
	// 每页数据条数
	PageSize int64 `json:"page_size"`
}

type SortDTO struct {
	// 排序规则，k：字段名，v：排序规则，1位升序，-1位降序
	Sort map[string]int `json:"sort" bingding:"oneof=-1 1"`
}

// PageInfo 分页信息
type PageInfo struct {
	PageDTO
	// 数据总量
	Total int64 `json:"total"`
	// 当前页的数据
	Items interface{} `json:"items"`
	// 总页数
	PageTotal int64 `json:"page_total"`
	// 是否存在上一页
	HasPrePage bool `json:"has_pre_page"`
	// 是否存在下一页
	HasNextPage bool `json:"has_next_page"`
}

// Build 构建分页信息
func (pageInfo *PageInfo) Build(pageDTO PageDTO, items interface{}, itemTotal int64) (*PageInfo, error) {
	if reflect.ValueOf(items).Kind() != reflect.Slice {
		return nil, errors.New("PageInfo items must be a slice")
	}
	if pageDTO.PageIndex == 0 {
		pageDTO.PageIndex = 1
	}
	if pageDTO.PageSize == 0 {
		pageDTO.PageSize = 10
	}
	pageTotal := itemTotal / pageDTO.PageSize
	if itemTotal%pageDTO.PageSize > 0 {
		pageTotal += 1
	}
	pageInfo.PageDTO = pageDTO
	pageInfo.Total = itemTotal
	pageInfo.Items = items
	pageInfo.PageTotal = pageTotal
	pageInfo.HasPrePage = pageDTO.PageIndex > 1 && pageDTO.PageIndex <= pageTotal
	pageInfo.HasNextPage = pageDTO.PageIndex >= 1 && pageDTO.PageIndex < pageTotal
	return pageInfo, nil
}
