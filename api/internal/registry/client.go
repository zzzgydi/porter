package registry

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/zzzgydi/porter/api/internal/registrytoken"
)

type Client struct {
	baseURL string
	signer  *registrytoken.Signer
	service string
	client  *http.Client
}

func NewClient(baseURL string, signer *registrytoken.Signer, service string) *Client {
	return &Client{
		baseURL: baseURL,
		signer:  signer,
		service: service,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) DeleteManifest(ctx context.Context, repoName, digest string) error {
	// Internal admin token for registry API calls
	token, err := c.signer.Sign(
		"admin",
		c.service,
		[]registrytoken.AccessEntry{
			{Type: "repository", Name: repoName, Actions: []string{"delete", "pull"}},
		},
		5*time.Minute,
	)
	if err != nil {
		return fmt.Errorf("sign admin token: %w", err)
	}

	url := fmt.Sprintf("%s/v2/%s/manifests/%s", c.baseURL, repoName, digest)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("registry delete request: %w", err)
	}
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("registry delete returned %d", resp.StatusCode)
	}
	return nil
}
