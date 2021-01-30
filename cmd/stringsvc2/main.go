package main

import (
	"context"
	"net/http"
	"os"

	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"

	"github.com/go-chi/chi"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
)

func loggingMiddleware(logger log.Logger) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			logger.Log("msg", "calling endpoint")
			defer logger.Log("msg", "called endpoint")
			return next(ctx, request)
		}
	}
}

func main() {
	fieldKeys := []string{"method", "error"}
	requestCount := kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Namespace: "my_group",
		Subsystem: "string_service",
		Name:      "request_count",
		Help:      "Number of requests received.",
	}, fieldKeys)
	requestLatency := kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
		Namespace: "my_group",
		Subsystem: "string_service",
		Name:      "request_latency_microseconds",
		Help:      "Total duration of requests in microseconds.",
	}, fieldKeys)
	countResult := kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
		Namespace: "my_group",
		Subsystem: "string_service",
		Name:      "count_result",
		Help:      "The result of each count method.",
	}, []string{}) // no fields here
	logger := log.NewLogfmtLogger(os.Stderr)
	var svc StringService = stringService{}
	svc = serviceLoggingMiddleware{
		logger: logger,
		next:   svc,
	}
	svc = instrumentingMiddleware{requestCount, requestLatency, countResult, svc}

	var uppercase endpoint.Endpoint
	uppercase = makeUppercaseEndpoint(svc)
	uppercase = loggingMiddleware(log.With(logger, "method", "uppercase"))(uppercase)

	var count endpoint.Endpoint
	count = makeCountEndpoint(svc)
	count = loggingMiddleware(log.With(logger, "method", "count"))(count)

	uppercaseHandler := httptransport.NewServer(
		uppercase,
		decodeUppercaseRequest,
		encodeResponse,
	)

	countHandler := httptransport.NewServer(
		count,
		decodeCountRequest,
		encodeResponse,
	)

	r := chi.NewRouter()
	r.Post("/uppercase", uppercaseHandler.ServeHTTP)
	r.Post("/count", countHandler.ServeHTTP)
	http.ListenAndServe(":3090", r)
}
