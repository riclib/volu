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

func TestRadioSeriesWithArtistFields(t *testing.T) {
	// Test album-based series with artist pattern
	albumSeries := RadioSeries{
		Name:          "Armin Albums",
		AlbumArtist:   "Armin van Buuren",
		AlbumPattern:  "^ASOT\\s+\\d+",
	}

	if albumSeries.AlbumArtist != "Armin van Buuren" {
		t.Errorf("Expected album artist 'Armin van Buuren', got '%s'", albumSeries.AlbumArtist)
	}

	if albumSeries.AlbumPattern != "^ASOT\\s+\\d+" {
		t.Errorf("Expected album pattern '^ASOT\\s+\\d+', got '%s'", albumSeries.AlbumPattern)
	}

	// Test track-based series
	trackSeries := RadioSeries{
		Name:        "Arash Tracks",
		TrackArtist: "Arash",
		TrackCount:  50,
	}

	if trackSeries.TrackArtist != "Arash" {
		t.Errorf("Expected track artist 'Arash', got '%s'", trackSeries.TrackArtist)
	}

	if trackSeries.TrackCount != 50 {
		t.Errorf("Expected track count 50, got %d", trackSeries.TrackCount)
	}
}

func TestBackwardCompatibility(t *testing.T) {
	// Old format should still work
	oldFormat := `
host: volumio.local
radio:
  asot:
    name: "A State of Trance"
    search_query: "ASOT"
    pattern: "^ASOT\\s+\\d+"
`

	cfg := &Config{}
	err := yaml.Unmarshal([]byte(oldFormat), cfg)
	if err != nil {
		t.Fatalf("Unmarshal() failed for old format: %v", err)
	}

	asot, exists := cfg.Radio["asot"]
	if !exists {
		t.Fatal("Expected 'asot' radio series to exist")
	}

	if asot.SearchQuery != "ASOT" {
		t.Errorf("Expected search query 'ASOT', got '%s'", asot.SearchQuery)
	}

	if asot.Pattern != "^ASOT\\s+\\d+" {
		t.Errorf("Expected pattern '^ASOT\\s+\\d+', got '%s'", asot.Pattern)
	}
}

func TestNewAlbumArtistFormat(t *testing.T) {
	newFormat := `
host: volumio.local
radio:
  armin:
    name: "Armin Albums"
    album_artist: "Armin van Buuren"
    album_pattern: "^ASOT\\s+\\d+"
`

	cfg := &Config{}
	err := yaml.Unmarshal([]byte(newFormat), cfg)
	if err != nil {
		t.Fatalf("Unmarshal() failed for new album format: %v", err)
	}

	armin, exists := cfg.Radio["armin"]
	if !exists {
		t.Fatal("Expected 'armin' radio series to exist")
	}

	if armin.AlbumArtist != "Armin van Buuren" {
		t.Errorf("Expected album artist 'Armin van Buuren', got '%s'", armin.AlbumArtist)
	}

	if armin.AlbumPattern != "^ASOT\\s+\\d+" {
		t.Errorf("Expected album pattern '^ASOT\\s+\\d+', got '%s'", armin.AlbumPattern)
	}
}

func TestNewTrackArtistFormat(t *testing.T) {
	newFormat := `
host: volumio.local
radio:
  arash:
    name: "Arash Tracks"
    track_artist: "Arash"
    track_count: 50
`

	cfg := &Config{}
	err := yaml.Unmarshal([]byte(newFormat), cfg)
	if err != nil {
		t.Fatalf("Unmarshal() failed for new track format: %v", err)
	}

	arash, exists := cfg.Radio["arash"]
	if !exists {
		t.Fatal("Expected 'arash' radio series to exist")
	}

	if arash.TrackArtist != "Arash" {
		t.Errorf("Expected track artist 'Arash', got '%s'", arash.TrackArtist)
	}

	if arash.TrackCount != 50 {
		t.Errorf("Expected track count 50, got %d", arash.TrackCount)
	}
}
