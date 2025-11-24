package handlers

import (
	"net/http"

	"SpaceBookProject/internal/domain"
	"SpaceBookProject/internal/services"

	"github.com/gin-gonic/gin"
)

type SpaceHandler struct {
	svc *services.SpaceService
}

func NewSpaceHandler(svc *services.SpaceService) *SpaceHandler {
	return &SpaceHandler{svc: svc}
}

func (h *SpaceHandler) ListSpaces(c *gin.Context) {
	spaces, err := h.svc.ListSpaces()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load spaces"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": spaces})
}

func (h *SpaceHandler) CreateSpace(c *gin.Context) {
	var req domain.CreateSpaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	rawID, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	ownerID, ok := rawID.(int)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id in context"})
		return
	}

	space, err := h.svc.CreateSpace(ownerID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create space"})
		return
	}

	c.JSON(http.StatusCreated, space)
}
