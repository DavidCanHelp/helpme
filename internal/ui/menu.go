package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/david/helpme/internal/data"
	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
)

func ShowMainMenu(locationCode string) error {
	location, err := data.GetLocation(locationCode)
	if err != nil {
		return err
	}

	for {
		color.New(color.FgCyan, color.Bold).Printf("\n📍 Location: %s\n", location.Name)
		fmt.Println(strings.Repeat("=", 50))

		categories := []string{
			"🚨 Emergency Services",
			"🧠 Mental Health & Crisis Support",
			"☠️  Poison Control",
			"🏠 Domestic Violence",
			"👶 Child Abuse",
			"🤝 Sexual Assault",
			"🏥 Substance Abuse",
			"🌈 LGBTQ Support",
			"🎖️  Veterans Support",
			"🔍 Search Resources",
			"❌ Exit",
		}

		prompt := promptui.Select{
			Label: "Select a category",
			Items: categories,
			Size:  11,
		}

		idx, _, err := prompt.Run()
		if err != nil {
			if err == promptui.ErrInterrupt {
				return nil
			}
			return err
		}

		switch idx {
		case 0:
			showEmergencyServices(location)
		case 1:
			showResources(location, "mental_health", "Mental Health & Crisis Support")
		case 2:
			showResources(location, "poison_control", "Poison Control")
		case 3:
			showResources(location, "domestic_violence", "Domestic Violence")
		case 4:
			showResources(location, "child_abuse", "Child Abuse")
		case 5:
			showResources(location, "sexual_assault", "Sexual Assault")
		case 6:
			showResources(location, "substance_abuse", "Substance Abuse")
		case 7:
			showResources(location, "lgbtq", "LGBTQ Support")
		case 8:
			showResources(location, "veterans", "Veterans Support")
		case 9:
			searchInteractive(location)
		case 10:
			return nil
		}
	}
}

func showEmergencyServices(location *data.Location) {
	color.Red("\n🚨 EMERGENCY SERVICES - %s 🚨\n", location.Name)
	fmt.Println(strings.Repeat("=", 50))
	
	for name, service := range location.Emergency {
		displayName := strings.Title(strings.ReplaceAll(name, "_", " "))
		color.New(color.FgYellow, color.Bold).Printf("%-25s: ", displayName)
		color.New(color.FgGreen, color.Bold).Printf("%s\n", service.Number)
		fmt.Printf("   %s\n\n", service.Description)
	}
	
	waitForEnter()
}

func showResources(location *data.Location, category, title string) {
	resources, ok := location.Resources[category]
	if !ok || len(resources) == 0 {
		fmt.Printf("\nNo %s resources available for %s\n", title, location.Name)
		if location.Note != "" {
			fmt.Printf("\nNote: %s\n", location.Note)
		}
		waitForEnter()
		return
	}

	color.New(color.FgYellow, color.Bold).Printf("\n%s - %s\n", title, location.Name)
	fmt.Println(strings.Repeat("=", 50))
	
	for _, resource := range resources {
		DisplayResource(resource)
	}
	
	waitForEnter()
}

func DisplayResource(resource data.Resource) {
	color.New(color.FgCyan, color.Bold).Printf("\n%s\n", resource.Name)
	if resource.Number != "" {
		color.New(color.FgGreen, color.Bold).Printf("📞 Phone: %s\n", resource.Number)
	}
	if resource.Text != "" {
		color.New(color.FgGreen).Printf("💬 Text: %s\n", resource.Text)
	}
	if resource.SMS != "" {
		color.New(color.FgGreen).Printf("📱 SMS: %s\n", resource.SMS)
	}
	fmt.Printf("📝 %s\n", resource.Description)
	if resource.Website != "" {
		color.New(color.FgBlue).Printf("🌐 %s\n", resource.Website)
	}
}

func searchInteractive(location *data.Location) {
	prompt := promptui.Prompt{
		Label: "Search for resources",
	}

	query, err := prompt.Run()
	if err != nil {
		return
	}

	results := data.SearchResources(location, query)
	
	if len(results) == 0 {
		fmt.Printf("\nNo resources found for '%s' in %s\n", query, location.Name)
	} else {
		color.New(color.FgCyan, color.Bold).Printf("\nSearch Results for '%s':\n", query)
		fmt.Println(strings.Repeat("=", 50))
		
		for _, resource := range results {
			DisplayResource(resource)
		}
	}
	
	waitForEnter()
}

func waitForEnter() {
	fmt.Printf("\nPress Enter to continue...")
	fmt.Scanln()
}

func ClearScreen() {
	fmt.Print("\033[H\033[2J")
}

func init() {
	if os.Getenv("TERM") == "" {
		os.Setenv("TERM", "xterm")
	}
}