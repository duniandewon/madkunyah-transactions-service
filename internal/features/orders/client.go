package orders

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type MenuClient struct {
	httpClient *http.Client
	baseURL    string
}

func NewMenuClient(baseURL string) *MenuClient {
	return &MenuClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

type MenuResponse struct {
	ID             int                     `json:"id"`
	Name           string                  `json:"name"`
	Image          string                  `json:"image"`
	Price          float64                 `json:"price"`
	ModifierGroups []ModifierGroupResponse `json:"modifier_groups"`
}

type ModifierGroupResponse struct {
	ID    int                    `json:"id"`
	Name  string                 `json:"name"`
	Items []ModifierItemResponse `json:"items"`
}

type ModifierItemResponse struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

func (c *MenuClient) FetchMenu(ctx context.Context, menuID int) (*MenuResponse, error) {
	url := fmt.Sprintf("%s/menus/%d", c.baseURL, menuID)

	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch menu, status: %s", resp.Status)
	}
	defer resp.Body.Close()

	var menu MenuResponse
	if err := json.NewDecoder(resp.Body).Decode(&menu); err != nil {
		return nil, err
	}

	return &menu, nil
}
