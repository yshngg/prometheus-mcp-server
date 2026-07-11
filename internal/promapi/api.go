package promapi

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

const APIVersion = "/api/v1"

type PrometheusAPI interface {
	QueryingAPI
	ManagementAPI
}

type prometheusAPI struct {
	QueryingAPI
	ManagementAPI
}

func New(addr string, client *http.Client, roundTripper http.RoundTripper) (PrometheusAPI, error) {
	cli, err := api.NewClient(api.Config{
		Address:      addr,
		Client:       client,
		RoundTripper: roundTripper,
	})
	if err != nil {
		return nil, fmt.Errorf("new client, err: %w", err)
	}

	return &prometheusAPI{
		QueryingAPI:   v1.NewAPI(cli),
		ManagementAPI: NewManagementAPI(cli),
	}, nil
}
