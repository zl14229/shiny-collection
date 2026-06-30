package repository

import (
	"shiny-collection/internal/model"

	"gorm.io/gorm"
)

type MethodRepository struct {
	db *gorm.DB
}

func NewMethodRepository(db *gorm.DB) *MethodRepository {
	return &MethodRepository{db: db}
}

func (r *MethodRepository) ListAll() ([]model.Method, error) {
	var methods []model.Method
	err := r.db.Find(&methods).Error
	return methods, err
}
