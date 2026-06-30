package model

import "time"

type Game struct {
	ID         uint   `gorm:"primarykey" json:"id"`
	Name       string `gorm:"size:100;not null;uniqueIndex" json:"name"`
	NameCN     string `gorm:"size:100;not null" json:"nameCN"`
	Generation int    `gorm:"not null" json:"generation"`
	Platform   string `gorm:"size:50;not null" json:"platform"`
	ShortName  string `gorm:"size:20;not null;index" json:"shortName"`
	ReleaseYear int   `json:"releaseYear"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

func (Game) TableName() string {
	return "games"
}
