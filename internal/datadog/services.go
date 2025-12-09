package datadog

import (
	"context"
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
)

// ServiceInfo represents information about a service.
type ServiceInfo struct {
	Name        string           `json:"name"`
	Description string           `json:"description,omitempty"`
	Team        string           `json:"team,omitempty"`
	Tier        string           `json:"tier,omitempty"`
	Lifecycle   string           `json:"lifecycle,omitempty"`
	Languages   []string         `json:"languages,omitempty"`
	Tags        []string         `json:"tags,omitempty"`
	Links       []ServiceLink    `json:"links,omitempty"`
	Contacts    []ServiceContact `json:"contacts,omitempty"`
}

// ServiceLink represents a link associated with a service.
type ServiceLink struct {
	Name string `json:"name"`
	Type string `json:"type"`
	URL  string `json:"url"`
}

// ServiceContact represents a contact for a service.
type ServiceContact struct {
	Name    string `json:"name,omitempty"`
	Type    string `json:"type"`
	Contact string `json:"contact"`
}

// ListServicesResult contains the result of listing services.
type ListServicesResult struct {
	Services []ServiceInfo `json:"services"`
	Total    int           `json:"total"`
}

// ListServices lists all services from the Datadog service catalog.
func (c *Client) ListServices(ctx context.Context) (*ListServicesResult, error) {
	opts := datadogV2.NewListServiceDefinitionsOptionalParameters().
		WithSchemaVersion(datadogV2.SERVICEDEFINITIONSCHEMAVERSIONS_V2_2)

	resp, _, err := c.serviceAPI.ListServiceDefinitions(c.ctx, *opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}

	result := &ListServicesResult{
		Services: make([]ServiceInfo, 0),
	}

	if resp.Data != nil {
		for _, svcData := range resp.Data {
			attrs := svcData.GetAttributes()

			svc := ServiceInfo{
				Name: svcData.GetId(),
			}

			schema := attrs.GetSchema()
			if v22, ok := schema.GetActualInstance().(*datadogV2.ServiceDefinitionV2Dot2); ok && v22 != nil {
				svc.Description = v22.GetDescription()
				svc.Team = v22.GetTeam()
				svc.Tier = v22.GetTier()
				svc.Lifecycle = v22.GetLifecycle()
				svc.Languages = v22.GetLanguages()
				svc.Tags = v22.GetTags()

				if links := v22.GetLinks(); links != nil {
					for _, link := range links {
						svc.Links = append(svc.Links, ServiceLink{
							Name: link.GetName(),
							Type: link.GetType(),
							URL:  link.GetUrl(),
						})
					}
				}

				if contacts := v22.GetContacts(); contacts != nil {
					for _, contact := range contacts {
						svc.Contacts = append(svc.Contacts, ServiceContact{
							Name:    contact.GetName(),
							Type:    contact.GetType(),
							Contact: contact.GetContact(),
						})
					}
				}
			}

			result.Services = append(result.Services, svc)
		}
		result.Total = len(result.Services)
	}

	return result, nil
}
