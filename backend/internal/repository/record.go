package repository

import (
	"shiny-collection/internal/model"

	"gorm.io/gorm"
)

type RecordFilter struct {
	Status   string
	GameID   uint
	MethodID uint
	PokemonID uint
	Keyword  string
	TagID    uint
}

type RecordRepository struct {
	db *gorm.DB
}

func NewRecordRepository(db *gorm.DB) *RecordRepository {
	return &RecordRepository{db: db}
}

func (r *RecordRepository) List(filter RecordFilter, page, pageSize int) ([]model.Record, int64, error) {
	var records []model.Record
	var total int64

	query := r.db.Model(&model.Record{}).
		Preload("Pokemon").
		Preload("Game").
		Preload("Method").
		Preload("Tags")

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.GameID > 0 {
		query = query.Where("game_id = ?", filter.GameID)
	}
	if filter.MethodID > 0 {
		query = query.Where("method_id = ?", filter.MethodID)
	}
	if filter.PokemonID > 0 {
		query = query.Where("pokemon_id = ?", filter.PokemonID)
	}
	if filter.TagID > 0 {
		query = query.Joins("JOIN record_tags ON record_tags.record_id = records.id").
			Where("record_tags.tag_id = ?", filter.TagID)
	}
	if filter.Keyword != "" {
		like := "%" + filter.Keyword + "%"
		query = query.Where(
			"notes LIKE ? OR EXISTS (SELECT 1 FROM pokemon WHERE pokemon.id = records.pokemon_id AND pokemon.name LIKE ?)",
			like, like,
		)
	}

	// count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// paginated query
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&records).Error; err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

func (r *RecordRepository) GetByID(id uint) (*model.Record, error) {
	var record model.Record
	err := r.db.Preload("Pokemon").
		Preload("Game").
		Preload("Method").
		Preload("Tags").
		First(&record, id).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *RecordRepository) Create(record *model.Record) error {
	return r.db.Create(record).Error
}

func (r *RecordRepository) Update(record *model.Record) error {
	return r.db.Session(&gorm.Session{FullSaveAssociations: false}).
		Model(record).
		Select("*").
		Omit("CreatedAt").
		Updates(record).Error
}

func (r *RecordRepository) Delete(id uint) error {
	return r.db.Delete(&model.Record{}, id).Error
}

// --- stats ---

type StatsOverview struct {
	TotalRecords     int64 `json:"totalRecords"`
	TotalShiny       int64 `json:"totalShiny"`
	HuntingRecords   int64 `json:"huntingRecords"`
	TotalEncounters  int64 `json:"totalEncounters"`
	MethodBreakdown  []MethodStat  `json:"methodBreakdown"`
	MonthlyTrend     []MonthlyStat `json:"monthlyTrend"`
}

type MethodStat struct {
	MethodID   uint   `json:"methodId"`
	MethodName string `json:"methodName"`
	Count      int64  `json:"count"`
}

type MonthlyStat struct {
	Year  int   `json:"year"`
	Month int   `json:"month"`
	Count int64 `json:"count"`
}

type GameStat struct {
	GameID   uint   `json:"gameId"`
	GameName string `json:"gameName"`
	Total    int64  `json:"total"`
	Shiny    int64  `json:"shiny"`
}

func (r *RecordRepository) GetStatsOverview() (*StatsOverview, error) {
	stats := &StatsOverview{}

	// total records
	r.db.Model(&model.Record{}).Count(&stats.TotalRecords)

	// total shiny obtained
	r.db.Model(&model.Record{}).Where("status = ? AND shiny_appearance = ?", model.RecordStatusObtained, true).Count(&stats.TotalShiny)

	// currently hunting
	r.db.Model(&model.Record{}).Where("status = ?", model.RecordStatusHunting).Count(&stats.HuntingRecords)

	// total encounters (all records)
	r.db.Model(&model.Record{}).Select("COALESCE(SUM(total_encounters), 0)").Scan(&stats.TotalEncounters)

	// method breakdown
	r.db.Model(&model.Record{}).
		Select("method_id, COUNT(*) as count").
		Group("method_id").
		Scan(&stats.MethodBreakdown)

	// fill method names
	for i, ms := range stats.MethodBreakdown {
		var m model.Method
		if err := r.db.First(&m, ms.MethodID).Error; err == nil {
			stats.MethodBreakdown[i].MethodName = m.NameCN
		}
	}

	// monthly trend (obtained records)
	r.db.Model(&model.Record{}).
		Where("status = ? AND shiny_appearance = ?", model.RecordStatusObtained, true).
		Select("strftime('%Y', end_date) as year, strftime('%m', end_date) as month, COUNT(*) as count").
		Group("strftime('%Y-%m', end_date)").
		Order("year DESC, month DESC").
		Limit(12).
		Scan(&stats.MonthlyTrend)

	return stats, nil
}

func (r *RecordRepository) GetStatsByGame() ([]GameStat, error) {
	var stats []GameStat
	r.db.Model(&model.Record{}).
		Select("game_id, COUNT(*) as total, SUM(CASE WHEN shiny_appearance = 1 AND status = 'obtained' THEN 1 ELSE 0 END) as shiny").
		Group("game_id").
		Scan(&stats)

	for i, s := range stats {
		var g model.Game
		if err := r.db.First(&g, s.GameID).Error; err == nil {
			stats[i].GameName = g.NameCN
		}
	}

	return stats, nil
}
