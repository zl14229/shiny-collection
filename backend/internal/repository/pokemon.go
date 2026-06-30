package repository

import (
	"shiny-collection/internal/model"

	"gorm.io/gorm"
)

type PokemonRepository struct {
	db *gorm.DB
}

func NewPokemonRepository(db *gorm.DB) *PokemonRepository {
	return &PokemonRepository{db: db}
}

func (r *PokemonRepository) ListAll() ([]model.Pokemon, error) {
	var pokemon []model.Pokemon
	err := r.db.Order("national_no ASC").Find(&pokemon).Error
	return pokemon, err
}

func (r *PokemonRepository) Search(keyword string, limit int) ([]model.Pokemon, error) {
	var pokemon []model.Pokemon
	like := "%" + keyword + "%"
	err := r.db.Where("name LIKE ? OR name_cn LIKE ? OR national_no LIKE ?", like, like, like).
		Order("national_no ASC").
		Limit(limit).
		Find(&pokemon).Error
	return pokemon, err
}

func (r *PokemonRepository) GetByID(id uint) (*model.Pokemon, error) {
	var p model.Pokemon
	err := r.db.First(&p, id).Error
	return &p, err
}

func (r *PokemonRepository) Create(pokemon *model.Pokemon) error {
	return r.db.Create(pokemon).Error
}

func (r *PokemonRepository) GetByNationalNo(no int) (*model.Pokemon, error) {
	var p model.Pokemon
	err := r.db.Where("national_no = ?", no).First(&p).Error
	return &p, err
}
