// Package mimir provides a simple client for the required Mimir API resources.
package mimir

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// Client is a simple client for the required Mimir API resources.
type Client struct {
	client           http.Client
	address          *url.URL
	prometheusPrefix string
}

// Config is used to configure the client.
type Config struct {
	Address           string
	PrometheusPrefix  string
	BasicAuthUsername string
	BasicAuthPassword string
}

// NewClient creates a new client with the given configuration.
func NewClient(config Config) (*Client, error) {
	addr, err := url.Parse(config.Address)
	if err != nil {
		return nil, err
	}

	if config.PrometheusPrefix == "" {
		config.PrometheusPrefix = "prometheus"
	}

	httpClient := http.Client{}
	if config.BasicAuthUsername != "" && config.BasicAuthPassword != "" {
		httpClient.Transport = &http.Transport{}
		httpClient.Transport = &BasicAuthTransport{
			Username:  config.BasicAuthUsername,
			Password:  config.BasicAuthPassword,
			Transport: httpClient.Transport,
		}
	}

	return &Client{
		client:           httpClient,
		address:          addr,
		prometheusPrefix: config.PrometheusPrefix,
	}, nil
}

// BasicAuthTransport is a transport that adds basic auth to the request.
type BasicAuthTransport struct {
	Username  string
	Password  string
	Transport http.RoundTripper
}

// RoundTrip adds basic auth to the request.
func (t *BasicAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.SetBasicAuth(t.Username, t.Password)
	return t.Transport.RoundTrip(req)
}

// Ready checks if mimir is ready to serve traffic.
func (c *Client) Ready(ctx context.Context) error {
	path := c.address.JoinPath("/ready")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path.String(), nil)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("mimir not ready, unexpected status code: %d, expected %d", resp.StatusCode, http.StatusOK)
	}
	return nil
}
