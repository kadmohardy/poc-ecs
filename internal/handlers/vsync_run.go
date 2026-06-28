package handlers

import (
	"math/rand"
	"net/http"
	"poc-ecs/internal/telemetry"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

func VsyncRun(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("vsync")

	ctx, span := tracer.Start(
		r.Context(),
		"vsync.run",
	)

	defer span.End()

	telemetry.VsyncExecutions.Add(
		ctx,
		1,
	)

	result := rand.Intn(100)

	switch {

	case result < 70:

		telemetry.VsyncPartnerOK.Add(
			ctx,
			1,
		)

		span.SetAttributes(
			attribute.String(
				"vsync.result",
				"success",
			),
		)

		w.WriteHeader(http.StatusOK)

		_, _ = w.Write(
			[]byte("vsync success"),
		)

	case result < 90:

		telemetry.VsyncInvalid.Add(
			ctx,
			1,
		)

		span.SetAttributes(
			attribute.String(
				"vsync.result",
				"invalid",
			),
			attribute.String(
				"vsync.reason",
				"account_blocked",
			),
		)

		w.WriteHeader(http.StatusConflict)

		_, _ = w.Write(
			[]byte("invalid vsync"),
		)

	default:

		telemetry.VsyncPartnerFailed.Add(
			ctx,
			1,
		)

		span.RecordError(
			http.ErrHandlerTimeout,
		)

		span.SetAttributes(
			attribute.String(
				"vsync.result",
				"partner_error",
			),
		)

		w.WriteHeader(
			http.StatusInternalServerError,
		)

		_, _ = w.Write(
			[]byte("btg unavailable"),
		)
	}
}
