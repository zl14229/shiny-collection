package model

import "time"

type Pokemon struct {
	ID         uint   `gorm:"primarykey" json:"id"`
	Name       string `gorm:"size:50;not null;uniqueIndex" json:"name"`
	NameCN     string `gorm:"size:50;not null" json:"nameCN"`
	NationalNo int    `gorm:"uniqueIndex;not null" json:"nationalNo"`
	Type1      string `gorm:"size:20;not null" json:"type1"`
	Type2      string `gorm:"size:20" json:"type2"`
	ImageURL   string `gorm:"size:255" json:"imageUrl"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

func (Pokemon) TableName() string {
	return "pokemon"
}
