package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/spf13/cobra"
)

type SearchResult struct {
	Title string `json:"title"`
	Info  string `json:"info"`
	URL   string `json:"url"`
}

func SearchCmd() *cobra.Command {
	var limit int
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "search [query]",
		Short: "Search DuckDuckGo",
		Long:  "Search DuckDuckGo and return top results with title, description, and URL",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := args[0]

			results, err := searchDuckDuckGo(cmd.Context(), query)
			if err != nil {
				return fmt.Errorf("search failed: %w", err)
			}

			if limit > 0 && len(results) > limit {
				results = results[:limit]
			}

			if jsonOutput {
				return printJSON(results)
			}

			for i, r := range results {
				if i > 0 {
					fmt.Printf("\n---\n\n")
				}
				fmt.Printf("# [%s](%s)\n\n%s\n", r.Title, r.URL, r.Info)
			}

			return nil
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "l", 10, "Max number of results")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	return cmd
}

func searchDuckDuckGo(ctx context.Context, query string) ([]SearchResult, error) {
	searchURL := fmt.Sprintf("https://html.duckduckgo.com/html/?q=%s", url.QueryEscape(query))

	resp, err := doGet(ctx, searchURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 202 {
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

func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
