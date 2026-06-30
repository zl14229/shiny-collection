package service

import (
	"errors"

	"shiny-collection/internal/model"
	"shiny-collection/internal/repository"

	"gorm.io/gorm"
)

type RecordService struct {
	repo *repository.RecordRepository
}

func NewRecordService(repo *repository.RecordRepository) *RecordService {
	return &RecordService{repo: repo}
}

func (s *RecordService) List(filter repository.RecordFilter, page, pageSize int) ([]model.Record, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.List(filter, page, pageSize)
}

func (s *RecordService) GetByID(id uint) (*model.Record, error) {
	if id == 0 {
		return nil, errors.New("invalid record id")
	}
	return s.repo.GetByID(id)
}

func (s *RecordService) Create(record *model.Record) error {
	if record.PokemonID == 0 {
		return errors.New("pokemon is required")
	}
	if record.GameID == 0 {
		return errors.New("game is required")
	}
	if record.MethodID == 0 {
		return errors.New("hunting method is required")
	}
	return s.repo.Create(record)
}

func (s *RecordService) Update(record *model.Record) error {
	if record.ID == 0 {
		return errors.New("invalid record id")
	}
	// verify record exists
	existing, err := s.repo.GetByID(record.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("record not found")
		}
		return err
	}
	// preserve creation timestamp
	record.CreatedAt = existing.CreatedAt
	return s.repo.Update(record)
}

func (s *RecordService) Delete(id uint) error {
	if id == 0 {
		return errors.New("invalid record id")
	}
	return s.repo.Delete(id)
}

func (s *RecordService) GetStatsOverview() (*repository.StatsOverview, error) {
	return s.repo.GetStatsOverview()
}

func (s *RecordService) GetStatsByGame() ([]repository.GameStat, error) {
	return s.repo.GetStatsByGame()
}
