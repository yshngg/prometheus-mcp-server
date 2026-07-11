package binding

import (
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yshngg/prometheus-mcp-server/internal/alertmanagerdiscover"
	"github.com/yshngg/prometheus-mcp-server/internal/alertquery"
	"github.com/yshngg/prometheus-mcp-server/internal/expressionquery"
	"github.com/yshngg/prometheus-mcp-server/internal/manage"
	"github.com/yshngg/prometheus-mcp-server/internal/metadataquery"
	"github.com/yshngg/prometheus-mcp-server/internal/rulequery"
	"github.com/yshngg/prometheus-mcp-server/internal/targetdiscover"
	"github.com/yshngg/prometheus-mcp-server/internal/tsdbadmin"
)

var promqlSchema = json.RawMessage(`{
	"type": "object",
	"properties": {
		"value": {
			"oneOf": [
				{"type": "object", "description": "Scalar or string: {value: number|string, timestamp: number}"},
				{"type": "array", "description": "Vector or matrix: [{metric: object, value: [number,string] | values: [[number,string]]}]"}
			]
		},
		"warnings": {"type": "array", "items": {"type": "string"}}
	}
}`)

func ptrBool(v bool) *bool {
	return &v
}

func readOnlyTool(name, desc, title string) *mcp.Tool {
	return &mcp.Tool{
		Name:        name,
		Description: desc,
		Annotations: &mcp.ToolAnnotations{Title: title, ReadOnlyHint: true},
	}
}

func destructiveTool(name, desc, title string) *mcp.Tool {
	return &mcp.Tool{
		Name:        name,
		Description: desc,
		Annotations: &mcp.ToolAnnotations{
			Title:           title,
			DestructiveHint: ptrBool(true),
			ReadOnlyHint:    false,
		},
	}
}

func idempotentTool(name, desc, title string) *mcp.Tool {
	return &mcp.Tool{
		Name:        name,
		Description: desc,
		Annotations: &mcp.ToolAnnotations{
			Title:          title,
			ReadOnlyHint:   true,
			IdempotentHint: true,
		},
	}
}

func withOutputSchema(t *mcp.Tool, s any) *mcp.Tool {
	t.OutputSchema = s
	return t
}

// addTools makes Prometheus capabilities discoverable by LLMs as MCP tool calls.
func (b *binder) addTools() {
	// Expression queries accept PromQL which can be evaluated at a point in time or
	// over a range. The output schema avoids model.Value (interface) and uses
	// oneOf for scalar, vector, and matrix shapes so LLMs can parse results.
	{
		expressionQuerier := expressionquery.NewExpressionQuerier(b.api)
		mcp.AddTool(b.server, withOutputSchema(readOnlyTool(
			"instant-query", "Evaluate an instant query at a single point in time.", "Instant Query"),
		promqlSchema), expressionQuerier.InstantQueryHandler)

		mcp.AddTool(b.server, withOutputSchema(readOnlyTool(
			"range-query", "Evaluate an expression query over a range of time.", "Range Query"),
		promqlSchema), expressionQuerier.RangeQueryHandler)
	}

	// LLMs need label and series metadata to construct accurate PromQL expressions.
	// These tools expose the label namespace and target metadata without executing
	// full queries, which helps the LLM discover metric names and label values.
	{
		metadataQuerier := metadataquery.NewMetadataQuerier(b.api)
		mcp.AddTool(b.server, readOnlyTool("find-series-by-labels",
			"Return the list of time series that match a certain label set.", "Find Series by Labels",
		), metadataQuerier.SeriesHandler)

		mcp.AddTool(b.server, readOnlyTool("list-label-names",
			"Return a list of label names.", "List Label Names",
		), metadataQuerier.LabelNamesHandler)

		mcp.AddTool(b.server, readOnlyTool("list-label-values",
			"Return a list of label values for a provided label name.", "List Label Values",
		), metadataQuerier.LabelValuesHandler)

		mcp.AddTool(b.server, readOnlyTool("target-metadata-query",
			"Return metadata about metrics currently scraped from targets.", "Target Metadata Query",
		), metadataQuerier.TargetMetadataQueryHandler)

		mcp.AddTool(b.server, readOnlyTool("metric-metadata-query",
			"Return metadata about metrics currently scraped from targets. However, it does not provide any target information.", "Metric Metadata Query",
		), metadataQuerier.MetricsMetadataQueryHandler)
	}

	// Target discovery displays scrape pool states (active/dropped) so LLMs can
	// diagnose missing metrics without parsing raw Prometheus API responses.
	{
		targetDiscoverer := targetdiscover.NewTargetDiscoverer(b.api)
		mcp.AddTool(b.server, readOnlyTool("target-discovery",
			"Return an overview of the current state of the Prometheus target discovery.", "Target Discovery",
		), targetDiscoverer.TargetDiscoverHandler)
	}

	// Rule introspection helps LLMs understand what alerting and recording rules
	// are active, enabling them to explain firing alerts and suggest threshold changes.
	{
		ruleQuerier := rulequery.NewRuleQuerier(b.api)
		mcp.AddTool(b.server, readOnlyTool("rule-query",
			"Return a list of alerting and recording rules that are currently loaded. In addition it returns the currently active alerts fired by the Prometheus instance of each alerting rule.",
			"Rule Query",
		), ruleQuerier.RuleQueryHandler)
	}

	// Current alerts give LLMs a snapshot of what is firing, which is useful for
	// incident response and root-cause analysis alongside rule definitions.
	{
		alertQuerier := alertquery.NewAlertQuerier(b.api)
		mcp.AddTool(b.server, readOnlyTool("alert-query",
			"Return a list of all active alerts.", "Alert Query",
		), alertQuerier.AlertQueryHandler)
	}

	// Alertmanager discovery reveals which alertmanagers are reachable so LLMs can
	// diagnose notification delivery issues.
	{
		alertmanagerDiscoverer := alertmanagerdiscover.NewAlertmanagerDiscoverer(b.api)
		mcp.AddTool(b.server, readOnlyTool("alertmanager-discovery",
			"Return an overview of the current state of the Prometheus alertmanager discovery.", "Alertmanager Discovery",
		), alertmanagerDiscoverer.AlertmanagerDiscoverHandler)
	}


	// TSDB write operations cannot be undone. The destructive hint annotation
	// triggers the elicitation middleware to confirm before executing.
	{
		tsdbAdmin := tsdbadmin.NewTSDBAdmin(b.api)
		mcp.AddTool(b.server, destructiveTool("tsdb-snapshot",
			"Create a snapshot of all current data into snapshots/<datetime>-<rand> under the TSDB's data directory and returns the directory as response. It will optionally skip snapshotting data that is only present in the head block, and which has not yet been compacted to disk.",
			"TSDB Snapshot",
		), tsdbAdmin.SnapshotHandler)

		mcp.AddTool(b.server, destructiveTool("delete-series",
			"Delete data for a selection of series in a time range. The actual data still exists on disk and is cleaned up in future compactions or can be explicitly cleaned up by hitting the Clean Tombstones endpoint. Not mentioning both start and end times would clear all the data for the matched series in the database.",
			"Delete Series",
		), tsdbAdmin.DeleteSeriesHandler)

		mcp.AddTool(b.server, destructiveTool("clean-tombstones",
			"Remove the deleted data from disk and cleans up the existing tombstones. This can be used after deleting series to free up space.",
			"Clean Tombstones",
		), tsdbAdmin.CleanTombstonesHandler)
	}

	// Reload and quit affect the running Prometheus process. The destructive hint
	// annotation triggers the elicitation middleware so users confirm before these
	// operations execute. Health checks are read-only and safe for repeated calls.
	{
		manager := manage.NewManager(b.api)
		mcp.AddTool(b.server, idempotentTool("health-check",
			"Check Prometheus health.", "Health Check",
		), manager.HealthCheckHandler)

		mcp.AddTool(b.server, idempotentTool("readiness-check",
			"Check if Prometheus is ready to serve traffic (i.e. respond to queries).", "Readiness Check",
		), manager.ReadinessCheckHandler)

		mcp.AddTool(b.server, destructiveTool("reload",
			"Trigger a reload of the Prometheus configuration and rule files.", "Reload",
		), manager.ReloadHandler)

		mcp.AddTool(b.server, destructiveTool("quit",
			"Trigger a graceful shutdown of Prometheus.", "Quit",
		), manager.QuitHandler)
	}
}
