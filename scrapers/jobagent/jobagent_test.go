package jobagent

import "testing"

func TestIsTargetLocation(t *testing.T) {
	tests := []struct {
		name     string
		location string
		want     bool
	}{
		{name: "indiaCountry", location: "Bangalore, India", want: true},
		{name: "indianCityOnly", location: "Gurgaon", want: true},
		{name: "remoteIndia", location: "Remote - India", want: true},
		{name: "blockedCountry", location: "San Mateo, CA, United States", want: false},
		{name: "ambiguousStateCode", location: "Delhi, NY", want: false},
		{name: "unitedKingdom", location: "London, United Kingdom", want: false},
		{name: "canada", location: "Toronto, Canada", want: false},
		{name: "remoteWorldwide", location: "Remote - Worldwide", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isTargetLocation(tt.location); got != tt.want {
				t.Errorf("isTargetLocation(%q) = %v, want %v", tt.location, got, tt.want)
			}
		})
	}
}
