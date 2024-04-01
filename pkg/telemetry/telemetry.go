/*
Copyright paskal.maksim@gmail.com
Licensed under the Apache License, Version 2.0 (the "License")
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package telemetry

import (
	"context"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

const serviceName = "kubernetes-manager"

func Init(ctx context.Context) error {
	otelExporter, err := otlptracegrpc.New(ctx)
	if err != nil {
		return errors.Wrap(err, "error creating exporter")
	}

	traceResource, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
		),
	)
	if err != nil {
		return errors.Wrap(err, "error creating resource")
	}

	traceProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(otelExporter),
		sdktrace.WithResource(traceResource),
	)

	go func() {
		<-ctx.Done()
		_ = traceProvider.Shutdown(ctx)
	}()

	otel.SetTracerProvider(traceProvider)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return nil
}

func Start(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) { //nolint:lll
	return otel.Tracer("").Start(ctx, spanName, opts...) //nolint:spancheck
}

func Attributes(span trace.Span, att map[string]string) {
	for k, v := range att {
		span.SetAttributes(attribute.String(k, v))
	}
}

func Event(span trace.Span, name string, att map[string]string) {
	opts := make([]trace.EventOption, 0)

	for k, v := range att {
		opts = append(opts, trace.WithAttributes(attribute.String(k, v)))
	}

	span.AddEvent(name, opts...)
}
