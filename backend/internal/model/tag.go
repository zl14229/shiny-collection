package model

import "time"

type Tag struct {
	ID        uint   `gorm:"primarykey" json:"id"`
	Name      string `gorm:"size:30;not null;uniqueIndex" json:"name"`
	Color     string `gorm:"size:7;default:#409EFF" json:"color"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (Tag) TableName() string {
	return "tags"
}
