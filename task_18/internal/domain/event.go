package domain

import "time"

type Event struct {
	ID          int       `json:"id"`          // Уникальный идентификатор события (UUID)
	Title       string    `json:"title"`       // Название события
	Description string    `json:"description"` // Описание события
	StartTime   time.Time `json:"start_time"`  // Время начала события
	EndTime     time.Time `json:"end_time"`    // Время окончания события
	Location    string    `json:"location"`    // Местоположение события
	CreatedAt   time.Time `json:"created_at"`  // Время создания события
	UpdatedAt   time.Time `json:"updated_at"`  // Время последнего обновления
}

type UpdateEventRequest struct {
	Title       *string    `json:"title,omitempty"`
	Description *string    `json:"description,omitempty"`
	StartTime   *time.Time `json:"start_time,omitempty"`
	EndTime     *time.Time `json:"end_time,omitempty"`
	Location    *string    `json:"location,omitempty"`
}

type Request struct {
	Title       string    `json:"title"`       // Название события
	Description string    `json:"description"` // Описание события
	StartTime   time.Time `json:"start_time"`  // Время начала события
	EndTime     time.Time `json:"end_time"`    // Время окончания события
	Location    string    `json:"location"`
}

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}
