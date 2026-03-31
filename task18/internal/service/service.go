package service

import (
    "errors"
    "time"

    "task18/internal/repository"
)

var (
    ErrInvalidDate = errors.New("invalid date format, expected YYYY-MM-DD")
    ErrBadRequest = errors.New("bad request")
)

func parseDate(dateStr string) (time.Time, error) {
    if dateStr == "" {
        return time.Time{}, ErrInvalidDate
    }
    d, err := time.Parse("2006-01-02", dateStr)
    if err != nil {
        return time.Time{}, ErrInvalidDate
    }
    return d, nil
}

type Service struct {
    repo repository.Repository
}

func NewService(r repository.Repository) *Service {
    return &Service{repo: r}
}

func (s *Service) CreateEvent(userID int64, date string, text string) (repository.Event, error) {
    if userID <= 0 || text == "" {
        return repository.Event{}, ErrBadRequest
    }
    d, err := parseDate(date)
    if err != nil {
        return repository.Event{}, err
    }
    ev := repository.Event{UserID: userID, Date: d, Text: text}
    return s.repo.Create(ev)
}

func (s *Service) UpdateEvent(id, userID int64, date string, text string) error {
    if id <= 0 || userID <= 0 || text == "" {
        return ErrBadRequest
    }
    d, err := parseDate(date)
    if err != nil {
        return err
    }
    ev := repository.Event{ID: id, UserID: userID, Date: d, Text: text}
    return s.repo.Update(ev)
}

func (s *Service) DeleteEvent(id, userID int64) error {
    if id <= 0 || userID <= 0 {
        return ErrBadRequest
    }
    return s.repo.Delete(userID, id)
}

func (s *Service) EventsForDay(userID int64, date string) ([]repository.Event, error) {
    if userID <= 0 {
        return nil, ErrBadRequest
    }
    d, err := parseDate(date)
    if err != nil {
        return nil, err
    }
    from := d
    to := d
    return s.repo.EventsForPeriod(userID, from, to)
}

func (s *Service) EventsForWeek(userID int64, date string) ([]repository.Event, error) {
    if userID <= 0 {
        return nil, ErrBadRequest
    }
    d, err := parseDate(date)
    if err != nil {
        return nil, err
    }
    from := d
    to := d.AddDate(0, 0, 6)
    return s.repo.EventsForPeriod(userID, from, to)
}

func (s *Service) EventsForMonth(userID int64, date string) ([]repository.Event, error) {
    if userID <= 0 {
        return nil, ErrBadRequest
    }
    d, err := parseDate(date)
    if err != nil {
        return nil, err
    }

    from := time.Date(d.Year(), d.Month(), 1, 0, 0, 0, 0, time.UTC)
    to := from.AddDate(0, 1, -1)
    return s.repo.EventsForPeriod(userID, from, to)
}
