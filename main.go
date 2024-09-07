package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/ryuichi1208/otel-echo/lib/calc"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	shutdown := initProvider(ctx)
	defer cancel()
	defer shutdown()

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(otelecho.Middleware("my-server"))

	e.GET("/", hello)
	e.GET("/2", hello2)
	e.GET("/3", hello3)

	// 引数で渡したポートでサーバーを起動する
	port, _ := strconv.Atoi(os.Args[1])
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", port)))
}

func hello(c echo.Context) error {
	//// _, span := tracer.Start(c.Request().Context(), "getUser")
	// defer span.End()
	time.Sleep(1 * time.Second)
	httpRequest(c.Request().Context())
	time.Sleep(1 * time.Second)
	hello2(c)
	return c.String(http.StatusOK, "Hello, World!")
}

func hello2(c echo.Context) error {
	c2 := calc.NewCalc(tracer)
	_, span := tracer.Start(c.Request().Context(), "getUser")
	defer span.End()
	span.SetStatus(codes.Error, "error")
	time.Sleep(1 * time.Second)
	c2.Add(c.Request().Context(), 1, 2)
	return c.String(http.StatusNotFound, "Hello, World!2")
}

func hello3(c echo.Context) error {
	_, span := tracer.Start(c.Request().Context(), "getUser")
	defer span.End()
	return c.String(http.StatusNotFound, "Hello, World!3")
}

func httpRequest(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "httpRequest")
	defer span.End()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8081/2", http.NoBody)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "example-service/1.0.0")
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

	for k, v := range req.Header {
		fmt.Println(k, v)
	}
	cli := &http.Client{
		Transport: otelhttp.NewTransport(
			http.DefaultTransport,
		),
	}
	resp, err := cli.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
