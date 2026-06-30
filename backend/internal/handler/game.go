package handler

import (
	"strconv"

	"shiny-collection/internal/repository"
	"shiny-collection/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type GameHandler struct {
	repo *repository.GameRepository
	log  *zap.Logger
}

func NewGameHandler(repo *repository.GameRepository, log *zap.Logger) *GameHandler {
	return &GameHandler{repo: repo, log: log}
}

func (h *GameHandler) List(c *gin.Context) {
	games, err := h.repo.ListAll()
	if err != nil {
		h.log.Error("failed to list games", zap.Error(err))
		response.InternalError(c, "failed to list games")
		return
	}
	response.Success(c, games)
}

func (h *GameHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid game id")
		return
	}

	game, err := h.repo.GetByID(uint(id))
	if err != nil {
		response.NotFound(c, "game not found")
		return
	}
	response.Success(c, game)
}
