package tui

import (
	"fmt"
	"strings"
)

// ModalRequest holds info for an active permission approval prompt.
type ModalRequest struct {
	ToolName  string
	RiskLevel string
	Arguments string
}

// RenderModal formats the permission modal popup box.
func RenderModal(req *ModalRequest) string {
	if req == nil {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("⚠️  PERMISSION REQUIRED FOR TOOL EXECUTION\n\n")
	sb.WriteString(fmt.Sprintf("Tool:       %s\n", req.ToolName))
	sb.WriteString(fmt.Sprintf("Risk Level: %s\n", req.RiskLevel))
	sb.WriteString(fmt.Sprintf("Arguments:  %s\n\n", req.Arguments))
	sb.WriteString("Options: [Y]es (allow once)  |  [N]o (deny)  |  [A]llow (entire session)")
	return styleModalBox.Render(sb.String())
}
