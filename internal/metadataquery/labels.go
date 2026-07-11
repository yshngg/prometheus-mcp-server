package metadataquery

import (
	"context"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/yshngg/prometheus-mcp-server/internal/timeutil"
	"k8s.io/klog/v2"
)

type LabelNamesArguments struct {
	Start string   `json:"start,omitzero" jsonschema:"<rfc3339 | unix_timestamp>: Start timestamp. Optional."`
	End   string   `json:"end,omitzero" jsonschema:"<rfc3339 | unix_timestamp>: End timestamp. Optional."`
	Match []string `json:"match[],omitzero" jsonschema:"<series_selector>: Repeated series selector argument that selects the series from which to read the label names. Optional."`
	Limit uint64   `json:"limit,omitzero" jsonschema:"<number>: Maximum number of returned series. Optional. 0 means disabled."`
}

type LabelNamesResult struct {
	LabelNames model.LabelNames `json:"names" jsonschema:"Names is a list of string label names."`
	Warnings   v1.Warnings      `json:"warnings,omitempty"`
}

func (q *metadataQuerier) LabelNamesHandler(ctx context.Context, request *mcp.CallToolRequest, input *LabelNamesArguments) (*mcp.CallToolResult, *LabelNamesResult, error) {
	if len(input.Match) == 0 && input.Start == "" && input.End == "" {
		if v, ok := q.cache.Get("labelnames"); ok {
			result := v.(LabelNamesResult)
			return nil, &result, nil
		}
	}

	var (
		start, end time.Time
		err        error
	)
	if input.Start != "" {
		if start, err = timeutil.ParseTime(input.Start); err != nil {
			klog.InfoS("parse start time", "err", err)
		}
	}
	if input.End != "" {
		if end, err = timeutil.ParseTime(input.End); err != nil {
			klog.InfoS("parse end time", "err", err)
		}
	}

	opts := make([]v1.Option, 0)
	if input.Limit != 0 {
		opts = append(opts, v1.WithLimit(input.Limit))
	}

	result := &LabelNamesResult{}
	var names []string
	if names, result.Warnings, err = q.API.LabelNames(
		ctx,
		input.Match,
		start,
		end,
		opts...,
	); err != nil {
		return nil, nil, err
	}
	result.LabelNames = make(model.LabelNames, len(names))
	for i, n := range names {
		result.LabelNames[i] = model.LabelName(n)
	}

	if len(input.Match) == 0 && input.Start == "" && input.End == "" {
		q.cache.Set("labelnames", *result)
	}
	return nil, result, nil
}
