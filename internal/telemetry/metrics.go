package telemetry

import (
	"go.opentelemetry.io/otel/metric"
)

var (
	VsyncExecutions    metric.Int64Counter
	VsyncPartnerOK     metric.Int64Counter
	VsyncPartnerFailed metric.Int64Counter
	VsyncInvalid       metric.Int64Counter
	PixDuration        metric.Int64Histogram
)
