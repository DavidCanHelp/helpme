package location

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// IPGeolocation represents the response from IP geolocation API
type IPGeolocation struct {
	Country     string `json:"country"`
	CountryCode string `json:"country_code"`
	Region      string `json:"region"`
	RegionName  string `json:"region_name"`
	City        string `json:"city"`
	Timezone    string `json:"timezone"`
}

// DetectLocation attempts to detect the user's location using various methods
func DetectLocation() (string, error) {
	// Try methods in order of reliability/speed
	
	// 1. Try system locale
	if loc := detectFromLocale(); loc != "" {
		return loc, nil
	}
	
	// 2. Try timezone
	if loc := detectFromTimezone(); loc != "" {
		return loc, nil
	}
	
	// 3. Try IP geolocation (requires internet)
	if loc := detectFromIP(); loc != "" {
		return loc, nil
	}
	
	// 4. macOS specific: try system preferences
	if runtime.GOOS == "darwin" {
		if loc := detectFromMacOS(); loc != "" {
			return loc, nil
		}
	}
	
	return "", fmt.Errorf("could not detect location automatically")
}

// detectFromLocale tries to determine location from system locale settings
func detectFromLocale() string {
	// Check LANG environment variable
	lang := os.Getenv("LANG")
	if lang == "" {
		lang = os.Getenv("LC_ALL")
	}
	
	// Parse locale format: en_US.UTF-8
	if lang != "" {
		parts := strings.Split(lang, "_")
		if len(parts) >= 2 {
			country := strings.Split(parts[1], ".")[0]
			return normalizeCountryCode(country)
		}
	}
	
	return ""
}

// detectFromTimezone tries to determine location from system timezone
func detectFromTimezone() string {
	zone, _ := time.Now().Zone()
	
	// Map common timezones to countries
	tzMap := map[string]string{
		"EST": "US", "EDT": "US", "CST": "US", "CDT": "US",
		"MST": "US", "MDT": "US", "PST": "US", "PDT": "US",
		"GMT": "UK", "BST": "UK",
		"CET": "EU", "CEST": "EU",
		"JST": "JP",
		"KST": "KR",
		"AEST": "AU", "AEDT": "AU",
		"NZST": "NZ", "NZDT": "NZ",
	}
	
	if country, ok := tzMap[zone]; ok {
		return country
	}
	
	// Try to read /etc/timezone on Linux
	if runtime.GOOS == "linux" {
		if data, err := os.ReadFile("/etc/timezone"); err == nil {
			tz := strings.TrimSpace(string(data))
			if strings.Contains(tz, "America/") {
				return "US"
			} else if strings.Contains(tz, "Europe/London") {
				return "UK"
			} else if strings.Contains(tz, "Europe/") {
				return "EU"
			} else if strings.Contains(tz, "Asia/Tokyo") {
				return "JP"
			} else if strings.Contains(tz, "Australia/") {
				return "AU"
			}
		}
	}
	
	return ""
}

// detectFromIP tries to determine location from IP geolocation
func detectFromIP() string {
	client := &http.Client{
		Timeout: 3 * time.Second,
	}
	
	// Try multiple free IP geolocation services
	services := []string{
		"http://ip-api.com/json/",
		"https://ipapi.co/json/",
		"https://api.ipify.org?format=json", // This one only gives IP, not location
	}
	
	for _, service := range services {
		resp, err := client.Get(service)
		if err != nil {
			continue
		}
		defer resp.Body.Close()
		
		if resp.StatusCode == http.StatusOK {
			var geo map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&geo); err == nil {
				// Try different field names used by different services
				if cc, ok := geo["countryCode"].(string); ok {
					return normalizeCountryCode(cc)
				}
				if cc, ok := geo["country_code"].(string); ok {
					return normalizeCountryCode(cc)
				}
				if cc, ok := geo["country"].(string); ok && len(cc) == 2 {
					return normalizeCountryCode(cc)
				}
			}
		}
	}
	
	return ""
}

// detectFromMacOS uses macOS-specific commands to detect location
func detectFromMacOS() string {
	// Try to get country from system preferences
	cmd := exec.Command("defaults", "read", "NSGlobalDomain", "AppleLocale")
	if output, err := cmd.Output(); err == nil {
		locale := strings.TrimSpace(string(output))
		parts := strings.Split(locale, "_")
		if len(parts) >= 2 {
			return normalizeCountryCode(parts[1])
		}
	}
	
	return ""
}

// normalizeCountryCode ensures the country code is uppercase and valid
func normalizeCountryCode(code string) string {
	code = strings.ToUpper(strings.TrimSpace(code))
	
	// Map some common variations
	switch code {
		case "USA":
		return "US"
	case "GBR", "GB":
		return "UK"
	case "DEU":
		return "DE"
	case "FRA":
		return "FR"
	case "JPN":
		return "JP"
	case "AUS":
		return "AU"
	case "CAN":
		return "CA"
	case "NZL":
		return "NZ"
	default:
		if len(code) == 2 {
			return code
		}
		return ""
	}
}

// GetLocationWithFallback tries to detect location, falls back to default
func GetLocationWithFallback(defaultLocation string) string {
	if detected, err := DetectLocation(); err == nil {
		return detected
	}
	return defaultLocation
}