package main

import (
	"fmt"
	"os"

	"frame/config"
	"frame/hello"
	"frame/logging"
	"frame/server"

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
	viper.BindPFlag("database.host", rootCmd.PersistentFlags().Lookup("db-host"))
	viper.BindPFlag("database.port", rootCmd.PersistentFlags().Lookup("db-port"))
	viper.BindPFlag("database.user", rootCmd.PersistentFlags().Lookup("db-user"))
	viper.BindPFlag("database.password", rootCmd.PersistentFlags().Lookup("db-password"))
	viper.BindPFlag("database.name", rootCmd.PersistentFlags().Lookup("db-name"))
	viper.BindPFlag("server.port", rootCmd.PersistentFlags().Lookup("server-port"))

	rootCmd.AddCommand(&cobra.Command{
		Use:   "serve",
		Short: "Start HTTP API server",
		Run: func(cmd *cobra.Command, args []string) {
			if err := server.Start(); err != nil {
				fmt.Println("Server error:", err)
			}
		},
	})

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
