package model

type AdminTplRes struct {
}
type AdminApiRes struct {
	Count int64 `p:"count"`
}

type ApiPageRes struct {
	Total int `json:"total"`
}

type ApiPageReq struct {
	Search string `p:"search"`
	Sort   string `p:"sort"`
	Order  string `p:"order"`
	Offset int    `p:"offset"`
	Limit  int    `p:"limit"`
	Filter string `p:"filter"`
	Op     string `p:"op"`
}

type ApiDelReq struct {
	Ids string `v:"required#ID参数必填"`
}
