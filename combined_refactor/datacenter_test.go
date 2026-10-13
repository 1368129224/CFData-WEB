package main

import (
	"reflect"
	"testing"
)

func TestParseDataCenterList(t *testing.T) {
	got := parseDataCenterList(" hkg, NRT；hkg | LAX\tSFO ")
	want := []string{"HKG", "NRT", "LAX", "SFO"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("parseDataCenterList() = %#v, want %#v", got, want)
	}
}

func TestMatchesDataCenterFilter(t *testing.T) {
	tests := []struct {
		name       string
		dataCenter string
		filter     string
		want       bool
	}{
		{name: "empty filter allows all", dataCenter: "HKG", filter: "", want: true},
		{name: "single match is case insensitive", dataCenter: "HKG", filter: "hkg", want: true},
		{name: "multiple match", dataCenter: "NRT", filter: "HKG,NRT", want: true},
		{name: "non match", dataCenter: "SFO", filter: "HKG,NRT", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matchesDataCenterFilter(tt.dataCenter, tt.filter); got != tt.want {
				t.Fatalf("matchesDataCenterFilter(%q, %q) = %v, want %v", tt.dataCenter, tt.filter, got, tt.want)
			}
		})
	}
}

func TestMatchesDataCenterOrCountryFilter(t *testing.T) {
	tests := []struct {
		name       string
		dataCenter string
		country    string
		filter     string
		want       bool
	}{
		{name: "colo matches", dataCenter: "HKG", country: "HK", filter: "HKG", want: true},
		{name: "country matches multiple colos", dataCenter: "MFM", country: "HK", filter: "HK", want: true},
		{name: "country miss", dataCenter: "LAX", country: "US", filter: "HK", want: false},
		{name: "mixed colo and country filters", dataCenter: "LAX", country: "US", filter: "HKG,US", want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matchesDataCenterOrCountryFilter(tt.dataCenter, tt.country, tt.filter); got != tt.want {
				t.Fatalf("matchesDataCenterOrCountryFilter(%q, %q, %q) = %v, want %v", tt.dataCenter, tt.country, tt.filter, got, tt.want)
			}
		})
	}
}

func TestDataCenterQuotaGroupsByCountryWhenConfigured(t *testing.T) {
	if got := dataCenterQuotaGroup("HKG", "HK", "HK,US"); got != "country:HK" {
		t.Fatalf("country-filter quota group = %q, want country:HK", got)
	}
	if got := dataCenterQuotaGroup("MFM", "HK", "HK,US"); got != "country:HK" {
		t.Fatalf("same-country colo quota group = %q, want country:HK", got)
	}
	if got := dataCenterQuotaGroup("HKG", "HK", "HKG"); got != "colo:HKG" {
		t.Fatalf("colo-filter quota group = %q, want colo:HKG", got)
	}
}
