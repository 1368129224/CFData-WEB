package main

import "testing"

func TestApplyCLIExportFiltersAddsQualifiedQuotaRowsPerDataCenter(t *testing.T) {
	rows := []cliResultRow{
		{"ip": "1.1.1.1", "port": "443", "ipport": "1.1.1.1:443", "dc": "HKG", "speed": "10MB/s"},
		{"ip": "2.2.2.1", "port": "443", "ipport": "2.2.2.1:443", "dc": "LAX", "speed": "11MB/s"},
		{"ip": "1.1.1.2", "port": "443", "ipport": "1.1.1.2:443", "dc": "HKG", "speed": "12MB/s"},
		{"ip": "2.2.2.2", "port": "443", "ipport": "2.2.2.2:443", "dc": "LAX", "speed": "13MB/s"},
		{"ip": "1.1.1.3", "port": "443", "ipport": "1.1.1.3:443", "dc": "HKG", "speed": "0.01MB/s"},
	}
	cfg := &cliConfig{
		outIPType:    "all",
		outQualified: "unqualified",
		startRow:     1,
		endRow:       1,
		export:       cliExportConfig{MinPerDataCenter: 2},
	}

	got := applyCLIExportFilters(cfg, rows, 0.1)
	counts := map[string]int{}
	for _, row := range got {
		if getNSBSpeedStatus(row["speed"], 0.1) != "qualified" {
			t.Errorf("export contains an unqualified row: %#v", row)
		}
		counts[row["dc"]]++
	}
	if counts["HKG"] < 2 || counts["LAX"] < 2 {
		t.Fatalf("minimum quota not met: counts=%v rows=%#v", counts, got)
	}
	if counts["HKG"] != 2 || counts["LAX"] != 2 {
		t.Fatalf("expected only the page row plus missing quota rows: counts=%v", counts)
	}
}

func TestMinPerDataCenterConfigMerge(t *testing.T) {
	fileConfig := defaultCLIFileConfig()
	if fileConfig.OutDCMinCount != 0 {
		t.Fatalf("default outdcmincount = %d, want 0", fileConfig.OutDCMinCount)
	}
	fileConfig.OutDCMinCount = 4
	exportConfig := fileConfig.Export()
	if exportConfig.MinPerDataCenter != 4 {
		t.Fatalf("file export config lost minimum: got %d", exportConfig.MinPerDataCenter)
	}

	merged := defaultCLIExportConfig()
	mergeCLIExportConfig(&merged, exportConfig, false)
	if merged.MinPerDataCenter != 4 {
		t.Fatalf("config-file merge set minimum %d, want 4", merged.MinPerDataCenter)
	}
	mergeCLIExportConfig(&merged, cliExportConfig{MinPerDataCenter: 0}, true, map[string]bool{"outdcmincount": true})
	if merged.MinPerDataCenter != 0 {
		t.Fatalf("explicit CLI zero did not override the config value: got %d", merged.MinPerDataCenter)
	}
}

func TestMinimumQuotaCountsUniqueIPsNotPorts(t *testing.T) {
	selected := []cliResultRow{
		{"ip": "1.1.1.1", "port": "443", "dc": "HKG"},
		{"ip": "1.1.1.1", "port": "8443", "dc": "HKG"},
	}
	got, counts := ensureMinimumPerDataCenter(selected, selected, 2, "")
	if len(got) != 2 || counts["colo:HKG"] != 1 {
		t.Fatalf("same address on two ports must count once: rows=%#v counts=%v", got, counts)
	}
}

func TestMinimumQuotaAggregatesColosByConfiguredCountry(t *testing.T) {
	qualifiedRows := []cliResultRow{
		{"ip": "1.1.1.1", "port": "443", "dc": "HKG", "dcCountry": "HK", "speed": "5MB/s"},
		{"ip": "1.1.1.2", "port": "443", "dc": "MFM", "dcCountry": "HK", "speed": "5MB/s"},
		{"ip": "2.2.2.1", "port": "443", "dc": "LAX", "dcCountry": "US", "speed": "5MB/s"},
		{"ip": "2.2.2.2", "port": "443", "dc": "SJC", "dcCountry": "US", "speed": "5MB/s"},
	}
	got, counts := ensureMinimumPerDataCenter(nil, qualifiedRows, 2, "HK,US")
	if len(got) != 4 || counts["country:HK"] != 2 || counts["country:US"] != 2 {
		t.Fatalf("country quota did not aggregate colos: rows=%#v counts=%v", got, counts)
	}
	cfg := &cliConfig{mode: "official", dc: "HK,US", outIPType: "all", outQualified: "all", startRow: 1, endRow: 1, export: cliExportConfig{MinPerDataCenter: 2}}
	got = applyCLIExportFilters(cfg, qualifiedRows, 0.1)
	if len(got) != 4 {
		t.Fatalf("CLI country selection did not produce per-country quotas beyond row range: %#v", got)
	}
}

func TestApplyCLIExportFiltersDoesNotUseUnqualifiedRowsToFillQuota(t *testing.T) {
	rows := []cliResultRow{
		{"ip": "1.1.1.1", "port": "443", "ipport": "1.1.1.1:443", "dc": "HKG", "speed": "5MB/s"},
		{"ip": "1.1.1.2", "port": "443", "ipport": "1.1.1.2:443", "dc": "HKG", "speed": "0.01MB/s"},
	}
	cfg := &cliConfig{
		outIPType:    "all",
		outQualified: "all",
		startRow:     1,
		endRow:       20,
		export:       cliExportConfig{MinPerDataCenter: 3},
	}

	got := applyCLIExportFilters(cfg, rows, 0.1)
	if len(got) != 1 || got[0]["speed"] != "5MB/s" {
		t.Fatalf("expected only the available qualified row, got %#v", got)
	}
}
