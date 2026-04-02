package main

import (
	"log"
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

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
