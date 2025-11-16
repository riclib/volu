package gui

import (
	"testing"

	"github.com/riclib/volu/internal/volumio"
)

func TestNowPlaying_LoadState(t *testing.T) {
	// Mock client that returns known state
	mockState := &volumio.PlayerState{
		Status:   "play",
		Title:    "Astronomy Domine",
		Artist:   "Pink Floyd",
		Album:    "The Piper at the Gates of Dawn",
		AlbumArt: "/albumart?path=music/Pink%20Floyd",
		Duration: 252,
		Position: 45,
		Volume:   75,
	}

	np := NewNowPlaying(mockState)

	if np.Title != "Astronomy Domine" {
		t.Errorf("Expected title 'Astronomy Domine', got '%s'", np.Title)
	}

	if np.Artist != "Pink Floyd" {
		t.Errorf("Expected artist 'Pink Floyd', got '%s'", np.Artist)
	}

	if np.Album != "The Piper at the Gates of Dawn" {
		t.Errorf("Expected album 'The Piper at the Gates of Dawn', got '%s'", np.Album)
	}

	if np.Status != "play" {
		t.Errorf("Expected status 'play', got '%s'", np.Status)
	}

	if np.Progress != "0:45 / 4:12" {
		t.Errorf("Expected progress '0:45 / 4:12', got '%s'", np.Progress)
	}

	if np.Volume != 75 {
		t.Errorf("Expected volume 75, got %d", np.Volume)
	}
}

func TestNowPlaying_FormatTime(t *testing.T) {
	tests := []struct {
		seconds  int
		expected string
	}{
		{45, "0:45"},
		{252, "4:12"},
		{3661, "1:01:01"},
		{0, "0:00"},
		{59, "0:59"},
	}

	for _, tt := range tests {
		result := formatTime(tt.seconds)
		if result != tt.expected {
			t.Errorf("formatTime(%d) = %s; want %s", tt.seconds, result, tt.expected)
		}
	}
}

func TestNowPlaying_AlbumArtURL(t *testing.T) {
	tests := []struct {
		name     string
		albumArt string
		host     string
		expected string
	}{
		{
			name:     "relative path",
			albumArt: "/albumart?path=music/Pink%20Floyd",
			host:     "volumio.local",
			expected: "http://volumio.local:3000/albumart?path=music/Pink%20Floyd",
		},
		{
			name:     "absolute URL",
			albumArt: "http://volumio.local:3000/albumart?path=music",
			host:     "volumio.local",
			expected: "http://volumio.local:3000/albumart?path=music",
		},
		{
			name:     "empty album art",
			albumArt: "",
			host:     "volumio.local",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getAlbumArtURL(tt.albumArt, tt.host)
			if result != tt.expected {
				t.Errorf("getAlbumArtURL(%s, %s) = %s; want %s",
					tt.albumArt, tt.host, result, tt.expected)
			}
		})
	}
}

func TestNowPlaying_StatusIcon(t *testing.T) {
	tests := []struct {
		status   string
		expected string
	}{
		{"play", "⏸️"},
		{"pause", "▶️"},
		{"stop", "▶️"},
		{"", "▶️"},
	}

	for _, tt := range tests {
		result := getStatusIcon(tt.status)
		if result != tt.expected {
			t.Errorf("getStatusIcon(%s) = %s; want %s", tt.status, result, tt.expected)
		}
	}
}
