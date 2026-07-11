package bindingblocks

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

// addTools registers Prometheus query tools with the MCP server.
// It adds tools for expression queries (instant and range) and metadata queries.
func (b *binder) addTools() {
	// Expression queries
	// Query language expressions may be evaluated at a single instant or over a range of time.
	{
		expressionQuerier := expressionquery.NewExpressionQuerier(b.api)
		mcp.AddTool(b.server, &mcp.Tool{
			Name:         "instant-query",
			Description:  "Evaluate an instant query at a single point in time.",
			OutputSchema: promqlSchema,
			Annotations: &mcp.ToolAnnotations{
				Title:        "Instant Query",
				ReadOnlyHint: true,
			},
		}, expressionQuerier.InstantQueryHandler)

		mcp.AddTool(b.server, &mcp.Tool{
			Name:         "range-query",
			Description:  "Evaluate an expression query over a range of time.",
			OutputSchema: promqlSchema,
			Annotations: &mcp.ToolAnnotations{
				Title:        "Range Query",
				ReadOnlyHint: true,
			},
		}, expressionQuerier.RangeQueryHandler)
	}

	// Querying metadata
	// Prometheus offers a set of API endpoints to query metadata about series and their labels.
	{
		metadataQuerier := metadataquery.NewMetadataQuerier(b.api)
		mcp.AddTool(b.server, &mcp.Tool{
			Name:        "find-series-by-labels",
			Description: "Return the list of time series that match a certain label set.",
			Annotations: &mcp.ToolAnnotations{
				Title:        "Find Series by Labels",
				ReadOnlyHint: true,
			},
		}, metadataQuerier.SeriesHandler)

		mcp.AddTool(b.server, &mcp.Tool{
			Name:        "list-label-names",
			Description: "Return a list of label names.",
			Annotations: &mcp.ToolAnnotations{
				Title:        "List Label Names",
				ReadOnlyHint: true,
			},
		}, metadataQuerier.LabelNamesHandler)

		mcp.AddTool(b.server, &mcp.Tool{
			Name:        "list-label-values",
			Description: "Return a list of label values for a provided label name.",
			Annotations: &mcp.ToolAnnotations{
				Title:        "List Label Values",
				ReadOnlyHint: true,
			},
		}, metadataQuerier.LabelValuesHandler)

		mcp.AddTool(b.server, &mcp.Tool{
			Name:        "target-metadata-query",
			Description: "Return metadata about metrics currently scraped from targets.",
			Annotations: &mcp.ToolAnnotations{
				Title:        "Target Metadata Query",
				ReadOnlyHint: true,
			},
		}, metadataQuerier.TargetMetadataQueryHandler)

		mcp.AddTool(b.server, &mcp.Tool{
			Name:        "metric-metadata-query",
			Description: "Return metadata about metrics currently scraped from targets. However, it does not provide any target information.",
			Annotations: &mcp.ToolAnnotations{
				Title:        "Metric Metadata Query",
				ReadOnlyHint: true,
			},
		}, metadataQuerier.MetricsMetadataQueryHandler)
	}

	// Targets
	// An overview of the current state of the Prometheus target discovery.
	{
		targetDiscoverer := targetdiscover.NewTargetDiscoverer(b.api)
		mcp.AddTool(b.server, &mcp.Tool{
			Name:        "target-discovery",
			Description: "Return an overview of the current state of the Prometheus target discovery.",
			Annotations: &mcp.ToolAnnotations{
				Title:        "Target Discovery",
				ReadOnlyHint: true,
			},
		}, targetDiscoverer.TargetDiscoverHandler)
	}

	// Rules
	// A list of alerting and recording rules that are currently loaded.
	{
		ruleQuerier := rulequery.NewRuleQuerier(b.api)
		mcp.AddTool(b.server, &mcp.Tool{
			Name:        "rule-query",
			Description: "Return a list of alerting and recording rules that are currently loaded. In addition it returns the currently active alerts fired by the Prometheus instance of each alerting rule.",
			Annotations: &mcp.ToolAnnotations{
				Title:        "Rule Query",
				ReadOnlyHint: true,
			},
		}, ruleQuerier.RuleQueryHandler)
	}

	// Alerts
	// A list of all active alerts.
	{
		alertQuerier := alertquery.NewAlertQuerier(b.api)
		mcp.AddTool(b.server, &mcp.Tool{
			Name:        "alert-query",
			Description: "Return a list of all active alerts.",
			Annotations: &mcp.ToolAnnotations{
				Title:        "Alert Query",
				ReadOnlyHint: true,
			},
		}, alertQuerier.AlertQueryHandler)
	}

	// Alertmanagers
	// An overview of the current state of the Prometheus alertmanager discovery.
	{
		alertmanagerDiscoverer := alertmanagerdiscover.NewAlertmanagerDiscoverer(b.api)
		mcp.AddTool(b.server, &mcp.Tool{
			Name:        "alertmanager-discovery",
			Description: "Return an overview of the current state of the Prometheus alertmanager discovery.",
			Annotations: &mcp.ToolAnnotations{
				Title:        "Alertmanager Discovery",
				ReadOnlyHint: true,
			},
		}, alertmanagerDiscoverer.AlertmanagerDiscoverHandler)
	}


	// TSDB Admin APIs
	// Expose database functionalities for the advanced user.
	{
		tsdbAdmin := tsdbadmin.NewTSDBAdmin(b.api)
		mcp.AddTool(b.server, &mcp.Tool{
			Name:        "tsdb-snapshot",
			Description: "Create a snapshot of all current data into snapshots/<datetime>-<rand> under the TSDB's data directory and returns the directory as response. It will optionally skip snapshotting data that is only present in the head block, and which has not yet been compacted to disk.",
			Annotations: &mcp.ToolAnnotations{
				Title:           "TSDB Snapshot",
				DestructiveHint: ptrBool(true),
			},
		}, tsdbAdmin.SnapshotHandler)

		mcp.AddTool(b.server, &mcp.Tool{
			Name:        "delete-series",
			Description: "Delete data for a selection of series in a time range. The actual data still exists on disk and is cleaned up in future compactions or can be explicitly cleaned up by hitting the Clean Tombstones endpoint. Not mentioning both start and end times would clear all the data for the matched series in the database.",
			Annotations: &mcp.ToolAnnotations{
				Title:           "Delete Series",
				DestructiveHint: ptrBool(true),
				ReadOnlyHint:    false,
			},
		}, tsdbAdmin.DeleteSeriesHandler)

		mcp.AddTool(b.server, &mcp.Tool{
			Name:        "clean-tombstones",
			Description: "Remove the deleted data from disk and cleans up the existing tombstones. This can be used after deleting series to free up space.",
			Annotations: &mcp.ToolAnnotations{
				Title:           "Clean Tombstones",
				DestructiveHint: ptrBool(true),
				ReadOnlyHint:    false,
			},
		}, tsdbAdmin.CleanTombstonesHandler)
	}

	// Management API
	// Prometheus provides a set of management APIs to facilitate automation and integration.
	{
		manager := manage.NewManager(b.api)
		mcp.AddTool(b.server, &mcp.Tool{
			Name:        "health-check",
			Description: "Check Prometheus health.",
			Annotations: &mcp.ToolAnnotations{
				Title:           "Health Check",
				ReadOnlyHint:    true,
				IdempotentHint:  true,
			},
		}, manager.HealthCheckHandler)

		mcp.AddTool(b.server, &mcp.Tool{
			Name:        "readiness-check",
			Description: "Check if Prometheus is ready to serve traffic (i.e. respond to queries).",
			Annotations: &mcp.ToolAnnotations{
				Title:           "Readiness Check",
				ReadOnlyHint:    true,
				IdempotentHint:  true,
			},
		}, manager.ReadinessCheckHandler)

		mcp.AddTool(b.server, &mcp.Tool{
			Name:        "reload",
			Description: "Trigger a reload of the Prometheus configuration and rule files.",
			Annotations: &mcp.ToolAnnotations{
				Title:           "Reload",
				DestructiveHint: ptrBool(true),
				ReadOnlyHint:    false,
			},
		}, manager.ReloadHandler)

		mcp.AddTool(b.server, &mcp.Tool{
			Name:        "quit",
			Description: "Trigger a graceful shutdown of Prometheus.",
			Annotations: &mcp.ToolAnnotations{
				Title:           "Quit",
				DestructiveHint: ptrBool(true),
				ReadOnlyHint:    false,
			},
		}, manager.QuitHandler)
	}
}
