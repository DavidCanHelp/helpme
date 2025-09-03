package local

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// OnlineZipInfo contains enhanced ZIP code information from APIs
type OnlineZipInfo struct {
	ZipCode   string  `json:"zip_code"`
	City      string  `json:"city"`
	State     string  `json:"state"`
	County    string  `json:"county"`
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lng"`
	Timezone  string  `json:"timezone"`
}

// LocalService represents a local service with contact info
type LocalService struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Phone    string `json:"phone"`
	Address  string `json:"address,omitempty"`
	Distance float64 `json:"distance,omitempty"`
}

// Cache for API responses
type CachedResponse struct {
	Data      interface{}
	Timestamp time.Time
}

var responseCache = make(map[string]CachedResponse)
const cacheExpiry = 24 * time.Hour

// GetComprehensiveLocalInfo attempts to get comprehensive local info for any US ZIP
func GetComprehensiveLocalInfo(zipCode string) (*OnlineZipInfo, []LocalService, error) {
	// Clean the ZIP code
	zipCode = strings.TrimSpace(zipCode)
	if len(zipCode) > 5 {
		zipCode = zipCode[:5]
	}
	
	// Validate ZIP format
	if !IsZipCode(zipCode) {
		return nil, nil, fmt.Errorf("invalid ZIP code format")
	}
	
	// Check cache first
	cacheKey := "zip:" + zipCode
	if cached, ok := responseCache[cacheKey]; ok {
		if time.Since(cached.Timestamp) < cacheExpiry {
			if data, ok := cached.Data.(*OnlineZipInfo); ok {
				services := searchLocalServices(data)
				return data, services, nil
			}
		}
	}
	
	// Try multiple data sources
	zipInfo, err := fetchZipInfo(zipCode)
	if err != nil {
		return nil, nil, err
	}
	
	// Cache the response
	responseCache[cacheKey] = CachedResponse{
		Data:      zipInfo,
		Timestamp: time.Now(),
	}
	
	// Search for local services
	services := searchLocalServices(zipInfo)
	
	return zipInfo, services, nil
}

// fetchZipInfo fetches ZIP code information from multiple sources
func fetchZipInfo(zipCode string) (*OnlineZipInfo, error) {
	// Try ZipCodeAPI.com (free tier: 10 requests/hour)
	if info := fetchFromZipCodeAPI(zipCode); info != nil {
		return info, nil
	}
	
	// Try Zippopotam.us (free, unlimited)
	if info := fetchFromZippopotamus(zipCode); info != nil {
		return info, nil
	}
	
	// Try OpenDataSoft (free public dataset)
	if info := fetchFromOpenDataSoft(zipCode); info != nil {
		return info, nil
	}
	
	return nil, fmt.Errorf("unable to fetch ZIP code information")
}

// fetchFromZippopotamus uses the free Zippopotam.us API
func fetchFromZippopotamus(zipCode string) *OnlineZipInfo {
	client := &http.Client{Timeout: 5 * time.Second}
	
	url := fmt.Sprintf("http://api.zippopotam.us/us/%s", zipCode)
	resp, err := client.Get(url)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil
	}
	
	var data struct {
		PostCode string `json:"post_code"`
		Country  string `json:"country"`
		Places   []struct {
			PlaceName  string  `json:"place name"`
			State      string  `json:"state"`
			StateAbbr  string  `json:"state abbreviation"`
			Latitude   string  `json:"latitude"`
			Longitude  string  `json:"longitude"`
		} `json:"places"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil
	}
	
	if len(data.Places) == 0 {
		return nil
	}
	
	place := data.Places[0]
	lat := 0.0
	lng := 0.0
	fmt.Sscanf(place.Latitude, "%f", &lat)
	fmt.Sscanf(place.Longitude, "%f", &lng)
	
	return &OnlineZipInfo{
		ZipCode:   zipCode,
		City:      place.PlaceName,
		State:     place.StateAbbr,
		County:    "", // Not provided by this API
		Latitude:  lat,
		Longitude: lng,
	}
}

// fetchFromZipCodeAPI uses zipcodeapi.com (requires free API key)
func fetchFromZipCodeAPI(zipCode string) *OnlineZipInfo {
	// You can get a free API key at https://www.zipcodeapi.com/Register
	apiKey := os.Getenv("ZIPCODE_API_KEY")
	if apiKey == "" {
		return nil
	}
	
	client := &http.Client{Timeout: 5 * time.Second}
	
	url := fmt.Sprintf("https://www.zipcodeapi.com/rest/%s/info.json/%s/degrees", apiKey, zipCode)
	resp, err := client.Get(url)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil
	}
	
	var data struct {
		ZipCode   string  `json:"zip_code"`
		City      string  `json:"city"`
		State     string  `json:"state"`
		County    string  `json:"county"`
		Latitude  float64 `json:"lat"`
		Longitude float64 `json:"lng"`
		Timezone  struct {
			Identifier string `json:"timezone_identifier"`
		} `json:"timezone"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil
	}
	
	return &OnlineZipInfo{
		ZipCode:   data.ZipCode,
		City:      data.City,
		State:     data.State,
		County:    data.County,
		Latitude:  data.Latitude,
		Longitude: data.Longitude,
		Timezone:  data.Timezone.Identifier,
	}
}

// fetchFromOpenDataSoft uses public US ZIP code dataset
func fetchFromOpenDataSoft(zipCode string) *OnlineZipInfo {
	client := &http.Client{Timeout: 5 * time.Second}
	
	query := fmt.Sprintf("zip=%s", zipCode)
	apiURL := fmt.Sprintf("https://public.opendatasoft.com/api/records/1.0/search/?dataset=us-zip-code-latitude-and-longitude&q=%s", url.QueryEscape(query))
	
	resp, err := client.Get(apiURL)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil
	}
	
	var data struct {
		Records []struct {
			Fields struct {
				Zip       string    `json:"zip"`
				City      string    `json:"city"`
				State     string    `json:"state"`
				Latitude  float64   `json:"latitude"`
				Longitude float64   `json:"longitude"`
				County    string    `json:"county"`
			} `json:"fields"`
		} `json:"records"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil
	}
	
	if len(data.Records) == 0 {
		return nil
	}
	
	record := data.Records[0].Fields
	return &OnlineZipInfo{
		ZipCode:   record.Zip,
		City:      record.City,
		State:     record.State,
		County:    record.County,
		Latitude:  record.Latitude,
		Longitude: record.Longitude,
	}
}

// searchLocalServices searches for police, fire, and hospitals near the location
func searchLocalServices(zipInfo *OnlineZipInfo) []LocalService {
	var services []LocalService
	
	// For comprehensive data, we'd integrate with:
	// 1. Google Places API (requires API key)
	// 2. Overpass API (OpenStreetMap data)
	// 3. Local government APIs
	
	// Try to search using coordinates if available
	if zipInfo.Latitude != 0 && zipInfo.Longitude != 0 {
		// Search OpenStreetMap for nearby services
		if osm := searchOpenStreetMap(zipInfo.Latitude, zipInfo.Longitude); osm != nil {
			services = append(services, osm...)
		}
		
		// Search Google Places if API key available
		if places := searchGooglePlaces(zipInfo.Latitude, zipInfo.Longitude); places != nil {
			services = append(services, places...)
		}
	}
	
	// Fallback to general city/state numbers
	if len(services) == 0 {
		services = getGenericCityServices(zipInfo.City, zipInfo.State)
	}
	
	return services
}

// searchOpenStreetMap searches for services using OpenStreetMap's Overpass API
func searchOpenStreetMap(lat, lng float64) []LocalService {
	// This would query Overpass API for police stations, fire stations, hospitals
	// within a certain radius of the coordinates
	// Implementation would be similar to:
	/*
	query := fmt.Sprintf(`
		[out:json];
		(
		  node["amenity"="police"](around:10000,%f,%f);
		  node["amenity"="fire_station"](around:10000,%f,%f);
		  node["amenity"="hospital"](around:10000,%f,%f);
		);
		out body;
	`, lat, lng, lat, lng, lat, lng)
	*/
	
	// For now, return nil (would need full implementation)
	return nil
}

// searchGooglePlaces searches using Google Places API
func searchGooglePlaces(lat, lng float64) []LocalService {
	apiKey := os.Getenv("GOOGLE_PLACES_API_KEY")
	if apiKey == "" {
		return nil
	}
	
	// Would search for police, fire, hospitals using Places API
	// Implementation would query:
	// https://maps.googleapis.com/maps/api/place/nearbysearch/json
	
	return nil
}

// getGenericCityServices returns generic services based on city/state
func getGenericCityServices(city, state string) []LocalService {
	services := []LocalService{
		{
			Name:  fmt.Sprintf("%s Police Department", city),
			Type:  "police",
			Phone: "Call 411 for local non-emergency",
		},
		{
			Name:  fmt.Sprintf("%s Fire Department", city),
			Type:  "fire",
			Phone: "Call 411 for local non-emergency",
		},
	}
	
	// Add state-specific resources
	stateResources := getStateResources(state)
	services = append(services, stateResources...)
	
	return services
}

// getStateResources returns state-level resources
func getStateResources(state string) []LocalService {
	// State police non-emergency numbers
	statePolice := map[string]string{
		"AL": "334-242-4371", "AK": "907-269-5511", "AZ": "602-223-2000",
		"AR": "501-618-8000", "CA": "916-843-3000", "CO": "303-239-4501",
		"CT": "860-685-8000", "DE": "302-739-4321", "FL": "850-617-2000",
		"GA": "404-624-7000", "HI": "808-586-1352", "ID": "208-884-7000",
		"IL": "217-782-6302", "IN": "317-232-8248", "IA": "515-725-6010",
		"KS": "785-296-6800", "KY": "502-782-1800", "LA": "225-925-6006",
		"ME": "207-624-7200", "MD": "410-486-3101", "MA": "508-820-2300",
		"MI": "517-332-2521", "MN": "651-201-7000", "MS": "601-987-1212",
		"MO": "573-751-3313", "MT": "406-444-2800", "NE": "402-471-4545",
		"NV": "775-687-5000", "NH": "603-271-2461", "NJ": "609-882-2000",
		"NM": "505-827-3370", "NY": "518-436-3600", "NC": "919-733-7952",
		"ND": "701-328-2455", "OH": "614-466-2660", "OK": "405-425-2424",
		"OR": "503-378-3720", "PA": "717-783-5517", "RI": "401-444-1000",
		"SC": "803-896-7979", "SD": "605-773-3178", "TN": "615-744-4000",
		"TX": "512-424-2000", "UT": "801-887-3800", "VT": "802-244-8727",
		"VA": "804-674-2000", "WA": "360-596-4000", "WV": "304-746-2100",
		"WI": "608-266-3212", "WY": "307-777-4321",
	}
	
	var services []LocalService
	if phone, ok := statePolice[state]; ok {
		services = append(services, LocalService{
			Name:  fmt.Sprintf("%s State Police", state),
			Type:  "police",
			Phone: phone,
		})
	}
	
	return services
}