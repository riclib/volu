package waybar

import (
	"testing"

	"github.com/riclib/volu/internal/volumio"
)

func TestEscapeMarkup(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "ampersand",
			input:    "Prophet & Janga",
			expected: "Prophet &amp; Janga",
		},
		{
			name:     "less than",
			input:    "Volume < 50",
			expected: "Volume &lt; 50",
		},
		{
			name:     "greater than",
			input:    "Volume > 50",
			expected: "Volume &gt; 50",
		},
		{
			name:     "quotes",
			input:    `Song "Title"`,
			expected: "Song &quot;Title&quot;",
		},
		{
			name:     "apostrophe",
			input:    "Don't Stop",
			expected: "Don&apos;t Stop",
		},
		{
			name:     "multiple special chars",
			input:    `Artist & "Band" <Live>`,
			expected: `Artist &amp; &quot;Band&quot; &lt;Live&gt;`,
		},
		{
			name:     "no special chars",
			input:    "Normal Text",
			expected: "Normal Text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EscapeMarkup(tt.input)
			if result != tt.expected {
				t.Errorf("EscapeMarkup(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCreateOutput_EscapesMarkup(t *testing.T) {
	state := &volumio.PlayerState{
		Status: "play",
		Title:  "Ha Te Tudnád... / Love Gone Wrong (feat. Prophet & Janga)",
		Artist: "Balkan Fanatik, Prophet & Janga",
		Album:  "Test Album <2024>",
	}

	output := CreateOutput(state, "http://volumio.local:3000")

	// Check that ampersands are escaped in text
	expectedText := "♫ Balkan Fanatik, Prophet &amp; Janga - Ha Te Tudnád... / Love Gone Wrong (feat. Prophet &amp; Janga) ▶"
	if output.Text != expectedText {
		t.Errorf("Text not properly escaped.\nGot:  %q\nWant: %q", output.Text, expectedText)
	}

	// Check that ampersands are escaped in tooltip
	if !containsString(output.Tooltip, "Artist: Balkan Fanatik, Prophet &amp; Janga") {
		t.Errorf("Tooltip artist not properly escaped: %q", output.Tooltip)
	}
	if !containsString(output.Tooltip, "Title: Ha Te Tudnád... / Love Gone Wrong (feat. Prophet &amp; Janga)") {
		t.Errorf("Tooltip title not properly escaped: %q", output.Tooltip)
	}
	if !containsString(output.Tooltip, "Album: Test Album &lt;2024&gt;") {
		t.Errorf("Tooltip album not properly escaped: %q", output.Tooltip)
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
