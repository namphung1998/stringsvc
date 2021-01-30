package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/pkg/errors"
)

func main() {
	svc := stringService{}

	uppercaseHandler := httptransport.NewServer(
		makeUppercaseEndpoint(svc),
		decodeUppercaseRequest,
		encodeResponse,
	)

	countHandler := httptransport.NewServer(
		makeCountEndpoint(svc),
		decodeCountRequest,
		encodeResponse,
	)

	r := chi.NewRouter()
	r.Post("/uppercase", uppercaseHandler.ServeHTTP)
	r.Post("/count", countHandler.ServeHTTP)
	log.Fatal(http.ListenAndServe(":3090", nil))
}

func decodeUppercaseRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var request uppercaseRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}
	return request, nil
}

func decodeCountRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var request countRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}
	return request, nil
}

func encodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	return json.NewEncoder(w).Encode(response)
}

// StringService provides operations on strings
type StringService interface {
	Uppercase(s string) (string, error)
	Count(s string) int
}

var (
	ErrEmpty = errors.New("empty string")
)

type uppercaseRequest struct {
	S string `json:"s"`
}

type uppercaseResponse struct {
	V   string `json:"v"`
	Err string `json:"err,omitempty"`
}

type countRequest struct {
	S string `json:"s"`
}

type countResponse struct {
	V int `json:"v"`
}

type stringService struct{}

func (stringService) Uppercase(s string) (string, error) {
	if s == "" {
		return "", ErrEmpty
	}

	return strings.ToUpper(s), nil
}

func (stringService) Count(s string) int {
	return len(s)
}

func makeUppercaseEndpoint(service StringService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(uppercaseRequest)
		v, err := service.Uppercase(req.S)
		if err != nil {
			return uppercaseResponse{V: v, Err: err.Error()}, nil
		}

		return uppercaseResponse{V: v}, nil
	}
}

func makeCountEndpoint(service StringService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(countRequest)
		v := service.Count(req.S)
		return countResponse{V: v}, nil
	}
}
