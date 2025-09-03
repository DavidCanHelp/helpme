package contacts

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Contact struct {
	Name         string    `json:"name"`
	Relationship string    `json:"relationship"`
	Phone        string    `json:"phone"`
	Email        string    `json:"email,omitempty"`
	Medical      bool      `json:"medical,omitempty"`
	Notes        string    `json:"notes,omitempty"`
	AddedAt      time.Time `json:"added_at"`
}

type MedicalInfo struct {
	BloodType   string   `json:"blood_type,omitempty"`
	Allergies   []string `json:"allergies,omitempty"`
	Medications []string `json:"medications,omitempty"`
	Conditions  []string `json:"conditions,omitempty"`
	DoctorName  string   `json:"doctor_name,omitempty"`
	DoctorPhone string   `json:"doctor_phone,omitempty"`
	Insurance   string   `json:"insurance,omitempty"`
	PolicyNum   string   `json:"policy_number,omitempty"`
}

type EmergencyProfile struct {
	Contacts    []Contact    `json:"contacts"`
	MedicalInfo MedicalInfo  `json:"medical_info"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// GetProfilePath returns the path to the emergency profile file
func GetProfilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, ".config", "helpme", "emergency_profile.json"), nil
}

// LoadProfile loads the user's emergency profile
func LoadProfile() (*EmergencyProfile, error) {
	profilePath, err := GetProfilePath()
	if err != nil {
		return nil, err
	}
	
	data, err := os.ReadFile(profilePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty profile if doesn't exist
			return &EmergencyProfile{
				Contacts:  []Contact{},
				UpdatedAt: time.Now(),
			}, nil
		}
		return nil, fmt.Errorf("failed to read profile: %w", err)
	}
	
	var profile EmergencyProfile
	if err := json.Unmarshal(data, &profile); err != nil {
		return nil, fmt.Errorf("failed to parse profile: %w", err)
	}
	
	return &profile, nil
}

// SaveProfile saves the emergency profile
func (p *EmergencyProfile) Save() error {
	profilePath, err := GetProfilePath()
	if err != nil {
		return err
	}
	
	profileDir := filepath.Dir(profilePath)
	if err := os.MkdirAll(profileDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	
	p.UpdatedAt = time.Now()
	
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal profile: %w", err)
	}
	
	if err := os.WriteFile(profilePath, data, 0600); err != nil { // 0600 for privacy
		return fmt.Errorf("failed to write profile: %w", err)
	}
	
	return nil
}

// AddContact adds a new emergency contact
func (p *EmergencyProfile) AddContact(contact Contact) {
	contact.AddedAt = time.Now()
	p.Contacts = append(p.Contacts, contact)
}

// RemoveContact removes a contact by name
func (p *EmergencyProfile) RemoveContact(name string) bool {
	for i, contact := range p.Contacts {
		if strings.EqualFold(contact.Name, name) {
			p.Contacts = append(p.Contacts[:i], p.Contacts[i+1:]...)
			return true
		}
	}
	return false
}

// FindContact finds a contact by name
func (p *EmergencyProfile) FindContact(name string) *Contact {
	for _, contact := range p.Contacts {
		if strings.Contains(strings.ToLower(contact.Name), strings.ToLower(name)) {
			return &contact
		}
	}
	return nil
}

// GetMedicalContacts returns only medical contacts
func (p *EmergencyProfile) GetMedicalContacts() []Contact {
	var medical []Contact
	for _, contact := range p.Contacts {
		if contact.Medical {
			medical = append(medical, contact)
		}
	}
	return medical
}

// FormatContact formats a contact for display
func FormatContact(c Contact) string {
	var lines []string
	lines = append(lines, fmt.Sprintf("👤 %s (%s)", c.Name, c.Relationship))
	lines = append(lines, fmt.Sprintf("   📞 %s", c.Phone))
	if c.Email != "" {
		lines = append(lines, fmt.Sprintf("   📧 %s", c.Email))
	}
	if c.Medical {
		lines = append(lines, "   🏥 Medical Contact")
	}
	if c.Notes != "" {
		lines = append(lines, fmt.Sprintf("   📝 %s", c.Notes))
	}
	return strings.Join(lines, "\n")
}

// FormatMedicalInfo formats medical information for display
func FormatMedicalInfo(m MedicalInfo) string {
	var lines []string
	
	if m.BloodType != "" {
		lines = append(lines, fmt.Sprintf("🩸 Blood Type: %s", m.BloodType))
	}
	
	if len(m.Allergies) > 0 {
		lines = append(lines, fmt.Sprintf("⚠️  Allergies: %s", strings.Join(m.Allergies, ", ")))
	}
	
	if len(m.Medications) > 0 {
		lines = append(lines, fmt.Sprintf("💊 Medications: %s", strings.Join(m.Medications, ", ")))
	}
	
	if len(m.Conditions) > 0 {
		lines = append(lines, fmt.Sprintf("🏥 Conditions: %s", strings.Join(m.Conditions, ", ")))
	}
	
	if m.DoctorName != "" {
		lines = append(lines, fmt.Sprintf("👨‍⚕️ Doctor: %s (%s)", m.DoctorName, m.DoctorPhone))
	}
	
	if m.Insurance != "" {
		lines = append(lines, fmt.Sprintf("📋 Insurance: %s (Policy: %s)", m.Insurance, m.PolicyNum))
	}
	
	return strings.Join(lines, "\n")
}

// ExportCard exports emergency info as a printable card format
func (p *EmergencyProfile) ExportCard() string {
	var card []string
	
	card = append(card, "╔══════════════════════════════════════════════════════╗")
	card = append(card, "║               EMERGENCY INFORMATION CARD              ║")
	card = append(card, "╠══════════════════════════════════════════════════════╣")
	
	if len(p.Contacts) > 0 {
		card = append(card, "║ EMERGENCY CONTACTS:                                   ║")
		for i, contact := range p.Contacts {
			if i >= 3 { break } // Limit to 3 for card size
			line := fmt.Sprintf("║ %s (%s): %s", contact.Name, contact.Relationship, contact.Phone)
			if len(line) > 55 {
				line = line[:55] + "║"
			} else {
				line = fmt.Sprintf("%-56s║", line)
			}
			card = append(card, line)
		}
	}
	
	if p.MedicalInfo.BloodType != "" || len(p.MedicalInfo.Allergies) > 0 {
		card = append(card, "╠══════════════════════════════════════════════════════╣")
		card = append(card, "║ MEDICAL INFO:                                         ║")
		
		if p.MedicalInfo.BloodType != "" {
			line := fmt.Sprintf("║ Blood Type: %s", p.MedicalInfo.BloodType)
			card = append(card, fmt.Sprintf("%-56s║", line))
		}
		
		if len(p.MedicalInfo.Allergies) > 0 {
			line := fmt.Sprintf("║ Allergies: %s", strings.Join(p.MedicalInfo.Allergies, ", "))
			if len(line) > 55 {
				line = line[:55] + "║"
			} else {
				line = fmt.Sprintf("%-56s║", line)
			}
			card = append(card, line)
		}
		
		if len(p.MedicalInfo.Medications) > 0 {
			line := fmt.Sprintf("║ Medications: %s", strings.Join(p.MedicalInfo.Medications, ", "))
			if len(line) > 55 {
				line = line[:55] + "║"
			} else {
				line = fmt.Sprintf("%-56s║", line)
			}
			card = append(card, line)
		}
	}
	
	card = append(card, "╠══════════════════════════════════════════════════════╣")
	card = append(card, "║ IN EMERGENCY CALL 911                                 ║")
	card = append(card, "║ Generated by HelpMe CLI                               ║")
	card = append(card, fmt.Sprintf("║ Updated: %-45s║", p.UpdatedAt.Format("2006-01-02")))
	card = append(card, "╚══════════════════════════════════════════════════════╝")
	
	return strings.Join(card, "\n")
}