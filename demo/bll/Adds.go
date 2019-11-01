package bll

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-kit/kit/endpoint"
)

type AddsRequest struct {
	A int64 `json:"a" validate:"required"`
	B int64 `json:"b" validate:"required"`
}

type AddsResponse struct {
	C int64 `json:"c" validate:"required"`
}

func (*Bei) Adds(request *AddsRequest) *AddsResponse {
	return &AddsResponse{C: request.A + request.B}
}

func MakeEndpointAdds(svc *Bei) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*AddsRequest)
		resp := svc.Adds(req)
		return resp, nil
	}
}

func DecodeRequestAdds(ctx context.Context, r *http.Request) (interface{}, error) {
	request := new(AddsRequest)
	if err := json.NewDecoder(r.Body).Decode(request); err != nil {
		return nil, err
	}
	return request, nil
}

func EncodeResponseAdds(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	return json.NewEncoder(w).Encode(response)

}
