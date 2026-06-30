package model

import "time"

type Method struct {
	ID        uint   `gorm:"primarykey" json:"id"`
	Name      string `gorm:"size:50;not null;uniqueIndex" json:"name"`
	NameCN    string `gorm:"size:50;not null" json:"nameCN"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (Method) TableName() string {
	return "methods"
}
