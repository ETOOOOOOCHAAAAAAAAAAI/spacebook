package handlers

import (
	"net/http"
	"strconv"

	"SpaceBookProject/internal/domain"
	"SpaceBookProject/internal/repository"
	"SpaceBookProject/internal/services"

	"github.com/gin-gonic/gin"
)

type BookingHandler struct {
	svc *services.BookingService
}

func NewBookingHandler(svc *services.BookingService) *BookingHandler {
	return &BookingHandler{svc: svc}
}

func (h *BookingHandler) CreateBooking(c *gin.Context) {
	var req domain.CreateBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	tenantIDRaw, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}
	tenantID := tenantIDRaw.(int)

	booking, err := h.svc.CreateBooking(tenantID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, booking)
}

func (h *BookingHandler) ListMyBookings(c *gin.Context) {
	tenantIDRaw, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}
	tenantID := tenantIDRaw.(int)

	bookings, err := h.svc.ListMyBookings(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load bookings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": bookings})
}

func (h *BookingHandler) CancelBooking(c *gin.Context) {
	idStr := c.Param("id")
	bookingID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid booking id"})
		return
	}

	tenantIDRaw, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}
	tenantID := tenantIDRaw.(int)

	err = h.svc.CancelBooking(tenantID, bookingID)
	if err != nil {
		switch err {
		case services.ErrForbiddenAction:
			c.JSON(http.StatusForbidden, gin.H{"error": "you can cancel only your own bookings"})
		case services.ErrInvalidBookingStatus:
			c.JSON(http.StatusBadRequest, gin.H{"error": "booking cannot be cancelled in this status"})
		case repository.ErrBookingNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to cancel booking"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "booking cancelled"})
}

func (h *BookingHandler) ListOwnerBookings(c *gin.Context) {
	ownerIDRaw, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}
	ownerID := ownerIDRaw.(int)

	bookings, err := h.svc.ListOwnerBookings(ownerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load bookings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": bookings})
}

func (h *BookingHandler) ApproveBooking(c *gin.Context) {
	idStr := c.Param("id")
	bookingID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid booking id"})
		return
	}

	ownerIDRaw, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}
	ownerID := ownerIDRaw.(int)

	err = h.svc.ApproveBooking(ownerID, bookingID)
	if err != nil {
		switch err {
		case services.ErrForbiddenAction:
			c.JSON(http.StatusForbidden, gin.H{"error": "you can manage only bookings for your spaces"})
		case services.ErrInvalidBookingStatus:
			c.JSON(http.StatusBadRequest, gin.H{"error": "only pending bookings can be approved"})
		case repository.ErrBookingNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "booking approved"})
}

func (h *BookingHandler) RejectBooking(c *gin.Context) {
	idStr := c.Param("id")
	bookingID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid booking id"})
		return
	}

	ownerIDRaw, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}
	ownerID := ownerIDRaw.(int)

	err = h.svc.RejectBooking(ownerID, bookingID)
	if err != nil {
		switch err {
		case services.ErrForbiddenAction:
			c.JSON(http.StatusForbidden, gin.H{"error": "you can manage only bookings for your spaces"})
		case services.ErrInvalidBookingStatus:
			c.JSON(http.StatusBadRequest, gin.H{"error": "only pending bookings can be rejected"})
		case repository.ErrBookingNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "booking rejected"})
}
