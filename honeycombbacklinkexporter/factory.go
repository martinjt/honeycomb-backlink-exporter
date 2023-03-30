// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package honeycombbacklinkexporter

import (
	"context"
	"fmt"

	"github.com/honeycombio/libhoney-go"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
)

type honeycombExporter struct {
	builder             *libhoney.Builder
	onError             func(error)
	logger              *zap.Logger
	sampleRateAttribute string
}

type link struct {
	TraceID        string `json:"trace.trace_id"`
	ParentID       string `json:"trace.parent_id,omitempty"`
	LinkTraceID    string `json:"trace.link.trace_id"`
	LinkSpanID     string `json:"trace.link.span_id"`
	AnnotationType string `json:"meta.annotation_type"`
}

const (
	// The value of "type" key in configuration.
	typeStr = "honeycombbacklinkexporter"
	// The stability level of the exporter.
	stability = component.StabilityLevelAlpha
)

// NewFactory creates a factory for OTLP exporter.
func NewFactory() exporter.Factory {
	return exporter.NewFactory(
		typeStr,
		createDefaultConfig,
		exporter.WithTraces(createTracesExporter, stability))
}

func createDefaultConfig() component.Config {
	return &Config{}
}

func createTracesExporter(
	_ context.Context,
	set exporter.CreateSettings,
	cfg component.Config,
) (exporter.Traces, error) {
	eCfg := cfg.(*Config)
	exporter, err := newHoneycombTracesExporter(eCfg, set.Logger)
	if err != nil {
		return nil, err
	}

	return exporterhelper.NewTracesExporter(
		context.TODO(),
		set,
		cfg,
		exporter.pushTraceData,
	)
}

func newHoneycombTracesExporter(cfg *Config, logger *zap.Logger) (*honeycombExporter, error) {
	libhoneyConfig := libhoney.Config{
		APIKey: cfg.APIKey,
	}

	if err := libhoney.Init(libhoneyConfig); err != nil {
		return nil, err
	}
	builder := libhoney.NewBuilder()
	exporter := &honeycombExporter{
		builder: builder,
		logger:  logger,
		onError: func(err error) {
			logger.Warn(err.Error())
		},
	}

	return exporter, nil
}

func (e *honeycombExporter) pushTraceData(ctx context.Context, td ptrace.Traces) error {
	var errs error
	rs := td.ResourceSpans()
	for i := 0; i < rs.Len(); i++ {
		rsSpan := rs.At(i)

		// Extract Resource attributes, they will be added to every span.
		resourceAttrs := spanAttributesToMap(rsSpan.Resource().Attributes())

		ils := rsSpan.ScopeSpans()
		for j := 0; j < ils.Len(); j++ {
			ilsSpan := ils.At(j)
			spans := ilsSpan.Spans()
			for k := 0; k < spans.Len(); k++ {
				span := spans.At(k)
				ev := e.builder.NewEvent()

				for k, v := range resourceAttrs {
					ev.AddField(k, v)
				}

				lib := ilsSpan.Scope()
				if name := lib.Name(); name != "" {
					ev.AddField("library.name", name)
				}
				if version := lib.Version(); version != "" {
					ev.AddField("library.version", version)
				}

				ev.Timestamp = timestampToTime(span.StartTimestamp())

				e.sendSpanLinks(span, resourceAttrs)
			}
		}
	}

	return errs
}

// sendSpanLinks gets the list of links associated with this span and sends them as
// separate events to Honeycomb, with a span type "link".
func (e *honeycombExporter) sendSpanLinks(span ptrace.Span, resourceAttrs map[string]interface{}) {
	links := span.Links()

	for i := 0; i < links.Len(); i++ {
		l := links.At(i)

		ev := e.builder.NewEvent()

		if err := ev.Add(link{
			TraceID:        getHoneycombTraceID(l.TraceID()),
			ParentID:       l.SpanID().String(),
			LinkTraceID:    getHoneycombTraceID(span.TraceID()),
			LinkSpanID:     span.SpanID().String(),
			AnnotationType: "link",
		}); err != nil {
			e.logger.Error(err.Error())
			e.onError(err)
		}

		for k, v := range resourceAttrs {
			ev.AddField(k, v)
			if k == "service.name" {
				ev.Dataset = fmt.Sprintf("%v", v)
			}
		}
		attrs := spanAttributesToMap(l.Attributes())
		for k, v := range attrs {
			ev.AddField(k, v)
		}
		if err := ev.Send(); err != nil {
			e.logger.Error(err.Error())
			e.onError(err)
		}
	}
}
