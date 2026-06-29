package handlers

import (
	"context"
	"encoding/json"
	"log"
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

	log.Printf("[PIX CONFIRM] TraceParent=%s", traceparent)
	// Reconstrói o contexto do trace
	carrier := propagation.MapCarrier{
		"traceparent": traceparent,
	}

	ctx := otel.GetTextMapPropagator().Extract(
		context.Background(),
		carrier,
	)

	spanCtx := trace.SpanContextFromContext(ctx)

	log.Printf("[PIX CONFIRM] Extracted TraceID=%s SpanID=%s",
		spanCtx.TraceID().String(),
		spanCtx.SpanID().String(),
	)

	tracer := otel.Tracer("pix")

	ctx, span := tracer.Start(
		ctx,
		"pix.confirm",
	)

	spanCtx = span.SpanContext()

	log.Printf(
		"[PIX CONFIRM] New Span TraceID=%s SpanID=%s",
		spanCtx.TraceID().String(),
		spanCtx.SpanID().String(),
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

	log.Printf(
		"[PIX CONFIRM] Duration=%dms",
		duration.Milliseconds(),
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
		attribute.String("pix.started_at", time.Now().UTC().Format(time.RFC3339)),
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

	log.Printf(
		"[PIX CONFIRM] Removing TraceContext EndToEndID=%s",
		req.EndToEndID,
	)

	// Limpa o contexto da memória
	h.TraceRepository.Delete(
		req.EndToEndID,
	)

	log.Printf(
		"[PIX CONFIRM] Pix confirmado EndToEndID=%s",
		req.EndToEndID,
	)

	w.WriteHeader(http.StatusOK)

	_, _ = w.Write(
		[]byte("confirmed"),
	)
}
