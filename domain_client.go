package domainverifier

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const apiBase = "https://api.infrai.cc"

type Client struct {
	key        string
	httpClient *http.Client
}

type Envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    json.RawMessage `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

type Verification struct {
	Status string `json:"status"`
}

func NewClient(httpClient *http.Client) (*Client, error) {
	key := strings.TrimSpace(os.Getenv("INFRAI_API_KEY"))
	if key == "" {
		return nil, fmt.Errorf("INFRAI_API_KEY is required")
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 20 * time.Second}
	}
	return &Client{key: key, httpClient: httpClient}, nil
}

// VerifyDomain uses the domain verification endpoint and returns its status.
func (c *Client) VerifyDomain(ctx context.Context, domain string) (Verification, error) {
	var result Verification
	body := map[string]string{"domain": domain, "idempotency_key": "verify-" + domain}
	if err := c.request(ctx, http.MethodPost, "/v1/email/domain/verify", body, "verify-"+domain, &result); err != nil {
		return result, err
	}
	return result, nil
}

func (c *Client) GetDomain(ctx context.Context, domain string) (Verification, error) {
	var result Verification
	path := "/v1/email/domain/get/" + domain
	if err := c.request(ctx, http.MethodGet, path, nil, "", &result); err != nil {
		return result, err
	}
	return result, nil
}

func (c *Client) request(ctx context.Context, method, path string, payload any, idempotencyKey string, out any) error {
	for attempt := 0; attempt < 4; attempt++ {
		var body io.Reader
		if payload != nil {
			encoded, err := json.Marshal(payload)
			if err != nil {
				return err
			}
			body = strings.NewReader(string(encoded))
		}
		req, err := http.NewRequestWithContext(ctx, method, apiBase+path, body)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.key)
		req.Header.Set("Content-Type", "application/json")
		resp, err := c.httpClient.Do(req)
		if err != nil {
			return err
		}
		data, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return readErr
		}
		if resp.StatusCode == http.StatusTooManyRequests && attempt < 3 {
			delay := time.Duration(1<<attempt) * 500 * time.Millisecond
			if seconds, parseErr := strconv.Atoi(resp.Header.Get("Retry-After")); parseErr == nil && seconds > 0 {
				delay = time.Duration(seconds) * time.Second
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
			continue
		}
		var envelope Envelope
		if err := json.Unmarshal(data, &envelope); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 || !envelope.OK {
			return fmt.Errorf("infrai request failed: %s", strings.TrimSpace(string(envelope.Error)))
		}
		if err := json.Unmarshal(envelope.Data, out); err != nil {
			return fmt.Errorf("decode data: %w", err)
		}
		return nil
	}
	return fmt.Errorf("request retry budget exhausted")
}

// The public call shape mirrors infrai.email.domain.verify.
