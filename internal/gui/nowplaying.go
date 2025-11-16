package gui

import (
	"fmt"
	"strings"

	"github.com/riclib/volu/internal/volumio"
)

// NowPlaying represents the current playback state for display
type NowPlaying struct {
	Title     string
	Artist    string
	Album     string
	AlbumArt  string
	Status    string // "play", "pause", "stop"
	Progress  string // "3:45 / 75:00"
	Volume    int    // 0-100
	VoluHost  string // Volumio host for album art URLs
}

// NewNowPlaying creates a NowPlaying display model from PlayerState
func NewNowPlaying(state *volumio.PlayerState) *NowPlaying {
	if state == nil {
		return &NowPlaying{
			Title:    "No track playing",
			Artist:   "",
			Album:    "",
			Status:   "stop",
			Progress: "0:00 / 0:00",
			Volume:   0,
		}
	}

	progress := fmt.Sprintf("%s / %s",
		formatTime(state.Position),
		formatTime(state.Duration))

	return &NowPlaying{
		Title:    state.Title,
		Artist:   state.Artist,
		Album:    state.Album,
		AlbumArt: state.AlbumArt,
		Status:   state.Status,
		Progress: progress,
		Volume:   state.Volume,
	}
}

// formatTime converts seconds to MM:SS or H:MM:SS format
func formatTime(seconds int) string {
	if seconds < 0 {
		seconds = 0
	}

	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	secs := seconds % 60

	if hours > 0 {
		return fmt.Sprintf("%d:%02d:%02d", hours, minutes, secs)
	}
	return fmt.Sprintf("%d:%02d", minutes, secs)
}

// getAlbumArtURL converts album art path to full URL
func getAlbumArtURL(albumArt, host string) string {
	if albumArt == "" {
		return ""
	}

	// If already absolute URL, return as-is
	if strings.HasPrefix(albumArt, "http://") || strings.HasPrefix(albumArt, "https://") {
		return albumArt
	}

	// Convert relative path to absolute URL
	return fmt.Sprintf("http://%s:3000%s", host, albumArt)
}

// getStatusIcon returns the appropriate icon for playback status
func getStatusIcon(status string) string {
	if status == "play" {
		return "⏸️" // Playing: show pause icon
	}
	return "▶️" // Paused/stopped: show play icon
}
