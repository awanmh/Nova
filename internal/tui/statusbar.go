package tui

import (
	"fmt"
	"strings"
)

// RenderStatusBar returns the styled header/footer status bar.
func RenderStatusBar(status, modelName, persona string, tokenCount int) string {
	left := fmt.Sprintf(" NOVA [%s] ", strings.ToUpper(status))
	mid := fmt.Sprintf(" Persona: %-10s | Model: %-12s ", persona, modelName)
	right := fmt.Sprintf(" Tokens: %-6d ", tokenCount)

	bar := left + mid + right
	return styleStatusBar.Render(bar)
}
