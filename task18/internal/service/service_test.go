package service_test

import (
    "testing"

    "task18/internal/repository"
    "task18/internal/service"
)

func TestCreateAndQueryEvents(t *testing.T) {
    repo := repository.NewInMemoryRepo()
    svc := service.NewService(repo)

    ev, err := svc.CreateEvent(1, "2023-12-31", "year-end party")
    if err != nil {
        t.Fatalf("create event failed: %v", err)
    }
    if ev.ID == 0 {
        t.Fatal("expected non-zero id")
    }

    events, err := svc.EventsForDay(1, "2023-12-31")
    if err != nil {
        t.Fatalf("events_for_day failed: %v", err)
    }
    if len(events) != 1 {
        t.Fatalf("expected 1 event, got %d", len(events))
    }
    if events[0].Text != "year-end party" {
        t.Fatalf("unexpected event text: %s", events[0].Text)
    }
}

func TestUpdateEvent(t *testing.T) {
    repo := repository.NewInMemoryRepo()
    svc := service.NewService(repo)

    ev, err := svc.CreateEvent(2, "2024-01-01", "new year")
    if err != nil {
        t.Fatalf("create event failed: %v", err)
    }

    err = svc.UpdateEvent(ev.ID, 2, "2024-01-02", "new year modified")
    if err != nil {
        t.Fatalf("update event failed: %v", err)
    }

    events, err := svc.EventsForDay(2, "2024-01-02")
    if err != nil {
        t.Fatalf("events_for_day failed: %v", err)
    }
    if len(events) != 1 || events[0].Text != "new year modified" {
        t.Fatalf("expected updated event, got %+v", events)
    }
}

func TestDeleteEvent(t *testing.T) {
    repo := repository.NewInMemoryRepo()
    svc := service.NewService(repo)

    ev, err := svc.CreateEvent(3, "2024-02-14", "valentines")
    if err != nil {
        t.Fatalf("create event failed: %v", err)
    }

    err = svc.DeleteEvent(ev.ID, 3)
    if err != nil {
        t.Fatalf("delete failed: %v", err)
    }

    events, err := svc.EventsForDay(3, "2024-02-14")
    if err != nil {
        t.Fatalf("events_for_day failed: %v", err)
    }
    if len(events) != 0 {
        t.Fatalf("expected 0 events after delete, got %d", len(events))
    }
}
