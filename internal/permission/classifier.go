package permission

import (
	"encoding/json"
	"strings"
)

// ClassifyRisk determines the effective risk level of a tool call based on tool type and arguments.
func ClassifyRisk(toolName, argsJSON string) RiskLevel {
	switch toolName {
	case "read_file", "list_dir", "search_files", "git_status", "git_diff":
		return RiskReadOnly
	case "write_file", "apply_patch":
		return RiskLowModify
	case "run_command":
		if isDestructiveCommand(argsJSON) {
			return RiskHighImpact
		}
		return RiskLowModify
	default:
		return RiskHighImpact
	}
}

func isDestructiveCommand(argsJSON string) bool {
	var payload struct {
		Command string `json:"command"`
	}
	_ = json.Unmarshal([]byte(argsJSON), &payload)
	lower := strings.ToLower(payload.Command)

	destructiveKeywords := []string{
		"rm -rf", "rm -r", "del /s", "rd /s",
		"git reset --hard", "git clean -fd",
		"sudo", "mkfs", "dd ", "format ",
		"drop database", "truncate table",
	}
	for _, kw := range destructiveKeywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}
