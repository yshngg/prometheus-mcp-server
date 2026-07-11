package mock

import (
	"context"
	"time"

	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

type PrometheusAPI struct {
	QueryFunc              func(ctx context.Context, query string, ts time.Time, opts ...v1.Option) (model.Value, v1.Warnings, error)
	QueryRangeFunc         func(ctx context.Context, query string, r v1.Range, opts ...v1.Option) (model.Value, v1.Warnings, error)
	AlertsFunc             func(ctx context.Context) (v1.AlertsResult, error)
	AlertManagersFunc      func(ctx context.Context) (v1.AlertManagersResult, error)
	CleanTombstonesFunc    func(ctx context.Context) error
	ConfigFunc             func(ctx context.Context) (v1.ConfigResult, error)
	DeleteSeriesFunc       func(ctx context.Context, matches []string, startTime, endTime time.Time) error
	FlagsFunc              func(ctx context.Context) (v1.FlagsResult, error)
	LabelNamesFunc         func(ctx context.Context, matches []string, startTime, endTime time.Time, opts ...v1.Option) ([]string, v1.Warnings, error)
	LabelValuesFunc        func(ctx context.Context, label string, matches []string, startTime, endTime time.Time, opts ...v1.Option) (model.LabelValues, v1.Warnings, error)
	SeriesFunc             func(ctx context.Context, matches []string, startTime, endTime time.Time, opts ...v1.Option) ([]model.LabelSet, v1.Warnings, error)
	RulesFunc              func(ctx context.Context) (v1.RulesResult, error)
	SnapshotFunc           func(ctx context.Context, skipHead bool) (v1.SnapshotResult, error)
	TargetsFunc            func(ctx context.Context) (v1.TargetsResult, error)
	TargetsMetadataFunc    func(ctx context.Context, matchTarget, metric, limit string) ([]v1.MetricMetadata, error)
	MetadataFunc           func(ctx context.Context, metric, limit string) (map[string][]v1.Metadata, error)
	TSDBFunc               func(ctx context.Context, opts ...v1.Option) (v1.TSDBResult, error)
	WalReplayFunc          func(ctx context.Context) (v1.WalReplayStatus, error)
	BuildinfoFunc          func(ctx context.Context) (v1.BuildinfoResult, error)
	RuntimeinfoFunc        func(ctx context.Context) (v1.RuntimeinfoResult, error)
	QueryExemplarsFunc     func(ctx context.Context, query string, startTime, endTime time.Time) ([]v1.ExemplarQueryResult, error)
	HealthCheckFunc        func(ctx context.Context) error
	ReadinessCheckFunc     func(ctx context.Context) error
	ReloadFunc             func(ctx context.Context) error
	QuitFunc               func(ctx context.Context) error
}

func (m *PrometheusAPI) Alerts(ctx context.Context) (v1.AlertsResult, error) {
	if m.AlertsFunc == nil {
		return v1.AlertsResult{}, nil
	}
	return m.AlertsFunc(ctx)
}

func (m *PrometheusAPI) AlertManagers(ctx context.Context) (v1.AlertManagersResult, error) {
	if m.AlertManagersFunc == nil {
		return v1.AlertManagersResult{}, nil
	}
	return m.AlertManagersFunc(ctx)
}

func (m *PrometheusAPI) CleanTombstones(ctx context.Context) error {
	if m.CleanTombstonesFunc == nil {
		return nil
	}
	return m.CleanTombstonesFunc(ctx)
}

func (m *PrometheusAPI) Config(ctx context.Context) (v1.ConfigResult, error) {
	if m.ConfigFunc == nil {
		return v1.ConfigResult{}, nil
	}
	return m.ConfigFunc(ctx)
}

func (m *PrometheusAPI) DeleteSeries(ctx context.Context, matches []string, startTime, endTime time.Time) error {
	if m.DeleteSeriesFunc == nil {
		return nil
	}
	return m.DeleteSeriesFunc(ctx, matches, startTime, endTime)
}

func (m *PrometheusAPI) Flags(ctx context.Context) (v1.FlagsResult, error) {
	if m.FlagsFunc == nil {
		return v1.FlagsResult{}, nil
	}
	return m.FlagsFunc(ctx)
}

func (m *PrometheusAPI) LabelNames(ctx context.Context, matches []string, startTime, endTime time.Time, opts ...v1.Option) ([]string, v1.Warnings, error) {
	if m.LabelNamesFunc == nil {
		return nil, nil, nil
	}
	return m.LabelNamesFunc(ctx, matches, startTime, endTime, opts...)
}

func (m *PrometheusAPI) LabelValues(ctx context.Context, label string, matches []string, startTime, endTime time.Time, opts ...v1.Option) (model.LabelValues, v1.Warnings, error) {
	if m.LabelValuesFunc == nil {
		return nil, nil, nil
	}
	return m.LabelValuesFunc(ctx, label, matches, startTime, endTime, opts...)
}

func (m *PrometheusAPI) Query(ctx context.Context, query string, ts time.Time, opts ...v1.Option) (model.Value, v1.Warnings, error) {
	if m.QueryFunc == nil {
		return nil, nil, nil
	}
	return m.QueryFunc(ctx, query, ts, opts...)
}

func (m *PrometheusAPI) QueryRange(ctx context.Context, query string, r v1.Range, opts ...v1.Option) (model.Value, v1.Warnings, error) {
	if m.QueryRangeFunc == nil {
		return nil, nil, nil
	}
	return m.QueryRangeFunc(ctx, query, r, opts...)
}

func (m *PrometheusAPI) Series(ctx context.Context, matches []string, startTime, endTime time.Time, opts ...v1.Option) ([]model.LabelSet, v1.Warnings, error) {
	if m.SeriesFunc == nil {
		return nil, nil, nil
	}
	return m.SeriesFunc(ctx, matches, startTime, endTime, opts...)
}

func (m *PrometheusAPI) Rules(ctx context.Context) (v1.RulesResult, error) {
	if m.RulesFunc == nil {
		return v1.RulesResult{}, nil
	}
	return m.RulesFunc(ctx)
}

func (m *PrometheusAPI) Snapshot(ctx context.Context, skipHead bool) (v1.SnapshotResult, error) {
	if m.SnapshotFunc == nil {
		return v1.SnapshotResult{}, nil
	}
	return m.SnapshotFunc(ctx, skipHead)
}

func (m *PrometheusAPI) Targets(ctx context.Context) (v1.TargetsResult, error) {
	if m.TargetsFunc == nil {
		return v1.TargetsResult{}, nil
	}
	return m.TargetsFunc(ctx)
}

func (m *PrometheusAPI) TargetsMetadata(ctx context.Context, matchTarget, metric, limit string) ([]v1.MetricMetadata, error) {
	if m.TargetsMetadataFunc == nil {
		return nil, nil
	}
	return m.TargetsMetadataFunc(ctx, matchTarget, metric, limit)
}

func (m *PrometheusAPI) Metadata(ctx context.Context, metric, limit string) (map[string][]v1.Metadata, error) {
	if m.MetadataFunc == nil {
		return nil, nil
	}
	return m.MetadataFunc(ctx, metric, limit)
}

func (m *PrometheusAPI) TSDB(ctx context.Context, opts ...v1.Option) (v1.TSDBResult, error) {
	if m.TSDBFunc == nil {
		return v1.TSDBResult{}, nil
	}
	return m.TSDBFunc(ctx, opts...)
}

func (m *PrometheusAPI) WalReplay(ctx context.Context) (v1.WalReplayStatus, error) {
	if m.WalReplayFunc == nil {
		return v1.WalReplayStatus{}, nil
	}
	return m.WalReplayFunc(ctx)
}

func (m *PrometheusAPI) Buildinfo(ctx context.Context) (v1.BuildinfoResult, error) {
	if m.BuildinfoFunc == nil {
		return v1.BuildinfoResult{}, nil
	}
	return m.BuildinfoFunc(ctx)
}

func (m *PrometheusAPI) Runtimeinfo(ctx context.Context) (v1.RuntimeinfoResult, error) {
	if m.RuntimeinfoFunc == nil {
		return v1.RuntimeinfoResult{}, nil
	}
	return m.RuntimeinfoFunc(ctx)
}

func (m *PrometheusAPI) QueryExemplars(ctx context.Context, query string, startTime, endTime time.Time) ([]v1.ExemplarQueryResult, error) {
	if m.QueryExemplarsFunc == nil {
		return nil, nil
	}
	return m.QueryExemplarsFunc(ctx, query, startTime, endTime)
}

func (m *PrometheusAPI) HealthCheck(ctx context.Context) error {
	if m.HealthCheckFunc == nil {
		return nil
	}
	return m.HealthCheckFunc(ctx)
}

func (m *PrometheusAPI) ReadinessCheck(ctx context.Context) error {
	if m.ReadinessCheckFunc == nil {
		return nil
	}
	return m.ReadinessCheckFunc(ctx)
}

func (m *PrometheusAPI) Reload(ctx context.Context) error {
	if m.ReloadFunc == nil {
		return nil
	}
	return m.ReloadFunc(ctx)
}

func (m *PrometheusAPI) Quit(ctx context.Context) error {
	if m.QuitFunc == nil {
		return nil
	}
	return m.QuitFunc(ctx)
}
