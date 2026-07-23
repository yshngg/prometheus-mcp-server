package promapi

import (
	"context"
	"time"

	"github.com/yshngg/prometheus-mcp-server/internal/cache"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

// CachingPrometheusAPI wraps a PrometheusAPI with an in-memory TTL cache for
// selected query methods. Uncacheable calls (with match/start/end filters) pass
// through to the inner implementation.
type CachingPrometheusAPI struct {
	PrometheusAPI
	cache *cache.Cache
}

// NewCachingAPI wraps a PrometheusAPI with an in-memory TTL cache for
// LabelNames and LabelValues queries without match/start/end filters.
// Other methods pass through to the inner implementation unchanged.
func NewCachingAPI(inner PrometheusAPI, ttl time.Duration) PrometheusAPI {
	return &CachingPrometheusAPI{
		PrometheusAPI: inner,
		cache:         cache.New(ttl),
	}
}

func (c *CachingPrometheusAPI) LabelNames(ctx context.Context, matches []string, startTime, endTime time.Time, opts ...v1.Option) (model.LabelNames, v1.Warnings, error) {
	key := "labelnames"
	if len(matches) > 0 || !startTime.IsZero() || !endTime.IsZero() {
		return c.PrometheusAPI.LabelNames(ctx, matches, startTime, endTime, opts...)
	}
	if v, ok := c.cache.Get(key); ok {
		return v.(model.LabelNames), nil, nil
	}
	names, warnings, err := c.PrometheusAPI.LabelNames(ctx, matches, startTime, endTime, opts...)
	if err != nil {
		return nil, warnings, err
	}
	c.cache.Set(key, names)
	return names, warnings, nil
}

func (c *CachingPrometheusAPI) LabelValues(ctx context.Context, label string, matches []string, startTime, endTime time.Time, opts ...v1.Option) (model.LabelValues, v1.Warnings, error) {
	key := "labelvalues:" + label
	if len(matches) > 0 || !startTime.IsZero() || !endTime.IsZero() {
		return c.PrometheusAPI.LabelValues(ctx, label, matches, startTime, endTime, opts...)
	}
	if v, ok := c.cache.Get(key); ok {
		return v.(model.LabelValues), nil, nil
	}
	values, warnings, err := c.PrometheusAPI.LabelValues(ctx, label, matches, startTime, endTime, opts...)
	if err != nil {
		return nil, warnings, err
	}
	c.cache.Set(key, values)
	return values, warnings, nil
}

// Query is uncacheable — pass through.
func (c *CachingPrometheusAPI) Query(ctx context.Context, query string, ts time.Time, opts ...v1.Option) (model.Value, v1.Warnings, error) {
	return c.PrometheusAPI.Query(ctx, query, ts, opts...)
}

// QueryRange is uncacheable — pass through.
func (c *CachingPrometheusAPI) QueryRange(ctx context.Context, query string, r v1.Range, opts ...v1.Option) (model.Value, v1.Warnings, error) {
	return c.PrometheusAPI.QueryRange(ctx, query, r, opts...)
}

// Series is uncacheable — pass through.
func (c *CachingPrometheusAPI) Series(ctx context.Context, matches []string, startTime, endTime time.Time, opts ...v1.Option) ([]model.LabelSet, v1.Warnings, error) {
	return c.PrometheusAPI.Series(ctx, matches, startTime, endTime, opts...)
}

// QueryExemplars is uncacheable — pass through.
func (c *CachingPrometheusAPI) QueryExemplars(ctx context.Context, query string, startTime, endTime time.Time) ([]v1.ExemplarQueryResult, error) {
	return c.PrometheusAPI.QueryExemplars(ctx, query, startTime, endTime)
}

// Alerts is uncacheable — pass through.
func (c *CachingPrometheusAPI) Alerts(ctx context.Context) (v1.AlertsResult, error) {
	return c.PrometheusAPI.Alerts(ctx)
}

// AlertManagers is uncacheable — pass through.
func (c *CachingPrometheusAPI) AlertManagers(ctx context.Context) (v1.AlertManagersResult, error) {
	return c.PrometheusAPI.AlertManagers(ctx)
}

// Rules is uncacheable — pass through.
func (c *CachingPrometheusAPI) Rules(ctx context.Context, matches []string) (v1.RulesResult, error) {
	return c.PrometheusAPI.Rules(ctx, matches)
}

// Targets is uncacheable — pass through.
func (c *CachingPrometheusAPI) Targets(ctx context.Context) (v1.TargetsResult, error) {
	return c.PrometheusAPI.Targets(ctx)
}

// TargetsMetadata is uncacheable — pass through.
func (c *CachingPrometheusAPI) TargetsMetadata(ctx context.Context, matchTarget, metric, limit string) ([]v1.MetricMetadata, error) {
	return c.PrometheusAPI.TargetsMetadata(ctx, matchTarget, metric, limit)
}

// Metadata is uncacheable — pass through.
func (c *CachingPrometheusAPI) Metadata(ctx context.Context, metric, limit string) (map[string][]v1.Metadata, error) {
	return c.PrometheusAPI.Metadata(ctx, metric, limit)
}

// TSDB is uncacheable — pass through.
func (c *CachingPrometheusAPI) TSDB(ctx context.Context, opts ...v1.Option) (v1.TSDBResult, error) {
	return c.PrometheusAPI.TSDB(ctx, opts...)
}

// TSDBBlocks is uncacheable — pass through.
func (c *CachingPrometheusAPI) TSDBBlocks(ctx context.Context) (v1.TSDBBlocksResult, error) {
	return c.PrometheusAPI.TSDBBlocks(ctx)
}

// WalReplay is uncacheable — pass through.
func (c *CachingPrometheusAPI) WalReplay(ctx context.Context) (v1.WalReplayStatus, error) {
	return c.PrometheusAPI.WalReplay(ctx)
}

// Buildinfo is uncacheable — pass through.
func (c *CachingPrometheusAPI) Buildinfo(ctx context.Context) (v1.BuildinfoResult, error) {
	return c.PrometheusAPI.Buildinfo(ctx)
}

// Runtimeinfo is uncacheable — pass through.
func (c *CachingPrometheusAPI) Runtimeinfo(ctx context.Context) (v1.RuntimeinfoResult, error) {
	return c.PrometheusAPI.Runtimeinfo(ctx)
}

func (c *CachingPrometheusAPI) Snapshot(ctx context.Context, skipHead bool) (v1.SnapshotResult, error) {
	return c.PrometheusAPI.Snapshot(ctx, skipHead)
}

func (c *CachingPrometheusAPI) FormatQuery(ctx context.Context, query string) (string, error) {
	return c.PrometheusAPI.FormatQuery(ctx, query)
}

func (c *CachingPrometheusAPI) CleanTombstones(ctx context.Context) error {
	return c.PrometheusAPI.CleanTombstones(ctx)
}

func (c *CachingPrometheusAPI) DeleteSeries(ctx context.Context, matches []string, startTime, endTime time.Time) error {
	return c.PrometheusAPI.DeleteSeries(ctx, matches, startTime, endTime)
}

func (c *CachingPrometheusAPI) Config(ctx context.Context) (v1.ConfigResult, error) {
	return c.PrometheusAPI.Config(ctx)
}

func (c *CachingPrometheusAPI) Flags(ctx context.Context) (v1.FlagsResult, error) {
	return c.PrometheusAPI.Flags(ctx)
}

func (c *CachingPrometheusAPI) HealthCheck(ctx context.Context) error {
	return c.PrometheusAPI.HealthCheck(ctx)
}

func (c *CachingPrometheusAPI) ReadinessCheck(ctx context.Context) error {
	return c.PrometheusAPI.ReadinessCheck(ctx)
}

func (c *CachingPrometheusAPI) Reload(ctx context.Context) error {
	return c.PrometheusAPI.Reload(ctx)
}

func (c *CachingPrometheusAPI) Quit(ctx context.Context) error {
	return c.PrometheusAPI.Quit(ctx)
}

var (
	_ v1.API         = &CachingPrometheusAPI{}
	_ ManagementAPI  = &CachingPrometheusAPI{}
)
