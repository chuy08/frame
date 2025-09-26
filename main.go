package main

import (
	"fmt"
	"os"

	"frame/config"
	"frame/hello"
	"frame/logging"
	"frame/server"
	"frame/version"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "frame",
		Short: "A simple CLI that prints hello world",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Load configuration before any command runs
			cfg, err := config.Load()
			if err != nil {
				fmt.Printf("Error loading config: %v\n", err)
				os.Exit(1)
			}
			// Store config in viper for use by other packages
			viper.Set("config", cfg)

			// Initialize logging before watching config
			if err := logging.Initialize(); err != nil {
				fmt.Printf("Error initializing logging: %v\n", err)
				os.Exit(1)
			}

			// Start watching config file for changes
			config.WatchConfig()
		},
		Run: func(cmd *cobra.Command, args []string) {
			hello.Print()
		},
	}

	// Add command line flags
	rootCmd.PersistentFlags().String("db-host", "", "Database host")
	rootCmd.PersistentFlags().Int("db-port", 0, "Database port")
	rootCmd.PersistentFlags().String("db-user", "", "Database user")
	rootCmd.PersistentFlags().String("db-password", "", "Database password")
	rootCmd.PersistentFlags().String("db-name", "", "Database name")
	rootCmd.PersistentFlags().Int("server-port", 0, "Server port")

	// Bind flags to viper
	bindFlags := []struct {
		key      string
		flagName string
	}{
		{"database.host", "db-host"},
		{"database.port", "db-port"},
		{"database.user", "db-user"},
		{"database.password", "db-password"},
		{"database.name", "db-name"},
		{"server.port", "server-port"},
	}

	for _, flag := range bindFlags {
		if err := viper.BindPFlag(flag.key, rootCmd.PersistentFlags().Lookup(flag.flagName)); err != nil {
			fmt.Printf("Error binding flag %s: %v\n", flag.flagName, err)
			os.Exit(1)
		}
	}

	// Add serve command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "serve",
		Short: "Start HTTP API server",
		Run: func(cmd *cobra.Command, args []string) {
			if err := server.Start(); err != nil {
				fmt.Println("Server error:", err)
			}
		},
	})

	// Add version command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Version    : %s\n", version.Version)
			fmt.Printf("Git Commit : %s\n", version.GitCommit)
			fmt.Printf("Git Branch : %s\n", version.GitBranch)
			fmt.Printf("Build Time : %s\n", version.BuildTime)
		},
	})

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
