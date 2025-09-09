package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Событие в календаре
type Event struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	Date        time.Time `json:"date"`
	Description string    `json:"event"`
}

var (
	ErrEventNotFound = errors.New("event not found")
)

// Интерфейс бизнес-логики
type CalendarService interface {
	CreateEvent(e Event) (int, error)
	UpdateEvent(id int, e Event) error
	DeleteEvent(id int) error
	GetEventsForDay(userID int, day time.Time) ([]Event, error)
	GetEventsForWeek(userID int, day time.Time) ([]Event, error)
	GetEventsForMonth(userID int, day time.Time) ([]Event, error)
}

// Реализация CalendarService с in-memory storage
type MemoryCalendar struct {
	mu     sync.RWMutex
	events map[int]Event
	nextID int
}

func NewMemoryCalendar() *MemoryCalendar {
	return &MemoryCalendar{
		events: make(map[int]Event),
		nextID: 1,
	}
}

func (m *MemoryCalendar) CreateEvent(e Event) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	e.ID = m.nextID
	m.nextID++
	m.events[e.ID] = e
	return e.ID, nil
}

func (m *MemoryCalendar) UpdateEvent(id int, e Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, exists := m.events[id]
	if !exists {
		return ErrEventNotFound
	}
	e.ID = id
	m.events[id] = e
	return nil
}

func (m *MemoryCalendar) DeleteEvent(id int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, exists := m.events[id]
	if !exists {
		return ErrEventNotFound
	}
	delete(m.events, id)
	return nil
}

func (m *MemoryCalendar) GetEventsInRange(userID int, start, end time.Time) []Event {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []Event
	for _, e := range m.events {
		if e.UserID == userID && !e.Date.Before(start) && e.Date.Before(end) {
			result = append(result, e)
		}
	}
	return result
}

func (m *MemoryCalendar) GetEventsForDay(userID int, day time.Time) ([]Event, error) {
	start := day.Truncate(24 * time.Hour)
	end := start.Add(24 * time.Hour)
	return m.GetEventsInRange(userID, start, end), nil
}

func (m *MemoryCalendar) GetEventsForWeek(userID int, day time.Time) ([]Event, error) {
	day = day.Truncate(24 * time.Hour)
	start := day.AddDate(0, 0, -6) // За 7 дней включая день
	end := day.AddDate(0, 0, 1)
	return m.GetEventsInRange(userID, start, end), nil
}

func (m *MemoryCalendar) GetEventsForMonth(userID int, day time.Time) ([]Event, error) {
	start := time.Date(day.Year(), day.Month(), 1, 0, 0, 0, 0, day.Location())
	end := start.AddDate(0, 1, 0)
	return m.GetEventsInRange(userID, start, end), nil
}

// --- HTTP Handlers и конфигурация ---

type Server struct {
	calendar CalendarService
}

func (s *Server) CreateEventHandler(c *gin.Context) {
	var input struct {
		UserID      int    `json:"user_id" form:"user_id" binding:"required"`
		DateStr     string `json:"date" form:"date" binding:"required"`
		Description string `json:"event" form:"event" binding:"required"`
	}
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input: " + err.Error()})
		return
	}
	date, err := time.Parse("2006-01-02", input.DateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format, expected YYYY-MM-DD"})
		return
	}

	event := Event{
		UserID:      input.UserID,
		Date:        date,
		Description: input.Description,
	}
	id, err := s.calendar.CreateEvent(event)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create event"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": fmt.Sprintf("event created with id %d", id)})
}

func (s *Server) UpdateEventHandler(c *gin.Context) {
	var input struct {
		ID          int    `json:"id" form:"id" binding:"required"`
		UserID      int    `json:"user_id" form:"user_id" binding:"required"`
		DateStr     string `json:"date" form:"date" binding:"required"`
		Description string `json:"event" form:"event" binding:"required"`
	}
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input: " + err.Error()})
		return
	}
	date, err := time.Parse("2006-01-02", input.DateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format, expected YYYY-MM-DD"})
		return
	}
	event := Event{
		UserID:      input.UserID,
		Date:        date,
		Description: input.Description,
	}
	err = s.calendar.UpdateEvent(input.ID, event)
	if err != nil {
		if errors.Is(err, ErrEventNotFound) {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update event"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": fmt.Sprintf("event %d updated", input.ID)})
}

func (s *Server) DeleteEventHandler(c *gin.Context) {
	var input struct {
		ID int `json:"id" form:"id" binding:"required"`
	}
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input, missing id"})
		return
	}
	err := s.calendar.DeleteEvent(input.ID)
	if err != nil {
		if errors.Is(err, ErrEventNotFound) {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete event"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": fmt.Sprintf("event %d deleted", input.ID)})
}

func (s *Server) parseUserIDAndDate(c *gin.Context) (int, time.Time, error) {
	userIDStr := c.Query("user_id")
	dateStr := c.Query("date")

	if userIDStr == "" {
		return 0, time.Time{}, errors.New("missing user_id parameter")
	}
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		return 0, time.Time{}, fmt.Errorf("invalid user_id: %w", err)
	}
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return 0, time.Time{}, fmt.Errorf("invalid date: %w", err)
	}
	return userID, date, nil
}

func (s *Server) GetEventsForDayHandler(c *gin.Context) {
	userID, date, err := s.parseUserIDAndDate(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	events, err := s.calendar.GetEventsForDay(userID, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": events})
}

func (s *Server) GetEventsForWeekHandler(c *gin.Context) {
	userID, date, err := s.parseUserIDAndDate(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	events, err := s.calendar.GetEventsForWeek(userID, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": events})
}

func (s *Server) GetEventsForMonthHandler(c *gin.Context) {
	userID, date, err := s.parseUserIDAndDate(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	events, err := s.calendar.GetEventsForMonth(userID, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": events})
}

// Logger middleware для логирования запросов
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)
		log.Printf("%s %s %s", c.Request.Method, c.Request.URL, duration)
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	router := gin.New()
	router.Use(Logger(), gin.Recovery())

	service := NewMemoryCalendar()
	server := &Server{calendar: service}

	router.POST("/create_event", server.CreateEventHandler)
	router.PATCH("/update_event", server.UpdateEventHandler)
	router.DELETE("/delete_event", server.DeleteEventHandler)
	router.GET("/events_for_day", server.GetEventsForDayHandler)
	router.GET("/events_for_week", server.GetEventsForWeekHandler)
	router.GET("/events_for_month", server.GetEventsForMonthHandler)

	log.Printf("Starting server on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
