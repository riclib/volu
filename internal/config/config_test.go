package config

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Host != "volumio.local" {
		t.Errorf("Expected default host 'volumio.local', got '%s'", cfg.Host)
	}

	if cfg.Radio == nil {
		t.Error("Expected Radio map to be initialized")
	}

	if len(cfg.Radio) != 0 {
		t.Errorf("Expected empty Radio map, got %d entries", len(cfg.Radio))
	}
}

func TestYAMLMarshalUnmarshal(t *testing.T) {
	// Create test config
	testCfg := &Config{
		Host: "192.168.1.100",
		Radio: map[string]RadioSeries{
			"asot": {
				Name:        "A State of Trance",
				SearchQuery: "ASOT",
				Pattern:     "^ASOT\\s+\\d+",
			},
			"grouptherapy": {
				Name:        "Group Therapy",
				SearchQuery: "Group Therapy",
				Pattern:     "Group Therapy\\s+\\d+",
			},
		},
	}

	// Marshal to YAML
	data, err := yaml.Marshal(testCfg)
	if err != nil {
		t.Fatalf("Marshal() failed: %v", err)
	}

	// Unmarshal back to Config
	loadedCfg := &Config{}
	err = yaml.Unmarshal(data, loadedCfg)
	if err != nil {
		t.Fatalf("Unmarshal() failed: %v", err)
	}

	// Verify loaded config matches original
	if loadedCfg.Host != testCfg.Host {
		t.Errorf("Expected host '%s', got '%s'", testCfg.Host, loadedCfg.Host)
	}

	if len(loadedCfg.Radio) != 2 {
		t.Fatalf("Expected 2 radio series, got %d", len(loadedCfg.Radio))
	}

	asot, exists := loadedCfg.Radio["asot"]
	if !exists {
		t.Fatal("Expected 'asot' radio series to exist")
	}

	if asot.Name != "A State of Trance" {
		t.Errorf("Expected name 'A State of Trance', got '%s'", asot.Name)
	}

	if asot.SearchQuery != "ASOT" {
		t.Errorf("Expected search query 'ASOT', got '%s'", asot.SearchQuery)
	}

	if asot.Pattern != "^ASOT\\s+\\d+" {
		t.Errorf("Expected pattern '^ASOT\\s+\\d+', got '%s'", asot.Pattern)
	}

	gt, exists := loadedCfg.Radio["grouptherapy"]
	if !exists {
		t.Fatal("Expected 'grouptherapy' radio series to exist")
	}

	if gt.Name != "Group Therapy" {
		t.Errorf("Expected name 'Group Therapy', got '%s'", gt.Name)
	}
}

func TestRadioSeriesStruct(t *testing.T) {
	series := RadioSeries{
		Name:        "Test Series",
		SearchQuery: "test",
		Pattern:     "^test\\d+",
	}

	if series.Name != "Test Series" {
		t.Errorf("Expected name 'Test Series', got '%s'", series.Name)
	}

	if series.SearchQuery != "test" {
		t.Errorf("Expected search query 'test', got '%s'", series.SearchQuery)
	}

	if series.Pattern != "^test\\d+" {
		t.Errorf("Expected pattern '^test\\d+', got '%s'", series.Pattern)
	}
}
