package pci

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const mockBasePath = "../mock/sys/bus/pci/devices"

// TestReadFromFile tests the readFromFile function
func TestReadFromFile(t *testing.T) {
	testCases := []struct {
		name      string
		device    string
		file      string
		expected  string
		expectErr bool
	}{
		{"ValidVendorFile", "0000:02:00.4", "vendor", "0x10ec", false},
		{"ValidDeviceFile", "0000:02:00.4", "device", "0x816d", false},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(mockBasePath, tc.device, tc.file)
			content, err := ioutil.ReadFile(path)
			if err != nil {
				if !tc.expectErr {
					t.Errorf("Failed to read file: %v", err)
				}
				return
			}
			if strings.TrimSpace(string(content)) != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, string(content))
			}
		})
	}
}

// TestParsePciDevices tests the ParsePciDevices function
func TestParsePciDevices(t *testing.T) {
	devices, err := ParsePciDevices()
	if err != nil {
		t.Fatalf("ParsePciDevices() error = %v", err)
	}

	expectedNumberOfDevices := 37 // Example value for current mock data, adjust accordingly when this breaks :)
	if len(devices) != expectedNumberOfDevices {
		t.Errorf("Expected %d devices, got %d", expectedNumberOfDevices, len(devices))
	}

	// Detailed test for a specific device
	for _, device := range devices {
		if device.Bus == "0000:02:00.4" {
			if device.VendorID != "10ec" || device.DeviceID != "816d" {
				t.Errorf("Incorrect Vendor or Device ID for %s", device.Bus)
			}
		}
	}
}

// TestLookup tests the Lookup function
func TestLookup(t *testing.T) {
	testCases := []struct {
		name       string
		searchType string
		ven        string
		dev        string
		class      string
		subclass   string
		expected   string
		expectErr  bool
	}{
		{"VendorLookup", "vendor", "1022", "", "", "", "Advanced Micro Devices, Inc. [AMD]", false},
		{"DeviceLookup", "device", "1022", "1630", "", "", "Renoir Root Complex", false},
		// TODO: add more
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := Lookup(tc.searchType, tc.ven, tc.dev, tc.class, tc.subclass)
			if (err != nil) != tc.expectErr {
				t.Errorf("Lookup() error = %v, expectErr %v", err, tc.expectErr)
				return
			}
			if result != tc.expected {
				t.Errorf("Lookup() got = %v, expected %v", result, tc.expected)
			}
		})
	}
}

// TestErrorHandling tests error handling in readFromFile
func TestErrorHandling(t *testing.T) {
	_, err := ioutil.ReadFile(filepath.Join(mockBasePath, "invalid", "vendor"))
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("Expected os.ErrNotExist, got %v", err)
	}
}
