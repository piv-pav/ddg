package tools

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	maxRetries = 3
	timeout    = 60 * time.Second
	userAgent  = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"
	acceptLang = "en-US,en;q=0.5"
	acceptHTML = "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8"
)

func newHTTPClient() *http.Client {
	return &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
	}
}

func doGet(ctx context.Context, rawURL string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", acceptHTML)
	req.Header.Set("Accept-Language", acceptLang)

	client := newHTTPClient()

	var resp *http.Response
	for attempt := 0; attempt <= maxRetries; attempt++ {
		resp, err = client.Do(req)
		if err == nil {
			break
		}
		if attempt < maxRetries {
			time.Sleep(time.Duration(1<<uint(attempt)) * time.Second)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed after %d retries: %w", maxRetries, err)
	}

	return resp, nil
}

func doGetOK(ctx context.Context, rawURL string) ([]byte, error) {
	resp, err := doGet(ctx, rawURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status: %d %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read failed: %w", err)
	}

	return body, nil
}

// doRequest performs a request with retry/backoff and returns the body of a
// 200 response.
func doRequest(ctx context.Context, req *http.Request) ([]byte, error) {
	client := newHTTPClient()

	var resp *http.Response
	var err error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		resp, err = client.Do(req)
		if err == nil {
			break
		}
		if attempt < maxRetries {
			time.Sleep(time.Duration(1<<uint(attempt)) * time.Second)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed after %d retries: %w", maxRetries, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status: %d %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read failed: %w", err)
	}

	return body, nil
}

// doGetOKWithCookies is a GET that returns the body, attaching optional cookies.
func doGetOKWithCookies(ctx context.Context, rawURL string, cookies []*http.Cookie) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", acceptHTML)
	req.Header.Set("Accept-Language", acceptLang)
	for _, c := range cookies {
		req.AddCookie(c)
	}

	return doRequest(ctx, req)
}

// doPostJSON performs a POST with a JSON body and returns the response body.
func doPostJSON(ctx context.Context, rawURL string, payload []byte) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", rawURL, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent)

	return doRequest(ctx, req)
}

func validateURL(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return fmt.Errorf("invalid URL: %s", rawURL)
	}
	return nil
}
