package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/riclib/volu/internal/config"
	"github.com/riclib/volu/internal/elephant"
	"github.com/riclib/volu/internal/radio"
	"github.com/riclib/volu/internal/volumio"
	"github.com/riclib/volu/internal/walker"
	"github.com/riclib/volu/internal/waybar"
	"github.com/spf13/cobra"
)

var (
	volumioHost string
	client      *volumio.Client
	cfg         *config.Config
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "volu",
		Short: "Control your Volumio music player",
		Long:  `volu is a CLI tool for controlling Volumio music players from the command line.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Load config file
			var err error
			cfg, err = config.Load()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Could not load config: %v\n", err)
				cfg = config.DefaultConfig()
			}

			// Priority: flag → config file → env var → default
			if volumioHost == "" {
				volumioHost = cfg.Host
				if volumioHost == "" {
					volumioHost = os.Getenv("VOLUMIO_HOST")
					if volumioHost == "" {
						volumioHost = "volumio.local"
					}
				}
			}
			client = volumio.NewClientWithHost(volumioHost)
		},
	}

	rootCmd.PersistentFlags().StringVarP(&volumioHost, "host", "H", "", "Volumio host (default: $VOLUMIO_HOST or volumio.local)")

	// Playback commands
	rootCmd.AddCommand(playCmd)
	rootCmd.AddCommand(pauseCmd)
	rootCmd.AddCommand(toggleCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(nextCmd)
	rootCmd.AddCommand(prevCmd)
	rootCmd.AddCommand(skipCmd) // Alias for next

	// Volume commands
	rootCmd.AddCommand(volumeCmd)

	// Status command
	rootCmd.AddCommand(statusCmd)

	// Playback mode commands
	rootCmd.AddCommand(shuffleCmd)
	rootCmd.AddCommand(repeatCmd)

	// Radio command
	rootCmd.AddCommand(radioCmd)

	// Waybar command
	rootCmd.AddCommand(waybarCmd)

	// Walker command
	rootCmd.AddCommand(walkerCmd)

	// Elephant provider (placeholder)
	rootCmd.AddCommand(elephantCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// Helper function to send notifications
func notify(title, message, icon string, urgent bool) {
	args := []string{"-t", "2000", "-i", icon, title, message}
	if urgent {
		args = append([]string{"-u", "critical"}, args...)
	}
	exec.Command("notify-send", args...).Run()
}

// Playback commands

var playCmd = &cobra.Command{
	Use:   "play",
	Short: "Start playback",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := client.Play(); err != nil {
			notify("Volumio Error", "Could not start playback", "error", true)
			return err
		}
		notify("Volumio", "Playing", "media-playback-start", false)
		return nil
	},
}

var pauseCmd = &cobra.Command{
	Use:   "pause",
	Short: "Pause playback",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := client.Pause(); err != nil {
			notify("Volumio Error", "Could not pause playback", "error", true)
			return err
		}
		notify("Volumio", "Paused", "media-playback-pause", false)
		return nil
	},
}

var toggleCmd = &cobra.Command{
	Use:   "toggle",
	Short: "Toggle play/pause",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := client.TogglePlayPause(); err != nil {
			notify("Volumio Error", "Could not toggle playback", "error", true)
			return err
		}

		time.Sleep(300 * time.Millisecond)
		state, err := client.GetState()
		if err == nil {
			status := "Playing"
			if state.Status == "pause" {
				status = "Paused"
			}
			trackInfo := state.Title
			if state.Artist != "" {
				trackInfo = state.Artist + " - " + state.Title
			}
			notify("Volumio", fmt.Sprintf("%s\n%s", status, trackInfo), "media-playback-start", false)
		}
		return nil
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop playback",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := client.Stop(); err != nil {
			notify("Volumio Error", "Could not stop playback", "error", true)
			return err
		}
		notify("Volumio", "Stopped", "media-playback-stop", false)
		return nil
	},
}

var nextCmd = &cobra.Command{
	Use:   "next",
	Short: "Skip to next track",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := client.Next(); err != nil {
			notify("Volumio Error", "Could not skip to next track", "error", true)
			return err
		}

		time.Sleep(500 * time.Millisecond)
		state, err := client.GetState()
		if err == nil && state.Title != "" {
			trackInfo := state.Title
			if state.Artist != "" {
				trackInfo = state.Artist + " - " + state.Title
			}
			notify("Volumio - Next Track", trackInfo, "media-skip-forward", false)
		} else {
			notify("Volumio", "Next track", "media-skip-forward", false)
		}
		return nil
	},
}

var prevCmd = &cobra.Command{
	Use:   "prev",
	Short: "Go to previous track",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := client.Previous(); err != nil {
			notify("Volumio Error", "Could not go to previous track", "error", true)
			return err
		}

		time.Sleep(500 * time.Millisecond)
		state, err := client.GetState()
		if err == nil && state.Title != "" {
			trackInfo := state.Title
			if state.Artist != "" {
				trackInfo = state.Artist + " - " + state.Title
			}
			notify("Volumio - Previous Track", trackInfo, "media-skip-backward", false)
		} else {
			notify("Volumio", "Previous track", "media-skip-backward", false)
		}
		return nil
	},
}

var skipCmd = &cobra.Command{
	Use:   "skip",
	Short: "Skip to next track (alias for next)",
	RunE:  nextCmd.RunE,
}

// Volume commands

var volumeCmd = &cobra.Command{
	Use:   "volume [up|down|<level>]",
	Short: "Control volume",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		action := args[0]

		switch action {
		case "up":
			if err := client.VolumeUp(10); err != nil {
				notify("Volumio Error", "Could not change volume", "error", true)
				return err
			}
		case "down":
			if err := client.VolumeDown(10); err != nil {
				notify("Volumio Error", "Could not change volume", "error", true)
				return err
			}
		default:
			var level int
			if _, err := fmt.Sscanf(action, "%d", &level); err != nil {
				return fmt.Errorf("invalid volume level: %s (use 'up', 'down', or 0-100)", action)
			}
			if err := client.SetVolume(level); err != nil {
				notify("Volumio Error", "Could not set volume", "error", true)
				return err
			}
		}

		time.Sleep(300 * time.Millisecond)
		state, err := client.GetState()
		if err == nil {
			notify("Volumio Volume", fmt.Sprintf("%d%%", state.Volume), "audio-volume-high", false)
		}
		return nil
	},
}

// Status command

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current playback status",
	RunE: func(cmd *cobra.Command, args []string) error {
		state, err := client.GetState()
		if err != nil {
			return fmt.Errorf("failed to get status: %w", err)
		}

		fmt.Printf("Status: %s\n", state.Status)
		if state.Title != "" {
			fmt.Printf("Title: %s\n", state.Title)
		}
		if state.Artist != "" {
			fmt.Printf("Artist: %s\n", state.Artist)
		}
		if state.Album != "" {
			fmt.Printf("Album: %s\n", state.Album)
		}
		if state.Service != "" {
			fmt.Printf("Service: %s\n", state.Service)
		}
		fmt.Printf("Volume: %d%%", state.Volume)
		if state.Mute {
			fmt.Printf(" (Muted)")
		}
		fmt.Println()

		if state.Duration > 0 {
			fmt.Printf("Position: %d:%02d / %d:%02d\n",
				state.Position/60, state.Position%60,
				state.Duration/60, state.Duration%60)
		}

		modes := []string{}
		if state.Random {
			modes = append(modes, "Shuffle")
		}
		if state.Repeat {
			modes = append(modes, "Repeat")
		}
		if len(modes) > 0 {
			fmt.Printf("Modes: %s\n", modes)
		}

		return nil
	},
}

// Playback mode commands

var shuffleCmd = &cobra.Command{
	Use:   "shuffle",
	Short: "Toggle shuffle mode",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := client.ToggleRandom(); err != nil {
			notify("Volumio Error", "Could not toggle shuffle", "error", true)
			return err
		}

		time.Sleep(300 * time.Millisecond)
		state, err := client.GetState()
		if err == nil {
			status := "disabled"
			if state.Random {
				status = "enabled"
			}
			notify("Volumio", fmt.Sprintf("Shuffle %s", status), "media-playlist-shuffle", false)
		}
		return nil
	},
}

var repeatCmd = &cobra.Command{
	Use:   "repeat",
	Short: "Toggle repeat mode",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := client.ToggleRepeat(); err != nil {
			notify("Volumio Error", "Could not toggle repeat", "error", true)
			return err
		}

		time.Sleep(300 * time.Millisecond)
		state, err := client.GetState()
		if err == nil {
			status := "disabled"
			if state.Repeat {
				status = "enabled"
			}
			notify("Volumio", fmt.Sprintf("Repeat %s", status), "media-playlist-repeat", false)
		}
		return nil
	},
}

// Waybar command

var waybarCmd = &cobra.Command{
	Use:   "waybar",
	Short: "Output Waybar-compatible JSON status",
	Long:  `Output current playback status in Waybar JSON format for use as a custom module`,
	RunE: func(cmd *cobra.Command, args []string) error {
		state, err := client.GetState()
		if err != nil {
			output := waybar.CreateErrorOutput(err.Error())
			return waybar.PrintJSON(output)
		}

		output := waybar.CreateOutput(state, client.GetAlbumArtURL(""))
		return waybar.PrintJSON(output)
	},
}

// Walker command

var walkerCmd = &cobra.Command{
	Use:   "walker [action]",
	Short: "Walker plugin interface",
	Long:  `Walker plugin for browsing and controlling Volumio. Pass action from stdin or as argument.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		state, err := client.GetState()
		if err != nil {
			// Show error but still output menu
			state = nil
		}

		// Default: show main menu
		if len(args) == 0 {
			items := walker.CreateMainMenu(state)
			return walker.PrintItems(items)
		}

		action := args[0]

		// Handle actions
		if len(action) > 7 && action[:7] == "action:" {
			return handleWalkerAction(action[7:])
		}

		// Handle browse
		if len(action) > 7 && action[:7] == "browse:" {
			uri := action[7:]
			items, err := client.Browse(uri)
			if err != nil {
				return fmt.Errorf("failed to browse: %w", err)
			}
			walkerItems := walker.CreateBrowseMenu(items, uri)
			return walker.PrintItems(walkerItems)
		}

		// Handle play
		if len(action) > 5 && action[:5] == "play:" {
			// Parse uri|service
			data := action[5:]
			uri := data
			service := ""
			for i, ch := range data {
				if ch == '|' {
					uri = data[:i]
					service = data[i+1:]
					break
				}
			}
			return client.ReplaceAndPlay(uri, service)
		}

		return fmt.Errorf("unknown action: %s", action)
	},
}

func handleWalkerAction(action string) error {
	switch action {
	case "toggle":
		return client.TogglePlayPause()
	case "play":
		return client.Play()
	case "pause":
		return client.Pause()
	case "stop":
		return client.Stop()
	case "next":
		return client.Next()
	case "prev":
		return client.Previous()
	case "volup":
		return client.VolumeUp(10)
	case "voldown":
		return client.VolumeDown(10)
	case "mute":
		return client.ToggleMute()
	case "shuffle":
		return client.ToggleRandom()
	case "repeat":
		return client.ToggleRepeat()
	default:
		return fmt.Errorf("unknown action: %s", action)
	}
}

// Radio command

var radioCmd = &cobra.Command{
	Use:   "radio <series> [count]",
	Short: "Play random episodes from a radio series",
	Long: `Search for albums matching a radio series pattern, randomly select N albums,
and queue them for playback. Shuffle is automatically disabled.

Series are defined in the config file (~/.config/volu/config.yaml).

Example: volu radio asot 3
         volu radio bb      # defaults to 10 albums`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		seriesName := args[0]

		// Default to 10 albums if count not provided
		count := 10
		if len(args) == 2 {
			var err error
			count, err = strconv.Atoi(args[1])
			if err != nil || count < 1 {
				return fmt.Errorf("count must be a positive integer (got: %s)", args[1])
			}
		}

		// Get series config
		series, exists := cfg.Radio[seriesName]
		if !exists {
			return fmt.Errorf("unknown radio series: %s (check your config file at ~/.config/volu/config.yaml)", seriesName)
		}

		// Create radio player
		player := radio.NewPlayer(client)

		// Show search notification
		notify("Volumio Radio",
			fmt.Sprintf("Searching for %s episodes...", series.Name),
			"media-playlist-shuffle", false)

		// Play random episodes
		if err := player.PlayRandomEpisodes(series.SearchQuery, series.Pattern, count); err != nil {
			notify("Volumio Error", err.Error(), "error", true)
			return err
		}

		// Success notification
		notify("Volumio Radio",
			fmt.Sprintf("Playing %d random %s episodes", count, series.Name),
			"media-playback-start", false)

		return nil
	},
}

// Elephant provider

var elephantCmd = &cobra.Command{
	Use:   "elephant",
	Short: "Start elephant provider",
	Long:  `Start the elephant provider for integration with https://github.com/abenz1267/elephant`,
	RunE: func(cmd *cobra.Command, args []string) error {
		provider := elephant.NewProvider(volumioHost)
		return provider.Run()
	},
}
