package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"poc-ecs/internal/telemetry"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type ConfirmRequest struct {
	EndToEndID string `json:"endToEndId"`
}

func (h *PixHandler) PixConfirm(w http.ResponseWriter, r *http.Request) {
	var req ConfirmRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Busca o traceparent salvo na iniciação
	traceparent, ok := h.TraceRepository.Get(req.EndToEndID)

	if !ok {
		http.Error(w, "trace context not found", http.StatusNotFound)
		return
	}

	// Reconstrói o contexto do trace
	carrier := propagation.MapCarrier{
		"traceparent": traceparent,
	}

	ctx := otel.GetTextMapPropagator().Extract(
		context.Background(),
		carrier,
	)

	tracer := otel.Tracer("pix")

	ctx, span := tracer.Start(
		ctx,
		"pix.confirm",
	)

	defer span.End()

	// Simula o processamento que depois será feito pela Lambda
	start := time.Now()

	time.Sleep(800 * time.Millisecond)

	duration := time.Since(start)

	telemetry.PixDuration.Record(
		ctx,
		duration.Milliseconds(),
		metric.WithAttributes(
			attribute.String("flow", "pix"),
		),
	)

	span.SetAttributes(
		attribute.String(
			"pix.end_to_end_id",
			req.EndToEndID,
		),
		attribute.String(
			"pix.phase",
			"confirm",
		),
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

	// Limpa o contexto da memória
	h.TraceRepository.Delete(
		req.EndToEndID,
	)

	w.WriteHeader(http.StatusOK)

	_, _ = w.Write(
		[]byte("confirmed"),
	)
}
