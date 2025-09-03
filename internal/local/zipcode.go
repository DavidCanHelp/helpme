package local

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type ZipInfo struct {
	City   string `json:"city"`
	State  string `json:"state"`
	County string `json:"county"`
}

type Hospital struct {
	Name string `json:"name"`
	Main string `json:"main"`
}

type LocalResources struct {
	PoliceNonEmergency string     `json:"police_non_emergency"`
	FireNonEmergency   string     `json:"fire_non_emergency"`
	CityServices       string     `json:"city_services"`
	Hospitals          []Hospital `json:"hospitals"`
}

type USLocalData struct {
	ZipToCity      map[string]ZipInfo           `json:"zip_to_city"`
	LocalResources map[string]LocalResources    `json:"local_resources"`
}

var localData *USLocalData

func init() {
	loadLocalData()
}

func loadLocalData() {
	execPath, err := os.Executable()
	if err != nil {
		return
	}
	
	dataPath := filepath.Join(filepath.Dir(execPath), "data", "us_local.json")
	
	// Try relative path for development
	if _, err := os.Stat(dataPath); os.IsNotExist(err) {
		dataPath = "data/us_local.json"
	}
	
	data, err := os.ReadFile(dataPath)
	if err != nil {
		// Local data is optional, so we don't panic
		return
	}
	
	if err := json.Unmarshal(data, &localData); err != nil {
		return
	}
}

// IsZipCode checks if the input looks like a US ZIP code
func IsZipCode(input string) bool {
	// Match 5 digits or 5+4 format (12345 or 12345-6789)
	matched, _ := regexp.MatchString(`^\d{5}(-\d{4})?$`, input)
	return matched
}

// GetLocalResources returns non-emergency numbers for a ZIP code
func GetLocalResources(zipCode string) (*LocalResources, *ZipInfo, error) {
	if localData == nil {
		return nil, nil, fmt.Errorf("local data not available")
	}
	
	// Clean the ZIP code (remove any +4 extension)
	if len(zipCode) > 5 {
		zipCode = zipCode[:5]
	}
	
	// Look up the ZIP code
	zipInfo, ok := localData.ZipToCity[zipCode]
	if !ok {
		// Try to find by prefix (first 3 digits) for broader area coverage
		prefix := zipCode[:3]
		for zip, info := range localData.ZipToCity {
			if strings.HasPrefix(zip, prefix) {
				zipInfo = info
				break
			}
		}
		if zipInfo.City == "" {
			return nil, nil, fmt.Errorf("ZIP code %s not found", zipCode)
		}
	}
	
	// Create lookup key for resources (City_State)
	resourceKey := fmt.Sprintf("%s_%s", zipInfo.City, zipInfo.State)
	
	// Get local resources for this city
	resources, ok := localData.LocalResources[resourceKey]
	if !ok {
		// Try just the state as fallback
		for key, res := range localData.LocalResources {
			if strings.HasSuffix(key, "_"+zipInfo.State) {
				resources = res
				break
			}
		}
		if resources.PoliceNonEmergency == "" {
			return nil, &zipInfo, fmt.Errorf("local resources not available for %s, %s", zipInfo.City, zipInfo.State)
		}
	}
	
	return &resources, &zipInfo, nil
}

// GetNearbyZipCodes returns nearby ZIP codes for a given ZIP
func GetNearbyZipCodes(zipCode string) []string {
	if localData == nil {
		return nil
	}
	
	var nearby []string
	baseZip := zipCode[:3] // Use first 3 digits as area prefix
	
	for zip := range localData.ZipToCity {
		if strings.HasPrefix(zip, baseZip) && zip != zipCode {
			nearby = append(nearby, zip)
			if len(nearby) >= 5 { // Limit to 5 nearby zips
				break
			}
		}
	}
	
	return nearby
}

// FormatLocalResources formats the resources for display
func FormatLocalResources(resources *LocalResources, zipInfo *ZipInfo) string {
	var output []string
	
	output = append(output, fmt.Sprintf("📍 %s, %s %s County", zipInfo.City, zipInfo.State, zipInfo.County))
	output = append(output, strings.Repeat("=", 50))
	output = append(output, "")
	output = append(output, "🚔 POLICE (Non-Emergency)")
	output = append(output, fmt.Sprintf("   %s", resources.PoliceNonEmergency))
	output = append(output, "")
	output = append(output, "🚒 FIRE (Non-Emergency)")
	output = append(output, fmt.Sprintf("   %s", resources.FireNonEmergency))
	output = append(output, "")
	
	if resources.CityServices != "" {
		output = append(output, "🏛️ CITY SERVICES")
		output = append(output, fmt.Sprintf("   %s", resources.CityServices))
		output = append(output, "")
	}
	
	if len(resources.Hospitals) > 0 {
		output = append(output, "🏥 HOSPITALS (Main Lines)")
		for _, hospital := range resources.Hospitals {
			output = append(output, fmt.Sprintf("   %-30s %s", hospital.Name+":", hospital.Main))
		}
		output = append(output, "")
	}
	
	output = append(output, strings.Repeat("=", 50))
	output = append(output, "⚠️  For emergencies, always call 911")
	
	return strings.Join(output, "\n")
}