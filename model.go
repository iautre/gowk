package gowk

import (
	"gorm.io/gorm"
)

type Model struct {
	ID      uint           `gorm:"primarykey" json:"id"`
	Created Time           `json:"created"`
	Updated Time           `json:"updated"`
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

type PageModel[T any] struct {
	Size  int64 `json:"size" form:"size"`
	Page  int64 `json:"page" form:"page"`
	Pages int64 `json:"pages"`
	Total int64 `json:"total"`
	List  []*T  `json:"list"`
}

type M = map[string]interface{}
type A = []interface{}

func Paginate[T any](page *PageModel[T]) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page.Page <= 0 {
			page.Page = 0
		}
		if page.Size <= 0 {
			page.Size = 10
		}
		page.Pages = page.Total / page.Size
		if page.Total%page.Size != 0 {
			page.Pages++
		}
		p := page.Page
		if page.Page > page.Pages {
			p = page.Pages
		}
		size := page.Page
		offset := int((p - 1) * size)
		return db.Offset(offset).Limit(int(size))
	}
}
