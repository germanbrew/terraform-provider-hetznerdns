package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

// Zone represents a DNS Zone.
type Zone struct {
	ID   string   `json:"id"`
	Name string   `json:"name"`
	NS   []string `json:"ns"`
	TTL  int64    `json:"ttl"`
}

// CreateZoneOpts covers all parameters used to create a new DNS zone.
type CreateZoneOpts struct {
	Name string `json:"name"`
	TTL  int64  `json:"ttl"`
}

// CreateZoneRequest represents the body of a POST Zone request.
type CreateZoneRequest struct {
	Name string `json:"name"`
	TTL  int64  `json:"ttl"`
}

// CreateZoneResponse represents the content of a POST Zone response.
type CreateZoneResponse struct {
	Zone Zone `json:"zone"`
}

// GetZoneResponse represents the content of a GET Zone request.
type GetZoneResponse struct {
	Zone Zone `json:"zone"`
}

// ZoneResponse represents the content of response containing a Zone.
type ZoneResponse struct {
	Zone Zone `json:"zone"`
}

// GetZones represents the content of a GET Zones response.
type GetZones struct {
	Zones []Zone `json:"zones"`
}

// GetZonesByNameResponse represents the content of a GET Zones response.
type GetZonesByNameResponse struct {
	Zones []Zone `json:"zones"`
}

// GetZones reads the current state of a DNS zone.
func (c *Client) GetZones(ctx context.Context) ([]Zone, error) {
	resp, err := c.request(ctx, http.MethodGet, "/api/v1/zones", nil)
	if err != nil {
		return nil, fmt.Errorf("error getting zones: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusNotFound:
		// Undocumented API behavior: Hetzner DNS API returns 404 when there are no zones
		return nil, fmt.Errorf("zones: %w", ErrNotFound)
	case http.StatusOK:
		var response GetZones

		err = readAndParseJSONBody(resp, &response)
		if err != nil {
			return nil, err
		}

		return response.Zones, nil
	default:
		return nil, fmt.Errorf("error getting zones. HTTP status %d unhandled", resp.StatusCode)
	}
}

// GetZone reads the current state of a DNS zone.
func (c *Client) GetZone(ctx context.Context, id string) (*Zone, error) {
	resp, err := c.request(ctx, http.MethodGet, "/api/v1/zones/"+id, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting zone %s: %w", id, err)
	}

	switch resp.StatusCode {
	case http.StatusNotFound:
		return nil, fmt.Errorf("zone %s: %w", id, ErrNotFound)
	case http.StatusOK:
		var response GetZoneResponse

		err = readAndParseJSONBody(resp, &response)
		if err != nil {
			return nil, err
		}

		return &response.Zone, nil
	default:
		return nil, fmt.Errorf("http status %d unhandled", resp.StatusCode)
	}
}

// UpdateZone takes the passed state and updates the respective Zone.
func (c *Client) UpdateZone(ctx context.Context, zone Zone) (*Zone, error) {
	resp, err := c.request(ctx, http.MethodPut, "/api/v1/zones/"+zone.ID, zone)
	if err != nil {
		return nil, fmt.Errorf("error updating zone %s: %s", zone.ID, err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		var response ZoneResponse

		err = readAndParseJSONBody(resp, &response)
		if err != nil {
			return nil, err
		}

		return &response.Zone, nil
	default:
		return nil, fmt.Errorf("http status %d unhandled", resp.StatusCode)
	}
}

// DeleteZone deletes a given DNS zone.
func (c *Client) DeleteZone(ctx context.Context, id string) error {
	resp, err := c.request(ctx, http.MethodDelete, "/api/v1/zones/"+id, nil)
	if err != nil {
		return fmt.Errorf("error deleting zone %s: %s", id, err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	default:
		return fmt.Errorf("http status %d unhandled", resp.StatusCode)
	}
}

// GetZoneByName reads the current state of a DNS zone with a given name.
func (c *Client) GetZoneByName(ctx context.Context, name string) (*Zone, error) {
	resp, err := c.request(ctx, http.MethodGet, "/api/v1/zones?name="+name, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting zones: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusNotFound:
		return nil, fmt.Errorf("zone %s: %w", name, ErrNotFound)
	case http.StatusOK:
		var response GetZones

		err = readAndParseJSONBody(resp, &response)
		if err != nil {
			return nil, err
		}

		if len(response.Zones) != 1 {
			return nil, fmt.Errorf("error getting zone '%s'. No matching zone or multiple matching zones found", name)
		}

		return &response.Zones[0], nil
	default:
		return nil, fmt.Errorf("error getting zones. HTTP status %d unhandled", resp.StatusCode)
	}
}

// CreateZone creates a new DNS zone.
func (c *Client) CreateZone(ctx context.Context, opts CreateZoneOpts) (*Zone, error) {
	if !strings.Contains(opts.Name, ".") {
		return nil, fmt.Errorf("error creating zone. The name '%s' is not a valid domain. It must correspond to the schema <domain>.<tld>", opts.Name)
	}

	reqBody := CreateZoneRequest(opts)

	resp, err := c.request(ctx, http.MethodPost, "/api/v1/zones", reqBody)
	if err != nil {
		return nil, fmt.Errorf("error getting zones: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		var response CreateZoneResponse

		err = readAndParseJSONBody(resp, &response)
		if err != nil {
			return nil, err
		}

		return &response.Zone, nil
	default:
		return nil, fmt.Errorf("http status %d unhandled", resp.StatusCode)
	}
}
