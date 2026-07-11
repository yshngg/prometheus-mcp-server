package promapi

import v1 "github.com/prometheus/client_golang/api/prometheus/v1"

type QueryingAPI interface {
	v1.API
}
