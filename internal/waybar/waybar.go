package waybar

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/riclib/volu/internal/volumio"
)

// Output represents the JSON output for Waybar
type Output struct {
	Text       string `json:"text"`
	Tooltip    string `json:"tooltip"`
	Class      string `json:"class"`
	Percentage int    `json:"percentage"`
}

// EscapeMarkup escapes special characters for Pango markup
func EscapeMarkup(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}

// FormatTime formats seconds as MM:SS
func FormatTime(seconds int) string {
	if seconds < 0 {
		return "0:00"
	}
	mins := seconds / 60
	secs := seconds % 60
	return fmt.Sprintf("%d:%02d", mins, secs)
}

// GetStatusIcon returns an icon for the playback status
func GetStatusIcon(status string) string {
	switch status {
	case "play":
		return "▶"
	case "pause":
		return "⏸"
	case "stop":
		return "⏹"
	default:
		return "⏹"
	}
}

// GetVolumeIcon returns an icon for the volume level
func GetVolumeIcon(volume int, muted bool) string {
	if muted || volume == 0 {
		return "🔇"
	} else if volume < 33 {
		return "🔈"
	} else if volume < 66 {
		return "🔉"
	}
	return "🔊"
}

// CreateOutput creates Waybar JSON output from player state
func CreateOutput(state *volumio.PlayerState, baseURL string) Output {
	statusIcon := GetStatusIcon(state.Status)

	// Build main text
	text := "♫ Volumio"
	if state.Title != "" {
		artist := state.Artist
		if artist == "" {
			artist = "Unknown Artist"
		}
		text = fmt.Sprintf("♫ %s - %s", EscapeMarkup(artist), EscapeMarkup(state.Title))
	}
	text = fmt.Sprintf("%s %s", text, statusIcon)

	// Build tooltip
	tooltip := ""
	if state.Title != "" {
		tooltip += fmt.Sprintf("Title: %s\n", EscapeMarkup(state.Title))
	}
	if state.Artist != "" {
		tooltip += fmt.Sprintf("Artist: %s\n", EscapeMarkup(state.Artist))
	}
	if state.Album != "" {
		tooltip += fmt.Sprintf("Album: %s\n", EscapeMarkup(state.Album))
	}
	if state.Service != "" {
		tooltip += fmt.Sprintf("Service: %s\n", state.Service)
	}

	// Add playback info
	if state.Duration > 0 {
		currentTime := FormatTime(state.Position)
		totalTime := FormatTime(state.Duration)
		progress := 0
		if state.Duration > 0 {
			progress = (state.Position * 100) / state.Duration
		}
		tooltip += fmt.Sprintf("Time: %s / %s (%d%%)\n", currentTime, totalTime, progress)
	}

	// Add volume info
	volumeIcon := GetVolumeIcon(state.Volume, state.Mute)
	muteText := ""
	if state.Mute {
		muteText = " (Muted)"
	}
	tooltip += fmt.Sprintf("Volume: %s %d%%%s\n", volumeIcon, state.Volume, muteText)

	// Add playback modes
	modes := []string{}
	if state.Random {
		modes = append(modes, "🔀 Shuffle")
	}
	if state.Repeat {
		modes = append(modes, "🔁 Repeat")
	}
	if len(modes) > 0 {
		tooltip += fmt.Sprintf("Modes: %s", modes)
	}

	// Remove trailing newline
	if len(tooltip) > 0 && tooltip[len(tooltip)-1] == '\n' {
		tooltip = tooltip[:len(tooltip)-1]
	}

	if tooltip == "" {
		tooltip = "No track playing"
	}

	// CSS class based on status
	class := fmt.Sprintf("volumio-%s", state.Status)

	// Calculate percentage for progress bar
	percentage := 0
	if state.Duration > 0 {
		percentage = (state.Position * 100) / state.Duration
	}

	return Output{
		Text:       text,
		Tooltip:    tooltip,
		Class:      class,
		Percentage: percentage,
	}
}

// CreateErrorOutput creates error output for Waybar
func CreateErrorOutput(errorMsg string) Output {
	return Output{
		Text:       "♫ Volumio (disconnected)",
		Tooltip:    fmt.Sprintf("Error: %s", errorMsg),
		Class:      "volumio-error",
		Percentage: 0,
	}
}

// PrintJSON outputs the result as JSON
func PrintJSON(output Output) error {
	data, err := json.Marshal(output)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}
