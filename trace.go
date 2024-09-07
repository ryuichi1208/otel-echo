package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

var tracer = otel.Tracer("otel-echo")

func initProvider(ctx context.Context) func() {
	fmt.Println("this1")
	// リソース情報（プロセス、ホスト、サービス名など）を設定
	res, err := resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithProcess(),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
		resource.WithOSType(),
		resource.WithProcessOwner(),
		resource.WithTelemetrySDK(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(os.Args[2]),
		),
	)
	if err != nil {
		log.Fatalf("failed to create resource: %v", err)
	}

	otelAgentAddr, ok := os.LookupEnv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if !ok {
		otelAgentAddr = "0.0.0.0:4317"
	}

	httpHeader := map[string]string{
		"X-Scope-OrgID": "1",
	}

	traceClient := otlptracehttp.NewClient(
		otlptracehttp.WithHeaders(httpHeader),
		otlptracehttp.WithInsecure(),
		otlptracehttp.WithEndpoint(otelAgentAddr),
		otlptracehttp.WithTimeout(5*time.Second),
		otlptracehttp.WithRetry(otlptracehttp.RetryConfig{ // エクスポート失敗時にバッチ送信をリトライするための設定
			Enabled:         true,
			InitialInterval: 500 * time.Millisecond, // 最初の失敗後にリトライするまでの待ち時間
			MaxInterval:     5 * time.Second,        // 最大待ち時間
			MaxElapsedTime:  30 * time.Second,       // 最大経過時間
		}),
	)
	traceExp, err := otlptrace.New(ctx, traceClient)
	if err != nil {
		log.Fatalf("failed to create trace exporter: %v", err)
	}

	bsp := sdktrace.NewBatchSpanProcessor(
		traceExp,
		sdktrace.WithMaxQueueSize(5000),
		sdktrace.WithMaxExportBatchSize(512),
	)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)

	// リクエストヘッダーからトレースIDとスパンIDを取得するための設定
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)
	otel.SetTracerProvider(tracerProvider)

	return func() {
		cxt, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		if err := traceExp.Shutdown(cxt); err != nil {
			otel.Handle(err)
		}
	}

	// fix コメント
}
