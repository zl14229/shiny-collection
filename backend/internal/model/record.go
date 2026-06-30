package model

import "time"

// RecordStatus 狩猎记录状态
type RecordStatus string

const (
	RecordStatusHunting   RecordStatus = "hunting"
	RecordStatusObtained  RecordStatus = "obtained"
	RecordStatusAbandoned RecordStatus = "abandoned"
)

type Record struct {
	ID              uint         `gorm:"primarykey" json:"id"`
	PokemonID       uint         `gorm:"not null;index" json:"pokemonId"`
	GameID          uint         `gorm:"not null;index" json:"gameId"`
	MethodID        uint         `gorm:"not null;index" json:"methodId"`
	Status          RecordStatus `gorm:"size:20;default:hunting;index" json:"status"`
	TotalEncounters int          `gorm:"default:0" json:"totalEncounters"`
	StartDate       time.Time    `json:"startDate"`
	EndDate         *time.Time   `json:"endDate,omitempty"`
	ShinyAppearance bool         `gorm:"default:false" json:"shinyAppearance"`
	Nature          string       `gorm:"size:20" json:"nature"`
	Gender          string       `gorm:"size:10" json:"gender"`
	BallUsed        string       `gorm:"size:30" json:"ballUsed"`
	Level           int          `gorm:"default:1" json:"level"`
	IsAlpha         bool         `gorm:"default:false" json:"isAlpha"`       // 阿尔宙斯中的头目/大王
	IsMarked        bool         `gorm:"default:false" json:"isMarked"`      // 是否有证章
	MarkName        string       `gorm:"size:50" json:"markName"`            // 证章名
	Notes           string       `gorm:"type:text" json:"notes"`
	ShinyVideo      string       `gorm:"size:255" json:"shinyVideo"`         // 出闪时刻视频
	Tags            []Tag        `gorm:"many2many:record_tags;constraint:OnDelete:CASCADE;" json:"tags,omitempty"`

	// 关联预加载
	Pokemon *Pokemon `gorm:"foreignKey:PokemonID" json:"pokemon,omitempty"`
	Game    *Game    `gorm:"foreignKey:GameID" json:"game,omitempty"`
	Method  *Method  `gorm:"foreignKey:MethodID" json:"method,omitempty"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (Record) TableName() string {
	return "records"
}
