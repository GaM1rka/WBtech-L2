package handler

import (
    "encoding/json"
    "io"
    "log"
    "net/http"
    "strconv"

    "task18/internal/repository"
    "task18/internal/service"
)

type Service interface {
    CreateEvent(userID int64, date string, text string) (repository.Event, error)
    UpdateEvent(id, userID int64, date string, text string) error
    DeleteEvent(id, userID int64) error
    EventsForDay(userID int64, date string) ([]repository.Event, error)
    EventsForWeek(userID int64, date string) ([]repository.Event, error)
    EventsForMonth(userID int64, date string) ([]repository.Event, error)
}

type Handler struct {
    svc *service.Service
}

type EventRequest struct {
    ID     int64  `json:"id"`
    UserID int64  `json:"user_id"`
    Date   string `json:"date"`
    Text   string `json:"event"`
}

func New(svc *service.Service) *Handler {
    return &Handler{svc: svc}
}

func jsonResponse(w http.ResponseWriter, status int, payload interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    _ = json.NewEncoder(w).Encode(payload)
}

func errorResponse(w http.ResponseWriter, status int, msg string) {
    jsonResponse(w, status, map[string]string{"error": msg})
}

func successResponse(w http.ResponseWriter, msg string) {
    jsonResponse(w, http.StatusOK, map[string]string{"result": msg})
}

func parseBody(r *http.Request, req *EventRequest) error {
    ct := r.Header.Get("Content-Type")
    if ct == "application/json" || ct == "application/json; charset=utf-8" {
        body, err := io.ReadAll(r.Body)
        if err != nil {
            return err
        }
        return json.Unmarshal(body, req)
    }

    if err := r.ParseForm(); err != nil {
        return err
    }

    if val := r.FormValue("id"); val != "" {
        if v, err := strconv.ParseInt(val, 10, 64); err == nil {
            req.ID = v
        }
    }
    if val := r.FormValue("user_id"); val != "" {
        if v, err := strconv.ParseInt(val, 10, 64); err == nil {
            req.UserID = v
        }
    }
    req.Date = r.FormValue("date")
    req.Text = r.FormValue("event")
    return nil
}

func (h *Handler) CreateEvent(w http.ResponseWriter, r *http.Request) {
    var req EventRequest
    if err := parseBody(r, &req); err != nil {
        errorResponse(w, http.StatusBadRequest, "invalid request")
        return
    }

    ev, err := h.svc.CreateEvent(req.UserID, req.Date, req.Text)
    if err != nil {
        status := http.StatusBadRequest
        if err == service.ErrInvalidDate {
            status = http.StatusBadRequest
        }
        jsonResponse(w, status, map[string]string{"error": err.Error()})
        return
    }

    jsonResponse(w, http.StatusOK, map[string]interface{}{"result": "event created", "event": map[string]interface{}{"id": ev.ID, "user_id": ev.UserID, "date": ev.Date.Format("2006-01-02"), "event": ev.Text}})
}

func (h *Handler) UpdateEvent(w http.ResponseWriter, r *http.Request) {
    var req EventRequest
    if err := parseBody(r, &req); err != nil {
        errorResponse(w, http.StatusBadRequest, "invalid request")
        return
    }

    err := h.svc.UpdateEvent(req.ID, req.UserID, req.Date, req.Text)
    if err != nil {
        status := http.StatusBadRequest
        if err == repository.ErrNotFound {
            status = http.StatusServiceUnavailable
        }
        jsonResponse(w, status, map[string]string{"error": err.Error()})
        return
    }
    successResponse(w, "event updated")
}

func (h *Handler) DeleteEvent(w http.ResponseWriter, r *http.Request) {
    var req EventRequest
    if err := parseBody(r, &req); err != nil {
        errorResponse(w, http.StatusBadRequest, "invalid request")
        return
    }

    err := h.svc.DeleteEvent(req.ID, req.UserID)
    if err != nil {
        status := http.StatusBadRequest
        if err == repository.ErrNotFound {
            status = http.StatusServiceUnavailable
        }
        jsonResponse(w, status, map[string]string{"error": err.Error()})
        return
    }
    successResponse(w, "event deleted")
}

func eventListResponse(w http.ResponseWriter, events []repository.Event) {
    out := make([]map[string]interface{}, 0, len(events))
    for _, e := range events {
        out = append(out, map[string]interface{}{
            "id":      e.ID,
            "user_id": e.UserID,
            "date":    e.Date.Format("2006-01-02"),
            "event":   e.Text,
        })
    }
    jsonResponse(w, http.StatusOK, map[string]interface{}{"result": "ok", "events": out})
}

func (h *Handler) EventsForDay(w http.ResponseWriter, r *http.Request) {
    q := r.URL.Query()
    uid, err := strconv.ParseInt(q.Get("user_id"), 10, 64)
    if err != nil || uid <= 0 {
        errorResponse(w, http.StatusBadRequest, "user_id required")
        return
    }
    date := q.Get("date")
    events, err := h.svc.EventsForDay(uid, date)
    if err != nil {
        status := http.StatusBadRequest
        if err == service.ErrInvalidDate {
            status = http.StatusBadRequest
        }
        jsonResponse(w, status, map[string]string{"error": err.Error()})
        return
    }
    eventListResponse(w, events)
}

func (h *Handler) EventsForWeek(w http.ResponseWriter, r *http.Request) {
    q := r.URL.Query()
    uid, err := strconv.ParseInt(q.Get("user_id"), 10, 64)
    if err != nil || uid <= 0 {
        errorResponse(w, http.StatusBadRequest, "user_id required")
        return
    }
    date := q.Get("date")
    events, err := h.svc.EventsForWeek(uid, date)
    if err != nil {
        status := http.StatusBadRequest
        if err == service.ErrInvalidDate {
            status = http.StatusBadRequest
        }
        jsonResponse(w, status, map[string]string{"error": err.Error()})
        return
    }
    eventListResponse(w, events)
}

func (h *Handler) EventsForMonth(w http.ResponseWriter, r *http.Request) {
    q := r.URL.Query()
    uid, err := strconv.ParseInt(q.Get("user_id"), 10, 64)
    if err != nil || uid <= 0 {
        errorResponse(w, http.StatusBadRequest, "user_id required")
        return
    }
    date := q.Get("date")
    events, err := h.svc.EventsForMonth(uid, date)
    if err != nil {
        status := http.StatusBadRequest
        if err == service.ErrInvalidDate {
            status = http.StatusBadRequest
        }
        jsonResponse(w, status, map[string]string{"error": err.Error()})
        return
    }
    eventListResponse(w, events)
}

func LoggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        log.Printf("%s %s %s", r.Method, r.URL.String(), r.RemoteAddr)
        next.ServeHTTP(w, r)
    })
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
    mux.HandleFunc("/create_event", h.CreateEvent)
    mux.HandleFunc("/update_event", h.UpdateEvent)
    mux.HandleFunc("/delete_event", h.DeleteEvent)
    mux.HandleFunc("/events_for_day", h.EventsForDay)
    mux.HandleFunc("/events_for_week", h.EventsForWeek)
    mux.HandleFunc("/events_for_month", h.EventsForMonth)
}
