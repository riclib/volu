package radio

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/riclib/volu/internal/volumio"
)

// TestPlayRandomTracks tests the new track-based radio mode
func TestPlayRandomTracks(t *testing.T) {
	queuedURIs := []string{}
	replaceCalled := false

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/search":
			// Return mock search response with tracks
			response := map[string]interface{}{
				"navigation": map[string]interface{}{
					"isSearchResult": true,
					"lists": []map[string]interface{}{
						{
							"title": "Found 5 Tracks 'Arash'",
							"items": []map[string]interface{}{
								{"title": "Track 1", "artist": "Arash", "type": "song", "service": "mpd", "uri": "track1.flac"},
								{"title": "Track 2", "artist": "Arash", "type": "song", "service": "mpd", "uri": "track2.flac"},
								{"title": "Track 3", "artist": "Arash", "type": "song", "service": "mpd", "uri": "track3.flac"},
								{"title": "Track 4", "artist": "OtherArtist", "type": "song", "service": "mpd", "uri": "track4.flac"},
								{"title": "Track 5", "artist": "Arash", "type": "song", "service": "mpd", "uri": "track5.flac"},
							},
						},
					},
				},
			}
			json.NewEncoder(w).Encode(response)

		case "/api/v1/getState":
			// Return mock state with shuffle off
			state := volumio.PlayerState{Random: false}
			json.NewEncoder(w).Encode(state)

		case "/api/v1/replaceAndPlay":
			replaceCalled = true
			w.WriteHeader(http.StatusOK)

		case "/api/v1/addToQueue":
			// Track queued URIs
			if err := r.ParseForm(); err != nil {
				t.Errorf("Failed to parse form: %v", err)
			}
			uri := r.Form.Get("uri")
			queuedURIs = append(queuedURIs, uri)
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	client := volumio.NewClient(server.URL)
	player := NewPlayer(client)

	// Test: Play 3 random tracks by Arash
	err := player.PlayRandomTracks("Arash", "(?i)arash", 3)
	if err != nil {
		t.Fatalf("PlayRandomTracks() error = %v", err)
	}

	// Verify ReplaceAndPlay was called for first track
	if !replaceCalled {
		t.Error("Expected ReplaceAndPlay to be called for first track")
	}

	// Verify remaining tracks were queued (3 total - 1 replaced = 2 queued)
	if len(queuedURIs) != 2 {
		t.Errorf("Expected 2 tracks to be queued, got %d", len(queuedURIs))
	}
}

// TestFilterTracksByArtist tests artist-based filtering for tracks
func TestFilterTracksByArtist(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/search" {
			response := map[string]interface{}{
				"navigation": map[string]interface{}{
					"isSearchResult": true,
					"lists": []map[string]interface{}{
						{
							"title": "Found Tracks",
							"items": []map[string]interface{}{
								{"title": "Track 1", "artist": "Arash", "type": "song", "service": "mpd", "uri": "track1.flac"},
								{"title": "Track 2", "artist": "Ferry Corsten", "type": "song", "service": "mpd", "uri": "track2.flac"},
								{"title": "Track 3", "artist": "Arash", "type": "song", "service": "mpd", "uri": "track3.flac"},
							},
						},
					},
				},
			}
			json.NewEncoder(w).Encode(response)
		}
	}))
	defer server.Close()

	client := volumio.NewClient(server.URL)
	player := NewPlayer(client)

	// Filter by artist name (case-insensitive regex)
	tracks, err := player.findMatchingTracks("Arash", "(?i)^arash$")
	if err != nil {
		t.Fatalf("findMatchingTracks() error = %v", err)
	}

	// Should return only 2 tracks with artist "Arash"
	expected := 2
	if len(tracks) != expected {
		t.Errorf("Expected %d tracks, got %d", expected, len(tracks))
	}

	// Verify all returned tracks match the artist pattern
	for _, track := range tracks {
		if track.Artist != "Arash" {
			t.Errorf("Expected artist 'Arash', got '%s' for track '%s'", track.Artist, track.Title)
		}
	}
}

// TestQueueTracks tests individual track queueing
func TestQueueTracks(t *testing.T) {
	queuedURIs := []string{}
	replaceCalled := false
	var firstURI string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/replaceAndPlay":
			// Parse JSON body
			var payload map[string]string
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Errorf("Failed to decode JSON: %v", err)
			}
			firstURI = payload["uri"]
			replaceCalled = true
			w.WriteHeader(http.StatusOK)

		case "/api/v1/addToQueue":
			// Parse JSON body
			var payload map[string]string
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Errorf("Failed to decode JSON: %v", err)
			}
			uri := payload["uri"]
			queuedURIs = append(queuedURIs, uri)
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	client := volumio.NewClient(server.URL)
	player := NewPlayer(client)

	// Create test tracks
	tracks := []volumio.BrowseItem{
		{Title: "Track 1", URI: "track1.flac", Service: "mpd"},
		{Title: "Track 2", URI: "track2.flac", Service: "mpd"},
		{Title: "Track 3", URI: "track3.flac", Service: "mpd"},
	}

	err := player.queueTracks(tracks)
	if err != nil {
		t.Fatalf("queueTracks() error = %v", err)
	}

	// Verify first track was replaced
	if !replaceCalled {
		t.Error("Expected ReplaceAndPlay to be called")
	}

	if firstURI != "track1.flac" {
		t.Errorf("Expected first URI 'track1.flac', got '%s'", firstURI)
	}

	// Verify remaining tracks were queued
	if len(queuedURIs) != 2 {
		t.Errorf("Expected 2 queued tracks, got %d", len(queuedURIs))
	}

	expectedQueuedURIs := []string{"track2.flac", "track3.flac"}
	for i, uri := range queuedURIs {
		if uri != expectedQueuedURIs[i] {
			t.Errorf("Expected queued URI[%d] '%s', got '%s'", i, expectedQueuedURIs[i], uri)
		}
	}
}
