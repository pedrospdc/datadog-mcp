package datadog

import (
	"context"
	"fmt"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
)

// Span represents an APM span.
type Span struct {
	TraceID    string            `json:"trace_id"`
	SpanID     string            `json:"span_id"`
	ParentID   string            `json:"parent_id,omitempty"`
	Service    string            `json:"service"`
	Name       string            `json:"name"`
	Resource   string            `json:"resource"`
	Type       string            `json:"type,omitempty"`
	Start      time.Time         `json:"start"`
	Duration   int64             `json:"duration_ns"`
	Status     string            `json:"status"`
	Error      int32             `json:"error"`
	Tags       map[string]string `json:"tags,omitempty"`
}

// QuerySpansResult contains the result of a spans query.
type QuerySpansResult struct {
	Spans      []Span `json:"spans"`
	TotalCount int    `json:"total_count"`
	NextCursor string `json:"next_cursor,omitempty"`
}

// QuerySpans queries APM spans from Datadog.
func (c *Client) QuerySpans(ctx context.Context, query string, from, to string, limit int32, cursor string) (*QuerySpansResult, error) {
	if from == "" {
		from = "now-15m"
	}
	if to == "" {
		to = "now"
	}
	if limit <= 0 {
		limit = 50
	}
	if limit > 1000 {
		limit = 1000
	}

	page := &datadogV2.SpansListRequestPage{
		Limit: datadog.PtrInt32(limit),
	}
	if cursor != "" {
		page.Cursor = datadog.PtrString(cursor)
	}

	body := datadogV2.SpansListRequest{
		Data: &datadogV2.SpansListRequestData{
			Attributes: &datadogV2.SpansListRequestAttributes{
				Filter: &datadogV2.SpansQueryFilter{
					From:  datadog.PtrString(from),
					Query: datadog.PtrString(query),
					To:    datadog.PtrString(to),
				},
				Options: &datadogV2.SpansQueryOptions{
					Timezone: datadog.PtrString("UTC"),
				},
				Page: page,
				Sort: datadogV2.SPANSSORT_TIMESTAMP_DESCENDING.Ptr(),
			},
			Type: datadogV2.SPANSLISTREQUESTTYPE_SEARCH_REQUEST.Ptr(),
		},
	}

	resp, _, err := c.spansAPI.ListSpans(c.ctx, body)
	if err != nil {
		return nil, fmt.Errorf("failed to query spans: %w", err)
	}

	result := &QuerySpansResult{
		Spans: make([]Span, 0),
	}

	if resp.Data != nil {
		for _, spanData := range resp.Data {
			attrs := spanData.GetAttributes()
			span := Span{
				SpanID:   attrs.GetSpanId(),
				TraceID:  attrs.GetTraceId(),
				Service:  attrs.GetService(),
				Name:     attrs.GetResourceName(),
				Resource: attrs.GetResourceName(),
				Type:     attrs.GetType(),
				Status:   "ok",
				Tags:     make(map[string]string),
			}

			startTime := attrs.GetStartTimestamp()
			if !startTime.IsZero() {
				span.Start = startTime
			}
			endTime := attrs.GetEndTimestamp()
			if !startTime.IsZero() && !endTime.IsZero() {
				span.Duration = endTime.Sub(startTime).Nanoseconds()
			}

			if ingestionReason := attrs.GetIngestionReason(); ingestionReason != "" {
				span.Tags["ingestion_reason"] = ingestionReason
			}

			result.Spans = append(result.Spans, span)
		}
		result.TotalCount = len(result.Spans)
	}

	// Extract next cursor from response metadata
	if resp.Meta != nil {
		if page := resp.Meta.GetPage(); page.After != nil {
			result.NextCursor = *page.After
		}
	}

	return result, nil
}
