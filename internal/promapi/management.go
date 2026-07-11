package promapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/api"
)

const (
	epHealthCheck    = "/-/healthy"
	epReadinessCheck = "/-/ready"
	epReload         = "/-/reload"
	epQuit           = "/-/quit"
)

type ManagementAPI interface {
	HealthCheck(ctx context.Context) error
	ReadinessCheck(ctx context.Context) error
	Reload(ctx context.Context) error
	Quit(ctx context.Context) error
}

type managementAPI struct {
	cli api.Client
}

func NewManagementAPI(cli api.Client) ManagementAPI {
	return &managementAPI{cli: cli}
}

// This endpoint always returns 200 and should be used to check Prometheus health.
// GET /-/healthy
// HEAD /-/healthy
func (m *managementAPI) HealthCheck(ctx context.Context) error {
	resp, body, err := m.cli.Do(ctx, &http.Request{
		URL: m.cli.URL(epHealthCheck, nil),
	})
	if err != nil {
		return fmt.Errorf("health check, err: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New(string(body))
	}
	return nil
}

// This endpoint returns 200 when Prometheus is ready to serve traffic (i.e. respond to queries).
// GET /-/ready
// HEAD /-/ready
func (m *managementAPI) ReadinessCheck(ctx context.Context) error {
	resp, body, err := m.cli.Do(ctx, &http.Request{
		URL: m.cli.URL(epReadinessCheck, nil),
	})
	if err != nil {
		return fmt.Errorf("readiness check, err: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New(string(body))
	}
	return nil
}

// This endpoint triggers a reload of the Prometheus configuration and rule files. It's disabled by default and can be enabled via the --web.enable-lifecycle flag.
// Alternatively, a configuration reload can be triggered by sending a SIGHUP to the Prometheus process.
// PUT  /-/reload
// POST /-/reload
func (m *managementAPI) Reload(ctx context.Context) error {
	resp, body, err := m.cli.Do(ctx, &http.Request{
		Method: http.MethodPut,
		URL:    m.cli.URL(epReload, nil),
	})
	if err != nil {
		return fmt.Errorf("reload configuration and rule files, err: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New(string(body))
	}
	return nil
}

// This endpoint triggers a graceful shutdown of Prometheus. It's disabled by default and can be enabled via the --web.enable-lifecycle flag.
// Alternatively, a graceful shutdown can be triggered by sending a SIGTERM to the Prometheus process.
// PUT  /-/quit
// POST /-/quit
func (m *managementAPI) Quit(ctx context.Context) error {
	resp, body, err := m.cli.Do(ctx, &http.Request{
		Method: http.MethodPut,
		URL:    m.cli.URL(epQuit, nil),
	})
	if err != nil {
		return fmt.Errorf("graceful shutdown, err: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New(string(body))
	}
	return nil
}
