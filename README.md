# HelpMe - Emergency & Crisis Support CLI

A life-saving command-line tool for instant access to emergency and crisis helplines, location-aware and designed for speed when every second counts.

## Features

### 🚨 Quick Emergency Access
```bash
helpme 911          # Instant emergency numbers
helpme emergency    # Same as above
helpme -e           # Emergency flag shortcut
helpme sos          # Works with various emergency keywords
```

### 📍 ZIP Code Local Non-Emergency (NEW!)
```bash
helpme 10001        # NYC non-emergency numbers
helpme 94102        # San Francisco local numbers
helpme zip 60601    # Chicago police, fire, hospitals
```
Get instant access to:
- Police non-emergency line
- Fire department non-emergency
- City services (311 where available)
- Major hospital main numbers

### 🔍 Smart Search
```bash
helpme suicide      # Search for suicide prevention
helpme mental       # Mental health resources
helpme poison       # Poison control
helpme domestic     # Domestic violence support
helpme "child abuse" # Search any topic
```

### 📍 Auto-Location Detection
- Automatically detects your location on first run using:
  - System locale settings
  - Timezone information
  - IP geolocation (if online)
  - macOS system preferences

### 📋 Clipboard Support
- Press 'c' when viewing any resource to copy phone numbers
- Works across all platforms (macOS, Linux, Windows)
- Fallback display if clipboard unavailable

### ⚡ Command Shortcuts
```bash
helpme mental       # or: helpme m, helpme crisis, helpme suicide
helpme poison       # or: helpme p
helpme domestic     # or: helpme dv, helpme violence
```

### 🌍 Multi-Location Support
```bash
helpme -l UK 911                    # Override location temporarily
helpme config set-location CA        # Set default location
helpme list-locations               # See all available locations
```

## Installation

```bash
go build -o helpme cmd/helpme/main.go
sudo mv helpme /usr/local/bin/       # Install system-wide (optional)
```

## Usage Examples

### Emergency Situations
```bash
# Fastest emergency access
helpme 911

# Mental health crisis
helpme suicide
helpme crisis

# Specific emergencies
helpme poison
helpme domestic violence
```

### Local Non-Emergency
```bash
# Get non-emergency numbers by ZIP code
helpme 10001        # Manhattan, NYC
helpme 90210        # Beverly Hills
helpme 60601        # Chicago Loop
helpme 98101        # Seattle Downtown
helpme 33139        # Miami Beach

# Or use the explicit command
helpme zip 10001
helpme local 94102
```

### Interactive Mode
```bash
helpme              # Opens interactive menu
```

### Search Mode
```bash
helpme veterans     # Search for veteran support
helpme lgbtq        # LGBTQ resources
helpme substance    # Substance abuse help
```

## Supported Locations

Currently supports 30+ countries/regions including:
- North America: US, CA
- Europe: UK, IE, DE, FR, ES, IT, NL, BE, CH, AT, SE, NO, DK, FI, IS, PT, LU, GR, EU
- Asia-Pacific: JP, KR, SG, AU, NZ
- Middle East: IL, AE
- Africa: ZA

## Privacy & Security

- **100% Offline**: All data stored locally, works without internet
- **No Tracking**: No analytics, no data collection
- **Secure**: No external dependencies for core functionality
- **Private**: No logs of what resources you access

## Contributing

This is a life-saving tool. Contributions to expand coverage, add resources, or improve functionality are welcome and could literally save lives.

### Priority Improvements Needed:
- More country coverage
- Regional/state-specific resources
- Multi-language support
- Time-aware filtering (24/7 vs business hours)
- Accessibility features (TTY, text-only services)

## Data Structure

Resources are stored in `data/resources.json` with the following structure:
```json
{
  "locations": {
    "US": {
      "name": "United States",
      "emergency": {
        "general": {
          "number": "911",
          "description": "Police, Fire, Medical"
        }
      },
      "resources": {
        "mental_health": [...],
        "poison_control": [...],
        ...
      }
    }
  }
}
```

## Important Disclaimers

### Medical & Legal Disclaimer
**THIS TOOL IS NOT A SUBSTITUTE FOR PROFESSIONAL EMERGENCY SERVICES**

- **In life-threatening emergencies, ALWAYS call 911 (or your local emergency number) directly**
- This tool provides directory information only - it does not dispatch emergency services
- Medical information stored is for personal reference only and should not replace professional medical advice
- Always verify phone numbers are current and correct for your area
- Response times and availability of services may vary by location

### Privacy & Security Notice
- All personal data (contacts, medical info) is stored locally on YOUR device only
- We do not collect, transmit, or store any personal information on servers
- You are responsible for securing access to your device and this data
- Emergency profile information should be kept up-to-date for accuracy

### Accuracy & Availability
- Phone numbers and resources are provided as-is and may change without notice
- Not all services are available 24/7 - verify availability when non-urgent
- ZIP code data covers major US cities; rural areas may have limited information
- International resources may vary significantly by country

### No Warranty
This software is provided "AS IS" without warranty of any kind, express or implied. The authors and contributors are not liable for any damages or consequences arising from the use of this tool.

### Responsible Use
- Test the tool when you're NOT in an emergency to familiarize yourself
- Keep your emergency contacts and medical information current
- Please help save lives responsibly by sharing this tool with others who might benefit, but ensure they understand its limitations
- Report issues or incorrect information via GitHub Issues to help improve the tool

## Contributing

**Your contributions could save lives.** Please help by:
- Adding/updating emergency numbers for your country/region
- Reporting incorrect or outdated information
- Translating resources to other languages
- Improving accessibility features
- Testing and reporting bugs

## License

MIT - Free to use, modify, and distribute. See [LICENSE](LICENSE).

---

**Remember: When in doubt, call 911 (or your local emergency number) directly. This tool is meant to supplement, not replace, official emergency services.**