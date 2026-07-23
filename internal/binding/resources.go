package binding

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func (b *binder) addResources() {
	b.addStaticResources()
	b.addResourceTemplates()
}

func (b *binder) addStaticResources() {
	type staticResource struct {
		uri         string
		name        string
		title       string
		description string
		mimeType    string
		fetch       func(ctx context.Context) (string, error)
	}

	resources := []staticResource{
		{
			uri:         "prom:///config",
			name:        "config",
			title:       "Prometheus Configuration",
			description: "Currently loaded Prometheus configuration file.",
			mimeType:    "application/yaml",
			fetch: func(ctx context.Context) (string, error) {
				result, err := b.api.Config(ctx)
				if err != nil {
					return "", err
				}
				return result.YAML, nil
			},
		},
		{
			uri:         "prom:///flags",
			name:        "flags",
			title:       "Prometheus Flags",
			description: "Flag values that Prometheus was configured with.",
			mimeType:    "application/json",
			fetch: func(ctx context.Context) (string, error) {
				result, err := b.api.Flags(ctx)
				if err != nil {
					return "", err
				}
			return marshalJSON("marshal response", result)
		},
		},
		{
			uri:         "prom:///runtime-info",
			name:        "runtime-info",
			title:       "Prometheus Runtime Information",
			description: "Runtime information properties about the Prometheus server.",
			mimeType:    "application/json",
			fetch: func(ctx context.Context) (string, error) {
				result, err := b.api.Runtimeinfo(ctx)
				if err != nil {
					return "", err
				}
			return marshalJSON("marshal response", result)
		},
		},
		{
			uri:         "prom:///build-info",
			name:        "build-info",
			title:       "Prometheus Build Information",
			description: "Build information properties about the Prometheus server.",
			mimeType:    "application/json",
			fetch: func(ctx context.Context) (string, error) {
				result, err := b.api.Buildinfo(ctx)
				if err != nil {
					return "", err
				}
			return marshalJSON("marshal response", result)
		},
		},
		{
			uri:         "prom:///tsdb-stats",
			name:        "tsdb-stats",
			title:       "TSDB Statistics",
			description: "Cardinality statistics about the Prometheus TSDB.",
			mimeType:    "application/json",
			fetch: func(ctx context.Context) (string, error) {
				result, err := b.api.TSDB(ctx)
				if err != nil {
					return "", err
				}
			return marshalJSON("marshal response", result)
		},
		},
		{
			uri:         "prom:///tsdb-blocks",
			name:        "tsdb-blocks",
			title:       "TSDB Blocks",
			description: "Currently loaded TSDB blocks and their metadata.",
			mimeType:    "application/json",
			fetch: func(ctx context.Context) (string, error) {
				result, err := b.api.TSDBBlocks(ctx)
				if err != nil {
					return "", err
				}
			return marshalJSON("marshal response", result)
		},
		},
		{
			uri:         "prom:///wal-replay-stats",
			name:        "wal-replay-stats",
			title:       "WAL Replay Statistics",
			description: "Information about the WAL replay state.",
			mimeType:    "application/json",
			fetch: func(ctx context.Context) (string, error) {
				result, err := b.api.WalReplay(ctx)
				if err != nil {
					return "", err
				}
			return marshalJSON("marshal response", result)
		},
		},
	}

	for _, r := range resources {
		uri, name, title, desc, mime, fetch := r.uri, r.name, r.title, r.description, r.mimeType, r.fetch
		b.server.AddResource(&mcp.Resource{
			URI:         uri,
			Name:        name,
			Title:       title,
			Description: desc,
			MIMEType:    mime,
		}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
			text, err := fetch(ctx)
			if err != nil {
				return nil, fmt.Errorf("read resource %s: %w", req.Params.URI, err)
			}
			return &mcp.ReadResourceResult{
				Contents: []*mcp.ResourceContents{
					{
						URI:      req.Params.URI,
						MIMEType: mime,
						Text:     text,
					},
				},
			}, nil
		})
	}
}

func (b *binder) addResourceTemplates() {
	b.server.AddResourceTemplate(&mcp.ResourceTemplate{
		URITemplate: "prom:///api/v1/format_query?query={promql}",
		Name:        "format_query",
		Title:       "Prometheus Format Query",
		Description: "Result of formatting a PromQL expression. Replace {promql} with a URL-encoded PromQL expression.",
		MIMEType:    "text/plain",
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		uri := req.Params.URI
		parsed, err := url.Parse(uri)
		if err != nil {
			return nil, fmt.Errorf("invalid resource URI: %w", err)
		}
		promql := parsed.Query().Get("query")
		if promql == "" {
			return nil, fmt.Errorf("missing query parameter in resource URI")
		}

		result, err := b.api.FormatQuery(ctx, promql)
		if err != nil {
			return nil, fmt.Errorf("format query %q: %w", promql, err)
		}

		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{
					URI:      uri,
					MIMEType: "text/plain",
					Text:     result,
				},
			},
		}, nil
	})

	b.server.AddResourceTemplate(&mcp.ResourceTemplate{
		URITemplate: "prom:///api/v1/query?query={promql}",
		Name:        "instant_query",
		Title:       "Prometheus Instant Query",
		Description: "Result of an instant Prometheus query. Replace {promql} with a URL-encoded PromQL expression. Use prometheus time format for timestamps.",
		MIMEType:    "application/json",
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		uri := req.Params.URI
		parsed, err := url.Parse(uri)
		if err != nil {
			return nil, fmt.Errorf("invalid resource URI: %w", err)
		}
		promql := parsed.Query().Get("query")
		if promql == "" {
			return nil, fmt.Errorf("missing query parameter in resource URI")
		}

		value, warnings, err := b.api.Query(ctx, promql, time.Time{})
		if err != nil {
			return nil, fmt.Errorf("query %q: %w", promql, err)
		}

		data, err := json.MarshalIndent(map[string]interface{}{
			"status":   "success",
			"data":     value,
			"warnings": warnings,
		}, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("marshal query result: %w", err)
		}
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{
					URI:      uri,
					MIMEType: "application/json",
					Text:     string(data),
				},
			},
		}, nil
	})

	b.server.AddResourceTemplate(&mcp.ResourceTemplate{
		URITemplate: "prom:///api/v1/label/{name}/values",
		Name:        "label-values",
		Title:       "Prometheus Label Values",
		Description: "List of values for a given label name. Replace {name} with a label name such as __name__ or job.",
		MIMEType:    "application/json",
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		uri := req.Params.URI
		labelName := strings.TrimSuffix(strings.TrimPrefix(uri, "prom:///api/v1/label/"), "/values")
		if labelName == "" {
			return nil, fmt.Errorf("invalid resource URI: %s", uri)
		}

		values, warnings, err := b.api.LabelValues(ctx, labelName, nil, time.Time{}, time.Time{})
		if err != nil {
			return nil, fmt.Errorf("label values for %q: %w", labelName, err)
		}

		result := make([]string, len(values))
		for i, v := range values {
			result[i] = string(v)
		}
		data, err := json.MarshalIndent(map[string]interface{}{
			"status":    "success",
			"data":      result,
			"warnings":  warnings,
			"labelName": labelName,
		}, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("marshal label values: %w", err)
		}

		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{
					URI:      uri,
					MIMEType: "application/json",
					Text:     string(data),
				},
			},
		}, nil
	})
}

// marshalJSON calls json.MarshalIndent with standard formatting and wraps
// errors with a context prefix. The underlying types are always valid for
// JSON serialization, so this function only exists to avoid repeating the
// boilerplate at each call site.
func marshalJSON(contextPrefix string, v any) (string, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", fmt.Errorf("%s: %w", contextPrefix, err)
	}
	return string(data), nil
}

