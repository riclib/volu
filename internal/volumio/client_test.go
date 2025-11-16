package volumio

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetState(t *testing.T) {
	// Mock response
	mockState := PlayerState{
		Status:   "play",
		Position: 42,
		Title:    "Test Song",
		Artist:   "Test Artist",
		Album:    "Test Album",
		Volume:   75,
		Mute:     false,
		Random:   true,
		Repeat:   false,
	}

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/getState" {
			t.Errorf("Expected path /api/v1/getState, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockState)
	}))
	defer server.Close()

	// Create client pointing to test server
	client := NewClient(server.URL)

	// Test GetState
	state, err := client.GetState()
	if err != nil {
		t.Fatalf("GetState() error = %v", err)
	}

	if state.Status != mockState.Status {
		t.Errorf("Status = %v, want %v", state.Status, mockState.Status)
	}
	if state.Title != mockState.Title {
		t.Errorf("Title = %v, want %v", state.Title, mockState.Title)
	}
	if state.Artist != mockState.Artist {
		t.Errorf("Artist = %v, want %v", state.Artist, mockState.Artist)
	}
}

func TestPlaybackCommands(t *testing.T) {
	tests := []struct {
		name        string
		method      func(*Client) error
		expectedCmd string
	}{
		{"Play", (*Client).Play, "play"},
		{"Pause", (*Client).Pause, "pause"},
		{"Stop", (*Client).Stop, "stop"},
		{"Next", (*Client).Next, "next"},
		{"Previous", (*Client).Previous, "prev"},
		{"TogglePlayPause", (*Client).TogglePlayPause, "toggle"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/v1/commands/" {
					t.Errorf("Expected path /api/v1/commands/, got %s", r.URL.Path)
				}
				cmd := r.URL.Query().Get("cmd")
				if cmd != tt.expectedCmd {
					t.Errorf("Expected cmd=%s, got %s", tt.expectedCmd, cmd)
				}
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
			}))
			defer server.Close()

			client := NewClient(server.URL)
			err := tt.method(client)
			if err != nil {
				t.Errorf("%s() error = %v", tt.name, err)
			}
		})
	}
}

func TestSetVolume(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cmd := r.URL.Query().Get("cmd")
		volume := r.URL.Query().Get("volume")

		if cmd != "volume" {
			t.Errorf("Expected cmd=volume, got %s", cmd)
		}
		if volume != "50" {
			t.Errorf("Expected volume=50, got %s", volume)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	err := client.SetVolume(50)
	if err != nil {
		t.Errorf("SetVolume() error = %v", err)
	}
}

func TestRealVolumioConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Test against real Volumio instance
	client := NewClient("http://volumio.local:3000")

	state, err := client.GetState()
	if err != nil {
		t.Skipf("Could not connect to volumio.local: %v", err)
	}

	t.Logf("Connected to Volumio successfully")
	t.Logf("Status: %s", state.Status)
	if state.Title != "" {
		t.Logf("Now playing: %s - %s", state.Artist, state.Title)
	}
}

func TestArtistSearchExploration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := NewClient("http://volumio.local:3000")

	// Test with "Arash" - expect 1 album by Arash + 2 Corsten's Countdown episodes featuring Arash
	artist := "Arash"

	t.Logf("\n========== Searching for: %s ==========\n", artist)

	albums, err := client.SearchAlbums(artist)
	if err != nil {
		t.Fatalf("Error searching for %s: %v", artist, err)
	}

	t.Logf("Found %d albums for '%s':\n", len(albums), artist)

	for i, album := range albums {
		t.Logf("  [%d]", i+1)
		t.Logf("      Title:       %s", album.Title)
		t.Logf("      Name:        %s", album.Name)
		t.Logf("      Artist:      %s", album.Artist)
		t.Logf("      Album:       %s", album.Album)
		t.Logf("      URI:         %s", album.URI)
		t.Logf("      Service:     %s", album.Service)
		t.Logf("      Type:        %s", album.Type)
		t.Logf("      DisplayName: %s", album.DisplayName())
		t.Logf("")
	}

	// Summary analysis
	t.Logf("\n========== ANALYSIS ==========")
	t.Logf("Total albums found: %d", len(albums))

	// Count how many have "Arash" in Artist field
	arashAsArtist := 0
	for _, album := range albums {
		if album.Artist == "Arash" || album.Artist == artist {
			arashAsArtist++
		}
	}
	t.Logf("Albums with Artist field exactly 'Arash': %d", arashAsArtist)

	// Show unique Artist values
	artistSet := make(map[string]bool)
	for _, album := range albums {
		if album.Artist != "" {
			artistSet[album.Artist] = true
		}
	}
	t.Logf("\nUnique Artist values found:")
	for a := range artistSet {
		t.Logf("  - %s", a)
	}
}

func TestTrackSearchExploration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := NewClient("http://volumio.local:3000")

	// Test with "Arash" to see tracks (should find tracks in Corsten's Countdown)
	artist := "Arash"

	t.Logf("\n========== Full Search Response for: %s ==========\n", artist)

	response, err := client.Search(artist)
	if err != nil {
		t.Fatalf("Error searching for %s: %v", artist, err)
	}

	t.Logf("Found %d result lists:\n", len(response.Navigation.Lists))

	for i, list := range response.Navigation.Lists {
		t.Logf("\n[List %d] %s", i+1, list.Title)
		t.Logf("  Items in this list: %d", len(list.Items))

		// Show first 5 items from each list as samples
		maxItems := 5
		if len(list.Items) < maxItems {
			maxItems = len(list.Items)
		}

		for j := 0; j < maxItems; j++ {
			item := list.Items[j]
			t.Logf("    [%d] Title: %s", j+1, item.Title)
			t.Logf("        Artist: %s", item.Artist)
			t.Logf("        Album: %s", item.Album)
			t.Logf("        Type: %s, Service: %s", item.Type, item.Service)
			t.Logf("        URI: %s", item.URI)
		}

		if len(list.Items) > maxItems {
			t.Logf("    ... and %d more items", len(list.Items)-maxItems)
		}
	}

	// Now specifically look for tracks
	t.Logf("\n========== TRACK ANALYSIS ==========")
	for _, list := range response.Navigation.Lists {
		// Use strings.Contains with lowercase for simplicity
		listTitleLower := ""
		for _, c := range list.Title {
			if c >= 'A' && c <= 'Z' {
				listTitleLower += string(c + 32)
			} else {
				listTitleLower += string(c)
			}
		}

		if listTitleLower != "" && (hasSubstring(listTitleLower, "tracks") || hasSubstring(listTitleLower, "track")) {
			t.Logf("\nFound track list: %s", list.Title)
			t.Logf("Total tracks: %d", len(list.Items))

			// Count tracks by artist
			tracksByArtist := make(map[string]int)
			tracksWithArtistField := 0

			for _, item := range list.Items {
				if item.Artist != "" {
					tracksByArtist[item.Artist]++
					tracksWithArtistField++
				}
			}

			t.Logf("Tracks with Artist field populated: %d", tracksWithArtistField)
			t.Logf("\nTracks grouped by Artist field:")
			for artist, count := range tracksByArtist {
				t.Logf("  %s: %d tracks", artist, count)
			}
		}
	}
}

// Helper function for simple substring check
func hasSubstring(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func TestSearchTracks(t *testing.T) {
	// Mock search response with tracks
	mockResponse := map[string]interface{}{
		"navigation": map[string]interface{}{
			"isSearchResult": true,
			"lists": []map[string]interface{}{
				{
					"title": "Found 3 Tracks 'TestArtist'",
					"items": []map[string]interface{}{
						{
							"title":   "Track 1",
							"artist":  "TestArtist",
							"album":   "Album 1",
							"type":    "song",
							"service": "mpd",
							"uri":     "music-library/track1.flac",
						},
						{
							"title":   "Track 2",
							"artist":  "TestArtist",
							"album":   "Album 1",
							"type":    "song",
							"service": "mpd",
							"uri":     "music-library/track2.flac",
						},
						{
							"title":   "Track 3 (feat. TestArtist)",
							"artist":  "OtherArtist",
							"album":   "Other Album",
							"type":    "song",
							"service": "mpd",
							"uri":     "music-library/track3.flac",
						},
					},
				},
				{
					"title": "TIDAL Tracks",
					"items": []map[string]interface{}{
						{
							"title":   "TIDAL Track",
							"artist":  "TestArtist",
							"type":    "song",
							"service": "tidal",
							"uri":     "tidal://song/123",
						},
					},
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/search" {
			t.Errorf("Expected path /api/v1/search, got %s", r.URL.Path)
		}
		query := r.URL.Query().Get("query")
		if query != "TestArtist" {
			t.Errorf("Expected query=TestArtist, got %s", query)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	tracks, err := client.SearchTracks("TestArtist")
	if err != nil {
		t.Fatalf("SearchTracks() error = %v", err)
	}

	// Should return only local tracks (type=song, service=mpd), not TIDAL
	expectedCount := 3
	if len(tracks) != expectedCount {
		t.Errorf("Expected %d tracks, got %d", expectedCount, len(tracks))
	}

	// Check first track
	if tracks[0].Title != "Track 1" {
		t.Errorf("Expected title 'Track 1', got '%s'", tracks[0].Title)
	}
	if tracks[0].Artist != "TestArtist" {
		t.Errorf("Expected artist 'TestArtist', got '%s'", tracks[0].Artist)
	}
	if tracks[0].Type != "song" {
		t.Errorf("Expected type 'song', got '%s'", tracks[0].Type)
	}
	if tracks[0].Service != "mpd" {
		t.Errorf("Expected service 'mpd', got '%s'", tracks[0].Service)
	}

	// Verify TIDAL track was filtered out
	for _, track := range tracks {
		if track.Service == "tidal" {
			t.Errorf("TIDAL track should have been filtered out: %v", track)
		}
	}
}
