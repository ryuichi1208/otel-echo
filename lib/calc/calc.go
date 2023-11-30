package calc

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type Calc struct {
	tracer trace.Tracer
}

func NewCalc(tracer trace.Tracer) *Calc {
	return &Calc{
		tracer: tracer,
	}
}

func Error() error {
	return fmt.Errorf("error trace")
}

func (c *Calc) Add(ctx context.Context, x, y int) int {
	ctx, span := c.tracer.Start(ctx, "Add")
	defer span.End()
	span.SetAttributes(
		attribute.Int("x", x),
		attribute.Int("y", y),
	)

	err := Error()
	span.RecordError(err, trace.WithStackTrace(true))
	span.SetStatus(codes.Error, err.Error())

	time.Sleep(1 * time.Second)
	return x + y
}
