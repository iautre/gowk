package gowk

type PageParams struct {
	Size    int64 `json:"size" form:"size"`
	Current int64 `json:"current" form:"current"`
}

type PageModel[T any] struct {
	Size    int64 `json:"size" form:"size"`
	Current int64 `json:"current" form:"current"`
	Pages   int64 `json:"pages"`
	Total   int64 `json:"total"`
	Records []*T  `json:"records"`
}

// CalcPages 根据已设置的 Total 和 Size 计算总页数，应在 COUNT 查询之后调用。
func (p *PageModel[T]) CalcPages() {
	if p.Size <= 0 {
		p.Size = 10
	}
	p.Pages = p.Total / p.Size
	if p.Total%p.Size != 0 {
		p.Pages++
	}
}

type M = map[string]interface{}
type A = []interface{}
