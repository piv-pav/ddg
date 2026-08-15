package main

import (
	"fmt"
	"net/http"
	"os"
	"runtime/debug"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/piv-pav/ddg/tools"
	"github.com/spf13/cobra"
)

var version = "dev"

func init() {
	if version != "dev" {
		return // ldflags already set the version (e.g. v0.5.0-dev)
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		version = info.Main.Version
	}
}

func main() {
	var mcpMode bool
	var mcpHTTPPort string

	rootCmd := &cobra.Command{
		Use:     "ddg",
		Short:   "DuckDuckGo search and web fetch CLI",
		Version: version,
		RunE: func(cmd *cobra.Command, args []string) error {
			if mcpHTTPPort == "" && !mcpMode {
				return cmd.Help()
			}

			server := tools.BuildMCPServer("ddg", version)

			if mcpHTTPPort != "" {
				addr := mcpHTTPPort
				if !strings.HasPrefix(addr, ":") {
					addr = ":" + addr
				}
				handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return server }, nil)
				return http.ListenAndServe(addr, handler)
			}

			return server.Run(cmd.Context(), &mcp.StdioTransport{})
		},
	}

	rootCmd.Flags().BoolVar(&mcpMode, "mcp", false, "Run as an MCP server over stdio")
	rootCmd.Flags().StringVar(&mcpHTTPPort, "mcp-http", "", "Run as a StreamableHTTP MCP server on the given port")

	rootCmd.AddCommand(tools.SearchCmd())
	rootCmd.AddCommand(tools.ReadCmd())
	rootCmd.AddCommand(tools.UpgradeCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
