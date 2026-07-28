package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// ExportJSON exports a session and its message history as formatted JSON string.
func ExportJSON(ctx context.Context, store SessionStore, sessionID string) (string, error) {
	session, err := store.GetSession(ctx, sessionID)
	if err != nil {
		return "", err
	}
	history, err := store.GetHistory(ctx, sessionID)
	if err != nil {
		return "", err
	}
	data := struct {
		Session  Session       `json:"session"`
		Messages []interface{} `json:"messages"`
	}{
		Session:  session,
		Messages: make([]interface{}, len(history)),
	}
	for i, m := range history {
		data.Messages[i] = m
	}
	out, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// ExportMarkdown exports a session and its message history as a clean Markdown transcript.
func ExportMarkdown(ctx context.Context, store SessionStore, sessionID string) (string, error) {
	session, err := store.GetSession(ctx, sessionID)
	if err != nil {
		return "", err
	}
	history, err := store.GetHistory(ctx, sessionID)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# NOVA Session Transcript: %s\n\n", session.Title))
	sb.WriteString(fmt.Sprintf("- **Session ID:** `%s`\n", session.ID))
	sb.WriteString(fmt.Sprintf("- **Persona:** `%s`\n", session.Persona))
	sb.WriteString(fmt.Sprintf("- **Created At:** %s\n\n", session.CreatedAt.Format("2006-01-02 15:04:05 UTC")))
	sb.WriteString("---\n\n")

	for _, msg := range history {
		role := strings.ToUpper(msg.Role)
		sb.WriteString(fmt.Sprintf("## %s\n\n", role))
		sb.WriteString(strings.TrimSpace(msg.Content))
		sb.WriteString("\n\n---\n\n")
	}

	return strings.TrimSpace(sb.String()), nil
}
