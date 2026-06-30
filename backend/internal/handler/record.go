package handler

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"shiny-collection/internal/model"
	"shiny-collection/internal/repository"
	"shiny-collection/internal/service"
	"shiny-collection/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type RecordHandler struct {
	svc  *service.RecordService
	log  *zap.Logger
}

func NewRecordHandler(svc *service.RecordService, log *zap.Logger) *RecordHandler {
	return &RecordHandler{svc: svc, log: log}
}

type CreateRecordRequest struct {
	PokemonID       uint   `json:"pokemonId" binding:"required"`
	GameID          uint   `json:"gameId" binding:"required"`
	MethodID        uint   `json:"methodId" binding:"required"`
	Status          string `json:"status"`
	TotalEncounters int    `json:"totalEncounters"`
	StartDate       string `json:"startDate" binding:"required"`
	EndDate         string `json:"endDate"`
	ShinyAppearance *bool  `json:"shinyAppearance"`
	Nature          string `json:"nature"`
	Gender          string `json:"gender"`
	BallUsed        string `json:"ballUsed"`
	Level           int    `json:"level"`
	IsAlpha         bool   `json:"isAlpha"`
	IsMarked        bool   `json:"isMarked"`
	MarkName        string `json:"markName"`
	Notes           string `json:"notes"`
	TagIDs          []uint `json:"tagIds"`
}

type UpdateRecordRequest struct {
	PokemonID       uint   `json:"pokemonId"`
	GameID          uint   `json:"gameId"`
	MethodID        uint   `json:"methodId"`
	Status          string `json:"status"`
	TotalEncounters *int   `json:"totalEncounters"`
	StartDate       string `json:"startDate"`
	EndDate         string `json:"endDate"`
	ShinyAppearance *bool  `json:"shinyAppearance"`
	Nature          string `json:"nature"`
	Gender          string `json:"gender"`
	BallUsed        string `json:"ballUsed"`
	Level           int    `json:"level"`
	IsAlpha         bool   `json:"isAlpha"`
	IsMarked        bool   `json:"isMarked"`
	MarkName        string `json:"markName"`
	Notes           string `json:"notes"`
	TagIDs          []uint `json:"tagIds"`
}

// parseDate helper
func parseDate(s string) (time.Time, error) {
	if s == "" {
		return time.Now(), nil
	}
	formats := []string{
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t, nil
		}
	}
	return time.Now(), nil
}

func (h *RecordHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	filter := repository.RecordFilter{
		Status:   c.Query("status"),
		Keyword:  c.Query("keyword"),
	}

	if v := c.Query("gameId"); v != "" {
		id, _ := strconv.ParseUint(v, 10, 64)
		filter.GameID = uint(id)
	}
	if v := c.Query("methodId"); v != "" {
		id, _ := strconv.ParseUint(v, 10, 64)
		filter.MethodID = uint(id)
	}
	if v := c.Query("pokemonId"); v != "" {
		id, _ := strconv.ParseUint(v, 10, 64)
		filter.PokemonID = uint(id)
	}
	if v := c.Query("tagId"); v != "" {
		id, _ := strconv.ParseUint(v, 10, 64)
		filter.TagID = uint(id)
	}

	records, total, err := h.svc.List(filter, page, pageSize)
	if err != nil {
		h.log.Error("failed to list records", zap.Error(err))
		response.InternalError(c, "failed to list records")
		return
	}

	response.SuccessWithPage(c, records, total, page, pageSize)
}

func (h *RecordHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid record id")
		return
	}

	record, err := h.svc.GetByID(uint(id))
	if err != nil {
		h.log.Error("failed to get record", zap.Error(err))
		response.NotFound(c, "record not found")
		return
	}

	response.Success(c, record)
}

func (h *RecordHandler) Create(c *gin.Context) {
	var req CreateRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	startDate, err := parseDate(req.StartDate)
	if err != nil {
		response.BadRequest(c, "invalid start date")
		return
	}

	shiny := false
	if req.ShinyAppearance != nil {
		shiny = *req.ShinyAppearance
	}

	record := &model.Record{
		PokemonID:       req.PokemonID,
		GameID:          req.GameID,
		MethodID:        req.MethodID,
		Status:          model.RecordStatus(req.Status),
		TotalEncounters: req.TotalEncounters,
		StartDate:       startDate,
		ShinyAppearance: shiny,
		Nature:          req.Nature,
		Gender:          req.Gender,
		BallUsed:        req.BallUsed,
		Level:           req.Level,
		IsAlpha:         req.IsAlpha,
		IsMarked:        req.IsMarked,
		MarkName:        req.MarkName,
		Notes:           req.Notes,
	}

	// handle end date
	if req.EndDate != "" {
		endDate, err := parseDate(req.EndDate)
		if err == nil {
			record.EndDate = &endDate
		}
	}

	// set default status
	if record.Status == "" {
		record.Status = model.RecordStatusHunting
	}

	// handle tags
	if len(req.TagIDs) > 0 {
		for _, tagID := range req.TagIDs {
			record.Tags = append(record.Tags, model.Tag{ID: tagID})
		}
	}

	if err := h.svc.Create(record); err != nil {
		h.log.Error("failed to create record", zap.Error(err))
		response.BadRequest(c, err.Error())
		return
	}

	response.Created(c, record)
}

func (h *RecordHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid record id")
		return
	}

	var req UpdateRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	record, err := h.svc.GetByID(uint(id))
	if err != nil {
		response.NotFound(c, "record not found")
		return
	}

	if req.PokemonID > 0 {
		record.PokemonID = req.PokemonID
	}
	if req.GameID > 0 {
		record.GameID = req.GameID
	}
	if req.MethodID > 0 {
		record.MethodID = req.MethodID
	}
	if req.Status != "" {
		record.Status = model.RecordStatus(req.Status)
	}
	if req.TotalEncounters != nil {
		record.TotalEncounters = *req.TotalEncounters
	}
	if req.StartDate != "" {
		if d, err := parseDate(req.StartDate); err == nil {
			record.StartDate = d
		}
	}
	if req.EndDate != "" {
		if d, err := parseDate(req.EndDate); err == nil {
			record.EndDate = &d
		}
	}
	if req.ShinyAppearance != nil {
		record.ShinyAppearance = *req.ShinyAppearance
	}
	if req.Nature != "" {
		record.Nature = req.Nature
	}
	if req.Gender != "" {
		record.Gender = req.Gender
	}
	if req.BallUsed != "" {
		record.BallUsed = req.BallUsed
	}
	if req.Level > 0 {
		record.Level = req.Level
	}
	record.IsAlpha = req.IsAlpha
	record.IsMarked = req.IsMarked
	if req.MarkName != "" {
		record.MarkName = req.MarkName
	}
	if req.Notes != "" {
		record.Notes = req.Notes
	}

	// update tags
	if req.TagIDs != nil {
		record.Tags = nil
		for _, tagID := range req.TagIDs {
			record.Tags = append(record.Tags, model.Tag{ID: tagID})
		}
	}

	if err := h.svc.Update(record); err != nil {
		h.log.Error("failed to update record", zap.Error(err))
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, record)
}

func (h *RecordHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid record id")
		return
	}

	if err := h.svc.Delete(uint(id)); err != nil {
		h.log.Error("failed to delete record", zap.Error(err))
		response.InternalError(c, "failed to delete record")
		return
	}

	response.Success(c, nil)
}

// UploadVideo 上传出闪时刻视频
func (h *RecordHandler) UploadVideo(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid record id")
		return
	}

	record, err := h.svc.GetByID(uint(id))
	if err != nil {
		response.NotFound(c, "record not found")
		return
	}

	file, err := c.FormFile("video")
	if err != nil {
		response.BadRequest(c, "请选择要上传的视频文件")
		return
	}

	// 校验文件类型
	ext := filepath.Ext(file.Filename)
	allowedExt := map[string]bool{".mp4": true, ".webm": true, ".mov": true, ".avi": true, ".mkv": true}
	if !allowedExt[ext] {
		response.BadRequest(c, fmt.Sprintf("不支持的文件格式 %s，支持: mp4/webm/mov/avi/mkv", ext))
		return
	}

	// 校验文件大小（最大 500MB）
	const maxSize = 500 << 20
	if file.Size > maxSize {
		response.BadRequest(c, "文件大小超过 500MB 限制")
		return
	}

	// 确保上传目录存在
	uploadDir := filepath.Join("data", "videos")
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		h.log.Error("failed to create upload dir", zap.Error(err))
		response.InternalError(c, "failed to create upload directory")
		return
	}

	// 生成唯一文件名: record_id_时间戳.mp4
	filename := fmt.Sprintf("record_%d_%d%s", id, time.Now().Unix(), ext)
	savePath := filepath.Join(uploadDir, filename)

	if err := c.SaveUploadedFile(file, savePath); err != nil {
		h.log.Error("failed to save video", zap.Error(err))
		response.InternalError(c, "failed to save video")
		return
	}

	// 更新数据库记录的 shinyVideo 字段
	record.ShinyVideo = filename
	if err := h.svc.Update(record); err != nil {
		h.log.Error("failed to update record video field", zap.Error(err))
		response.InternalError(c, "failed to update record")
		return
	}

	response.Success(c, gin.H{
		"filename": filename,
		"url":      fmt.Sprintf("/uploads/videos/%s", filename),
	})
}

// DeleteVideo 删除出闪时刻视频
func (h *RecordHandler) DeleteVideo(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid record id")
		return
	}

	record, err := h.svc.GetByID(uint(id))
	if err != nil {
		response.NotFound(c, "record not found")
		return
	}

	if record.ShinyVideo == "" {
		response.Success(c, nil)
		return
	}

	// 删除文件
	videoPath := filepath.Join("data", "videos", record.ShinyVideo)
	if err := os.Remove(videoPath); err != nil && !os.IsNotExist(err) {
		h.log.Error("failed to delete video file", zap.Error(err))
	}

	// 清空数据库字段
	record.ShinyVideo = ""
	if err := h.svc.Update(record); err != nil {
		h.log.Error("failed to clear record video field", zap.Error(err))
		response.InternalError(c, "failed to update record")
		return
	}

	response.Success(c, nil)
}
