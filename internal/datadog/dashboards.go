package datadog

import (
	"context"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
)

// DashboardSummary represents a summary of a dashboard.
type DashboardSummary struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	LayoutType  string    `json:"layout_type"`
	URL         string    `json:"url,omitempty"`
	AuthorHandle string   `json:"author_handle,omitempty"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
	ModifiedAt  time.Time `json:"modified_at,omitempty"`
	IsReadOnly  bool      `json:"is_read_only"`
}

// ListDashboardsResult contains the result of listing dashboards.
type ListDashboardsResult struct {
	Dashboards []DashboardSummary `json:"dashboards"`
	Total      int                `json:"total"`
}

// DashboardWidget represents a widget in a dashboard.
type DashboardWidget struct {
	ID         int64                  `json:"id,omitempty"`
	Definition map[string]interface{} `json:"definition"`
}

// DashboardTemplateVariable represents a template variable.
type DashboardTemplateVariable struct {
	Name             string   `json:"name"`
	Prefix           string   `json:"prefix,omitempty"`
	Default          string   `json:"default,omitempty"`
	AvailableValues  []string `json:"available_values,omitempty"`
}

// Dashboard represents a full dashboard configuration.
type Dashboard struct {
	ID                string                      `json:"id"`
	Title             string                      `json:"title"`
	Description       string                      `json:"description,omitempty"`
	LayoutType        string                      `json:"layout_type"`
	URL               string                      `json:"url,omitempty"`
	AuthorHandle      string                      `json:"author_handle,omitempty"`
	AuthorName        string                      `json:"author_name,omitempty"`
	CreatedAt         time.Time                   `json:"created_at,omitempty"`
	ModifiedAt        time.Time                   `json:"modified_at,omitempty"`
	IsReadOnly        bool                        `json:"is_read_only"`
	Tags              []string                    `json:"tags,omitempty"`
	Widgets           []DashboardWidget           `json:"widgets"`
	TemplateVariables []DashboardTemplateVariable `json:"template_variables,omitempty"`
	WidgetCount       int                         `json:"widget_count"`
}

// ListDashboards retrieves all dashboards from Datadog.
func (c *Client) ListDashboards(ctx context.Context, filterShared, filterDeleted bool) (*ListDashboardsResult, error) {
	opts := datadogV1.NewListDashboardsOptionalParameters()
	if filterShared {
		opts = opts.WithFilterShared(filterShared)
	}
	if filterDeleted {
		opts = opts.WithFilterDeleted(filterDeleted)
	}

	resp, _, err := c.dashboardsAPI.ListDashboards(c.ctx, *opts)
	if err != nil {
		return nil, err
	}

	result := &ListDashboardsResult{
		Dashboards: make([]DashboardSummary, 0),
	}

	if resp.Dashboards != nil {
		for _, d := range resp.Dashboards {
			summary := DashboardSummary{}
			if d.Id != nil {
				summary.ID = *d.Id
			}
			if d.Title != nil {
				summary.Title = *d.Title
			}
			if d.Description.IsSet() && d.Description.Get() != nil {
				summary.Description = *d.Description.Get()
			}
			if d.LayoutType != nil {
				summary.LayoutType = string(*d.LayoutType)
			}
			if d.Url != nil {
				summary.URL = *d.Url
			}
			if d.AuthorHandle != nil {
				summary.AuthorHandle = *d.AuthorHandle
			}
			if d.CreatedAt != nil {
				summary.CreatedAt = *d.CreatedAt
			}
			if d.ModifiedAt != nil {
				summary.ModifiedAt = *d.ModifiedAt
			}
			if d.IsReadOnly != nil {
				summary.IsReadOnly = *d.IsReadOnly
			}
			result.Dashboards = append(result.Dashboards, summary)
		}
	}

	result.Total = len(result.Dashboards)
	return result, nil
}

// GetDashboard retrieves a specific dashboard by ID.
func (c *Client) GetDashboard(ctx context.Context, dashboardID string) (*Dashboard, error) {
	resp, _, err := c.dashboardsAPI.GetDashboard(c.ctx, dashboardID)
	if err != nil {
		return nil, err
	}

	dashboard := &Dashboard{
		Title:      resp.Title,
		LayoutType: string(resp.LayoutType),
		Widgets:    make([]DashboardWidget, 0),
	}

	if resp.Id != nil {
		dashboard.ID = *resp.Id
	}
	if resp.Description.IsSet() && resp.Description.Get() != nil {
		dashboard.Description = *resp.Description.Get()
	}
	if resp.Url != nil {
		dashboard.URL = *resp.Url
	}
	if resp.AuthorHandle != nil {
		dashboard.AuthorHandle = *resp.AuthorHandle
	}
	if resp.AuthorName.IsSet() && resp.AuthorName.Get() != nil {
		dashboard.AuthorName = *resp.AuthorName.Get()
	}
	if resp.CreatedAt != nil {
		dashboard.CreatedAt = *resp.CreatedAt
	}
	if resp.ModifiedAt != nil {
		dashboard.ModifiedAt = *resp.ModifiedAt
	}
	if resp.IsReadOnly != nil {
		dashboard.IsReadOnly = *resp.IsReadOnly
	}
	if resp.Tags.IsSet() && resp.Tags.Get() != nil {
		dashboard.Tags = *resp.Tags.Get()
	}

	// Process widgets
	for _, w := range resp.Widgets {
		widget := DashboardWidget{}
		if w.Id != nil {
			widget.ID = *w.Id
		}
		// Store widget definition as a map for flexibility
		widget.Definition = make(map[string]interface{})
		widget.Definition["type"] = "widget"
		dashboard.Widgets = append(dashboard.Widgets, widget)
	}
	dashboard.WidgetCount = len(dashboard.Widgets)

	// Process template variables
	if resp.TemplateVariables != nil {
		for _, tv := range resp.TemplateVariables {
			tmplVar := DashboardTemplateVariable{
				Name: tv.Name,
			}
			if tv.Prefix.IsSet() && tv.Prefix.Get() != nil {
				tmplVar.Prefix = *tv.Prefix.Get()
			}
			if tv.Default.IsSet() && tv.Default.Get() != nil {
				tmplVar.Default = *tv.Default.Get()
			}
			if tv.AvailableValues.IsSet() && tv.AvailableValues.Get() != nil {
				tmplVar.AvailableValues = *tv.AvailableValues.Get()
			}
			dashboard.TemplateVariables = append(dashboard.TemplateVariables, tmplVar)
		}
	}

	return dashboard, nil
}
