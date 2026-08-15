package tools

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"codeberg.org/readeck/go-readability/v2"
	md "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/spf13/cobra"
)

func ReadCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:     "read [url]",
		Aliases: []string{"fetch"},
		Short:   "Fetch and convert URL to markdown",
		Long:    "Fetch webpage, clean with go-readability, convert to markdown",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			targetURL := args[0]

			if jsonOutput {
				if videoID, ok := parseYouTubeID(targetURL); ok {
					info, err := fetchYouTube(cmd.Context(), videoID)
					if err != nil {
						return fmt.Errorf("could not retrieve transcript: %w", err)
					}
					return printJSON(youtubeJSON{
						URL:        targetURL,
						Title:      info.Title,
						Channel:    info.Channel,
						Duration:   info.Duration,
						Date:       info.Date,
						Transcript: info.Transcript,
					})
				}
			}

			content, err := readURL(cmd.Context(), targetURL)
			if err != nil {
				return err
			}

			if jsonOutput {
				return printJSON(map[string]string{"url": targetURL, "content": content})
			}

			fmt.Println(content)
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	return cmd
}

// readURL returns the markdown for a URL: YouTube videos use the transcript
// path, everything else uses readability conversion. Shared by the CLI and the
// MCP server.
func readURL(ctx context.Context, targetURL string) (string, error) {
	if videoID, ok := parseYouTubeID(targetURL); ok {
		info, err := fetchYouTube(ctx, videoID)
		if err != nil {
			return "", fmt.Errorf("could not retrieve transcript: %w", err)
		}
		return formatYouTubeMarkdown(info), nil
	}

	content, err := fetchAndConvert(ctx, targetURL)
	if err != nil {
		return "", fmt.Errorf("fetch failed: %w", err)
	}
	return content, nil
}

func fetchAndConvert(ctx context.Context, targetURL string) (string, error) {
	if err := validateURL(targetURL); err != nil {
		return "", err
	}

	body, err := doGetOK(ctx, targetURL)
	if err != nil {
		return "", err
	}

	parsedURL, _ := url.Parse(targetURL)

	// Clean with go-readability
	article, err := readability.FromReader(strings.NewReader(string(body)), parsedURL)
	if err != nil {
		return md.ConvertString(string(body))
	}

	// Render cleaned HTML
	var htmlBuilder strings.Builder
	if err := article.RenderHTML(&htmlBuilder); err != nil {
		return md.ConvertString(string(body))
	}
	htmlContent := htmlBuilder.String()

	// Convert to markdown
	markdownContent, err := md.ConvertString(htmlContent)
	if err != nil {
		return "", fmt.Errorf("markdown conversion failed: %w", err)
	}

	// Fallback: if readability extracted very little (<2% ratio), use raw HTML
	if len(body) > 10000 {
		extractionRatio := float64(len(htmlContent)) / float64(len(body))
		if extractionRatio < 0.02 {
			markdownContent, err = md.ConvertString(string(body))
			if err != nil {
				return "", fmt.Errorf("markdown conversion (fallback) failed: %w", err)
			}
		}
	}

	return markdownContent, nil
}
