// Package azuredevops provides an HTTP client for interacting with Azure DevOps REST APIs.
package azuredevops

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ashutosh/sprintgpt-backend/internal/cache"
	"github.com/ashutosh/sprintgpt-backend/pkg/models"
)

// Client is an HTTP client for Azure DevOps APIs.
type Client struct {
	httpClient   *http.Client
	organization string
	project      string
	pat          string
	baseURL      string
}

// NewClient creates a new Azure DevOps API client.
func NewClient(organization, project, pat string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		organization: organization,
		project:      project,
		pat:          pat,
		baseURL:      fmt.Sprintf("https://dev.azure.com/%s/%s", organization, project),
	}
}

// authHeader returns the Basic Auth header value for PAT authentication.
func (c *Client) authHeader() string {
	// Azure DevOps uses Basic Auth with an empty username and the PAT as the password.
	encoded := base64.StdEncoding.EncodeToString([]byte(":" + c.pat))
	return "Basic " + encoded
}

// doRequest performs an authenticated HTTP request to Azure DevOps.
func (c *Client) doRequest(method, url string, body io.Reader) ([]byte, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", c.authHeader())
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("authentication failed: invalid PAT or insufficient permissions")
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("resource not found (404)")
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API error (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// GetWorkItem fetches a single work item by its ID.
func (c *Client) GetWorkItem(id int) (*models.WorkItem, error) {
	cacheKey := fmt.Sprintf("workitem:%s:%s:%d", c.organization, c.project, id)
	var cachedItem models.WorkItem
	if err := cache.Get(context.Background(), cacheKey, &cachedItem); err == nil && cachedItem.ID != 0 {
		return &cachedItem, nil
	}

	url := fmt.Sprintf(
		"https://dev.azure.com/%s/%s/_apis/wit/workitems/%d?api-version=7.1&$expand=all",
		c.organization, c.project, id,
	)

	respBody, err := c.doRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get work item %d: %w", id, err)
	}

	// Azure DevOps work item response structure
	var adoResponse struct {
		ID     int `json:"id"`
		Fields struct {
			Title         string `json:"System.Title"`
			State         string `json:"System.State"`
			AssignedTo    struct {
				DisplayName string `json:"displayName"`
			} `json:"System.AssignedTo"`
			WorkItemType  string `json:"System.WorkItemType"`
			Priority      int    `json:"Microsoft.VSTS.Common.Priority"`
			IterationPath string `json:"System.IterationPath"`
			CreatedDate   string `json:"System.CreatedDate"`
			ChangedDate   string `json:"System.ChangedDate"`
			Description   string `json:"System.Description"`
		} `json:"fields"`
		Links struct {
			HTML struct {
				Href string `json:"href"`
			} `json:"html"`
		} `json:"_links"`
	}

	if err := json.Unmarshal(respBody, &adoResponse); err != nil {
		return nil, fmt.Errorf("failed to parse work item response: %w", err)
	}

	result := &models.WorkItem{
		ID:            adoResponse.ID,
		Title:         adoResponse.Fields.Title,
		State:         adoResponse.Fields.State,
		AssignedTo:    adoResponse.Fields.AssignedTo.DisplayName,
		WorkItemType:  adoResponse.Fields.WorkItemType,
		Priority:      adoResponse.Fields.Priority,
		IterationPath: adoResponse.Fields.IterationPath,
		CreatedDate:   adoResponse.Fields.CreatedDate,
		ChangedDate:   adoResponse.Fields.ChangedDate,
		Description:   adoResponse.Fields.Description,
		URL:           adoResponse.Links.HTML.Href,
	}

	_ = cache.Set(context.Background(), cacheKey, result, 2*time.Minute)
	return result, nil
}

// ValidateConnection checks if the provided credentials can access the project.
func (c *Client) ValidateConnection() (*models.ValidationResponse, error) {
	url := fmt.Sprintf(
		"https://dev.azure.com/%s/_apis/projects/%s?api-version=7.1",
		c.organization, c.project,
	)

	respBody, err := c.doRequest("GET", url, nil)
	if err != nil {
		return &models.ValidationResponse{
			Valid:   false,
			Message: fmt.Sprintf("Connection failed: %s", err.Error()),
		}, nil
	}

	var project struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(respBody, &project); err != nil {
		return &models.ValidationResponse{
			Valid:   false,
			Message: "Failed to parse project info",
		}, nil
	}

	return &models.ValidationResponse{
		Valid:   true,
		Message: "Successfully connected to Azure DevOps",
		Project: project.Name,
	}, nil
}

// QueryWorkItems executes a WIQL query and returns the matching work items.
func (c *Client) QueryWorkItems(wiql string, maxItems int) ([]models.WorkItem, error) {
	cacheKey := fmt.Sprintf("wiql:%s:%s:%s", c.organization, c.project, base64.StdEncoding.EncodeToString([]byte(wiql)))
	var cachedResults []models.WorkItem
	if err := cache.Get(context.Background(), cacheKey, &cachedResults); err == nil {
		return cachedResults, nil
	}

	url := fmt.Sprintf(
		"https://dev.azure.com/%s/%s/_apis/wit/wiql?$top=%d&api-version=7.1",
		c.organization, c.project, maxItems,
	)

	reqBody := fmt.Sprintf(`{"query": "%s"}`, wiql)
	respBody, err := c.doRequest("POST", url, strings.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("WIQL query failed: %w", err)
	}

	var wiqlResp struct {
		WorkItems []struct {
			ID int `json:"id"`
		} `json:"workItems"`
	}

	if err := json.Unmarshal(respBody, &wiqlResp); err != nil {
		return nil, fmt.Errorf("failed to parse WIQL response: %w", err)
	}

	if len(wiqlResp.WorkItems) == 0 {
		return []models.WorkItem{}, nil
	}

	// Limit the number of items
	if len(wiqlResp.WorkItems) > maxItems {
		wiqlResp.WorkItems = wiqlResp.WorkItems[:maxItems]
	}

	// ADO requires a comma-separated list of IDs to fetch multiple items
	var ids []string
	for _, item := range wiqlResp.WorkItems {
		ids = append(ids, fmt.Sprintf("%d", item.ID))
	}
	idString := strings.Join(ids, ",")

	// Second API call: Get actual work item details for these IDs
	detailsUrl := fmt.Sprintf(
		"https://dev.azure.com/%s/%s/_apis/wit/workitems?ids=%s&api-version=7.1&$expand=all",
		c.organization, c.project, idString,
	)

	detailsBody, err := c.doRequest("GET", detailsUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch work item details: %w", err)
	}

	var detailsResp struct {
		Value []struct {
			ID     int `json:"id"`
			Fields struct {
				Title         string `json:"System.Title"`
				State         string `json:"System.State"`
				AssignedTo    struct {
					DisplayName string `json:"displayName"`
				} `json:"System.AssignedTo"`
				WorkItemType  string `json:"System.WorkItemType"`
				Priority      int    `json:"Microsoft.VSTS.Common.Priority"`
				IterationPath string `json:"System.IterationPath"`
				CreatedDate   string `json:"System.CreatedDate"`
				ChangedDate   string `json:"System.ChangedDate"`
				Description   string `json:"System.Description"`
			} `json:"fields"`
			Links struct {
				HTML struct {
					Href string `json:"href"`
				} `json:"html"`
			} `json:"_links"`
		} `json:"value"`
	}

	if err := json.Unmarshal(detailsBody, &detailsResp); err != nil {
		return nil, fmt.Errorf("failed to parse work item details: %w", err)
	}

	var results []models.WorkItem
	for _, item := range detailsResp.Value {
		results = append(results, models.WorkItem{
			ID:            item.ID,
			Title:         item.Fields.Title,
			State:         item.Fields.State,
			AssignedTo:    item.Fields.AssignedTo.DisplayName,
			WorkItemType:  item.Fields.WorkItemType,
			Priority:      item.Fields.Priority,
			IterationPath: item.Fields.IterationPath,
			CreatedDate:   item.Fields.CreatedDate,
			ChangedDate:   item.Fields.ChangedDate,
			Description:   item.Fields.Description,
			URL:           item.Links.HTML.Href,
		})
	}

	_ = cache.Set(context.Background(), cacheKey, results, 1*time.Minute)
	return results, nil
}
