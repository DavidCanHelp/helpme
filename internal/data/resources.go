package data

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Resource struct {
	Name        string `json:"name"`
	Number      string `json:"number,omitempty"`
	Text        string `json:"text,omitempty"`
	SMS         string `json:"sms,omitempty"`
	Description string `json:"description"`
	Website     string `json:"website,omitempty"`
}

type EmergencyService struct {
	Number      string `json:"number"`
	Description string `json:"description"`
}

type Location struct {
	Name      string                       `json:"name"`
	Emergency map[string]EmergencyService `json:"emergency"`
	Resources map[string][]Resource       `json:"resources"`
	Note      string                       `json:"note,omitempty"`
}

type ResourceData struct {
	Locations map[string]Location `json:"locations"`
}

var Data *ResourceData

func init() {
	execPath, err := os.Executable()
	if err != nil {
		panic(fmt.Sprintf("failed to get executable path: %v", err))
	}
	
	dataPath := filepath.Join(filepath.Dir(execPath), "data", "resources.json")
	
	// Try relative path first (for development)
	if _, err := os.Stat(dataPath); os.IsNotExist(err) {
		dataPath = "data/resources.json"
	}
	
	resourcesJSON, err := os.ReadFile(dataPath)
	if err != nil {
		panic(fmt.Sprintf("failed to read resources.json: %v", err))
	}
	
	if err := json.Unmarshal(resourcesJSON, &Data); err != nil {
		panic(fmt.Sprintf("failed to unmarshal resources: %v", err))
	}
}

func GetLocation(code string) (*Location, error) {
	code = strings.ToUpper(code)
	if loc, ok := Data.Locations[code]; ok {
		return &loc, nil
	}
	return nil, fmt.Errorf("location '%s' not found", code)
}

func GetAvailableLocations() []string {
	locations := make([]string, 0, len(Data.Locations))
	for code := range Data.Locations {
		locations = append(locations, code)
	}
	return locations
}

func SearchResources(location *Location, query string) []Resource {
	query = strings.ToLower(query)
	var results []Resource
	
	for _, resources := range location.Resources {
		for _, resource := range resources {
			if strings.Contains(strings.ToLower(resource.Name), query) ||
				strings.Contains(strings.ToLower(resource.Description), query) {
				results = append(results, resource)
			}
		}
	}
	
	return results
}