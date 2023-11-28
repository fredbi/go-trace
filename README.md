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

Tested to be compatible wih Datadog tracing.

## Logging

The main idea is to unify logging and tracing, and insulate the app layer from the intricacies of tracing.

This repo provides a few wrappers around a `zap.Logger`:
* a logger factory based on zap logger, with a convenient builder to link logs to trace spans
* a logger builder, e.g. to initialize a root logger for your service

### Usage

```go
import (
    "context"

    "github.com/fredbi/go-trace/log"
	"go.opencensus.io/trace"
)

ctx := context.Background()
zlg, closer := log.MustGetLogger("root") // builds a named zap logger with sensible defaults
defer closer()

lgf := log.NewFactory(zlg) // builds a logger with trace propagation

ctx, span := trace.StartSpan(ctx, "span name")
defer span.End()
lg := lgf.For(ctx)

lg.Info("log propagated as a trace span")
```

## Tracing

Simple utilities to instrument tracing inside apps with minimal boiler-plate.


```go
import (
    "context"

    "github.com/fredbi/go-trace/log"
    "github.com/fredbi/go-trace/trace"
)

type loggable struct {
    lgf log.Factory
}

func (l *loggable) Logger() log.Factory {
    return l.lgf
}

zlg := log.MustGetLogger("root") // builds a named zap logger with sensible defaults
lgf := log.NewFactory(zlg) // builds a logger with trace propagation
component := loggable{lfg:  lgf}

ctx, span, lg := tracer.StartSpan(context.Background(), component) // the span is named automatically from the calling function
defer span.End()

lg.Info("log propagated as a trace span")
```

[Full example](https://github.com/fredbi/go-trace/blob/master/tracer/example_test.go)

## Middleware

* `log/middleware/LogRequests` logs all requests from a http server, using the logger factory
* `tracer.Middleware` wraps the `ochttp` opencensus plugin in a more convenient middleware.

TODOs:
* [] expose middlewares from a the top level package

## Exporters

Various opencensus exporters (as a separate module).
* influxdb: export opencensus metrics to an influxdb sink
* amplitude (experimental): propagate trace event to the amplitude API

TODOs:
* [] opentelemetry/opentracing
* [] reorganize module at the top level

## Credits

Much inspired by prior art from @casualjim. Thanks so much.
