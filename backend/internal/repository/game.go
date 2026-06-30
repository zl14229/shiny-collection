package repository

import (
	"shiny-collection/internal/model"

	"gorm.io/gorm"
)

type GameRepository struct {
	db *gorm.DB
}

func NewGameRepository(db *gorm.DB) *GameRepository {
	return &GameRepository{db: db}
}

func (r *GameRepository) ListAll() ([]model.Game, error) {
	var games []model.Game
	err := r.db.Order("generation ASC, release_year ASC").Find(&games).Error
	return games, err
}

func (r *GameRepository) GetByID(id uint) (*model.Game, error) {
	var game model.Game
	err := r.db.First(&game, id).Error
	return &game, err
}
