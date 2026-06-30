package model

import "time"

type Pokemon struct {
	ID         uint   `gorm:"primarykey" json:"id"`
	Name       string `gorm:"size:50;not null;index:idx_name_form" json:"name"`
	NameCN     string `gorm:"size:50;not null" json:"nameCN"`
	NationalNo int    `gorm:"not null;index" json:"nationalNo"`
	Type1      string `gorm:"size:20;not null" json:"type1"`
	Type2      string `gorm:"size:20" json:"type2"`
	Form       string `gorm:"size:30;index:idx_name_form" json:"form"`
	ImageURL   string `gorm:"size:255" json:"imageUrl"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

func (Pokemon) TableName() string {
	return "pokemon"
}
