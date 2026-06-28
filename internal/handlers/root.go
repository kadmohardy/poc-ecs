package handlers

import (
	"fmt"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

func Root(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("go-otel-demo")

	_, span := tracer.Start(r.Context(), "root-handler")
	defer span.End()

	span.SetAttributes(
		attribute.String("http.method", r.Method),
		attribute.String("http.path", r.URL.Path),
	)

	// simulate work
	time.Sleep(100 * time.Millisecond)

	message := "hello from golang + otel"

	span.AddEvent("sending response")

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)

	_, err := fmt.Fprintln(w, message)

	if err != nil {
		span.RecordError(err)

		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
}
