package walker

import (
	"encoding/json"
	"fmt"

	"github.com/riclib/volu/internal/volumio"
)

// Item represents a Walker plugin item
type Item struct {
	Label  string `json:"label"`
	Sub    string `json:"sub"`
	Icon   string `json:"icon"`
	Search string `json:"search"`
	Action string `json:"action"`
}

// CreateItem creates a Walker item
func CreateItem(label, sub, icon, action string, searchable bool) Item {
	search := ""
	if searchable {
		search = label
	}
	return Item{
		Label:  label,
		Sub:    sub,
		Icon:   icon,
		Search: search,
		Action: action,
	}
}

// GetIconForItem returns an appropriate icon for a browse item
func GetIconForItem(item *volumio.BrowseItem) string {
	switch item.Type {
	case "song", "track":
		return "🎵"
	case "album":
		return "💿"
	case "artist":
		return "👤"
	case "playlist":
		return "📋"
	case "webradio":
		return "📻"
	case "folder", "category":
		return "📁"
	default:
		return "🎶"
	}
}

// CreateMainMenu creates the main menu with quick controls
func CreateMainMenu(state *volumio.PlayerState) []Item {
	items := []Item{}

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

		nowPlaying := fmt.Sprintf("%s - %s", artist, state.Title)
		album := state.Album
		if album == "" {
			album = "Unknown"
		}

		items = append(items, CreateItem(
			fmt.Sprintf("%s Now Playing: %s", statusIcon, nowPlaying),
			fmt.Sprintf("Album: %s | %s", album, state.Status),
			"🎵",
			"action:toggle",
			false,
		))

		items = append(items, CreateItem(
			"────────────────────────────────────────",
			"",
			"",
			"",
			false,
		))
	}

	// Playback controls
	items = append(items,
		CreateItem("▶️  Play / Pause", "Toggle playback", "▶️", "action:toggle", true),
		CreateItem("⏭️  Next Track", "Skip to next track", "⏭️", "action:next", true),
		CreateItem("⏮️  Previous Track", "Go to previous track", "⏮️", "action:prev", true),
		CreateItem("⏹️  Stop", "Stop playback", "⏹️", "action:stop", true),
	)

	items = append(items, CreateItem(
		"────────────────────────────────────────",
		"",
		"",
		"",
		false,
	))

	// Volume controls
	if state != nil {
		volumeIcon := "🔊"
		if state.Mute {
			volumeIcon = "🔇"
		}
		muteText := ""
		if state.Mute {
			muteText = " (Muted)"
		}

		items = append(items,
			CreateItem(
				fmt.Sprintf("%s Volume: %d%%", volumeIcon, state.Volume),
				fmt.Sprintf("Current volume level%s", muteText),
				volumeIcon,
				"",
				false,
			),
			CreateItem("🔊 Volume Up (+10%)", "Increase volume", "🔊", "action:volup", true),
			CreateItem("🔉 Volume Down (-10%)", "Decrease volume", "🔉", "action:voldown", true),
			CreateItem("🔇 Toggle Mute", "Mute/unmute audio", "🔇", "action:mute", true),
		)
	}

	items = append(items, CreateItem(
		"────────────────────────────────────────",
		"",
		"",
		"",
		false,
	))

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

	items = append(items,
		CreateItem(
			fmt.Sprintf("🔀 Shuffle: %s", shuffleStatus),
			"Toggle shuffle mode",
			"🔀",
			"action:shuffle",
			true,
		),
		CreateItem(
			fmt.Sprintf("🔁 Repeat: %s", repeatStatus),
			"Toggle repeat mode",
			"🔁",
			"action:repeat",
			true,
		),
	)

	items = append(items, CreateItem(
		"────────────────────────────────────────",
		"",
		"",
		"",
		false,
	))

	// Browse sections
	items = append(items,
		CreateItem("📁 Browse Music Library", "Navigate your music collection", "📁", "browse:", true),
		CreateItem("📋 Browse Playlists", "View and play playlists", "📋", "browse:playlists", true),
		CreateItem("👤 Browse Artists", "Browse by artist", "👤", "browse:artists", true),
		CreateItem("💿 Browse Albums", "Browse by album", "💿", "browse:albums", true),
	)

	return items
}

// CreateBrowseMenu creates a browse menu for a given URI
func CreateBrowseMenu(items []volumio.BrowseItem, uri string) []Item {
	walkerItems := []Item{}

	// Add back button if not at root
	if uri != "" {
		walkerItems = append(walkerItems, CreateItem(
			"⬅️  Back",
			"Go back to previous level",
			"⬅️",
			"nav:back",
			true,
		))
	}

	if len(items) == 0 {
		walkerItems = append(walkerItems, CreateItem(
			"No items found",
			"This folder is empty",
			"❌",
			"",
			false,
		))
		return walkerItems
	}

	// Convert browse items to Walker items
	for _, item := range items {
		icon := GetIconForItem(&item)
		subtitle := item.Type
		if subtitle == "" {
			subtitle = "Item"
		}

		if item.Artist != "" {
			subtitle = item.Artist
			if item.Album != "" {
				subtitle += " • " + item.Album
			}
		} else if item.Album != "" {
			subtitle = item.Album
		}

		action := ""
		if item.IsBrowsable() {
			action = fmt.Sprintf("browse:%s", item.URI)
		} else if item.IsPlayable() {
			action = fmt.Sprintf("play:%s|%s", item.URI, item.Service)
		} else {
			action = fmt.Sprintf("browse:%s", item.URI)
		}

		walkerItems = append(walkerItems, CreateItem(
			item.DisplayName(),
			subtitle,
			icon,
			action,
			true,
		))
	}

	return walkerItems
}

// PrintItems outputs items as JSON, one per line
func PrintItems(items []Item) error {
	for _, item := range items {
		data, err := json.Marshal(item)
		if err != nil {
			return fmt.Errorf("failed to marshal item: %w", err)
		}
		fmt.Println(string(data))
	}
	return nil
}
