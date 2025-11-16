package gui

import (
	"sort"

	"github.com/riclib/volu/internal/config"
)

// Shortcut represents a radio series quick-access button
type Shortcut struct {
	Key  string // Config key (e.g., "asot", "queen")
	Name string // Display name (e.g., "A State of Trance")
	Icon string // Emoji icon
}

// LoadShortcuts creates shortcut buttons from radio config
func LoadShortcuts(cfg *config.Config) []Shortcut {
	shortcuts := make([]Shortcut, 0, len(cfg.Radio))

	for key, series := range cfg.Radio {
		shortcuts = append(shortcuts, Shortcut{
			Key:  key,
			Name: series.Name,
			Icon: getSeriesIcon(key),
		})
	}

	// Sort by key for consistent ordering
	sort.Slice(shortcuts, func(i, j int) bool {
		return shortcuts[i].Key < shortcuts[j].Key
	})

	return shortcuts
}

// getSeriesIcon returns an appropriate emoji for the series
func getSeriesIcon(seriesKey string) string {
	// Radio shows
	radioShows := map[string]bool{
		"asot":   true,
		"group":  true,
		"buddha": true,
	}

	// Rock/Metal artists
	rockArtists := map[string]bool{
		"queen":     true,
		"maiden":    true,
		"manowar":   true,
		"metallica": true,
	}

	if radioShows[seriesKey] {
		return "📻"
	}

	if rockArtists[seriesKey] {
		return "🎸"
	}

	// Default music note
	return "🎵"
}
