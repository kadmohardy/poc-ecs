package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"poc-ecs/internal/infra"
	"poc-ecs/internal/queue"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
)

type PixHandler struct {
	Queue           queue.Queue
	TraceRepository *infra.TraceRepository
}

type InitiateRequest struct {
	EndToEndID string `json:"endToEndId"`
}

func (h *PixHandler) PixInitiate(w http.ResponseWriter, r *http.Request) {

	tracer := otel.Tracer("pix")

	ctx, span := tracer.Start(r.Context(), "pix.initiate")
	defer span.End()

	var req InitiateRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	span.SetAttributes(attribute.String("pix.end_to_end_id", req.EndToEndID), attribute.String(
		"pix.phase",
		"initiate",
	),
	)

	// Cria o traceparent
	carrier := propagation.MapCarrier{}

	otel.GetTextMapPropagator().Inject(
		ctx,
		carrier,
	)

	log.Printf("[PIX INIT] TraceParent=%s", carrier["traceparent"])

	// Persiste para uso futuro pela Lambda
	if err := h.TraceRepository.Save(req.EndToEndID, carrier["traceparent"]); err != nil {
		span.RecordError(err)

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("[PIX INIT] Saved TraceParent for EndToEndID=%s", req.EndToEndID)

	// // TODO:
	// // Publicar mensagem no SQS

	// h.Queue.Publish(ctx, body,
	// 	map[string]string{
	// 		"traceparent": carrier["traceparent"],
	// 	},
	// )

	log.Printf("[PIX INIT] Pix iniciado EndToEndID=%s", req.EndToEndID)
	span.AddEvent("pix initiated")

	w.WriteHeader(http.StatusAccepted)

	_, _ = w.Write(
		[]byte("initiated"),
	)
}
