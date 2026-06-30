package handler

import (
	"strconv"

	"shiny-collection/internal/model"
	"shiny-collection/internal/repository"
	"shiny-collection/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type PokemonHandler struct {
	repo *repository.PokemonRepository
	log  *zap.Logger
}

func NewPokemonHandler(repo *repository.PokemonRepository, log *zap.Logger) *PokemonHandler {
	return &PokemonHandler{repo: repo, log: log}
}

func (h *PokemonHandler) List(c *gin.Context) {
	// support search
	if keyword := c.Query("keyword"); keyword != "" {
		limit := 20
		if l, err := strconv.Atoi(c.DefaultQuery("limit", "20")); err == nil && l > 0 && l <= 100 {
			limit = l
		}
		pokemon, err := h.repo.Search(keyword, limit)
		if err != nil {
			h.log.Error("failed to search pokemon", zap.Error(err))
			response.InternalError(c, "failed to search pokemon")
			return
		}
		response.Success(c, pokemon)
		return
	}

	pokemon, err := h.repo.ListAll()
	if err != nil {
		h.log.Error("failed to list pokemon", zap.Error(err))
		response.InternalError(c, "failed to list pokemon")
		return
	}
	response.Success(c, pokemon)
}

func (h *PokemonHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid pokemon id")
		return
	}

	p, err := h.repo.GetByID(uint(id))
	if err != nil {
		response.NotFound(c, "pokemon not found")
		return
	}
	response.Success(c, p)
}

func (h *PokemonHandler) Create(c *gin.Context) {
	var req struct {
		Name       string `json:"name" binding:"required"`
		NationalNo int    `json:"nationalNo" binding:"required"`
		Type1      string `json:"type1" binding:"required"`
		Type2      string `json:"type2"`
		ImageURL   string `json:"imageUrl"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	pokemon := &model.Pokemon{
		Name:       req.Name,
		NationalNo: req.NationalNo,
		Type1:      req.Type1,
		Type2:      req.Type2,
		ImageURL:   req.ImageURL,
	}

	if err := h.repo.Create(pokemon); err != nil {
		h.log.Error("failed to create pokemon", zap.Error(err))
		response.InternalError(c, "failed to create pokemon")
		return
	}

	response.Created(c, pokemon)
}
