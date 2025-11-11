package radio

import (
	"fmt"
	"math/rand"
	"regexp"
	"time"

	"github.com/riclib/volu/internal/volumio"
)

// Player handles radio series playback functionality.
type Player struct {
	client *volumio.Client
}

// NewPlayer creates a new radio player.
func NewPlayer(client *volumio.Client) *Player {
	return &Player{
		client: client,
	}
}

// PlayRandomEpisodes searches for albums matching the pattern, randomly selects count albums,
// and queues them for playback. Shuffle is automatically disabled.
func (p *Player) PlayRandomEpisodes(searchQuery, pattern string, count int) error {
	if count < 1 {
		return fmt.Errorf("count must be at least 1")
	}

	// 1. Find all albums matching the pattern
	albums, err := p.findMatchingAlbums(searchQuery, pattern)
	if err != nil {
		return fmt.Errorf("failed to find albums: %w", err)
	}

	if len(albums) == 0 {
		return fmt.Errorf("no albums found matching pattern: %s", pattern)
	}

	// 2. Randomly select N albums
	selected := p.randomSelect(albums, count)

	// 3. Ensure shuffle is off
	if err := p.ensureShuffleOff(); err != nil {
		return fmt.Errorf("failed to disable shuffle: %w", err)
	}

	// 4. Queue the selected albums
	if err := p.queueAlbums(selected); err != nil {
		return fmt.Errorf("failed to queue albums: %w", err)
	}

	return nil
}

// findMatchingAlbums searches for albums using the search API and filters by regex pattern.
func (p *Player) findMatchingAlbums(searchQuery, pattern string) ([]volumio.BrowseItem, error) {
	// Compile the regex pattern
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %w", err)
	}

	// Search for albums using the Volumio API
	albums, err := p.client.SearchAlbums(searchQuery)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	// Filter albums by regex pattern
	var matching []volumio.BrowseItem
	for _, album := range albums {
		displayName := album.DisplayName()
		if regex.MatchString(displayName) {
			matching = append(matching, album)
		}
	}

	return matching, nil
}

// randomSelect randomly selects up to count items from the albums slice.
// If count >= len(albums), returns all albums in random order.
func (p *Player) randomSelect(albums []volumio.BrowseItem, count int) []volumio.BrowseItem {
	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	if count >= len(albums) {
		// Return all albums in random order
		return shuffle(albums)
	}

	// Random selection without replacement using Fisher-Yates shuffle
	indices := make([]int, len(albums))
	for i := range indices {
		indices[i] = i
	}

	// Shuffle indices
	for i := len(indices) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		indices[i], indices[j] = indices[j], indices[i]
	}

	// Select first count indices
	selected := make([]volumio.BrowseItem, count)
	for i := 0; i < count; i++ {
		selected[i] = albums[indices[i]]
	}

	return selected
}

// shuffle returns a shuffled copy of the albums slice.
func shuffle(albums []volumio.BrowseItem) []volumio.BrowseItem {
	shuffled := make([]volumio.BrowseItem, len(albums))
	copy(shuffled, albums)

	for i := len(shuffled) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}

	return shuffled
}

// ensureShuffleOff checks if shuffle is enabled and disables it if necessary.
func (p *Player) ensureShuffleOff() error {
	state, err := p.client.GetState()
	if err != nil {
		return err
	}

	if state.Random {
		if err := p.client.ToggleRandom(); err != nil {
			return err
		}
		// Small delay to let the state update
		time.Sleep(300 * time.Millisecond)
	}

	return nil
}

// queueAlbums queues the selected albums for playback.
// The first album replaces the queue and starts playing.
// Subsequent albums are added to the queue.
func (p *Player) queueAlbums(albums []volumio.BrowseItem) error {
	if len(albums) == 0 {
		return fmt.Errorf("no albums to queue")
	}

	// First album: replace queue and play
	first := albums[0]
	if err := p.client.ReplaceAndPlay(first.URI, first.Service); err != nil {
		return fmt.Errorf("failed to play first album %q: %w", first.DisplayName(), err)
	}

	// Wait for the first album to start loading
	time.Sleep(500 * time.Millisecond)

	// Remaining albums: add to queue
	for i := 1; i < len(albums); i++ {
		album := albums[i]
		if err := p.client.AddToQueue(album.URI, album.Service); err != nil {
			return fmt.Errorf("failed to add album %q to queue: %w", album.DisplayName(), err)
		}
		// Small delay to avoid overwhelming the API
		time.Sleep(100 * time.Millisecond)
	}

	return nil
}
