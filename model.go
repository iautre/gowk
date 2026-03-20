package gowk

import (
	"gorm.io/gorm"
)

type Model struct {
	ID      uint64 `gorm:"primarykey" json:"id"`
	Created *Time  `json:"created" gorm:"autoCreateTime"`
	Updated *Time  `json:"updated" gorm:"autoUpdateTime"`
	Deleted *Time  `json:"-"`
}

func (m *Model) BeforeCreate(tx *gorm.DB) (err error) {
	m.Created = Now()
	m.Updated = Now()
	return
}

func (m *Model) BeforeUpdate(tx *gorm.DB) (err error) {
	m.Updated = Now()
	return
}

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

// Paginate 是 GORM scope，返回对应页的 Offset/Limit。
// 调用前须先执行 COUNT 查询并设置 page.Total，再调用 page.CalcPages()。
func Paginate[T any](page *PageModel[T]) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page.Current <= 0 {
			page.Current = 1
		}
		if page.Size <= 0 {
			page.Size = 10
		}
		p := page.Current
		if page.Pages > 0 && page.Current > page.Pages {
			p = page.Pages
		}
		offset := int((p - 1) * page.Size)
		return db.Offset(offset).Limit(int(page.Size))
	}
}
