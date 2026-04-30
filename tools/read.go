package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-shiori/go-readability"
	md "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/spf13/cobra"
)

func ReadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "read [url]",
		Short: "Fetch and convert URL to markdown",
		Long:  "Fetch webpage, clean with go-readability, convert to markdown",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			targetURL := args[0]

			content, err := fetchAndConvert(context.Background(), targetURL)
			if err != nil {
				return fmt.Errorf("fetch failed: %w", err)
			}

			fmt.Println(content)
			return nil
		},
	}

	return cmd
}

func fetchAndConvert(ctx context.Context, targetURL string) (string, error) {
	parsedURL, err := url.Parse(targetURL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return "", fmt.Errorf("invalid URL: %s", targetURL)
	}

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	}

	client := &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
	if err != nil {
		return "", fmt.Errorf("request creation failed: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")

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
		return "", fmt.Errorf("failed after %d retries: %w", maxRetries, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read failed: %w", err)
	}

	// Clean with go-readability
	article, err := readability.FromReader(strings.NewReader(string(body)), parsedURL)
	if err != nil {
		// If readability fails, convert raw HTML directly
		markdownContent, err := md.ConvertString(string(body))
		if err != nil {
			return "", fmt.Errorf("markdown conversion failed: %w", err)
		}
		return markdownContent, nil
	}

	// Convert cleaned content to markdown
	markdownContent, err := md.ConvertString(article.Content)
	if err != nil {
		return "", fmt.Errorf("markdown conversion failed: %w", err)
	}

	// If readability extracted very little compared to input (< 2% ratio),
	// it likely failed to find content. Fall back to raw HTML conversion.
	// This handles JS-heavy sites while preserving small valid responses.
	extractionRatio := float64(len(article.Content)) / float64(len(body))
	if len(body) > 10000 && extractionRatio < 0.02 {
		markdownContent, err = md.ConvertString(string(body))
		if err != nil {
			return "", fmt.Errorf("markdown conversion (fallback) failed: %w", err)
		}
	}

	return markdownContent, nil
}
