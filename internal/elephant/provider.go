package elephant

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/riclib/volu/internal/volumio"
)

// Provider implements the Elephant provider interface for Volumio
type Provider struct {
	client *volumio.Client
}

// NewProvider creates a new Elephant provider
func NewProvider(host string) *Provider {
	return &Provider{
		client: volumio.NewClientWithHost(host),
	}
}

// Entry represents an Elephant entry
type Entry struct {
	Label       string   `json:"label"`
	Sub         string   `json:"sub,omitempty"`
	Exec        string   `json:"exec,omitempty"`
	Image       string   `json:"image,omitempty"`
	Categories  []string `json:"categories,omitempty"`
	Searchable  bool     `json:"searchable"`
	Piped       string   `json:"piped,omitempty"`
}

// Response represents an Elephant provider response
type Response struct {
	Entries []Entry `json:"entries"`
}

// Run starts the Elephant provider
func (p *Provider) Run() error {
	log.Println("Starting Volumio Elephant provider...")

	// Elephant providers work by reading JSON from stdin and outputting JSON to stdout
	// For now, we'll implement a basic structure

	// Read input (if any) from stdin
	var input map[string]interface{}
	decoder := json.NewDecoder(os.Stdin)
	if err := decoder.Decode(&input); err != nil {
		// No input or invalid - show main menu
		return p.ShowMainMenu()
	}

	// Handle piped input (selected item)
	if piped, ok := input["piped"].(string); ok {
		return p.HandleAction(piped)
	}

	return p.ShowMainMenu()
}

// ShowMainMenu outputs the main menu entries
func (p *Provider) ShowMainMenu() error {
	state, err := p.client.GetState()
	if err != nil {
		log.Printf("Warning: Could not get Volumio state: %v", err)
		state = nil
	}

	entries := []Entry{}

	// Now playing section
	if state != nil && state.Title != "" {
		statusIcon := "▶"
		if state.Status == "pause" {
			statusIcon = "⏸"
		} else if state.Status == "stop" {
			statusIcon = "⏹"
		}

		artist := state.Artist
		if artist == "" {
			artist = "Unknown"
		}

		entries = append(entries, Entry{
			Label:      fmt.Sprintf("%s %s - %s", statusIcon, artist, state.Title),
			Sub:        fmt.Sprintf("%s | Vol: %d%%", state.Album, state.Volume),
			Searchable: false,
			Piped:      "action:toggle",
		})

		entries = append(entries, Entry{
			Label:      "────────────────────────",
			Searchable: false,
		})
	}

	// Playback controls
	entries = append(entries,
		Entry{
			Label:      "▶️ Play / Pause",
			Sub:        "Toggle playback",
			Searchable: true,
			Piped:      "action:toggle",
		},
		Entry{
			Label:      "⏭️ Next Track",
			Sub:        "Skip to next track",
			Searchable: true,
			Piped:      "action:next",
		},
		Entry{
			Label:      "⏮️ Previous Track",
			Sub:        "Go to previous track",
			Searchable: true,
			Piped:      "action:prev",
		},
		Entry{
			Label:      "⏹️ Stop",
			Sub:        "Stop playback",
			Searchable: true,
			Piped:      "action:stop",
		},
	)

	entries = append(entries, Entry{
		Label:      "────────────────────────",
		Searchable: false,
	})

	// Volume controls
	if state != nil {
		entries = append(entries,
			Entry{
				Label:      fmt.Sprintf("🔊 Volume: %d%%", state.Volume),
				Sub:        "Current volume",
				Searchable: false,
			},
			Entry{
				Label:      "🔊 Volume Up",
				Sub:        "Increase volume (+10%)",
				Searchable: true,
				Piped:      "action:volup",
			},
			Entry{
				Label:      "🔉 Volume Down",
				Sub:        "Decrease volume (-10%)",
				Searchable: true,
				Piped:      "action:voldown",
			},
			Entry{
				Label:      "🔇 Toggle Mute",
				Sub:        "Mute/unmute audio",
				Searchable: true,
				Piped:      "action:mute",
			},
		)
	}

	entries = append(entries, Entry{
		Label:      "────────────────────────",
		Searchable: false,
	})

	// Playback modes
	shuffleStatus := "OFF"
	repeatStatus := "OFF"
	if state != nil {
		if state.Random {
			shuffleStatus = "ON"
		}
		if state.Repeat {
			repeatStatus = "ON"
		}
	}

	entries = append(entries,
		Entry{
			Label:      fmt.Sprintf("🔀 Shuffle: %s", shuffleStatus),
			Sub:        "Toggle shuffle mode",
			Searchable: true,
			Piped:      "action:shuffle",
		},
		Entry{
			Label:      fmt.Sprintf("🔁 Repeat: %s", repeatStatus),
			Sub:        "Toggle repeat mode",
			Searchable: true,
			Piped:      "action:repeat",
		},
	)

	response := Response{Entries: entries}
	return json.NewEncoder(os.Stdout).Encode(response)
}

// HandleAction handles an action from a selected entry
func (p *Provider) HandleAction(action string) error {
	log.Printf("Handling action: %s", action)

	// Parse action type
	if len(action) > 7 && action[:7] == "action:" {
		cmd := action[7:]
		return p.executeAction(cmd)
	}

	return fmt.Errorf("unknown action: %s", action)
}

func (p *Provider) executeAction(cmd string) error {
	switch cmd {
	case "toggle":
		return p.client.TogglePlayPause()
	case "play":
		return p.client.Play()
	case "pause":
		return p.client.Pause()
	case "stop":
		return p.client.Stop()
	case "next":
		return p.client.Next()
	case "prev":
		return p.client.Previous()
	case "volup":
		return p.client.VolumeUp(10)
	case "voldown":
		return p.client.VolumeDown(10)
	case "mute":
		return p.client.ToggleMute()
	case "shuffle":
		return p.client.ToggleRandom()
	case "repeat":
		return p.client.ToggleRepeat()
	default:
		return fmt.Errorf("unknown command: %s", cmd)
	}
}
