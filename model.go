package gowk

import (
	"gorm.io/gorm"
)

type Model struct {
	ID      uint           `gorm:"primarykey" json:"id"`
	Created *Time          `json:"created"`
	Updated *Time          `json:"updated"`
	Deleted gorm.DeletedAt `gorm:"index" json:"-"`
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

type M = map[string]interface{}
type A = []interface{}

func Paginate[T any](page *PageModel[T]) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page.Current <= 0 {
			page.Current = 0
		}
		if page.Size <= 0 {
			page.Size = 10
		}
		page.Pages = page.Total / page.Size
		if page.Total%page.Size != 0 {
			page.Pages++
		}
		p := page.Current
		if page.Current > page.Pages {
			p = page.Pages
		}
		size := page.Size
		offset := int((p - 1) * size)
		return db.Offset(offset).Limit(int(size))
	}
}
