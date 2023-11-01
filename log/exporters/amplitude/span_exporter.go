// Package amplitude exports opencensus trace spans to the amplitude API,
// in some non-blocking way.
//
// A filter may be set on the span attributes.
//
// Usage: register the new exporter in your runtime.
//
//	 import(
//			"go.opencensus.io/trace"
//			"github.com/oneconcern/ocpkg/log/exporters/amplitude
//	 )
//
// exporter := amplitude.New("{api-key}") // you may add some filtered and encoding options here
//
//	if err := exporter.Start() ; err != nil {
//		log.Fatalf(err)
//	}
//
//	defer func() {
//	  _ = exporter.Stop()
//		}
//
// trace.RegisterExporter(exporter)
// ...
package amplitude

import (
	api "github.com/renatoaf/amplitude-go/amplitude"
	"go.opencensus.io/trace"
	"go.uber.org/zap"
)

var _ trace.Exporter = &TraceExporter{}

// New amplitude trace exporter, with an API key and some options.
//
// The created exporter must call Start() before actually start working.
func New(key string, opts ...Option) *TraceExporter {
	o := defaultOptions(opts...)

	return &TraceExporter{
		options: o,
		client:  newAPIClient(key, o),
	}
}

type TraceExporter struct {
	*options
	client *api.Client
}

func newAPIClient(key string, o *options) *api.Client {
	if o.clientOptions == nil {
		// use defaults from the amplitude API package

		return api.NewDefaultClient(key)
	}

	return api.NewClient(key, *o.clientOptions)
}

// Start the amplitude upload go routine
func (e *TraceExporter) Start() error {
	return e.client.Start()
}

// Stop the amplitude upload go routine, flushing all pending messages before shutting down
func (e *TraceExporter) Stop() error {
	if err := e.client.Flush(); err != nil {
		e.logger.Error(
			"amplitude exporter failed to flush events before shutting down",
			zap.Error(err),
		)
	}

	return e.client.Shutdown()
}

func (e *TraceExporter) ExportSpan(s *trace.SpanData) {
	if e.filters.IsFiltered(s) {
		return
	}

	event := e.spanEncoder(s)
	if event == nil {
		return
	}

	if err := e.client.LogEvent(event); err != nil {
		e.logger.Warn(
			"amplitude exporter log event error",
			zap.Error(err),
		)
	}
}
