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

This repo provides a few wrappers around a zap.Logger:
* a logger factory based on zap logger, with a convenient builder to link logs to trace spans
* a logger builder, e.g. to initialize a root logger for your service
* a simple middleware to trace a `http.Handler`

### Exporters

Various opencensus exporters.
* influxdb: export opencensus metrics to an influxdb sink
* amplitude (experimental): propagate trace event to the amplitude API

TODOs:
* [] opentelemetry/opentracing

## Tracing

Simple utilities to instrument tracing inside apps.

[Example Usage](https://github.com/fredbi/go-trace/blob/master/tracer/example_test.go)

## Credits

Much inspired by prior art from @casualjim. Thanks so much.
