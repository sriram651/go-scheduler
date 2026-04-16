package telegram

import "strings"

type continentGroup struct {
	Name  string
	Zones []string
}

var timezonesByContinent = []continentGroup{
	{Name: "Asia", Zones: []string{"Asia/Kolkata", "Asia/Dubai", "Asia/Karachi", "Asia/Dhaka", "Asia/Bangkok", "Asia/Jakarta", "Asia/Singapore", "Asia/Shanghai", "Asia/Hong_Kong", "Asia/Manila", "Asia/Tokyo",
		"Asia/Seoul", "Asia/Tehran", "Asia/Jerusalem", "Asia/Riyadh", "Indian/Maldives"}},
	{Name: "Africa", Zones: []string{"Africa/Cairo", "Africa/Lagos", "Africa/Johannesburg", "Africa/Nairobi", "Africa/Casablanca", "Africa/Accra", "Africa/Addis_Ababa", "Africa/Algiers", "Africa/Tunis",
		"Africa/Kinshasa", "Indian/Mauritius", "Indian/Reunion"}},
	{Name: "Americas", Zones: []string{"America/New_York", "America/Chicago", "America/Denver", "America/Los_Angeles", "America/Toronto", "America/Vancouver", "America/Anchorage", "America/Mexico_City",
		"America/Sao_Paulo", "America/Buenos_Aires", "America/Santiago", "America/Lima", "America/Bogota", "America/Caracas"}},
	{Name: "Europe", Zones: []string{"Europe/London", "Europe/Dublin", "Europe/Lisbon", "Europe/Paris", "Europe/Berlin", "Europe/Madrid", "Europe/Rome", "Europe/Amsterdam", "Europe/Warsaw", "Europe/Athens",
		"Europe/Istanbul", "Europe/Moscow"}},
	{Name: "Oceania", Zones: []string{"Australia/Sydney", "Australia/Melbourne", "Australia/Brisbane", "Australia/Perth", "Australia/Adelaide", "Australia/Darwin", "Pacific/Auckland", "Pacific/Fiji", "Pacific/Honolulu",
		"Pacific/Guam"}},
	{Name: "Other", Zones: []string{"UTC", "Atlantic/Reykjavik", "Atlantic/Azores"}},
}

func IsValidTimeZone(inputZone string) bool {
	for _, continent := range timezonesByContinent {
		for _, zone := range continent.Zones {
			if strings.EqualFold(inputZone, zone) {
				return true
			}
		}
	}

	return false
}
