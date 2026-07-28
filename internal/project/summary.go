package project

import (
	"fmt"
	"strings"
)

// GenerateSummary formats a natural-language architecture overview for prompt injection and CLI display.
func GenerateSummary(snap *Snapshot) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Project: %s\n", snap.Name))
	if len(snap.Languages) > 0 {
		sb.WriteString(fmt.Sprintf("Languages: %s\n", strings.Join(snap.Languages, ", ")))
	}
	if len(snap.Frameworks) > 0 {
		sb.WriteString(fmt.Sprintf("Frameworks: %s\n", strings.Join(snap.Frameworks, ", ")))
	}
	if len(snap.EntryPoints) > 0 {
		sb.WriteString(fmt.Sprintf("Entry Points: %s\n", strings.Join(snap.EntryPoints, ", ")))
	}

	// File classification counts
	counts := make(map[FileClass]int)
	for _, f := range snap.Files {
		counts[f.Class]++
	}

	sb.WriteString(fmt.Sprintf("Files: %d total (Source: %d, Test: %d, Config: %d, Doc: %d)\n",
		len(snap.Files),
		counts[FileClassSource],
		counts[FileClassTest],
		counts[FileClassConfig],
		counts[FileClassDoc],
	))

	return strings.TrimSpace(sb.String())
}
