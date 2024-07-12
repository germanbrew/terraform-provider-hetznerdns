package api

import (
	"context"
	"fmt"
	"net/http"
)

type PrimaryServer struct {
	ID      string `json:"id"`
	Port    uint16 `json:"port"`
	ZoneID  string `json:"zone_id"`
	Address string `json:"address"`
}

type CreatePrimaryServerRequest struct {
	Port    uint16 `json:"port"`
	ZoneID  string `json:"zone_id"`
	Address string `json:"address"`
}

type PrimaryServersResponse struct {
	PrimaryServers []PrimaryServer `json:"primary_servers"`
}

type PrimaryServerResponse struct {
	PrimaryServer PrimaryServer `json:"primary_server"`
}

func (c *Client) GetPrimaryServer(ctx context.Context, id string) (*PrimaryServer, error) {
	resp, err := c.request(ctx, http.MethodGet, "/api/v1/primary_servers/"+id, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting primary server %s: %w", id, err)
	}

	switch resp.StatusCode {
	case http.StatusNotFound:
		return nil, fmt.Errorf("primary server %s: %w", id, ErrNotFound)
	case http.StatusOK:
		var response *PrimaryServerResponse

		err = readAndParseJSONBody(resp, &response)
		if err != nil {
			return nil, fmt.Errorf("error Reading json response of get primary server %s request: %s", id, err)
		}

		return &response.PrimaryServer, nil
	default:
		return nil, fmt.Errorf("http status %d unhandled", resp.StatusCode)
	}
}

func (c *Client) GetPrimaryServers(ctx context.Context, zoneID string) ([]PrimaryServer, error) {
	resp, err := c.request(ctx, http.MethodGet, "/api/v1/primary_servers?zone_id="+zoneID, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting primary servers for zone %s: %w", zoneID, err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		var response PrimaryServersResponse

		err = readAndParseJSONBody(resp, &response)
		if err != nil {
			return nil, err
		}

		return response.PrimaryServers, nil
	default:
		return nil, fmt.Errorf("http status %d unhandled", resp.StatusCode)
	}
}

func (c *Client) CreatePrimaryServer(ctx context.Context, server CreatePrimaryServerRequest) (*PrimaryServer, error) {
	resp, err := c.request(ctx, http.MethodPost, "/api/v1/primary_servers", server)
	if err != nil {
		return nil, fmt.Errorf("error creating primary server %s: %w", server.Address, err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		var response PrimaryServerResponse

		err = readAndParseJSONBody(resp, &response)
		if err != nil {
			return nil, err
		}

		return &response.PrimaryServer, nil
	default:
		return nil, fmt.Errorf("http status %d unhandled", resp.StatusCode)
	}
}

func (c *Client) UpdatePrimaryServer(ctx context.Context, server PrimaryServer) (*PrimaryServer, error) {
	resp, err := c.request(ctx, http.MethodPut, "/api/v1/primary_servers/"+server.ID, server)
	if err != nil {
		return nil, fmt.Errorf("error updating primary server %s: %w", server.ID, err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		var response PrimaryServerResponse

		err = readAndParseJSONBody(resp, &response)
		if err != nil {
			return nil, err
		}

		return &response.PrimaryServer, nil
	default:
		return nil, fmt.Errorf("http status %d unhandled", resp.StatusCode)
	}
}

func (c *Client) DeletePrimaryServer(ctx context.Context, id string) error {
	resp, err := c.request(ctx, http.MethodDelete, "/api/v1/primary_servers/"+id, nil)
	if err != nil {
		return fmt.Errorf("error deleting primary server %s: %w", id, err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	default:
		return fmt.Errorf("http status %d unhandled", resp.StatusCode)
	}
}
