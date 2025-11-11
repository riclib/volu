package volumio

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// PlayerState represents the current state of the Volumio player
type PlayerState struct {
	Status   string `json:"status"`
	Position int    `json:"position"`
	Title    string `json:"title"`
	Artist   string `json:"artist"`
	Album    string `json:"album"`
	AlbumArt string `json:"albumart"`
	Duration int    `json:"duration"`
	Volume   int    `json:"volume"`
	Mute     bool   `json:"mute"`
	Service  string `json:"service"`
	Random   bool   `json:"random"`
	Repeat   bool   `json:"repeat"`
}

// BrowseItem represents an item in the Volumio music library
type BrowseItem struct {
	URI        string `json:"uri"`
	Title      string `json:"title"`
	Name       string `json:"name"`
	Service    string `json:"service"`
	Type       string `json:"type"`
	Artist     string `json:"artist"`
	Album      string `json:"album"`
	AlbumArt   string `json:"albumart"`
	PluginType string `json:"plugin_type"`
	PluginName string `json:"plugin_name"`
}

// DisplayName returns the best display name for a browse item
func (b *BrowseItem) DisplayName() string {
	if b.Title != "" {
		return b.Title
	}
	if b.Name != "" {
		return b.Name
	}
	return b.URI
}

// IsPlayable returns true if the item can be played
func (b *BrowseItem) IsPlayable() bool {
	return b.Type == "song" || b.Type == "track" || b.Type == "webradio" || b.Type == "folder"
}

// IsBrowsable returns true if the item can be browsed
func (b *BrowseItem) IsBrowsable() bool {
	return b.Type == "folder" || b.Type == "playlist-category" || b.Type == "category" || b.Type == ""
}

// Client is a Volumio REST API client
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new Volumio API client
func NewClient(baseURL string) *Client {
	if baseURL == "" {
		baseURL = "http://volumio.local:3000"
	}
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// NewClientWithHost creates a client for a specific host
func NewClientWithHost(host string) *Client {
	return NewClient(fmt.Sprintf("http://%s:3000", host))
}

func (c *Client) get(endpoint string, params url.Values) ([]byte, error) {
	u := c.baseURL + endpoint
	if params != nil {
		u += "?" + params.Encode()
	}

	resp, err := c.httpClient.Get(u)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return body, nil
}

func (c *Client) post(endpoint string, data interface{}) ([]byte, error) {
	// Volumio POST endpoints accept JSON
	// Implementation will be added when needed
	return nil, fmt.Errorf("not implemented")
}

// GetState retrieves the current player state
func (c *Client) GetState() (*PlayerState, error) {
	body, err := c.get("/api/v1/getState", nil)
	if err != nil {
		return nil, err
	}

	var state PlayerState
	if err := json.Unmarshal(body, &state); err != nil {
		return nil, fmt.Errorf("failed to parse state: %w", err)
	}

	return &state, nil
}

// GetAlbumArtURL returns the full URL for album artwork
func (c *Client) GetAlbumArtURL(albumart string) string {
	if albumart == "" {
		return ""
	}
	// Handle absolute URLs
	if len(albumart) > 7 && (albumart[:7] == "http://" || albumart[:8] == "https://") {
		return albumart
	}
	// Handle relative paths
	return c.baseURL + albumart
}

// Playback control commands

// Play starts playback
func (c *Client) Play() error {
	params := url.Values{}
	params.Set("cmd", "play")
	_, err := c.get("/api/v1/commands/", params)
	return err
}

// Pause pauses playback
func (c *Client) Pause() error {
	params := url.Values{}
	params.Set("cmd", "pause")
	_, err := c.get("/api/v1/commands/", params)
	return err
}

// TogglePlayPause toggles between play and pause
func (c *Client) TogglePlayPause() error {
	params := url.Values{}
	params.Set("cmd", "toggle")
	_, err := c.get("/api/v1/commands/", params)
	return err
}

// Stop stops playback
func (c *Client) Stop() error {
	params := url.Values{}
	params.Set("cmd", "stop")
	_, err := c.get("/api/v1/commands/", params)
	return err
}

// Next skips to the next track
func (c *Client) Next() error {
	params := url.Values{}
	params.Set("cmd", "next")
	_, err := c.get("/api/v1/commands/", params)
	return err
}

// Previous goes to the previous track
func (c *Client) Previous() error {
	params := url.Values{}
	params.Set("cmd", "prev")
	_, err := c.get("/api/v1/commands/", params)
	return err
}

// Volume control

// SetVolume sets the volume level (0-100)
func (c *Client) SetVolume(volume int) error {
	// Clamp to valid range
	if volume < 0 {
		volume = 0
	}
	if volume > 100 {
		volume = 100
	}

	params := url.Values{}
	params.Set("cmd", "volume")
	params.Set("volume", fmt.Sprintf("%d", volume))
	_, err := c.get("/api/v1/commands/", params)
	return err
}

// VolumeUp increases volume by the specified step
func (c *Client) VolumeUp(step int) error {
	state, err := c.GetState()
	if err != nil {
		return err
	}
	return c.SetVolume(state.Volume + step)
}

// VolumeDown decreases volume by the specified step
func (c *Client) VolumeDown(step int) error {
	state, err := c.GetState()
	if err != nil {
		return err
	}
	return c.SetVolume(state.Volume - step)
}

// Mute mutes audio
func (c *Client) Mute() error {
	params := url.Values{}
	params.Set("cmd", "mute")
	_, err := c.get("/api/v1/commands/", params)
	return err
}

// Unmute unmutes audio
func (c *Client) Unmute() error {
	params := url.Values{}
	params.Set("cmd", "unmute")
	_, err := c.get("/api/v1/commands/", params)
	return err
}

// ToggleMute toggles mute state
func (c *Client) ToggleMute() error {
	state, err := c.GetState()
	if err != nil {
		return err
	}

	if state.Mute {
		return c.Unmute()
	}
	return c.Mute()
}

// Playback mode control

// ToggleRandom toggles shuffle/random mode
func (c *Client) ToggleRandom() error {
	params := url.Values{}
	params.Set("cmd", "random")
	_, err := c.get("/api/v1/commands/", params)
	return err
}

// ToggleRepeat toggles repeat mode
func (c *Client) ToggleRepeat() error {
	params := url.Values{}
	params.Set("cmd", "repeat")
	_, err := c.get("/api/v1/commands/", params)
	return err
}

// Queue management

// ClearQueue clears the playback queue
func (c *Client) ClearQueue() error {
	params := url.Values{}
	params.Set("cmd", "clearQueue")
	_, err := c.get("/api/v1/commands/", params)
	return err
}

// Browse and playback

// BrowseResponse represents the complex nested response from the browse API
type BrowseResponse struct {
	Navigation *struct {
		Lists []json.RawMessage `json:"lists"`
	} `json:"navigation"`
	Lists []json.RawMessage `json:"lists"`
	List  []BrowseItem      `json:"list"`
	Items []BrowseItem      `json:"items"`
}

// Browse browses the music library at the given URI
func (c *Client) Browse(uri string) ([]BrowseItem, error) {
	params := url.Values{}
	if uri != "" {
		params.Set("uri", uri)
	}

	body, err := c.get("/api/v1/browse", params)
	if err != nil {
		return nil, err
	}

	// Parse the complex nested response structure
	var response BrowseResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse browse response: %w", err)
	}

	// Try different extraction paths based on Volumio's response structure
	var items []BrowseItem

	// Path 1: navigation.lists[0].items
	if response.Navigation != nil && len(response.Navigation.Lists) > 0 {
		var firstList struct {
			Items []BrowseItem `json:"items"`
		}
		if err := json.Unmarshal(response.Navigation.Lists[0], &firstList); err == nil && len(firstList.Items) > 0 {
			items = firstList.Items
		} else {
			// Try parsing as direct list of items
			var directItems []BrowseItem
			if err := json.Unmarshal(response.Navigation.Lists[0], &directItems); err == nil {
				items = directItems
			}
		}
	}

	// Path 2: lists[0].items (fallback)
	if len(items) == 0 && len(response.Lists) > 0 {
		var firstList struct {
			Items []BrowseItem `json:"items"`
		}
		if err := json.Unmarshal(response.Lists[0], &firstList); err == nil && len(firstList.Items) > 0 {
			items = firstList.Items
		} else {
			// Try parsing as direct list
			var directItems []BrowseItem
			if err := json.Unmarshal(response.Lists[0], &directItems); err == nil {
				items = directItems
			}
		}
	}

	// Path 3: list (direct)
	if len(items) == 0 && len(response.List) > 0 {
		items = response.List
	}

	// Path 4: items (direct)
	if len(items) == 0 && len(response.Items) > 0 {
		items = response.Items
	}

	return items, nil
}

// SearchList represents a list in the search response
type SearchList struct {
	Title string       `json:"title"`
	Items []BrowseItem `json:"items"`
}

// SearchResponse represents the response from the search API
type SearchResponse struct {
	Navigation struct {
		IsSearchResult bool         `json:"isSearchResult"`
		Lists          []SearchList `json:"lists"`
	} `json:"navigation"`
}

// Search searches the music library for the given query.
// Returns the raw search response with all result types (albums, tracks, artists, etc.)
func (c *Client) Search(query string) (*SearchResponse, error) {
	params := url.Values{}
	params.Set("query", query)

	body, err := c.get("/api/v1/search", params)
	if err != nil {
		return nil, err
	}

	var response SearchResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse search response: %w", err)
	}

	return &response, nil
}

// SearchAlbums searches for albums matching the given query.
// Filters the results to return only local albums (not TIDAL or other services).
func (c *Client) SearchAlbums(query string) ([]BrowseItem, error) {
	response, err := c.Search(query)
	if err != nil {
		return nil, err
	}

	// Find the local albums list
	// Typically the first list with "Albums" in the title and not containing "TIDAL"
	for _, list := range response.Navigation.Lists {
		if containsIgnoreCase(list.Title, "Albums") && !containsIgnoreCase(list.Title, "TIDAL") {
			// Filter to only include folder type items from mpd service
			var albums []BrowseItem
			for _, item := range list.Items {
				if item.Type == "folder" && item.Service == "mpd" {
					albums = append(albums, item)
				}
			}
			return albums, nil
		}
	}

	// No albums found
	return []BrowseItem{}, nil
}

// containsIgnoreCase checks if s contains substr, case-insensitive
func containsIgnoreCase(s, substr string) bool {
	sLower := make([]byte, len(s))
	substrLower := make([]byte, len(substr))
	for i := 0; i < len(s); i++ {
		sLower[i] = toLower(s[i])
	}
	for i := 0; i < len(substr); i++ {
		substrLower[i] = toLower(substr[i])
	}
	return bytesContains(sLower, substrLower)
}

// toLower converts a single ASCII character to lowercase
func toLower(c byte) byte {
	if c >= 'A' && c <= 'Z' {
		return c + ('a' - 'A')
	}
	return c
}

// bytesContains checks if b contains substr
func bytesContains(b, substr []byte) bool {
	if len(substr) == 0 {
		return true
	}
	if len(substr) > len(b) {
		return false
	}
	for i := 0; i <= len(b)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if b[i+j] != substr[j] {
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

// ReplaceAndPlay clears the queue and plays an item
func (c *Client) ReplaceAndPlay(uri, service string) error {
	payload := map[string]string{"uri": uri}
	if service != "" {
		payload["service"] = service
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL+"/api/v1/replaceAndPlay", bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// AddToQueue adds an item to the queue
func (c *Client) AddToQueue(uri, service string) error {
	payload := map[string]string{"uri": uri}
	if service != "" {
		payload["service"] = service
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL+"/api/v1/addToQueue", bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
