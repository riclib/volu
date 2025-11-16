package gui

import (
	"fmt"
	"log/slog"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	"github.com/riclib/volu/internal/config"
	"github.com/riclib/volu/internal/radio"
	"github.com/riclib/volu/internal/volumio"
)

// Popup represents the main GUI window
type Popup struct {
	app         fyne.App
	window      fyne.Window
	client      *volumio.Client
	config      *config.Config
	nowPlay     *NowPlaying
	selected    int // Currently selected shortcut index
	statusLabel *widget.Label
	titleLabel  *widget.Label
	artistLabel *widget.Label
	albumLabel  *widget.Label
	albumArt    *canvas.Image
}

// NewPopup creates a new popup window
func NewPopup(client *volumio.Client, cfg *config.Config) *Popup {
	return &Popup{
		app:      app.New(),
		client:   client,
		config:   cfg,
		selected: 0,
	}
}

// Show displays the popup window
func (p *Popup) Show() {
	p.window = p.app.NewWindow("Volu - Music Browser")
	p.window.Resize(fyne.NewSize(800, 600))

	// Load current state
	state, err := p.client.GetState()
	if err != nil {
		slog.Error("popup", "getstate", err)
		state = nil
	}

	p.nowPlay = NewNowPlaying(state)
	p.nowPlay.VoluHost = p.config.Host

	// Build UI
	content := p.buildUI()
	p.window.SetContent(content)

	// Setup keyboard bindings
	p.setupKeyBindings()

	// TODO: Add periodic state updates (needs proper thread-safe implementation)
	// p.startStateUpdates()

	p.window.ShowAndRun()
}

func (p *Popup) buildUI() fyne.CanvasObject {
	// Now Playing section
	nowPlayingSection := p.buildNowPlayingSection()

	// Shortcuts grid
	shortcutsSection := p.buildShortcutsSection()

	// Combine sections
	return container.NewVBox(
		nowPlayingSection,
		widget.NewSeparator(),
		shortcutsSection,
	)
}

func (p *Popup) buildNowPlayingSection() fyne.CanvasObject {
	// Album art
	p.albumArt = p.createAlbumArtImage()

	// Track info
	p.titleLabel = widget.NewLabelWithStyle(
		p.nowPlay.Title,
		fyne.TextAlignLeading,
		fyne.TextStyle{Bold: true},
	)
	p.titleLabel.Wrapping = fyne.TextWrapWord

	p.artistLabel = widget.NewLabel(p.nowPlay.Artist)
	p.albumLabel = widget.NewLabel(p.nowPlay.Album)

	infoVBox := container.NewVBox(
		widget.NewLabel("Now Playing"),
		p.titleLabel,
		p.artistLabel,
		p.albumLabel,
	)

	// Playback controls
	p.statusLabel = widget.NewLabel(fmt.Sprintf("%s  %s  🔊 %d%%",
		getStatusIcon(p.nowPlay.Status),
		p.nowPlay.Progress,
		p.nowPlay.Volume,
	))

	controlButtons := p.buildControlButtons()

	// Layout
	return container.NewBorder(
		nil,
		container.NewVBox(p.statusLabel, controlButtons),
		p.albumArt,
		nil,
		infoVBox,
	)
}

func (p *Popup) createAlbumArtImage() *canvas.Image {
	var img *canvas.Image

	artURL := getAlbumArtURL(p.nowPlay.AlbumArt, p.nowPlay.VoluHost)

	if artURL != "" {
		// Try to load album art from URL
		uri, err := storage.ParseURI(artURL)
		if err == nil {
			img = canvas.NewImageFromURI(uri)
		}
	}

	if img == nil {
		// Fallback to placeholder
		img = canvas.NewImageFromResource(nil)
	}

	img.FillMode = canvas.ImageFillContain
	img.SetMinSize(fyne.NewSize(200, 200))

	return img
}

func (p *Popup) buildControlButtons() fyne.CanvasObject {
	playPauseBtn := widget.NewButton(getStatusIcon(p.nowPlay.Status), func() {
		p.togglePlayPause()
	})

	nextBtn := widget.NewButton("⏭️", func() {
		p.next()
	})

	prevBtn := widget.NewButton("⏮️", func() {
		p.prev()
	})

	stopBtn := widget.NewButton("⏹️", func() {
		p.stop()
	})

	return container.NewHBox(
		layout.NewSpacer(),
		prevBtn,
		playPauseBtn,
		nextBtn,
		stopBtn,
		layout.NewSpacer(),
	)
}

func (p *Popup) buildShortcutsSection() fyne.CanvasObject {
	shortcuts := LoadShortcuts(p.config)

	if len(shortcuts) == 0 {
		return widget.NewLabel("No radio shortcuts configured")
	}

	// Create grid of shortcut cards
	cards := make([]fyne.CanvasObject, len(shortcuts))

	for i, sc := range shortcuts {
		idx := i // Capture for closure
		shortcut := sc

		// Create card with icon and name
		iconLabel := widget.NewLabelWithStyle(
			shortcut.Icon,
			fyne.TextAlignCenter,
			fyne.TextStyle{Bold: true},
		)
		// Make icon larger
		iconLabel.TextStyle.Symbol = true

		nameLabel := widget.NewLabel(shortcut.Name)
		nameLabel.Alignment = fyne.TextAlignCenter
		nameLabel.Wrapping = fyne.TextWrapWord

		cardContent := container.NewVBox(
			iconLabel,
			nameLabel,
		)

		card := widget.NewButton("", func() {
			p.playRadio(shortcut.Key)
		})

		// Overlay card content on button
		cards[idx] = container.NewStack(
			card,
			container.NewCenter(cardContent),
		)
	}

	// Create grid layout (3 columns)
	grid := container.NewGridWithColumns(3, cards...)

	return container.NewVBox(
		widget.NewLabel("Radio Shortcuts"),
		container.NewScroll(grid),
	)
}

func (p *Popup) setupKeyBindings() {
	p.window.Canvas().SetOnTypedKey(func(ev *fyne.KeyEvent) {
		switch ev.Name {
		case fyne.KeyEscape:
			p.window.Close()
		case fyne.KeySpace:
			p.togglePlayPause()
		case fyne.KeyReturn:
			p.activateSelected()
		}
	})

	p.window.Canvas().SetOnTypedRune(func(r rune) {
		switch r {
		case 'q':
			p.window.Close()
		}
	})
}

// Control actions

func (p *Popup) togglePlayPause() {
	if err := p.client.TogglePlayPause(); err != nil {
		slog.Error("popup", "toggle", err)
	}
}

func (p *Popup) next() {
	if err := p.client.Next(); err != nil {
		slog.Error("popup", "next", err)
	}
}

func (p *Popup) prev() {
	if err := p.client.Previous(); err != nil {
		slog.Error("popup", "prev", err)
	}
}

func (p *Popup) stop() {
	if err := p.client.Stop(); err != nil {
		slog.Error("popup", "stop", err)
	}
}

func (p *Popup) playRadio(seriesKey string) {
	player := radio.NewPlayer(p.client)

	albumCount := 10 // Default

	// Check if series exists in config
	series, exists := p.config.Radio[seriesKey]
	if !exists {
		slog.Error("popup", "playRadio", fmt.Sprintf("unknown series: %s", seriesKey))
		return
	}

	slog.Info("popup", "playing radio", series.Name)

	// Play the radio series
	// This uses the same logic as the CLI radio command
	if series.AlbumArtist != "" {
		// Album artist mode
		pattern := fmt.Sprintf("(?i)^%s$", series.AlbumArtist)
		if err := player.PlayRandomAlbumsByArtist(series.AlbumArtist, pattern, "", albumCount); err != nil {
			slog.Error("popup", "playRadio", err)
		}
	} else if series.TrackArtist != "" {
		// Track mode
		count := series.TrackCount
		if count == 0 {
			count = 20
		}
		pattern := fmt.Sprintf("(?i)%s", series.TrackArtist)
		if err := player.PlayRandomTracks(series.TrackArtist, pattern, count); err != nil {
			slog.Error("popup", "playRadio", err)
		}
	} else {
		// Legacy album search mode
		if err := player.PlayRandomEpisodes(series.SearchQuery, series.Pattern, albumCount); err != nil {
			slog.Error("popup", "playRadio", err)
		}
	}
}

func (p *Popup) activateSelected() {
	shortcuts := LoadShortcuts(p.config)
	if p.selected >= 0 && p.selected < len(shortcuts) {
		p.playRadio(shortcuts[p.selected].Key)
	}
}

func (p *Popup) startStateUpdates() {
	// Update every 2 seconds
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			p.updateNowPlaying()
		}
	}()
}

func (p *Popup) updateNowPlaying() {
	state, err := p.client.GetState()
	if err != nil {
		slog.Error("popup", "updateState", err)
		return
	}

	p.nowPlay = NewNowPlaying(state)
	p.nowPlay.VoluHost = p.config.Host

	// Update UI elements - SetText is thread-safe and triggers refresh
	p.titleLabel.SetText(p.nowPlay.Title)
	p.artistLabel.SetText(p.nowPlay.Artist)
	p.albumLabel.SetText(p.nowPlay.Album)
	p.statusLabel.SetText(fmt.Sprintf("%s  %s  🔊 %d%%",
		getStatusIcon(p.nowPlay.Status),
		p.nowPlay.Progress,
		p.nowPlay.Volume,
	))

	// Update album art if changed
	artURL := getAlbumArtURL(p.nowPlay.AlbumArt, p.nowPlay.VoluHost)
	if artURL != "" {
		uri, err := storage.ParseURI(artURL)
		if err == nil {
			p.albumArt.Resource = nil
			if img := canvas.NewImageFromURI(uri); img != nil {
				p.albumArt.Resource = img.Resource
				p.albumArt.Refresh()
			}
		}
	}
}
