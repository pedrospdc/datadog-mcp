package datadog

import (
	"context"
	"fmt"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
)

// MetricPoint represents a single data point in a metric series.
type MetricPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// MetricSeries represents a metric timeseries with its data points.
type MetricSeries struct {
	Metric     string        `json:"metric"`
	Tags       []string      `json:"tags,omitempty"`
	Unit       string        `json:"unit,omitempty"`
	DataPoints []MetricPoint `json:"data_points"`
}

// QueryMetricsResult contains the result of a metrics query.
type QueryMetricsResult struct {
	Series      []MetricSeries `json:"series"`
	Query       string         `json:"query"`
	From        time.Time      `json:"from"`
	To          time.Time      `json:"to"`
	TotalSeries int            `json:"total_series,omitempty"`
	Truncated   bool           `json:"truncated,omitempty"`
}

// QueryMetrics queries timeseries metrics from Datadog.
func (c *Client) QueryMetrics(ctx context.Context, query string, from, to time.Time) (*QueryMetricsResult, error) {
	resp, _, err := c.metricsV1.QueryMetrics(c.ctx, from.Unix(), to.Unix(), query)
	if err != nil {
		return nil, fmt.Errorf("failed to query metrics: %w", err)
	}

	result := &QueryMetricsResult{
		Series: make([]MetricSeries, 0),
		Query:  query,
		From:   from,
		To:     to,
	}

	for _, series := range resp.GetSeries() {
		ms := MetricSeries{
			Metric:     series.GetMetric(),
			Tags:       series.GetTagSet(),
			DataPoints: make([]MetricPoint, 0),
		}

		if unit := series.GetUnit(); len(unit) > 0 && unit[0].GetName() != "" {
			ms.Unit = unit[0].GetName()
		}

		if pointList := series.GetPointlist(); pointList != nil {
			for _, point := range pointList {
				if len(point) >= 2 && point[0] != nil && point[1] != nil {
					ms.DataPoints = append(ms.DataPoints, MetricPoint{
						Timestamp: time.UnixMilli(int64(*point[0])),
						Value:     *point[1],
					})
				}
			}
		}

		result.Series = append(result.Series, ms)
	}

	return result, nil
}

// ListMetricsResult contains the result of listing available metrics.
type ListMetricsResult struct {
	Metrics []string `json:"metrics"`
	From    int64    `json:"from"`
	Total   int      `json:"total"`
	Offset  int      `json:"offset"`
	HasMore bool     `json:"has_more"`
}

// ListMetrics lists active metrics from Datadog.
func (c *Client) ListMetrics(ctx context.Context, from time.Time, host string, tagFilter string) (*ListMetricsResult, error) {
	opts := datadogV1.NewListActiveMetricsOptionalParameters()
	if host != "" {
		opts = opts.WithHost(host)
	}
	if tagFilter != "" {
		opts = opts.WithTagFilter(tagFilter)
	}

	resp, _, err := c.metricsV1.ListActiveMetrics(c.ctx, from.Unix(), *opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list metrics: %w", err)
	}

	result := &ListMetricsResult{
		Metrics: make([]string, 0),
		From:    from.Unix(),
	}

	result.Metrics = resp.GetMetrics()

	return result, nil
}
