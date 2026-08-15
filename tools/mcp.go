package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// BuildMCPServer constructs an MCP server exposing ddg's search and read
// functionality. It reuses the same underlying functions as the CLI so
// behavior stays identical.
func BuildMCPServer(name, version string) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{Name: name, Version: version}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "web_search",
		Description: "Search DuckDuckGo and return structured results with explicit title, url, and info.",
	}, handleWebSearch)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "web_read",
		Description: "Fetch a URL and convert it to clean markdown. For YouTube URLs, returns the video transcript.",
	}, handleWebRead)

	return server
}

// webSearchInput is the input for the web_search tool.
type webSearchInput struct {
	Query string `json:"query" jsonschema:"the search query"`
	Limit int    `json:"limit,omitempty" jsonschema:"max number of results, default 10"`
}

// webSearchOutput is the structured output for the web_search tool.
type webSearchOutput struct {
	Results []SearchResult `json:"results"`
}

func handleWebSearch(ctx context.Context, _ *mcp.CallToolRequest, in webSearchInput) (*mcp.CallToolResult, webSearchOutput, error) {
	results, err := searchDuckDuckGo(ctx, in.Query)
	if err != nil {
		return nil, webSearchOutput{}, err
	}
	if in.Limit > 0 && len(results) > in.Limit {
		results = results[:in.Limit]
	}
	return nil, webSearchOutput{Results: results}, nil
}

// webReadInput is the input for the web_read tool.
type webReadInput struct {
	URL string `json:"url" jsonschema:"the URL to fetch and convert to markdown"`
}

func handleWebRead(ctx context.Context, _ *mcp.CallToolRequest, in webReadInput) (*mcp.CallToolResult, any, error) {
	content, err := readURL(ctx, in.URL)
	if err != nil {
		return nil, nil, err
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: content}},
	}, nil, nil
}
