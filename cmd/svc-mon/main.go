// Binary svc-mon
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jzer7/svc-mon/internal/core"

	"github.com/spf13/cobra"
)

var (
	version = "1.0.0"
)

func main() {

	// The root command (i.e., svc-mon)
	var rootCmd = &cobra.Command{
		Use:   "svc-mon",
		Short: "Service Monitor",
		Long:  `svc-mon is a tool to monitor the health of various network services.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Initialize the application. This runs before any subcommand.
		},
	}

	// The monitor command to start monitoring services (i.e., svc-mon monitor)
	var monitorCmd = &cobra.Command{
		Use:   "monitor",
		Short: "Start monitoring services",
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, _ := cmd.Flags().GetString("config")
			return core.Run(configPath)
		},
	}

	// The version command to print the version (i.e., svc-mon version)
	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("svc-mon version %s\n", version)
		},
	}

	rootCmd.AddCommand(monitorCmd)
	rootCmd.AddCommand(versionCmd)

	monitorCmd.Flags().StringP("config", "c", "config.yaml", "Path to configuration file")
	monitorCmd.Flags().BoolP("dry-run", "d", false, "Run in dry-run to validate configuration without monitoring")
	monitorCmd.Flags().BoolP("verbose", "v", false, "Enable verbose logging")

	if err := rootCmd.Execute(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
