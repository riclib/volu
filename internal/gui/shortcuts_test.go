package gui

import (
	"testing"

	"github.com/riclib/volu/internal/config"
)

func TestShortcuts_LoadFromConfig(t *testing.T) {
	cfg := &config.Config{
		Radio: map[string]config.RadioSeries{
			"asot": {
				Name:        "A State of Trance",
				SearchQuery: "ASOT",
				Pattern:     "^ASOT\\s+\\d+",
			},
			"queen": {
				Name:        "Queen",
				AlbumArtist: "Queen",
			},
			"maiden": {
				Name:        "Iron Maiden",
				AlbumArtist: "Iron Maiden",
			},
		},
	}

	shortcuts := LoadShortcuts(cfg)

	if len(shortcuts) != 3 {
		t.Errorf("Expected 3 shortcuts, got %d", len(shortcuts))
	}

	// Check shortcuts are sorted by key
	if shortcuts[0].Key != "asot" {
		t.Errorf("Expected first shortcut key 'asot', got '%s'", shortcuts[0].Key)
	}

	if shortcuts[0].Name != "A State of Trance" {
		t.Errorf("Expected first shortcut name 'A State of Trance', got '%s'", shortcuts[0].Name)
	}

	// Verify all shortcuts have required fields
	for _, sc := range shortcuts {
		if sc.Key == "" {
			t.Error("Shortcut has empty Key")
		}
		if sc.Name == "" {
			t.Error("Shortcut has empty Name")
		}
		if sc.Icon == "" {
			t.Error("Shortcut has empty Icon")
		}
	}
}

func TestShortcuts_GetIcon(t *testing.T) {
	tests := []struct {
		seriesKey string
		expected  string
	}{
		{"asot", "📻"},
		{"group", "📻"},
		{"buddha", "📻"},
		{"queen", "🎸"},
		{"maiden", "🎸"},
		{"manowar", "🎸"},
		{"metallica", "🎸"},
		{"enigma", "🎵"},
		{"straits", "🎵"},
		{"unknown", "🎵"}, // default
	}

	for _, tt := range tests {
		result := getSeriesIcon(tt.seriesKey)
		if result != tt.expected {
			t.Errorf("getSeriesIcon(%s) = %s; want %s", tt.seriesKey, result, tt.expected)
		}
	}
}

func TestShortcuts_EmptyConfig(t *testing.T) {
	cfg := &config.Config{
		Radio: map[string]config.RadioSeries{},
	}

	shortcuts := LoadShortcuts(cfg)

	if len(shortcuts) != 0 {
		t.Errorf("Expected 0 shortcuts for empty config, got %d", len(shortcuts))
	}
}

func TestShortcuts_SortingOrder(t *testing.T) {
	cfg := &config.Config{
		Radio: map[string]config.RadioSeries{
			"zebra":  {Name: "Zebra"},
			"alpha":  {Name: "Alpha"},
			"middle": {Name: "Middle"},
		},
	}

	shortcuts := LoadShortcuts(cfg)

	if shortcuts[0].Key != "alpha" {
		t.Errorf("Expected first key 'alpha', got '%s'", shortcuts[0].Key)
	}

	if shortcuts[1].Key != "middle" {
		t.Errorf("Expected second key 'middle', got '%s'", shortcuts[1].Key)
	}

	if shortcuts[2].Key != "zebra" {
		t.Errorf("Expected third key 'zebra', got '%s'", shortcuts[2].Key)
	}
}
