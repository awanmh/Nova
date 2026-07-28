package tui

import (
	"fmt"
	"strings"

	"github.com/awanmh/Nova/internal/llm"
)

// RenderChatHistory formats a list of chat messages into styled terminal strings.
func RenderChatHistory(messages []llm.Message) string {
	var sb strings.Builder
	for _, m := range messages {
		switch m.Role {
		case "user":
			sb.WriteString(styleUserMsg.Render("You > ") + m.Content + "\n\n")
		case "assistant":
			sb.WriteString(styleAssistantMsg.Render("NOVA > ") + m.Content + "\n\n")
		case "tool":
			sb.WriteString(styleToolMsg.Render(fmt.Sprintf("[Tool Output: %s]", m.ToolID)) + "\n" + m.Content + "\n\n")
		case "system":
			sb.WriteString(styleSystemMsg.Render("[System: "+m.Content+"]") + "\n\n")
		}
	}
	return strings.TrimSpace(sb.String())
}
