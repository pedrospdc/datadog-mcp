package tools

import (
	"context"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/pedrospdc/datadog-mcp/internal/datadog"
)

// QueryAPMStatsInput defines the input for the query_apm_stats tool.
type QueryAPMStatsInput struct {
	Service   string `json:"service" jsonschema:"The service name to query stats for"`
	Operation string `json:"operation,omitempty" jsonschema:"Specific operation/resource name to filter by"`
	Env       string `json:"env,omitempty" jsonschema:"Environment to filter by, e.g. production or staging"`
	From      string `json:"from,omitempty" jsonschema:"Start time. Defaults to 1 hour ago"`
	To        string `json:"to,omitempty" jsonschema:"End time. Defaults to now"`
}

// APMStatsResult contains APM statistics for a service.
type APMStatsResult struct {
	Service    string           `json:"service"`
	Operation  string           `json:"operation,omitempty"`
	Env        string           `json:"env,omitempty"`
	TimeRange  TimeRange        `json:"time_range"`
	Latency    *LatencyStats    `json:"latency,omitempty"`
	ErrorRate  *ErrorRateStats  `json:"error_rate,omitempty"`
	Throughput *ThroughputStats `json:"throughput,omitempty"`
}

// TimeRange represents a time range.
type TimeRange struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

// LatencyStats contains latency statistics.
type LatencyStats struct {
	Avg float64 `json:"avg_ms"`
	P50 float64 `json:"p50_ms"`
	P95 float64 `json:"p95_ms"`
	P99 float64 `json:"p99_ms"`
}

// ErrorRateStats contains error rate statistics.
type ErrorRateStats struct {
	ErrorCount   float64 `json:"error_count"`
	TotalCount   float64 `json:"total_count"`
	ErrorPercent float64 `json:"error_percent"`
}

// ThroughputStats contains throughput statistics.
type ThroughputStats struct {
	RequestsPerSecond float64 `json:"requests_per_second"`
	TotalRequests     float64 `json:"total_requests"`
}

func registerQueryAPMStats(server *mcp.Server, client *datadog.Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "query_apm_stats",
		Description: "Query APM statistics for a service including latency percentiles (p50, p95, p99), error rates, and throughput.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input QueryAPMStatsInput) (*mcp.CallToolResult, *APMStatsResult, error) {
		var from, to time.Time
		var err error

		if input.From == "" {
			from = time.Now().Add(-1 * time.Hour)
		} else {
			from, err = parseTime(input.From)
			if err != nil {
				return nil, nil, fmt.Errorf("invalid 'from' time: %w", err)
			}
		}

		if input.To == "" {
			to = time.Now()
		} else {
			to, err = parseTime(input.To)
			if err != nil {
				return nil, nil, fmt.Errorf("invalid 'to' time: %w", err)
			}
		}

		// Build tag filter
		tags := fmt.Sprintf("service:%s", input.Service)
		if input.Operation != "" {
			tags += fmt.Sprintf(",resource_name:%s", input.Operation)
		}
		if input.Env != "" {
			tags += fmt.Sprintf(",env:%s", input.Env)
		}

		result := &APMStatsResult{
			Service:   input.Service,
			Operation: input.Operation,
			Env:       input.Env,
			TimeRange: TimeRange{From: from, To: to},
		}

		// Query latency metrics (avg)
		latencyQuery := fmt.Sprintf("avg:trace.%s.duration{%s}", input.Service, tags)
		latencyResult, err := client.QueryMetrics(ctx, latencyQuery, from, to)
		if err == nil && len(latencyResult.Series) > 0 {
			result.Latency = &LatencyStats{}
			var sum float64
			for _, dp := range latencyResult.Series[0].DataPoints {
				sum += dp.Value
			}
			if len(latencyResult.Series[0].DataPoints) > 0 {
				result.Latency.Avg = sum / float64(len(latencyResult.Series[0].DataPoints)) / 1e6
			}
		}

		// Query p95 latency
		p95Query := fmt.Sprintf("p95:trace.%s.duration{%s}", input.Service, tags)
		p95Result, err := client.QueryMetrics(ctx, p95Query, from, to)
		if err == nil && len(p95Result.Series) > 0 && result.Latency != nil {
			var sum float64
			for _, dp := range p95Result.Series[0].DataPoints {
				sum += dp.Value
			}
			if len(p95Result.Series[0].DataPoints) > 0 {
				result.Latency.P95 = sum / float64(len(p95Result.Series[0].DataPoints)) / 1e6
			}
		}

		// Query error count
		errorQuery := fmt.Sprintf("sum:trace.%s.errors{%s}.as_count()", input.Service, tags)
		errorResult, err := client.QueryMetrics(ctx, errorQuery, from, to)
		if err == nil && len(errorResult.Series) > 0 {
			result.ErrorRate = &ErrorRateStats{}
			for _, dp := range errorResult.Series[0].DataPoints {
				result.ErrorRate.ErrorCount += dp.Value
			}
		}

		// Query hit count (total requests)
		hitsQuery := fmt.Sprintf("sum:trace.%s.hits{%s}.as_count()", input.Service, tags)
		hitsResult, err := client.QueryMetrics(ctx, hitsQuery, from, to)
		if err == nil && len(hitsResult.Series) > 0 {
			if result.ErrorRate == nil {
				result.ErrorRate = &ErrorRateStats{}
			}
			for _, dp := range hitsResult.Series[0].DataPoints {
				result.ErrorRate.TotalCount += dp.Value
			}
			if result.ErrorRate.TotalCount > 0 {
				result.ErrorRate.ErrorPercent = (result.ErrorRate.ErrorCount / result.ErrorRate.TotalCount) * 100
			}

			// Calculate throughput
			duration := to.Sub(from).Seconds()
			if duration > 0 {
				result.Throughput = &ThroughputStats{
					TotalRequests:     result.ErrorRate.TotalCount,
					RequestsPerSecond: result.ErrorRate.TotalCount / duration,
				}
			}
		}

		// Build summary text
		summary := fmt.Sprintf("APM Stats for service: %s\n", input.Service)
		if input.Operation != "" {
			summary += fmt.Sprintf("Operation: %s\n", input.Operation)
		}
		if input.Env != "" {
			summary += fmt.Sprintf("Environment: %s\n", input.Env)
		}
		summary += fmt.Sprintf("Time Range: %s to %s\n\n", from.Format(time.RFC3339), to.Format(time.RFC3339))

		if result.Latency != nil {
			summary += "Latency:\n"
			summary += fmt.Sprintf("  Avg: %.2f ms\n", result.Latency.Avg)
			summary += fmt.Sprintf("  P95: %.2f ms\n", result.Latency.P95)
		}

		if result.ErrorRate != nil {
			summary += "\nError Rate:\n"
			summary += fmt.Sprintf("  Errors: %.0f\n", result.ErrorRate.ErrorCount)
			summary += fmt.Sprintf("  Total: %.0f\n", result.ErrorRate.TotalCount)
			summary += fmt.Sprintf("  Rate: %.2f%%\n", result.ErrorRate.ErrorPercent)
		}

		if result.Throughput != nil {
			summary += "\nThroughput:\n"
			summary += fmt.Sprintf("  Requests/sec: %.2f\n", result.Throughput.RequestsPerSecond)
			summary += fmt.Sprintf("  Total Requests: %.0f\n", result.Throughput.TotalRequests)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: summary},
			},
		}, result, nil
	})
}
