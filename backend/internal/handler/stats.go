package handler

import (
	"shiny-collection/internal/service"
	"shiny-collection/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type StatsHandler struct {
	svc *service.RecordService
	log *zap.Logger
}

func NewStatsHandler(svc *service.RecordService, log *zap.Logger) *StatsHandler {
	return &StatsHandler{svc: svc, log: log}
}

func (h *StatsHandler) Overview(c *gin.Context) {
	stats, err := h.svc.GetStatsOverview()
	if err != nil {
		h.log.Error("failed to get stats overview", zap.Error(err))
		response.InternalError(c, "failed to get stats")
		return
	}
	response.Success(c, stats)
}

func (h *StatsHandler) ByGame(c *gin.Context) {
	stats, err := h.svc.GetStatsByGame()
	if err != nil {
		h.log.Error("failed to get stats by game", zap.Error(err))
		response.InternalError(c, "failed to get stats")
		return
	}
	response.Success(c, stats)
}
