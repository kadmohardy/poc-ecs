package main

import (
	"net/http"
	"time"

	"poc-ecs/internal/handlers"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func NewHTTPServer(h *Handlers) *http.Server {

	mux := http.NewServeMux()

	mux.Handle(
		"/",
		otelhttp.NewHandler(
			http.HandlerFunc(handlers.Root),
			"root",
		),
	)

	mux.Handle(
		"/pix/initiate",
		otelhttp.NewHandler(
			http.HandlerFunc(h.Pix.PixInitiate),
			"pix-initiate",
		),
	)

	mux.Handle(
		"/pix/confirm",
		otelhttp.NewHandler(
			http.HandlerFunc(h.Pix.PixConfirm),
			"pix-confirm",
		),
	)

	mux.Handle(
		"/vsync/run",
		otelhttp.NewHandler(
			http.HandlerFunc(handlers.VsyncRun),
			"vsync-run",
		),
	)

	mux.HandleFunc("/healthcheck", handlers.Health)

	return &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  30 * time.Second,
	}
}
