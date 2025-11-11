# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**volu** is a high-performance Go CLI tool for controlling Volumio music players. It provides command-line control, real-time status for Waybar, browse/control interface for Walker, and native Elephant launcher integration. Built as a complete rewrite of a Python implementation, delivering 75x faster startup with zero runtime dependencies.

**Technology Stack:**
- Go 1.25.3
- Cobra CLI framework (github.com/spf13/cobra)
- Standard library HTTP client
- YAML config (gopkg.in/yaml.v3)
- Cross-compilation support for Linux (AMD64, ARM64, ARM)

## Essential Commands

### Building
```bash
make build              # Development build → ./volu
make build-release      # Optimized build with stripped symbols
make build-all          # Cross-compile for all platforms
```

### Testing
```bash
make test               # Unit tests (with -short flag)
make test-all           # All tests including integration tests
make test-coverage      # Generate coverage.html report
go test -v ./...        # Verbose test output
go test -v -run TestName ./internal/volumio  # Run specific test
```

### Installation
```bash
make install-user       # Install to ~/.local/bin (recommended)
make install            # System-wide to /usr/local/bin (requires sudo)
```

### Running
```bash
make run                # Build and run with 'status' command
go run ./cmd/volu status
go run ./cmd/volu --host volumio.local play
```

### Code Quality
```bash
make fmt                # Format code
make vet                # Run go vet
make lint               # Run golangci-lint
```

## Architecture

### Component Structure

```
cmd/volu/main.go                 # CLI entry point with all commands
  ├─ Cobra command tree
  ├─ PersistentPreRun: loads config, initializes volumio.Client
  └─ Host resolution: flag → config file → VOLUMIO_HOST env → "volumio.local"

internal/config/config.go        # YAML configuration management
  ├─ Config struct (host + radio series)
  ├─ Load/Save functions
  └─ Default config path: ~/.config/volu/config.yaml

internal/volumio/client.go       # Core API client
  ├─ HTTP client (10s timeout)
  ├─ Base URL: http://{host}:3000/api/v1
  ├─ PlayerState and BrowseItem structs
  ├─ Browse() - 4-path response handling
  ├─ Search() - search API wrapper
  └─ SearchAlbums() - filtered album search

internal/radio/radio.go          # Radio series player
  ├─ FindAlbums() - search + regex filter
  ├─ RandomSelect() - Fisher-Yates shuffle
  └─ QueueAlbums() - ReplaceAndPlay + AddToQueue

internal/waybar/waybar.go        # Waybar JSON output
internal/walker/walker.go        # Walker plugin interface
internal/elephant/provider.go    # Elephant provider protocol
```

### Key Architectural Patterns

**1. Configuration and Host Resolution**

Config file location: `~/.config/volu/config.yaml` (see `config.example.yaml` for template)

Host resolution priority:
```go
// 1. CLI flag --host
// 2. Config file (cfg.Host)
// 3. VOLUMIO_HOST environment variable
// 4. Default "volumio.local"
// Implemented in PersistentPreRun hook (cmd/volu/main.go)
```

Config is loaded on every command execution. If the file doesn't exist or has errors, defaults are used gracefully.

**2. Complex Browse API Response Handling**

The Volumio Browse API returns 4 different JSON structures depending on the endpoint. The client must try all paths sequentially:

```go
// Path 1: response["navigation"]["lists"][0]["items"]
// Path 2: response["lists"][0]["items"]
// Path 3: response["list"]
// Path 4: response["items"]
```

See `client.go:Browse()` for implementation. This is non-obvious and critical for browse functionality.

**3. Timing Patterns for State Synchronization**

State-changing operations require delays before fetching updated state:

```go
// 300ms after: toggle, shuffle, repeat
// 500ms after: next, prev, clear queue
// Then call GetState() for accurate notifications
```

This is implemented across commands in `main.go`. Without these delays, state queries return stale data.

**4. Volume Control Race Prevention**

Volume up/down operations must fetch current state first:

```go
// 1. GetState() to get current volume
// 2. Calculate new level (clamp 0-100)
// 3. SetVolume(newLevel)
// Prevents race conditions from concurrent volume changes
```

**5. Desktop Notifications**

Commands use `notify-send` with specific patterns:
- 2-second duration (`-t 2000`)
- Urgent flag for errors (`-u critical`)
- Icon from theme (`--icon=multimedia-player`)
- Executed after sleep delay to show updated state

### Volumio REST API Endpoints

```
GET  /api/v1/getState              # Current player state
GET  /api/v1/commands/?cmd={...}   # play|pause|toggle|stop|next|prev|volume|random|repeat|clearQueue
GET  /api/v1/browse?uri={uri}      # Browse library (4 response formats!)
GET  /api/v1/search?query={query}  # Search library (returns albums, tracks, artists, etc.)
POST /api/v1/replaceAndPlay        # Clear queue and play item
POST /api/v1/addToQueue            # Add item to queue
```

**Search API Response Structure:**

The search API returns multiple lists grouped by type:
```json
{
  "navigation": {
    "isSearchResult": true,
    "lists": [
      {"title": "Found 184 Albums 'ASOT'", "items": [...]},
      {"title": "Found 9519 Tracks 'ASOT'", "items": [...]},
      {"title": "TIDAL Albums", "items": [...]},
      ...
    ]
  }
}
```

**Client-side filtering is required:**
- `SearchAlbums()` filters to local albums only (not TIDAL)
- Checks: `item.Type == "folder" && item.Service == "mpd"`
- First list with "Albums" (not "TIDAL") contains local albums

### Testing Approach

- **TDD methodology**: Tests written first, driving implementation
- **Mock HTTP servers**: `httptest.NewServer()` for unit tests
- **Integration tests**: Skip with `-short` flag, require real Volumio instance
- **Test file**: `internal/volumio/client_test.go` (153 lines)

### PlayerState Structure

```go
type PlayerState struct {
    Status   string // "play", "pause", "stop"
    Position int    // Current position in seconds
    Title    string
    Artist   string
    Album    string
    AlbumArt string // Relative path or full URL
    Duration int    // Track length in seconds
    Volume   int    // 0-100
    Mute     bool
    Service  string // "mpd", "webradio", etc.
    Random   bool   // Shuffle state
    Repeat   bool
}
```

## Development Guidelines

### Code Organization
- Use `internal/` packages for non-exported code
- Package-level separation: client logic separate from presentation (waybar/walker/elephant)
- Client/presenter pattern throughout

### Error Handling
- Return explicit errors, no panics in library code
- Wrap errors with `fmt.Errorf` and `%w` for context
- Graceful degradation (e.g., show menu even if API fails)

### When Adding New Commands
1. Add command in `cmd/volu/main.go` using Cobra
2. Implement API method in `internal/volumio/client.go` if needed
3. Add appropriate sleep delay before GetState() if state-changing
4. Use notify-send for user feedback
5. Write tests in `client_test.go` using mock HTTP server

### When Modifying Browse Functionality
Remember that browse responses have 4 different formats. Always test against:
- Root navigation
- Music library categories
- Playlist views
- Radio/streaming service listings

### Radio Series Feature

The `volu radio` command plays random episodes from configured series (e.g., ASOT, Group Therapy).

**Configuration:**
```yaml
# ~/.config/volu/config.yaml
radio:
  asot:
    name: "A State of Trance"
    search_query: "ASOT"
    pattern: "^ASOT\\s+\\d+"
```

**Implementation Flow:**
1. Search API with `search_query`
2. Filter albums by regex `pattern`
3. Random selection using Fisher-Yates shuffle
4. Disable shuffle mode
5. ReplaceAndPlay first album, AddToQueue rest
6. 100ms delays between queue operations

**Key Considerations:**
- Search returns both albums and tracks (9,519 tracks for "ASOT")
- Must filter to albums only: `Type == "folder" && Service == "mpd"`
- Regex pattern must use `\\s` in YAML (escaped backslash)
- Album URIs use `albums://Artist/Album` format
- Timing: 500ms after ReplaceAndPlay, 100ms between AddToQueue calls

## Performance Characteristics

- **Binary size**: ~10MB (optimized), ~9.3MB (debug)
- **Startup time**: ~1-5ms (vs 70-100ms for Python version)
- **Memory usage**: ~10-15MB runtime (vs 50-60MB for Python)
- **Build time**: ~2-3 seconds
