package repository

import (
    "errors"
    "sync"
    "time"
)

var (
    ErrNotFound = errors.New("event not found")
)

type Event struct {
    ID     int64     `json:"id"`
    UserID int64     `json:"user_id"`
    Date   time.Time `json:"date"`
    Text   string    `json:"event"`
}

type Repository interface {
    Create(event Event) (Event, error)
    Update(event Event) error
    Delete(userID, id int64) error
    EventsForPeriod(userID int64, from, to time.Time) ([]Event, error)
    EventByID(userID, id int64) (Event, error)
}

type InMemoryRepo struct {
    mu      sync.RWMutex
    events  map[int64]Event
    nextID  int64
}

func NewInMemoryRepo() *InMemoryRepo {
    return &InMemoryRepo{events: make(map[int64]Event), nextID: 1}
}

func (r *InMemoryRepo) Create(event Event) (Event, error) {
    r.mu.Lock()
    defer r.mu.Unlock()

    event.ID = r.nextID
    r.nextID++

    r.events[event.ID] = event
    return event, nil
}

func (r *InMemoryRepo) Update(event Event) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    existing, ok := r.events[event.ID]
    if !ok || existing.UserID != event.UserID {
        return ErrNotFound
    }

    r.events[event.ID] = event
    return nil
}

func (r *InMemoryRepo) Delete(userID, id int64) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    existing, ok := r.events[id]
    if !ok || existing.UserID != userID {
        return ErrNotFound
    }
    delete(r.events, id)
    return nil
}

func (r *InMemoryRepo) EventByID(userID, id int64) (Event, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()

    e, ok := r.events[id]
    if !ok || e.UserID != userID {
        return Event{}, ErrNotFound
    }
    return e, nil
}

func (r *InMemoryRepo) EventsForPeriod(userID int64, from, to time.Time) ([]Event, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()

    out := make([]Event, 0)
    for _, e := range r.events {
        if e.UserID != userID {
            continue
        }
        if (e.Date.Equal(from) || e.Date.After(from)) && (e.Date.Equal(to) || e.Date.Before(to)) {
            out = append(out, e)
        }
    }
    return out, nil
}
