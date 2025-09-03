package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/david/helpme/internal/clipboard"
	"github.com/david/helpme/internal/config"
	"github.com/david/helpme/internal/contacts"
	"github.com/david/helpme/internal/data"
	"github.com/david/helpme/internal/local"
	"github.com/david/helpme/internal/ui"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	locationFlag string
	cfg          *config.Config
)

var rootCmd = &cobra.Command{
	Use:   "helpme [query]",
	Short: "Emergency and crisis helpline directory",
	Long:  `A location-aware command-line tool for accessing emergency and crisis helplines.
	
Quick access:
  helpme 911        Show emergency numbers immediately
  helpme emergency  Show emergency numbers immediately  
  helpme 10001      Show local non-emergency for ZIP code
  helpme suicide    Search for suicide prevention resources
  helpme [query]    Search for specific resources`,
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if -e flag is set
		if emergency, _ := cmd.Flags().GetBool("emergency"); emergency {
			return showQuickEmergency(getActiveLocation())
		}
		
		// Quick emergency access: "helpme 911" or "helpme emergency"
		if len(args) > 0 {
			// Check if it's a ZIP code first
			if local.IsZipCode(args[0]) {
				return showLocalNonEmergency(args[0])
			}
			
			// Special SOS command shows personal emergency info
			if strings.EqualFold(args[0], "sos") {
				return showSOSInfo()
			}
			
			emergencyKeywords := []string{"911", "999", "112", "emergency", "help"}
			for _, keyword := range emergencyKeywords {
				if strings.EqualFold(args[0], keyword) {
					return showQuickEmergency(getActiveLocation())
				}
			}
			// If not emergency, treat as search query
			return quickSearch(getActiveLocation(), strings.Join(args, " "))
		}
		
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
		regions := map[string][]string{
			"North America": {"US", "CA"},
			"Europe": {"UK", "IE", "DE", "FR", "ES", "IT", "NL", "BE", "CH", "AT", 
			          "SE", "NO", "DK", "FI", "IS", "PT", "LU", "GR", "EU"},
			"Asia-Pacific": {"JP", "KR", "SG", "AU", "NZ"},
			"Middle East": {"IL", "AE"},
			"Africa": {"ZA"},
		}
		
		fmt.Println("Available locations by region:\n")
		for region, codes := range regions {
			color.New(color.FgCyan, color.Bold).Printf("%s:\n", region)
			for _, code := range codes {
				if loc, ok := data.Data.Locations[code]; ok {
					fmt.Printf("  %s - %s\n", code, loc.Name)
				}
			}
			fmt.Println()
		}
	},
}

func getActiveLocation() string {
	if locationFlag != "" {
		return locationFlag
	}
	return cfg.Location
}

func showQuickEmergency(locationCode string) error {
	location, err := data.GetLocation(locationCode)
	if err != nil {
		return err
	}

	// Clear screen for visibility
	fmt.Print("\033[H\033[2J")
	
	// Big, clear emergency header
	color.New(color.FgRed, color.Bold).Printf("\n🚨🚨🚨 EMERGENCY NUMBERS - %s 🚨🚨🚨\n\n", strings.ToUpper(location.Name))
	
	// Show all emergency services with large, clear formatting
	for name, service := range location.Emergency {
		displayName := strings.ToUpper(strings.ReplaceAll(name, "_", " "))
		color.New(color.FgYellow, color.Bold).Printf("%-20s ", displayName+":")
		color.New(color.FgGreen, color.Bold).Printf("%s\n", service.Number)
		color.New(color.FgWhite).Printf("                     %s\n\n", service.Description)
	}
	
	// Also show critical mental health line
	if resources, ok := location.Resources["mental_health"]; ok && len(resources) > 0 {
		color.New(color.FgCyan, color.Bold).Printf("\n🧠 CRISIS SUPPORT:\n")
		for i, resource := range resources {
			if i >= 2 { break } // Show top 2 crisis lines
			color.New(color.FgYellow, color.Bold).Printf("%-20s ", resource.Name+":")
			color.New(color.FgGreen, color.Bold).Printf("%s\n", resource.Number)
			if resource.Text != "" {
				color.New(color.FgWhite).Printf("                     Text: %s\n", resource.Text)
			}
		}
	}
	
	fmt.Println("\n" + strings.Repeat("=", 60))
	color.New(color.FgWhite).Printf("Location: %s | Change with: helpme config set-location [CODE]\n", location.Name)
	return nil
}

func quickSearch(locationCode string, query string) error {
	location, err := data.GetLocation(locationCode)
	if err != nil {
		return err
	}

	results := data.SearchResources(location, query)
	
	if len(results) == 0 {
		fmt.Printf("No resources found for '%s' in %s\n", query, location.Name)
		fmt.Println("\nTry: helpme (for menu) or helpme emergency (for emergency numbers)")
		return nil
	}

	color.New(color.FgCyan, color.Bold).Printf("\nResults for '%s' in %s:\n", query, location.Name)
	fmt.Println(strings.Repeat("=", 50))
	
	for _, resource := range results {
		ui.DisplayResource(resource)
	}
	
	return nil
}

func showSOSInfo() error {
	// Clear screen for maximum visibility
	fmt.Print("\033[H\033[2J")
	
	// Show emergency numbers first
	color.New(color.FgRed, color.Bold).Printf("\n🆘🆘🆘 SOS - EMERGENCY INFORMATION 🆘🆘🆘\n\n")
	
	location := getActiveLocation()
	loc, _ := data.GetLocation(location)
	
	// Emergency services
	color.New(color.FgRed, color.Bold).Println("📞 EMERGENCY: 911")
	fmt.Println()
	
	// Load and show personal emergency contacts
	profile, err := contacts.LoadProfile()
	if err == nil && len(profile.Contacts) > 0 {
		color.New(color.FgYellow, color.Bold).Println("👥 MY EMERGENCY CONTACTS:")
		for i, contact := range profile.Contacts {
			if i >= 3 { break } // Show top 3 for quick access
			fmt.Printf("   %s (%s): %s\n", contact.Name, contact.Relationship, contact.Phone)
		}
		fmt.Println()
	}
	
	// Medical info if available
	if err == nil && (profile.MedicalInfo.BloodType != "" || len(profile.MedicalInfo.Allergies) > 0) {
		color.New(color.FgRed, color.Bold).Println("🏥 MEDICAL INFO:")
		if profile.MedicalInfo.BloodType != "" {
			fmt.Printf("   Blood Type: %s\n", profile.MedicalInfo.BloodType)
		}
		if len(profile.MedicalInfo.Allergies) > 0 {
			fmt.Printf("   ⚠️  Allergies: %s\n", strings.Join(profile.MedicalInfo.Allergies, ", "))
		}
		if len(profile.MedicalInfo.Medications) > 0 {
			fmt.Printf("   💊 Medications: %s\n", strings.Join(profile.MedicalInfo.Medications, ", "))
		}
		if profile.MedicalInfo.DoctorName != "" {
			fmt.Printf("   👨‍⚕️ Doctor: %s (%s)\n", profile.MedicalInfo.DoctorName, profile.MedicalInfo.DoctorPhone)
		}
		fmt.Println()
	}
	
	// Crisis lines
	color.New(color.FgCyan, color.Bold).Println("🧠 CRISIS SUPPORT:")
	fmt.Println("   Suicide & Crisis: 988")
	fmt.Println("   Poison Control: 1-800-222-1222")
	if loc != nil && loc.Resources != nil {
		if dv := loc.Resources["domestic_violence"]; len(dv) > 0 {
			fmt.Printf("   Domestic Violence: %s\n", dv[0].Number)
		}
	}
	
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Press 's' to send SOS text (if configured), 'c' to copy info, or Enter to exit")
	
	var input string
	fmt.Scanln(&input)
	
	switch strings.ToLower(input) {
	case "c":
		// Copy all emergency info to clipboard
		var info []string
		info = append(info, "EMERGENCY: 911")
		if profile != nil && len(profile.Contacts) > 0 {
			info = append(info, "\nEMERGENCY CONTACTS:")
			for _, contact := range profile.Contacts {
				info = append(info, fmt.Sprintf("%s (%s): %s", contact.Name, contact.Relationship, contact.Phone))
			}
		}
		if err := clipboard.Copy(strings.Join(info, "\n")); err == nil {
			color.New(color.FgGreen).Println("✓ Emergency info copied to clipboard")
		}
	case "s":
		fmt.Println("SMS gateway support coming soon...")
		// Future: Implement SMS sending via Twilio or similar
	}
	
	return nil
}

func showLocalNonEmergency(zipCode string) error {
	// Clear screen for visibility
	fmt.Print("\033[H\033[2J")
	
	color.New(color.FgBlue, color.Bold).Printf("\n📞 LOCAL NON-EMERGENCY NUMBERS\n")
	fmt.Println(strings.Repeat("=", 50))
	
	// First try local database for speed
	resources, zipInfo, err := local.GetLocalResources(zipCode)
	if err != nil {
		// Try comprehensive online lookup
		fmt.Println("Searching for ZIP code information...")
		
		onlineInfo, services, onlineErr := local.GetComprehensiveLocalInfo(zipCode)
		if onlineErr != nil {
			// Show helpful error message
			color.New(color.FgYellow).Printf("\n⚠️  Unable to find information for ZIP code %s\n", zipCode)
			fmt.Println("\nPlease verify the ZIP code is correct.")
			fmt.Println("For emergency services, always call 911")
			fmt.Println("\nYou can also try:")
			fmt.Println("  • Call 411 for local directory assistance")
			fmt.Println("  • Search online for '[city name] non-emergency number'")
			
			return nil
		}
		
		// Display online results
		displayOnlineResults(onlineInfo, services)
		return nil
	}
	
	// Display the formatted local resources
	fmt.Println(local.FormatLocalResources(resources, zipInfo))
	
	// Offer clipboard support
	fmt.Println("\nPress 'c' to copy all numbers, or Enter to continue...")
	var input string
	fmt.Scanln(&input)
	if strings.ToLower(input) == "c" {
		clipboardText := fmt.Sprintf("%s, %s Non-Emergency Numbers\n", zipInfo.City, zipInfo.State)
		clipboardText += fmt.Sprintf("Police: %s\n", resources.PoliceNonEmergency)
		clipboardText += fmt.Sprintf("Fire: %s\n", resources.FireNonEmergency)
		if resources.CityServices != "" {
			clipboardText += fmt.Sprintf("City Services: %s\n", resources.CityServices)
		}
		
		if err := clipboard.Copy(clipboardText); err == nil {
			color.New(color.FgGreen).Println("✓ Copied non-emergency numbers to clipboard")
		} else {
			fmt.Println("Could not copy to clipboard. Please copy manually.")
		}
	}
	
	return nil
}

func displayOnlineResults(info *local.OnlineZipInfo, services []local.LocalService) {
	color.New(color.FgGreen).Printf("\n📍 %s, %s", info.City, info.State)
	if info.County != "" {
		fmt.Printf(" (%s County)", info.County)
	}
	fmt.Println()
	fmt.Println(strings.Repeat("=", 50))
	
	if len(services) > 0 {
		// Group services by type
		police := []local.LocalService{}
		fire := []local.LocalService{}
		hospitals := []local.LocalService{}
		other := []local.LocalService{}
		
		for _, svc := range services {
			switch svc.Type {
			case "police":
				police = append(police, svc)
			case "fire":
				fire = append(fire, svc)
			case "hospital":
				hospitals = append(hospitals, svc)
			default:
				other = append(other, svc)
			}
		}
		
		if len(police) > 0 {
			fmt.Println("\n🚔 POLICE (Non-Emergency)")
			for _, svc := range police {
				fmt.Printf("   %s: %s\n", svc.Name, svc.Phone)
			}
		}
		
		if len(fire) > 0 {
			fmt.Println("\n🚒 FIRE (Non-Emergency)")
			for _, svc := range fire {
				fmt.Printf("   %s: %s\n", svc.Name, svc.Phone)
			}
		}
		
		if len(hospitals) > 0 {
			fmt.Println("\n🏥 HOSPITALS")
			for _, svc := range hospitals {
				fmt.Printf("   %s: %s\n", svc.Name, svc.Phone)
			}
		}
		
		if len(other) > 0 {
			fmt.Println("\n📞 OTHER SERVICES")
			for _, svc := range other {
				fmt.Printf("   %s: %s\n", svc.Name, svc.Phone)
			}
		}
	} else {
		// Show generic guidance
		fmt.Printf("\n📍 Location confirmed: %s, %s\n", info.City, info.State)
		fmt.Println("\nFor local non-emergency numbers:")
		fmt.Printf("  • Police: Call 411 or search '%s police non-emergency'\n", info.City)
		fmt.Printf("  • Fire: Call 411 or search '%s fire non-emergency'\n", info.City)
		fmt.Println("  • City Services: Many cities use 311")
	}
	
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("⚠️  For emergencies, always call 911")
}

// Add mental health command shortcut
var mentalCmd = &cobra.Command{
	Use:     "mental",
	Aliases: []string{"m", "crisis", "suicide"},
	Short:   "Quick access to mental health and crisis support",
	RunE: func(cmd *cobra.Command, args []string) error {
		location := getActiveLocation()
		loc, err := data.GetLocation(location)
		if err != nil {
			return err
		}
		return showCategoryResources(loc, "mental_health", "🧠 Mental Health & Crisis Support")
	},
}

// Add poison control command shortcut
var poisonCmd = &cobra.Command{
	Use:     "poison",
	Aliases: []string{"p"},
	Short:   "Quick access to poison control",
	RunE: func(cmd *cobra.Command, args []string) error {
		location := getActiveLocation()
		loc, err := data.GetLocation(location)
		if err != nil {
			return err
		}
		return showCategoryResources(loc, "poison_control", "☠️ Poison Control")
	},
}

// Add domestic violence command shortcut
var domesticCmd = &cobra.Command{
	Use:     "domestic",
	Aliases: []string{"dv", "violence"},
	Short:   "Quick access to domestic violence resources",
	RunE: func(cmd *cobra.Command, args []string) error {
		location := getActiveLocation()
		loc, err := data.GetLocation(location)
		if err != nil {
			return err
		}
		return showCategoryResources(loc, "domestic_violence", "🏠 Domestic Violence Support")
	},
}

// Add ZIP code lookup command
var zipCmd = &cobra.Command{
	Use:     "zip [zipcode]",
	Aliases: []string{"local", "nonemergency"},
	Short:   "Get local non-emergency numbers for a US ZIP code",
	Long: `Get local non-emergency numbers for a US ZIP code.
	
Examples:
  helpme zip 10001     # New York
  helpme zip 90210     # Beverly Hills  
  helpme zip 60601     # Chicago
  helpme 10001         # Also works directly`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return showLocalNonEmergency(args[0])
	},
}

// Contacts management commands
var contactsCmd = &cobra.Command{
	Use:   "contacts",
	Short: "Manage emergency contacts",
	Long:  "Store and manage your personal emergency contacts for quick access",
}

var contactsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all emergency contacts",
	RunE: func(cmd *cobra.Command, args []string) error {
		profile, err := contacts.LoadProfile()
		if err != nil {
			return err
		}
		
		if len(profile.Contacts) == 0 {
			fmt.Println("No emergency contacts saved.")
			fmt.Println("\nAdd a contact with: helpme contacts add")
			return nil
		}
		
		color.New(color.FgYellow, color.Bold).Printf("\n👥 Emergency Contacts\n")
		fmt.Println(strings.Repeat("=", 50))
		
		for _, contact := range profile.Contacts {
			fmt.Println(contacts.FormatContact(contact))
			fmt.Println()
		}
		
		if profile.MedicalInfo.BloodType != "" || len(profile.MedicalInfo.Allergies) > 0 {
			color.New(color.FgRed, color.Bold).Printf("\n🏥 Medical Information\n")
			fmt.Println(strings.Repeat("=", 50))
			fmt.Println(contacts.FormatMedicalInfo(profile.MedicalInfo))
		}
		
		return nil
	},
}

var contactsAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new emergency contact",
	RunE: func(cmd *cobra.Command, args []string) error {
		profile, err := contacts.LoadProfile()
		if err != nil {
			return err
		}
		
		// Interactive prompts for contact info
		fmt.Print("Contact name: ")
		var name string
		fmt.Scanln(&name)
		
		fmt.Print("Relationship (e.g., spouse, parent, doctor): ")
		var relationship string
		fmt.Scanln(&relationship)
		
		fmt.Print("Phone number: ")
		var phone string
		fmt.Scanln(&phone)
		
		fmt.Print("Email (optional): ")
		var email string
		fmt.Scanln(&email)
		
		fmt.Print("Is this a medical contact? (y/n): ")
		var medicalStr string
		fmt.Scanln(&medicalStr)
		medical := strings.ToLower(medicalStr) == "y"
		
		fmt.Print("Notes (optional): ")
		var notes string
		fmt.Scanln(&notes)
		
		contact := contacts.Contact{
			Name:         name,
			Relationship: relationship,
			Phone:        phone,
			Email:        email,
			Medical:      medical,
			Notes:        notes,
		}
		
		profile.AddContact(contact)
		
		if err := profile.Save(); err != nil {
			return fmt.Errorf("failed to save contact: %w", err)
		}
		
		color.New(color.FgGreen).Printf("✓ Added %s to emergency contacts\n", name)
		return nil
	},
}

var contactsRemoveCmd = &cobra.Command{
	Use:   "remove [name]",
	Short: "Remove an emergency contact",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		profile, err := contacts.LoadProfile()
		if err != nil {
			return err
		}
		
		if profile.RemoveContact(args[0]) {
			if err := profile.Save(); err != nil {
				return err
			}
			color.New(color.FgGreen).Printf("✓ Removed %s from emergency contacts\n", args[0])
		} else {
			color.New(color.FgRed).Printf("Contact '%s' not found\n", args[0])
		}
		
		return nil
	},
}

var contactsCardCmd = &cobra.Command{
	Use:   "card",
	Short: "Export emergency info as printable card",
	RunE: func(cmd *cobra.Command, args []string) error {
		profile, err := contacts.LoadProfile()
		if err != nil {
			return err
		}
		
		card := profile.ExportCard()
		fmt.Println(card)
		
		fmt.Println("\nPress 'c' to copy, 's' to save to file, or Enter to exit...")
		var input string
		fmt.Scanln(&input)
		
		switch strings.ToLower(input) {
		case "c":
			if err := clipboard.Copy(card); err == nil {
				color.New(color.FgGreen).Println("✓ Copied to clipboard")
			}
		case "s":
			filename := fmt.Sprintf("emergency_card_%s.txt", profile.UpdatedAt.Format("2006-01-02"))
			if err := os.WriteFile(filename, []byte(card), 0644); err == nil {
				color.New(color.FgGreen).Printf("✓ Saved to %s\n", filename)
			}
		}
		
		return nil
	},
}

var contactsMedicalCmd = &cobra.Command{
	Use:   "medical",
	Short: "Update medical information",
	RunE: func(cmd *cobra.Command, args []string) error {
		profile, err := contacts.LoadProfile()
		if err != nil {
			return err
		}
		
		fmt.Println("Update Medical Information (press Enter to skip)")
		fmt.Println(strings.Repeat("=", 50))
		
		fmt.Print("Blood type (e.g., O+, A-, B+): ")
		var bloodType string
		fmt.Scanln(&bloodType)
		if bloodType != "" {
			profile.MedicalInfo.BloodType = bloodType
		}
		
		fmt.Print("Allergies (comma-separated): ")
		var allergiesStr string
		fmt.Scanln(&allergiesStr)
		if allergiesStr != "" {
			profile.MedicalInfo.Allergies = strings.Split(allergiesStr, ",")
			for i := range profile.MedicalInfo.Allergies {
				profile.MedicalInfo.Allergies[i] = strings.TrimSpace(profile.MedicalInfo.Allergies[i])
			}
		}
		
		fmt.Print("Current medications (comma-separated): ")
		var medsStr string
		fmt.Scanln(&medsStr)
		if medsStr != "" {
			profile.MedicalInfo.Medications = strings.Split(medsStr, ",")
			for i := range profile.MedicalInfo.Medications {
				profile.MedicalInfo.Medications[i] = strings.TrimSpace(profile.MedicalInfo.Medications[i])
			}
		}
		
		fmt.Print("Medical conditions (comma-separated): ")
		var conditionsStr string
		fmt.Scanln(&conditionsStr)
		if conditionsStr != "" {
			profile.MedicalInfo.Conditions = strings.Split(conditionsStr, ",")
			for i := range profile.MedicalInfo.Conditions {
				profile.MedicalInfo.Conditions[i] = strings.TrimSpace(profile.MedicalInfo.Conditions[i])
			}
		}
		
		fmt.Print("Primary doctor name: ")
		var doctorName string
		fmt.Scanln(&doctorName)
		if doctorName != "" {
			profile.MedicalInfo.DoctorName = doctorName
		}
		
		fmt.Print("Doctor phone: ")
		var doctorPhone string
		fmt.Scanln(&doctorPhone)
		if doctorPhone != "" {
			profile.MedicalInfo.DoctorPhone = doctorPhone
		}
		
		fmt.Print("Insurance provider: ")
		var insurance string
		fmt.Scanln(&insurance)
		if insurance != "" {
			profile.MedicalInfo.Insurance = insurance
		}
		
		fmt.Print("Policy number: ")
		var policyNum string
		fmt.Scanln(&policyNum)
		if policyNum != "" {
			profile.MedicalInfo.PolicyNum = policyNum
		}
		
		if err := profile.Save(); err != nil {
			return fmt.Errorf("failed to save medical info: %w", err)
		}
		
		color.New(color.FgGreen).Println("✓ Medical information updated")
		return nil
	},
}

func showCategoryResources(location *data.Location, category, title string) error {
	resources, ok := location.Resources[category]
	if !ok || len(resources) == 0 {
		fmt.Printf("\nNo %s resources available for %s\n", title, location.Name)
		if location.Note != "" {
			fmt.Printf("\nNote: %s\n", location.Note)
		}
		return nil
	}

	color.New(color.FgYellow, color.Bold).Printf("\n%s - %s\n", title, location.Name)
	fmt.Println(strings.Repeat("=", 50))
	
	for _, resource := range resources {
		ui.DisplayResource(resource)  // Use non-interactive display for shortcuts
	}
	
	fmt.Println("\nTip: Use 'helpme' for interactive mode with copy support")
	
	return nil
}

func init() {
	// Flags
	rootCmd.PersistentFlags().StringVarP(&locationFlag, "location", "l", "", "Override default location")
	rootCmd.PersistentFlags().BoolP("emergency", "e", false, "Show emergency numbers immediately")
	
	// Subcommands
	configCmd.AddCommand(setLocationCmd)
	rootCmd.AddCommand(emergencyCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(listLocationsCmd)
	
	// Quick access shortcuts
	rootCmd.AddCommand(mentalCmd)
	rootCmd.AddCommand(poisonCmd)
	rootCmd.AddCommand(domesticCmd)
	rootCmd.AddCommand(zipCmd)
	
	// Contacts management
	contactsCmd.AddCommand(contactsListCmd)
	contactsCmd.AddCommand(contactsAddCmd)
	contactsCmd.AddCommand(contactsRemoveCmd)
	contactsCmd.AddCommand(contactsCardCmd)
	contactsCmd.AddCommand(contactsMedicalCmd)
	rootCmd.AddCommand(contactsCmd)
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