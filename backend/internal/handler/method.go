package handler

import (
	"shiny-collection/internal/repository"
	"shiny-collection/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type MethodHandler struct {
	repo *repository.MethodRepository
	log  *zap.Logger
}

func NewMethodHandler(repo *repository.MethodRepository, log *zap.Logger) *MethodHandler {
	return &MethodHandler{repo: repo, log: log}
}

func (h *MethodHandler) List(c *gin.Context) {
	methods, err := h.repo.ListAll()
	if err != nil {
		h.log.Error("failed to list methods", zap.Error(err))
		response.InternalError(c, "failed to list methods")
		return
	}
	response.Success(c, methods)
}
