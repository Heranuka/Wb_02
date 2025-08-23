package rest

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"test_18/internal/domain"
	"time"

	"github.com/gin-gonic/gin"
)

type EventHandler interface {
	CreateEvent(ctx context.Context, event domain.Event) (int, error)
	UpdateEvent(ctx context.Context, id int, req domain.UpdateEventRequest) error
	DeleteEvent(ctx context.Context, id int) error
	GetEventsForTime(ctx context.Context, startDate time.Time, endDate time.Time) ([]domain.Event, error)
}

type Handler struct {
	logger       *slog.Logger
	eventHandler EventHandler
}

func NewHandler(logger *slog.Logger, eventHandler EventHandler) *Handler {
	return &Handler{
		logger:       logger,
		eventHandler: eventHandler,
	}
}
func (h *Handler) CreateEventHandler(c *gin.Context) {
	var req domain.Request

	if err := c.Bind(&req); err != nil {
		h.logger.Error("Failed to bind in CreateEvent", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var event = domain.Event{
		Title:       req.Title,
		Description: req.Description,
		StartTime:   time.Now(),
		EndTime:     time.Now(),
		Location:    req.Location,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	id, err := h.eventHandler.CreateEvent(c.Request.Context(), event)
	if err != nil {
		h.logger.Error("Failed to CreateEvent", slog.String("error", err.Error()), slog.Int("id", id))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"response": id})
}

func (h *Handler) UpdateEventHandler(c *gin.Context) {
	var req domain.UpdateEventRequest

	idstr := c.Param("id")
	id, err := strconv.Atoi(idstr)
	if err != nil {
		h.logger.Error("invalid id", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := c.Bind(&req); err != nil {
		h.logger.Error("Failed to bind UpdateEventRequest", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.eventHandler.UpdateEvent(c.Request.Context(), id, req); err != nil {
		h.logger.Error("Failed to UpdateEvent", slog.String("error", err.Error()), slog.Int("id", id))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"successfully updated": id})
}

func (h *Handler) DeleteEventHandler(c *gin.Context) {
	idstr := c.Param("id")
	id, err := strconv.Atoi(idstr)
	if err != nil {
		h.logger.Error("invalid id", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.eventHandler.DeleteEvent(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to DeleteEvent", slog.String("error", err.Error()), slog.Int("id", id))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"successfully deleted": id})
}

func (h *Handler) GetEventsForDayHandler(c *gin.Context) {
	// Получаем текущую дату и время в UTC
	now := time.Now().UTC()

	// Обрезаем время, чтобы получить только дату
	date := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	// Вычисляем дату следующего дня
	nextDay := date.AddDate(0, 0, 1)

	events, err := h.eventHandler.GetEventsForTime(c.Request.Context(), date, nextDay)
	h.logger.Error(" GetEventsForDay", slog.Any("events", events))

	if err != nil {
		h.logger.Error("Failed to GetEventsForDay", slog.String("error", err.Error()))
		statusCode := http.StatusInternalServerError // Default status
		if err.Error() == "not found" {              // Adjust based on your error types
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"events for day": events})
}

func (h *Handler) GetEventsForWeekHandler(c *gin.Context) {
	now := time.Now()

	// Обрезаем время, чтобы получить только дату
	date := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	nextDay := date.AddDate(0, 0, 7)
	events, err := h.eventHandler.GetEventsForTime(c.Request.Context(), date, nextDay)
	if err != nil {
		h.logger.Error("Failed toGetEventsForWeek", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"events for week": events})
}

func (h *Handler) GetEventsForMonthHandler(c *gin.Context) {
	now := time.Now()

	// Обрезаем время, чтобы получить только дату
	date := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	nextDay := date.AddDate(0, 0, 31)
	events, err := h.eventHandler.GetEventsForTime(c.Request.Context(), date, nextDay)
	if err != nil {
		h.logger.Error("Failed toGetEventsForMonth", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"events for month": events})
}
