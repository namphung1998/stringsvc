package main

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-kit/kit/endpoint"
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
