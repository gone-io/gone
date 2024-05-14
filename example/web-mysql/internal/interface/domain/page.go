package domain

// Page 分页
type Page[T any] struct {
	List  []*T  `json:"list,omitempty"`
	Total int64 `json:"total,omitempty"`
}

type PageQuery struct {
	Page     int `form:"page"`
	PageSize int `form:"pageSize"`
}
