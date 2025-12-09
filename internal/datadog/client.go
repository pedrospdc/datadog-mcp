package datadog

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"

	"github.com/pedrospdc/datadog-mcp/internal/config"
)

// Client wraps the Datadog API client with authentication.
type Client struct {
	apiClient     *datadog.APIClient
	metricsV1     *datadogV1.MetricsApi
	metricsV2     *datadogV2.MetricsApi
	spansAPI      *datadogV2.SpansApi
	serviceAPI    *datadogV2.ServiceDefinitionApi
	dashboardsAPI *datadogV1.DashboardsApi
	ctx           context.Context
}

// NewClient creates a new Datadog API client with the given configuration.
func NewClient(cfg *config.Config) *Client {
	ctx := context.WithValue(
		context.Background(),
		datadog.ContextAPIKeys,
		map[string]datadog.APIKey{
			"apiKeyAuth": {Key: cfg.APIKey},
			"appKeyAuth": {Key: cfg.AppKey},
		},
	)

	if cfg.Site != "datadoghq.com" {
		ctx = context.WithValue(
			ctx,
			datadog.ContextServerVariables,
			map[string]string{"site": cfg.Site},
		)
	}

	configuration := datadog.NewConfiguration()
	apiClient := datadog.NewAPIClient(configuration)

	return &Client{
		apiClient:     apiClient,
		metricsV1:     datadogV1.NewMetricsApi(apiClient),
		metricsV2:     datadogV2.NewMetricsApi(apiClient),
		spansAPI:      datadogV2.NewSpansApi(apiClient),
		serviceAPI:    datadogV2.NewServiceDefinitionApi(apiClient),
		dashboardsAPI: datadogV1.NewDashboardsApi(apiClient),
		ctx:           ctx,
	}
}

// Context returns the authenticated context for API calls.
func (c *Client) Context() context.Context {
	return c.ctx
}

// MetricsV1 returns the V1 Metrics API.
func (c *Client) MetricsV1() *datadogV1.MetricsApi {
	return c.metricsV1
}

// MetricsV2 returns the V2 Metrics API.
func (c *Client) MetricsV2() *datadogV2.MetricsApi {
	return c.metricsV2
}

// SpansAPI returns the Spans API.
func (c *Client) SpansAPI() *datadogV2.SpansApi {
	return c.spansAPI
}

// ServiceAPI returns the Service Definition API.
func (c *Client) ServiceAPI() *datadogV2.ServiceDefinitionApi {
	return c.serviceAPI
}
