package cli

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/dangrier/ocm-client/ocm"
	"github.com/spf13/cobra"
)

type state struct {
	client *ocm.Client
	output string
}

var s state

var rootCmd = &cobra.Command{
	Use:   "ocm",
	Short: "Query Queensland Police Service crime statistics",
	Long: `ocm queries the QPS Online Crime Map API for publicly available
crime statistics data. Not affiliated with the Queensland Police Service. See https://qps-ocm.s3-ap-southeast-2.amazonaws.com/index.html for the original web application.`,
	SilenceUsage: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&s.output, "output", "o", "table", "Output format: table, json, csv")
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		s.client = ocm.NewClient()
		return nil
	}

	rootCmd.AddCommand(locationsCmd)
	rootCmd.AddCommand(locationCmd)
	rootCmd.AddCommand(offencesCmd)
}

// clientWithTimeout returns a context with the default API timeout.
func clientWithTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 30*time.Second)
}
