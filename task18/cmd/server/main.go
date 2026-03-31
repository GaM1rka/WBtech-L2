package main

import (
    "context"
    "flag"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "task18/config"
    "task18/internal/handler"
    "task18/internal/repository"
    "task18/internal/service"
)

func main() {
    cfg := config.Load()

    var port = cfg.Port
    flag.StringVar(&port, "port", cfg.Port, "port for server")
    flag.Parse()

    if envPort := os.Getenv("CALENDAR_PORT"); envPort != "" {
        port = envPort
    }

    repo := repository.NewInMemoryRepo()
    svc := service.NewService(repo)
    h := handler.New(svc)

    mux := http.NewServeMux()
    h.RegisterRoutes(mux)
    logged := handler.LoggingMiddleware(mux)

    srv := &http.Server{
        Addr:    ":" + port,
        Handler: logged,
    }

    go func() {
        log.Printf("server starting on %s", srv.Addr)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("listen failed: %v", err)
        }
    }()

    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    log.Println("server shutting down")
    if err := srv.Shutdown(ctx); err != nil {
        log.Fatalf("shutdown failed: %v", err)
    }
    log.Println("server stopped")
}
