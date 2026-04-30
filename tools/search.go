package tools

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/spf13/cobra"
)

const (
	maxRetries = 3
	timeout    = 60 * time.Second
)

type SearchResult struct {
	Title string
	Info  string
	URL   string
}

func SearchCmd() *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "search [query]",
		Short: "Search DuckDuckGo",
		Long:  "Search DuckDuckGo and return top results with title, description, and URL",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := args[0]

			results, err := searchDuckDuckGo(context.Background(), query)
			if err != nil {
				return fmt.Errorf("search failed: %w", err)
			}

			// Limit results
			if limit > 0 && len(results) > limit {
				results = results[:limit]
			}

			// Output results
			for _, r := range results {
				fmt.Printf("%s\n%s\n%s\n\n", r.Title, r.Info, r.URL)
			}

			return nil
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "l", 10, "Max number of results")

	return cmd
}

func searchDuckDuckGo(ctx context.Context, query string) ([]SearchResult, error) {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	}

	client := &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}

	searchURL := fmt.Sprintf("https://html.duckduckgo.com/html/?q=%s", url.QueryEscape(query))

	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
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
		return nil, fmt.Errorf("failed after %d retries: %w", maxRetries, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("status: %d %s", resp.StatusCode, resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parse failed: %w", err)
	}

	var results []SearchResult
	doc.Find(".web-result").Each(func(i int, s *goquery.Selection) {
		titleNode := s.Find(".result__a")
		title := strings.TrimSpace(titleNode.Text())
		info := strings.TrimSpace(s.Find(".result__snippet").Text())

		var resultURL string
		if titleNode.Length() > 0 {
			if href, exists := titleNode.Attr("href"); exists {
				if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
					resultURL = href
				}
			}
		}

		if title != "" && resultURL != "" {
			results = append(results, SearchResult{
				Title: title,
				Info:  info,
				URL:   resultURL,
			})
		}
	})

	return results, nil
}
