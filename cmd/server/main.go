package main

import (
	"context"
	"log"
	"os"

	"poc-ecs/internal/handlers"
	"poc-ecs/internal/infra"
	"poc-ecs/internal/queue"
	"poc-ecs/internal/telemetry"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

func main() {
	ctx := context.Background()
	traceRepository := infra.NewTraceRepository()

	// AWS Config
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("failed to load AWS config: %v", err)
	}

	// OpenTelemetry
	tel, err := telemetry.Init(ctx)
	if err != nil {
		log.Fatalf("failed to initialize telemetry: %v", err)
	}

	defer func() {
		if err := tel.TracerShutdown(ctx); err != nil {
			log.Printf("failed to shutdown tracer: %v", err)
		}
	}()

	// SQS
	queueURL := os.Getenv("PIX_QUEUE_URL")

	sqsClient := sqs.NewFromConfig(cfg)

	pixQueue := queue.NewSQSQueue(
		sqsClient,
		queueURL,
	)

	// Handlers
	pixHandler := &handlers.PixHandler{
		Queue:           pixQueue,
		TraceRepository: traceRepository,
	}

	appHandlers := &Handlers{
		Pix: pixHandler,
	}

	// HTTP Server
	srv := NewHTTPServer(appHandlers)

	log.Println("server listening on :8080")

	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
