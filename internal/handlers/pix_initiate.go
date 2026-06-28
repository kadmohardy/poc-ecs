package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var Transactions = map[string]time.Time{}

type InitiateRequest struct {
	ID string `json:"id"`
}

func PixInitiate(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("pix")

	_, span := tracer.Start(
		r.Context(),
		"pix.initiate",
	)

	defer span.End()

	var req InitiateRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	Transactions[req.ID] = time.Now()

	span.SetAttributes(
		attribute.String("pix.id", req.ID),
		attribute.String("pix.phase", "initiate"),
	)

	span.AddEvent("pix initiated")

	w.WriteHeader(http.StatusAccepted)

	_, _ = w.Write([]byte("initiated"))
}
