package main

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func containsField(fields []string, expected string) bool {
	for _, field := range fields {
		if field == expected {
			return true
		}
	}
	return false
}

func TestCompactCLIExportsIncludeSpeedAndLandingCountry(t *testing.T) {
	officialRows := []cliResultRow{{
		"ip": "1.2.3.4", "port": "443", "latency": "20ms", "lossRate": "0%", "speed": "10MB/s",
		"dc": "HKG", "dcCountry": "HK", "region": "Hong Kong", "city": "Hong Kong",
	}}
	for _, format := range []string{"txt", "csv"} {
		fields := resolveCLIFields("compact", format, officialRows, nil)
		for _, required := range []string{"latency", "lossRate", "speed", "dcCountry"} {
			if !containsField(fields, required) {
				t.Errorf("official compact %s fields %v do not include %q", format, fields, required)
			}
		}
		content, err := formatCLIResults(officialRows, cliExportConfig{Format: format, Fields: "compact", Separator: "-"})
		if err != nil {
			t.Fatalf("formatCLIResults(%s): %v", format, err)
		}
		if format == "txt" {
			if !strings.Contains(content, "20ms-0%-10MB/s-HKG-HK") {
				t.Errorf("official compact TXT missing speed/landing country: %s", content)
			}
		} else if !strings.Contains(content, "网络延迟") || !strings.Contains(content, "丢包率") || !strings.Contains(content, "下载速度") || !strings.Contains(content, "落地区域") || !strings.Contains(content, "20ms") || !strings.Contains(content, "0%") || !strings.Contains(content, "10MB/s") || !strings.Contains(content, "HK") {
			t.Errorf("official compact CSV missing speed/landing country: %s", content)
		}
	}

	nsbRows := []cliResultRow{{
		"ip": "1.2.3.4", "port": "443", "tls": "true", "lossRate": "0%", "latency": "20ms", "speed": "10MB/s",
		"dc": "HKG", "dcCountry": "HK", "loc": "US", "ipType": "IPv4",
	}}
	for _, format := range []string{"txt", "csv"} {
		fields := resolveCLIFields("compact", format, nsbRows, nil)
		for _, required := range []string{"latency", "lossRate", "speed", "dcCountry"} {
			if !containsField(fields, required) {
				t.Errorf("NSB compact %s fields %v do not include %q", format, fields, required)
			}
		}
	}
}

func TestNSBExportsIncludeLandingCountryAndSpeed(t *testing.T) {
	result := iptestResult{
		ipAddr: "1.2.3.4", port: 443, dataCenter: "HKG", dcCountry: "HK",
		locCode: "US", latency: "20ms", lossRate: 0.05, speedText: "10.25MB/s",
	}
	for _, compact := range []bool{true, false} {
		headers := nsbCSVHeaders(compact, scanModeTCPing)
		row := nsbCSVRow(result, true, compact, scanModeTCPing)
		countryIndex := -1
		speedIndex := -1
		latencyIndex := -1
		lossRateIndex := -1
		for i, header := range headers {
			if header == "落地区域" {
				countryIndex = i
			}
			if header == "下载速度" {
				speedIndex = i
			}
			if header == "网络延迟" {
				latencyIndex = i
			}
			if header == "丢包率" {
				lossRateIndex = i
			}
		}
		if countryIndex < 0 || countryIndex >= len(row) || row[countryIndex] != "HK" {
			t.Errorf("compact=%v landing country missing or incorrect: headers=%v row=%v", compact, headers, row)
		}
		if speedIndex < 0 || speedIndex >= len(row) || row[speedIndex] != result.speedText {
			t.Errorf("compact=%v speed missing or incorrect: headers=%v row=%v", compact, headers, row)
		}
		if latencyIndex < 0 || latencyIndex >= len(row) || row[latencyIndex] != result.latency {
			t.Errorf("compact=%v latency missing or incorrect: headers=%v row=%v", compact, headers, row)
		}
		if lossRateIndex < 0 || lossRateIndex >= len(row) || row[lossRateIndex] != "5%" {
			t.Errorf("compact=%v loss rate missing or incorrect: headers=%v row=%v", compact, headers, row)
		}
	}
}

func TestNSBMessagesPreserveLandingCountry(t *testing.T) {
	message := (&iptestResult{ipAddr: "1.2.3.4", port: 443, dataCenter: "LAX", dcCountry: "US"}).toNSBMessage("12MB/s")
	if message.DCCountry != "US" || message.Speed != "12MB/s" {
		t.Fatalf("NSB message lost export values: %#v", message)
	}
	result, ok := nsbMessageToResult(message)
	if !ok || result.dcCountry != "US" || result.speedText != "12MB/s" {
		t.Fatalf("NSB result lost export values: %#v, ok=%v", result, ok)
	}
	encoded, err := json.Marshal(message)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(encoded), `"dcCountry":"US"`) {
		t.Fatalf("serialized NSB message missing dcCountry: %s", encoded)
	}
}

func TestOfficialFinalRowsIncludeLatencyAndLossRate(t *testing.T) {
	scanResults := []ScanResult{{IP: "1.2.3.4", Port: 443, DataCenter: "HKG", DCCountry: "HK", LatencyStr: "25ms"}}
	testResults := []TestResult{{IP: "1.2.3.4", Port: 443, AvgLatency: 25 * time.Millisecond, LossRate: 0.25}}
	rows := officialResultRows(scanResults, testResults, scanModeTCPing)
	if len(rows) != 1 || rows[0]["latency"] != "25ms" || rows[0]["lossRate"] != "25%" {
		t.Fatalf("official final export omitted latency/loss rate: %#v", rows)
	}

	scanOnly := officialScanRows(scanResults, scanModeTCPing)
	if len(scanOnly) != 1 || scanOnly[0]["latency"] == "" || scanOnly[0]["lossRate"] != "未测" {
		t.Fatalf("official scan-only export should identify unmeasured loss: %#v", scanOnly)
	}
}
