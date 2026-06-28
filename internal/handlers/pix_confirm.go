package handlers

import (
	"encoding/json"
	"net/http"
	"poc-ecs/internal/telemetry"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type ConfirmRequest struct {
	ID string `json:"id"`
}

func PixConfirm(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("pix")

	ctx, span := tracer.Start(
		r.Context(),
		"pix.confirm",
	)

	defer span.End()

	var req ConfirmRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	startedAt, ok := Transactions[req.ID]

	if !ok {
		http.Error(w, "transaction not found", http.StatusNotFound)
		return
	}

	duration := time.Since(startedAt)

	telemetry.PixDuration.Record(
		ctx,
		duration.Milliseconds(),
		metric.WithAttributes(
			attribute.String("flow", "pix"),
		),
	)

	span.SetAttributes(
		attribute.String("pix.id", req.ID),
		attribute.String("pix.phase", "confirm"),
		attribute.Int64(
			"pix.duration_ms",
			duration.Milliseconds(),
		),
	)

	span.AddEvent(
		"pix confirmed",
		trace.WithAttributes(
			attribute.Int64(
				"duration_ms",
				duration.Milliseconds(),
			),
		),
	)

	delete(Transactions, req.ID)

	w.WriteHeader(http.StatusOK)

	_, _ = w.Write([]byte("confirmed"))
}
