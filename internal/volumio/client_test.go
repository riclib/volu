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
