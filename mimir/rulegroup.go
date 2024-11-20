package mimir

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/prometheus/prometheus/model/rulefmt"
	"gopkg.in/yaml.v3"
)

// SetRuleGroup creates or updates a rule group.
func (c *Client) SetRuleGroup(ctx context.Context, namespace string, ruleGroup rulefmt.RuleGroup) error {
	path := c.address.JoinPath(c.prometheusPrefix, "/config/v1/rules/", namespace)

	body, err := yaml.Marshal(ruleGroup)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, path.String(), io.NopCloser(bytes.NewReader(body)))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/yaml")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("unexpected status code: %d, expected %d", resp.StatusCode, http.StatusAccepted)
	}
	return nil
}

// DeleteNamespace deletes all the rule groups in a namespace (including the namespace itself).
func (c *Client) DeleteNamespace(ctx context.Context, namespace string) error {
	path := c.address.JoinPath(c.prometheusPrefix, "/config/v1/rules/", namespace)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, path.String(), nil)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("unexpected status code: %d, expected %d", resp.StatusCode, http.StatusAccepted)
	}
	return nil
}
