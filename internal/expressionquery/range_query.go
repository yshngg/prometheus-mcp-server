package expressionquery

import (
	"context"
	"errors"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/yshngg/prometheus-mcp-server/internal/utils"
	"k8s.io/klog/v2"
)

type RangeQueryArguments struct {
	Query   string        `json:"query" jsonschema:"<string>: Prometheus expression query string."`
	Start   string        `json:"start" jsonschema:"<rfc3339 | unix_timestamp>: Start timestamp, inclusive."`
	End     string        `json:"end" jsonschema:"<rfc3339 | unix_timestamp>: End timestamp, inclusive."`
	Step    time.Duration `json:"step" jsonschema:"<duration | float>: Query resolution step width in duration format or float number of seconds."`
	Timeout time.Duration `json:"timeout,omitzero" jsonschema:"<duration>: Evaluation timeout. Optional. Defaults to and is capped by the value of the -query.timeout flag."`
	Limit   uint64        `json:"limit,omitzero" jsonschema:"<number>: Maximum number of returned series. Optional. 0 means disabled."`
}

type RangeQueryResult struct {
	Value    model.Value `json:"value" jsonschema:"<value> refers to the query result data, which has varying formats depending on the resultType. See the [range-vector result format](https://prometheus.io/docs/prometheus/latest/querying/api/#range-vectors)."`
	Warnings v1.Warnings `json:"warnings,omitempty"`
}

func (q *expressionQuerier) RangeQueryHandler(ctx context.Context, request *mcp.CallToolRequest, input *RangeQueryArguments) (*mcp.CallToolResult, *RangeQueryResult, error) {
	var (
		start, end time.Time
		step       time.Duration
		err        error
	)
	if input.Start != "" {
		if start, err = utils.ParseTime(input.Start); err != nil {
			klog.InfoS("parse start time", "err", err)
		}
	}
	if input.End != "" {
		if end, err = utils.ParseTime(input.End); err != nil {
			klog.InfoS("parse end time", "err", err)
		}
	}

	if input.Step == 0 {
		return nil, nil, errors.New("step cannot be 0")
	}
	step = input.Step * time.Second

	opts := make([]v1.Option, 0)
	if input.Timeout != 0 {
		opts = append(opts, v1.WithTimeout(input.Timeout))
	}
	if input.Limit != 0 {
		opts = append(opts, v1.WithLimit(input.Limit))
	}

	result := &RangeQueryResult{}
	if result.Value, result.Warnings, err = q.API.QueryRange(
		ctx,
		input.Query,
		v1.Range{
			Start: start,
			End:   end,
			Step:  step,
		},
		opts...,
	); err != nil {
		return nil, nil, err
	}
	return nil, result, nil
}
