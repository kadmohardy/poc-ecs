package main

import (
	"context"
	"log"

	"poc-ecs/internal/server"
	"poc-ecs/internal/telemetry"
)

func main() {
	ctx := context.Background()

	tel, err := telemetry.Init(ctx)

	if err != nil {
		log.Fatalf("failed to initialize telemetry: %v", err)
	}

	defer func() {
		if err := tel.TracerShutdown(ctx); err != nil {
			log.Printf("failed to shutdown tracer: %v", err)
		}
	}()

	srv := server.NewHTTPServer()

	log.Println("server listening on :8080")

	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
