# volu

A modern Go-based CLI tool for controlling [Volumio](https://volumio.com) music players from the command line, with integrated support for Waybar, Walker, and [Elephant](https://github.com/abenz1267/elephant).

## Features

- **CLI Control**: Simple command-line interface for all playback operations
- **Radio Series**: Play random episodes from your favorite radio shows (ASOT, Group Therapy, etc.)
- **YAML Configuration**: Optional config file for host and radio series settings
- **Waybar Integration**: Real-time status display in your status bar
- **Walker Plugin**: Browse and control music through Walker launcher
- **Elephant Provider**: Native integration with the Elephant launcher (coming soon)
- **Single Binary**: No runtime dependencies, just compile and run
- **TDD Approach**: Well-tested codebase with unit and integration tests

## Installation

### From Source

```bash
git clone https://github.com/riclib/volu
cd volu
go build -o volu ./cmd/volu
sudo mv volu /usr/local/bin/
```

### Configuration

#### Config File (Recommended)

Create `~/.config/volu/config.yaml`:

```yaml
# Volumio host configuration
host: volumio.local  # or IP address like 192.168.1.100

# Radio series for 'volu radio' command (optional)
radio:
  asot:
    name: "A State of Trance"
    search_query: "ASOT"
    pattern: "^ASOT\\s+\\d+"

  grouptherapy:
    name: "Group Therapy"
    search_query: "Group Therapy"
    pattern: "Group Therapy\\s+\\d+"
```

See `config.example.yaml` for more examples and regex pattern tips.

#### Environment Variable

Alternatively, set your Volumio host via environment variable:

```bash
export VOLUMIO_HOST="volumio.local"
# Or use an IP address
export VOLUMIO_HOST="192.168.1.100"
```

Add to your `~/.bashrc` or `~/.zshrc` to make it permanent.

**Priority:** CLI flag `--host` → Config file → `VOLUMIO_HOST` env → default `volumio.local`

## Usage

### Basic Commands

```bash
# Playback control
volu play
volu pause
volu toggle          # Toggle play/pause
volu stop
volu next           # or: volu skip
volu prev

# Volume control
volu volume up
volu volume down
volu volume 50      # Set to specific level (0-100)

# Playback modes
volu shuffle        # Toggle shuffle
volu repeat         # Toggle repeat

# Status
volu status         # Show current playback info
```

### Radio Series

Play random episodes from your favorite radio shows (requires configuration):

```bash
# Play 3 random A State of Trance episodes
volu radio asot 3

# Play 5 random Group Therapy episodes
volu radio grouptherapy 5

# Play 1 random episode from any configured series
volu radio tritonia 1
```

**How it works:**
1. Searches your Volumio library for albums matching the series pattern
2. Randomly selects N albums
3. Queues them for playback (in order, shuffle automatically disabled)
4. Each album plays in full before moving to the next

**Configuration example:**
```yaml
radio:
  asot:
    name: "A State of Trance"
    search_query: "ASOT"           # What to search for
    pattern: "^ASOT\\s+\\d+"       # Regex to match album names
```

The `pattern` field uses regular expressions to filter albums. Common patterns:
- `^ASOT\\s+\\d+` - Matches "ASOT 1090", "ASOT 600", etc.
- `Buddha Bar\\s+\\d+` - Matches "Buddha Bar 1", "Buddha Bar 25", etc.
- `Episode\\s+\\d+` - Matches "Episode 123", etc.

### Host Override

```bash
volu -H 192.168.1.100 status
volu --host volumio2.local play
```

## Waybar Integration

### Waybar Configuration

Add to your `~/.config/waybar/config`:

```jsonc
{
    "modules-center": ["custom/volumio"],

    "custom/volumio": {
        "exec": "volu waybar",
        "return-type": "json",
        "interval": 2,
        "format": "{}",
        "on-click": "volu toggle",
        "tooltip": true,
        "max-length": 50
    }
}
```

### Waybar Styling

Add to your `~/.config/waybar/style.css`:

```css
/* Volumio module base styling */
#custom-volumio {
    padding: 0 10px;
    color: #ffffff;
    background-color: #1a1a1a;
    border-radius: 5px;
    margin: 5px;
}

/* Playing state - green */
#custom-volumio.volumio-play {
    background-color: #2d5016;
    color: #a6e3a1;
}

/* Paused state - yellow */
#custom-volumio.volumio-pause {
    background-color: #5a4a1a;
    color: #f9e2af;
}

/* Stopped state - gray */
#custom-volumio.volumio-stop {
    background-color: #2a2a2a;
    color: #9399b2;
}

/* Error state - red */
#custom-volumio.volumio-error {
    background-color: #4a1a1a;
    color: #f38ba8;
}
```

Restart Waybar:

```bash
killall waybar && waybar &
```

## Walker Integration

### Walker Configuration

Add to your `~/.config/walker/config.toml`:

```toml
[[plugins]]
prefix = "vm"  # Type 'vm' in Walker to access Volumio controls
name = "volumio"
cmd = "volu walker"
```

Or without a prefix (always available):

```toml
[[plugins]]
name = "volumio"
cmd = "volu walker"
```

### Usage

1. Open Walker (usually `Super+Space`)
2. Type `vm` (if using prefix) or start typing to search
3. Navigate with arrow keys
4. Press Enter to execute actions

Features:
- **Quick Controls**: Play/pause, next, previous, stop
- **Volume Control**: Up, down, mute
- **Playback Modes**: Shuffle and repeat toggles
- **Browse Library**: Navigate music, playlists, artists, albums
- **Now Playing**: See current track info

## Hyprland Integration

Add keybindings to `~/.config/hypr/hyprland.conf`:

```conf
# Volumio media controls
bind = SUPER, F9, exec, volu toggle
bind = SUPER, F10, exec, volu next
bind = SUPER, F8, exec, volu prev
bind = SUPER, F7, exec, volu stop

# Volume controls
bind = SUPER SHIFT, F11, exec, volu volume up
bind = SUPER SHIFT, F10, exec, volu volume down

# Playback modes
bind = SUPER SHIFT, S, exec, volu shuffle

# Open Walker with Volumio
bind = SUPER, M, exec, walker --modules volumio
```

Or use standard media keys:

```conf
bind = , XF86AudioPlay, exec, volu toggle
bind = , XF86AudioNext, exec, volu next
bind = , XF86AudioPrev, exec, volu prev
bind = , XF86AudioStop, exec, volu stop
```

## Elephant Integration

**Coming Soon!** Native provider for the [Elephant](https://github.com/abenz1267/elephant) launcher.

```bash
volu elephant
```

This will start a background provider that Elephant can use to browse and control Volumio.

## Development

### Project Structure

```
volu/
├── cmd/
│   └── volu/          # Main CLI application
│       └── main.go
├── internal/
│   ├── volumio/       # Volumio REST API client
│   │   ├── client.go
│   │   └── client_test.go
│   ├── waybar/        # Waybar JSON output
│   │   └── waybar.go
│   ├── walker/        # Walker plugin interface
│   │   └── walker.go
│   └── elephant/      # Elephant provider (WIP)
│       └── provider.go
├── go.mod
├── go.sum
└── README.md
```

### Running Tests

```bash
# Run unit tests
go test ./...

# Run unit tests with verbose output
go test -v ./...

# Run integration tests (requires Volumio instance)
go test -v ./internal/volumio/...

# Run specific test
go test -v -run TestGetState ./internal/volumio/
```

### Building

```bash
# Development build
go build -o volu ./cmd/volu

# Optimized release build
go build -ldflags="-s -w" -o volu ./cmd/volu

# Cross-compile for different platforms
GOOS=linux GOARCH=amd64 go build -o volu-linux-amd64 ./cmd/volu
GOOS=linux GOARCH=arm64 go build -o volu-linux-arm64 ./cmd/volu
```

### API Client

The `internal/volumio` package provides a clean Go interface to the Volumio REST API:

```go
package main

import (
    "fmt"
    "github.com/riclib/volu/internal/volumio"
)

func main() {
    // Create client
    client := volumio.NewClientWithHost("volumio.local")

    // Get current state
    state, err := client.GetState()
    if err != nil {
        panic(err)
    }

    fmt.Printf("Now playing: %s - %s\n", state.Artist, state.Title)
    fmt.Printf("Status: %s\n", state.Status)
    fmt.Printf("Volume: %d%%\n", state.Volume)

    // Control playback
    client.TogglePlayPause()
    client.Next()
    client.SetVolume(50)

    // Browse library
    items, err := client.Browse("")
    if err != nil {
        panic(err)
    }

    for _, item := range items {
        fmt.Printf("%s: %s\n", item.Type, item.DisplayName())
    }
}
```

## Troubleshooting

### Cannot connect to Volumio

1. **Check Volumio is running:**
   ```bash
   ping volumio.local
   # Or your custom host
   ping 192.168.1.100
   ```

2. **Test API directly:**
   ```bash
   curl http://volumio.local:3000/api/v1/getState
   ```

3. **Check firewall:**
   Ensure port 3000 is accessible

4. **Override host:**
   ```bash
   volu -H 192.168.1.100 status
   ```

### Waybar module not showing

1. **Check Waybar config:**
   ```bash
   cat ~/.config/waybar/config | grep -A 10 volumio
   ```

2. **Test command directly:**
   ```bash
   volu waybar
   ```

3. **Check Waybar logs:**
   ```bash
   killall waybar
   waybar 2>&1 | grep volumio
   ```

### Walker plugin not working

1. **Check Walker config:**
   ```bash
   cat ~/.config/walker/config.toml | grep -A 5 volumio
   ```

2. **Test plugin directly:**
   ```bash
   volu walker
   ```

3. **Test with action:**
   ```bash
   volu walker "action:toggle"
   ```

## Comparison with Python Version

This Go implementation replaces the previous Python version with several advantages:

| Feature | Python | Go |
|---------|--------|-----|
| **Installation** | Requires Python + dependencies | Single binary, no dependencies |
| **Performance** | Slower (interpreted) | Fast (compiled) |
| **Memory** | Higher (~30MB+ with Python runtime) | Lower (~10MB single binary) |
| **Startup Time** | ~100-200ms | ~1-5ms |
| **Deployment** | Manage virtualenv, pip, etc. | Copy binary, done |
| **Testing** | Manual testing | TDD with unit + integration tests |
| **Elephant Support** | Not possible | Native Go provider |

## License

MIT License - Feel free to use, modify, and distribute.

## Contributing

Contributions welcome! This project follows TDD principles - please include tests with your PRs.

### TODO

- [ ] Complete Elephant provider implementation
- [ ] Add album art support to Waybar
- [ ] Add queue management commands
- [ ] Add playlist management
- [ ] Add search functionality
- [ ] Add bash/zsh completion scripts
- [ ] Create Arch Linux AUR package
- [ ] Add systemd service for elephant provider
- [ ] Add more comprehensive integration tests
- [ ] Add benchmarks

## Credits

Ported from the Python-based [volumio-linux-control](../README.md) project.

## Links

- [Volumio](https://volumio.com) - The Volumio music player
- [Walker](https://github.com/abenz1267/walker) - Application launcher for Linux
- [Elephant](https://github.com/abenz1267/elephant) - Provider-based launcher
- [Waybar](https://github.com/Alexays/Waybar) - Highly customizable status bar
- [Hyprland](https://hyprland.org/) - Dynamic tiling Wayland compositor
