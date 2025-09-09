package main

import (
	"testing"
	"time"
)

func TestCreateEvent(t *testing.T) {
	cal := NewMemoryCalendar()
	event := Event{
		UserID:      1,
		Date:        time.Date(2025, 9, 9, 0, 0, 0, 0, time.UTC),
		Description: "Test event",
	}
	id, err := cal.CreateEvent(event)
	if err != nil {
		t.Fatalf("CreateEvent error: %v", err)
	}
	if id == 0 {
		t.Fatalf("CreateEvent returned id=0")
	}
}

func TestUpdateEvent(t *testing.T) {
	cal := NewMemoryCalendar()
	event := Event{
		UserID:      1,
		Date:        time.Date(2025, 9, 9, 0, 0, 0, 0, time.UTC),
		Description: "Test event",
	}
	id, _ := cal.CreateEvent(event)

	updatedEvent := Event{
		UserID:      1,
		Date:        time.Date(2025, 9, 10, 0, 0, 0, 0, time.UTC),
		Description: "Updated event",
	}

	err := cal.UpdateEvent(id, updatedEvent)
	if err != nil {
		t.Fatalf("UpdateEvent error: %v", err)
	}

	storedEvent := cal.events[id]
	if storedEvent.Description != "Updated event" {
		t.Errorf("UpdateEvent did not update Description")
	}

	if !storedEvent.Date.Equal(updatedEvent.Date) {
		t.Errorf("UpdateEvent did not update Date")
	}
}

func TestUpdateEvent_NotFound(t *testing.T) {
	cal := NewMemoryCalendar()
	err := cal.UpdateEvent(999, Event{
		UserID:      1,
		Date:        time.Now(),
		Description: "Nothing",
	})
	if err != ErrEventNotFound {
		t.Fatalf("Expected ErrEventNotFound, got %v", err)
	}
}

func TestDeleteEvent(t *testing.T) {
	cal := NewMemoryCalendar()
	event := Event{
		UserID:      1,
		Date:        time.Now(),
		Description: "To delete",
	}
	id, _ := cal.CreateEvent(event)

	err := cal.DeleteEvent(id)
	if err != nil {
		t.Fatalf("DeleteEvent error: %v", err)
	}

	if _, exists := cal.events[id]; exists {
		t.Errorf("DeleteEvent did not remove event")
	}
}

func TestDeleteEvent_NotFound(t *testing.T) {
	cal := NewMemoryCalendar()
	err := cal.DeleteEvent(12345)
	if err != ErrEventNotFound {
		t.Fatalf("Expected ErrEventNotFound, got %v", err)
	}
}

func TestGetEventsForDay(t *testing.T) {
	cal := NewMemoryCalendar()
	userID := 1
	day := time.Date(2025, 9, 9, 0, 0, 0, 0, time.UTC)

	event := Event{
		UserID:      userID,
		Date:        day,
		Description: "Day event",
	}
	cal.CreateEvent(event)

	events, err := cal.GetEventsForDay(userID, day)
	if err != nil {
		t.Fatalf("GetEventsForDay error: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("GetEventsForDay expected 1 event, got %d", len(events))
	}
	if events[0].Description != "Day event" {
		t.Errorf("GetEventsForDay returned wrong event")
	}
}

func TestGetEventsForWeek(t *testing.T) {
	cal := NewMemoryCalendar()
	userID := 1
	day := time.Date(2025, 9, 9, 0, 0, 0, 0, time.UTC)

	// Событие 5 дней назад – должно попасть
	event1 := Event{
		UserID:      userID,
		Date:        day.AddDate(0, 0, -5),
		Description: "Week event 1",
	}
	// Событие 8 дней назад – не попадёт
	event2 := Event{
		UserID:      userID,
		Date:        day.AddDate(0, 0, -8),
		Description: "Week event 2",
	}
	cal.CreateEvent(event1)
	cal.CreateEvent(event2)

	events, err := cal.GetEventsForWeek(userID, day)
	if err != nil {
		t.Fatalf("GetEventsForWeek error: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("GetEventsForWeek expected 1 event, got %d", len(events))
	}
	if events[0].Description != "Week event 1" {
		t.Errorf("GetEventsForWeek returned wrong event")
	}
}

func TestGetEventsForMonth(t *testing.T) {
	cal := NewMemoryCalendar()
	userID := 1
	day := time.Date(2025, 9, 15, 0, 0, 0, 0, time.UTC)

	eventInMonth := Event{
		UserID:      userID,
		Date:        time.Date(2025, 9, 10, 0, 0, 0, 0, time.UTC),
		Description: "September event",
	}
	eventNotInMonth := Event{
		UserID:      userID,
		Date:        time.Date(2025, 8, 30, 0, 0, 0, 0, time.UTC),
		Description: "August event",
	}
	cal.CreateEvent(eventInMonth)
	cal.CreateEvent(eventNotInMonth)

	events, err := cal.GetEventsForMonth(userID, day)
	if err != nil {
		t.Fatalf("GetEventsForMonth error: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("GetEventsForMonth expected 1 event, got %d", len(events))
	}
	if events[0].Description != "September event" {
		t.Errorf("GetEventsForMonth returned wrong event")
	}
}
