# go-trace
![Lint](https://github.com/fredbi/go-trace/actions/workflows/01-golang-lint.yaml/badge.svg)
![CI](https://github.com/fredbi/go-trace/actions/workflows/02-test.yaml/badge.svg)
[![Coverage Status](https://coveralls.io/repos/github/fredbi/go-trace/badge.svg?branch=master)](https://coveralls.io/github/fredbi/go-trace?branch=master)
![Vulnerability Check](https://github.com/fredbi/go-trace/actions/workflows/03-govulncheck.yaml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/fredbi/go-trace)](https://goreportcard.com/report/github.com/fredbi/go-trace)

![GitHub tag (latest by date)](https://img.shields.io/github/v/tag/fredbi/go-trace)
[![Go Reference](https://pkg.go.dev/badge/github.com/fredbi/go-trace.svg)](https://pkg.go.dev/github.com/fredbi/go-trace)
[![license](https://img.shields.io/badge/license/License-Apache-yellow.svg)](https://raw.githubusercontent.com/fredbi/go-trace/master/LICENSE.md)

Logging & tracing utilities for micro services.

Based on:
* `go.uber.org/zap`
* `go.opencensus.io/trace`
* `go.opentelemetry.io/otel`

Tested to be compatible wih Datadog tracing.

## Logging

The main idea is to unify logging and tracing, and insulate the app layer from the intricacies of tracing.

This repositories provides a few wrappers around a `zap.Logger`:
* a logger factory based on zap logger, with a convenient builder to link logs to trace spans
* a logger builder, e.g. to initialize a root logger for your service

TODOs:
* [] explore how to expose zerolog as an alternative to zap

### Usage

With OpenCensus tracing.

```go
import (
    "context"

    "github.com/fredbi/go-trace/log"
	"go.opencensus.io/trace"
)

func tracedFunc() {
    ctx := context.Background()
    zlg, closer := log.MustGetLogger("root") // builds a named zap logger with sensible defaults
    defer closer()

    lgf := log.NewFactory(zlg) // builds a logger with trace propagation

    ctx, span := trace.StartSpan(ctx, "span name")
    defer span.End()
    lg := lgf.For(ctx)

    lg.Info("log propagated as a trace span")
}
```

With OpenTelemetry tracing.

```go
import (
    "context"

    "github.com/fredbi/go-trace/log"
	"go.opentelemetry.io/otel"
)


func tracedFunc() {
    tracer := otel.Tracer("") // returns the global default tracer
    ctx := context.Background()
    zlg, closer := log.MustGetLogger("root")
    defer closer()

    lgf := log.NewFactory(zlg) // builds a logger with trace propagation

    ctx, span := tracer.Start(ctx, "span name")
    defer span.End()
    lg := lgf.For(ctx)

    lg.Info("log propagated as a trace span")
}
```

[Full example](https://github.com/fredbi/go-trace/blob/master/log/examples_test.go)

## Tracing

A simple wrapper to instrument tracing in apps with minimal boiler-plate.


With OpenCensus tracing.

```go
import (
    "context"

    "github.com/fredbi/go-trace/log"
    "github.com/fredbi/go-trace/opencensus/tracer"
)

type loggable struct {
    lgf log.Factory
}

func (l *loggable) Logger() log.Factory {
    return l.lgf
}

func tracedFunc() {
    zlg := log.MustGetLogger("root") // builds a named zap logger with sensible defaults
    lgf := log.NewFactory(zlg) // builds a logger with trace propagation
    component := loggable{lfg:  lgf}

    ctx, span, lg := tracer.StartSpan(context.Background(), component) // the span is named automatically from the calling function
    defer span.End()

    lg.Info("log propagated as a trace span")
}
```

[Full example](https://github.com/fredbi/go-trace/blob/master/opencensus/tracer/example_test.go)

With OpenTelemetry tracing.

```go
import (
    "context"

    "github.com/fredbi/go-trace/log"
    "github.com/fredbi/go-trace/otel/tracer"
)

type loggable struct {
    lgf log.Factory
    tracer trace.Tracer
}

func (l *loggable) Logger() log.Factory {
    return l.lgf
}

// Tracer returns the configured tracer. If this method, is not provided,
// the default globally registered OTEL tracer is returned.
func (l *loggable) Tracer() trace.Tracer {
    return l.tracer 
}

func tracedFunc() {
    zlg := log.MustGetLogger("root") // builds a named zap logger with sensible defaults
    lgf := log.NewFactory(zlg, log.WithOTEL(true)) // builds a logger with trace propagation
    component := loggable{lfg:  lgf}

    ctx, span, lg := tracer.StartSpan(context.Background(), component) // the span is named automatically from the calling function
    defer span.End()

    lg.Info("log propagated as a trace span")
}
```

[Full example](https://github.com/fredbi/go-trace/blob/master/otel/tracer/example_test.go)

## Middleware

* `middleware.LogRequests` logs all requests from a http server, using the logger factory

* `opencensus/middleware.OCHTTP` wraps the `ochttp` opencensus plugin into a more convenient middleware function.
* `otel/middleware.OTELHTTP` wraps the `otelhttp` OTEL contrib plugin.

## Exporters

Mock trace exporters for opencensus and OTEL.

## Misc
Various opencensus exporters (as a separate module).
* misc/influxdb: export opencensus metrics to an influxdb sink
* misc/amplitude (experimental): propagate trace event to the amplitude API

## Credits

Much inspired by prior art from @casualjim. Thanks so much.
