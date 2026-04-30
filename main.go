package main

import (
	"fmt"
	"os"

	"github.com/ppivtorak/ddg/tools"
	"github.com/spf13/cobra"
)

var version = "dev"

func main() {
	rootCmd := &cobra.Command{
		Use:     "ddg",
		Short:   "DuckDuckGo search and web fetch CLI",
		Version: version,
	}

	rootCmd.AddCommand(tools.SearchCmd())
	rootCmd.AddCommand(tools.ReadCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
