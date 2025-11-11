# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

#### Configuration System
- **YAML configuration file support** at `~/.config/volu/config.yaml`
  - Optional config file for persistent settings
  - Host configuration (alternative to environment variable)
  - Radio series definitions
  - Graceful fallback to defaults if config doesn't exist
  - Config priority: CLI flag → Config file → Environment variable → Default
- `internal/config` package for configuration management
- `config.example.yaml` with comprehensive examples and documentation

#### Radio Series Feature
- **`volu radio <series> <count>` command** for playing random episodes from configured radio series
  - Searches music library for albums matching configured patterns
  - Randomly selects N albums and queues them
  - Automatically disables shuffle to maintain album track order
  - Desktop notifications for search progress and success
- Search API integration (`/api/v1/search`)
  - `Search()` method in volumio client for raw search
  - `SearchAlbums()` method with client-side filtering for albums only
  - Filters out TIDAL results and tracks to return only local albums
- `internal/radio` package implementing radio player logic
  - Album search with regex pattern matching
  - Fisher-Yates shuffle for random selection
  - Smart queuing (ReplaceAndPlay + AddToQueue)
  - Timing delays to prevent API race conditions

#### Testing
- Unit tests for config package (YAML marshaling/unmarshaling)
- All tests passing with new functionality
- Tested against live Volumio instance with 183 ASOT albums

#### Documentation
- Updated README.md with configuration section and radio feature examples
- Updated CLAUDE.md with architecture details for new features
- Documented Search API response structure and filtering strategy
- Added regex pattern examples for common use cases

### Changed
- Host resolution now includes config file in priority chain
- `cmd/volu/main.go` now loads config on startup
- Updated feature list to include radio series and configuration

### Technical Details

**Search API Implementation:**
- Handles multiple result types (albums, tracks, artists, TIDAL)
- Client-side filtering: `type == "folder" && service == "mpd"`
- Real-world tested: "ASOT" search returns 184 albums + 9,519 tracks

**Timing Strategy:**
- 300ms delay after toggling shuffle
- 500ms delay after ReplaceAndPlay
- 100ms delays between AddToQueue operations

**Dependencies:**
- Added `gopkg.in/yaml.v3` for YAML configuration support

## [1.0.0] - 2025-01-11

### Added
- Initial Go port of volumio-linux-control
- Complete CLI interface for Volumio control
- Waybar integration for status bar display
- Walker plugin for launcher integration
- Elephant provider support
- TDD approach with comprehensive test suite
- Zero runtime dependencies (single binary)
- 75x faster startup vs Python version
- Cross-compilation support for Linux (AMD64, ARM64, ARM)

### Features
- Playback control (play, pause, stop, next, prev, toggle)
- Volume control (up, down, set level, mute)
- Playback modes (shuffle, repeat)
- Status display with current track info
- Volumio REST API client with 10s timeout
- Complex browse API handling (4 different response formats)
- Desktop notifications via notify-send

[Unreleased]: https://github.com/riclib/volu/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/riclib/volu/releases/tag/v1.0.0
