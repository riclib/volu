# Quick Start Guide

Get up and running with volu in 5 minutes!

## Installation

```bash
# Clone and build
git clone https://github.com/riclib/volu
cd volu
make build

# Install (choose one)
make install          # Install to /usr/local/bin (requires sudo)
make install-user     # Install to ~/.local/bin (no sudo)
```

## Test Connection

```bash
# Check if you can connect to Volumio
volu status

# If volumio.local doesn't work, try IP address
volu -H 192.168.1.100 status
```

## Basic Usage

```bash
# Playback control
volu play
volu pause
volu toggle        # Most useful - toggle play/pause
volu next
volu stop

# Volume control
volu volume up
volu volume down
volu volume 75     # Set to 75%

# Check what's playing
volu status
```

## Waybar Setup (2 minutes)

1. **Add to waybar config** (`~/.config/waybar/config.jsonc`):

```jsonc
{
    "modules-center": ["custom/volumio"],

    "custom/volumio": {
        "exec": "volu waybar",
        "return-type": "json",
        "interval": 2,
        "on-click": "volu toggle",
        "tooltip": true
    }
}
```

2. **Add styling** (`~/.config/waybar/style.css`):

```css
#custom-volumio {
  min-width: 12px;
  margin: 0 7.5px;
}
```

3. **Restart Waybar**:

```bash
killall waybar && waybar &
```

## Hyprland Keybindings (1 minute)

Add to `~/.config/hypr/hyprland.conf`:

```conf
# Media controls
bind = SUPER, F9, exec, volu toggle
bind = SUPER, F10, exec, volu next
bind = SUPER, F8, exec, volu prev

# Volume
bind = SUPER SHIFT, F11, exec, volu volume up
bind = SUPER SHIFT, F10, exec, volu volume down
```

Reload config:
```bash
hyprctl reload
```

## Troubleshooting

**Can't connect?**

```bash
# Check if Volumio is reachable
ping volumio.local

# Try curl
curl http://volumio.local:3000/api/v1/getState

# Use IP address instead
export VOLUMIO_HOST="192.168.1.100"
volu status
```

**Waybar not showing?**

```bash
# Test directly
volu waybar

# Check waybar logs
killall waybar
waybar 2>&1 | grep volumio
```

## Next Steps

- Read the full [README.md](README.md) for all features
- Customize keybindings
- Try the radio feature: `volu radio asot 5`

Enjoy! 🎵
