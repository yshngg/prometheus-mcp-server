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

type LabelValuesArguments struct {
	Label string   `json:"label,omitzero" jsonschema:"<string>: Label names can optionally be encoded using the Values Escaping method, and is necessary if a name includes the / character. To encode a name in this way: 1. Prepend the label with U__. 2. Letters, numbers, and colons appear as-is. 3. Convert single underscores to double underscores. 4. For all other characters, use the UTF-8 codepoint as a hex integer, surrounded by underscores. So becomes _20_ and a . becomes _2e_."`
	Start string   `json:"start,omitzero" jsonschema:"<rfc3339 | unix_timestamp>: Start timestamp. Optional."`
	End   string   `json:"end,omitzero" jsonschema:"<rfc3339 | unix_timestamp>: End timestamp. Optional."`
	Match []string `json:"match[],omitzero" jsonschema:"<series_selector>: Repeated series selector argument that selects the series from which to read the label values. Optional."`
	Limit uint64   `json:"limit,omitzero" jsonschema:"<number>: Maximum number of returned series. Optional. 0 means disabled."`
}

type LabelValuesResult struct {
	LabelValues model.LabelValues `json:"labelvalues" jsonschema:"LabelSets consists of a list of objects that contain the label name/value pairs which identify each series."`
	Warnings    v1.Warnings       `json:"warnings,omitempty"`
}

func (q *metadataQuerier) LabelValuesHandler(ctx context.Context, request *mcp.CallToolRequest, input *LabelValuesArguments) (*mcp.CallToolResult, *LabelValuesResult, error) {
	key := "labelvalues:" + input.Label
	if len(input.Match) == 0 && input.Start == "" && input.End == "" {
		if v, ok := q.cache.Get(key); ok {
			result := v.(LabelValuesResult)
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

	result := &LabelValuesResult{}
	result.LabelValues, result.Warnings, err = q.API.LabelValues(
		ctx,
		input.Label,
		input.Match,
		start,
		end,
		opts...,
	)
	if err != nil {
		return nil, nil, err
	}

	if len(input.Match) == 0 && input.Start == "" && input.End == "" {
		q.cache.Set(key, *result)
	}
	return nil, result, nil
}
