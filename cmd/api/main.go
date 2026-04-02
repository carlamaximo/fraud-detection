package main

import (
	"log"
	"log/slog"
	"net/http"

	"fraud-platform/internal/handler"
	"fraud-platform/internal/logging"
	"fraud-platform/internal/worker"
)

func main() {
	logging.Init()
	worker.StartWorkers(3)

	http.HandleFunc("/events", handler.HandleEvent)
	http.HandleFunc("/metrics", handler.HandleMetrics)
	http.HandleFunc("/dlq", handler.HandleDLQ)

	slog.Info("starting server", "port", ":8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
