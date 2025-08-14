package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/david/helpme/internal/config"
	"github.com/david/helpme/internal/data"
	"github.com/david/helpme/internal/ui"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	locationFlag string
	cfg          *config.Config
)

var rootCmd = &cobra.Command{
	Use:   "helpme",
	Short: "Emergency and crisis helpline directory",
	Long:  "A location-aware command-line tool for accessing emergency and crisis helplines",
	RunE: func(cmd *cobra.Command, args []string) error {
		location := getActiveLocation()
		return ui.ShowMainMenu(location)
	},
}

var emergencyCmd = &cobra.Command{
	Use:   "emergency",
	Short: "Show emergency numbers for your location",
	RunE: func(cmd *cobra.Command, args []string) error {
		location := getActiveLocation()
		loc, err := data.GetLocation(location)
		if err != nil {
			return err
		}

		color.Red("\n🚨 EMERGENCY SERVICES - %s 🚨\n", loc.Name)
		fmt.Println(strings.Repeat("=", 50))
		
		for name, service := range loc.Emergency {
			color.New(color.FgYellow, color.Bold).Printf("%-20s: ", strings.Title(strings.ReplaceAll(name, "_", " ")))
			color.New(color.FgGreen, color.Bold).Printf("%s\n", service.Number)
			fmt.Printf("   %s\n", service.Description)
		}
		fmt.Println()
		return nil
	},
}

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search for specific help resources",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		location := getActiveLocation()
		loc, err := data.GetLocation(location)
		if err != nil {
			return err
		}

		query := strings.Join(args, " ")
		results := data.SearchResources(loc, query)
		
		if len(results) == 0 {
			fmt.Printf("No resources found for '%s' in %s\n", query, loc.Name)
			return nil
		}

		color.New(color.FgCyan, color.Bold).Printf("\nSearch Results for '%s' in %s:\n", query, loc.Name)
		fmt.Println(strings.Repeat("=", 50))
		
		for _, resource := range results {
			ui.DisplayResource(resource)
		}
		
		return nil
	},
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
}

var setLocationCmd = &cobra.Command{
	Use:   "set-location [location]",
	Short: "Set your default location",
	Long:  fmt.Sprintf("Set your default location. Available: %s", strings.Join(data.GetAvailableLocations(), ", ")),
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		location := strings.ToUpper(args[0])
		
		if _, err := data.GetLocation(location); err != nil {
			return fmt.Errorf("invalid location '%s'. Available: %s", 
				location, strings.Join(data.GetAvailableLocations(), ", "))
		}
		
		cfg.Location = location
		if err := cfg.Save(); err != nil {
			return err
		}
		
		fmt.Printf("Default location set to %s\n", location)
		return nil
	},
}

var listLocationsCmd = &cobra.Command{
	Use:   "list-locations",
	Short: "List all available locations",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Available locations:")
		for code, loc := range data.Data.Locations {
			fmt.Printf("  %s - %s\n", code, loc.Name)
		}
	},
}

func getActiveLocation() string {
	if locationFlag != "" {
		return locationFlag
	}
	return cfg.Location
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&locationFlag, "location", "l", "", "Override default location")
	
	configCmd.AddCommand(setLocationCmd)
	rootCmd.AddCommand(emergencyCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(listLocationsCmd)
}

func main() {
	var err error
	cfg, err = config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}