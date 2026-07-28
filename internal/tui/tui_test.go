package tui_test

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbletea"
	"github.com/awanmh/Nova/internal/llm"
	"github.com/awanmh/Nova/internal/tui"
)

func TestTUI_ModelInitAndUpdate(t *testing.T) {
	m := tui.NewModel("llama3", "/test/workspace", "general")

	if m.Persona != "general" || m.ModelName != "llama3" {
		t.Fatalf("unexpected init model state: %+v", m)
	}

	// 1. Switch persona via ctrl+p
	res, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("ctrl+p")})
	m2 := res.(tui.Model)
	if m2.Persona == "general" {
		t.Fatalf("expected persona change after ctrl+p, got %s", m2.Persona)
	}

	// Transition to Agent Workspace for direct chat test
	m2.State = tui.StateAgentWorkspace

	// 2. Type characters and press Enter
	for _, r := range "hello" {
		res, _ = m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m2 = res.(tui.Model)
	}
	if m2.InputText != "hello" {
		t.Fatalf("expected input 'hello', got '%s'", m2.InputText)
	}
	res, _ = m2.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m3 := res.(tui.Model)
	if len(m3.Messages) != 2 {
		t.Fatalf("expected 2 chat messages (user+assistant), got %d", len(m3.Messages))
	}

	// 3. Clear via /clear slash command
	res, _ = m3.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	m4 := res.(tui.Model)
	for _, r := range "clear" {
		res, _ = m4.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m4 = res.(tui.Model)
	}
	res, _ = m4.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m5 := res.(tui.Model)
	if len(m5.Messages) != 0 {
		t.Fatalf("expected 0 messages after /clear, got %d", len(m5.Messages))
	}
}

func TestTUI_MenuArchitecture(t *testing.T) {
	m := tui.NewModel("llama3", "/test/workspace", "general")
	if m.State != tui.StateMainMenu {
		t.Fatalf("expected default StateMainMenu, got %v", m.State)
	}

	view := m.View()
	if !strings.Contains(view, "N O V A") || !strings.Contains(view, "1. Agent") || !strings.Contains(view, "6. Extensions") || !strings.Contains(view, "0. Exit") || !strings.Contains(view, "Extensions  : 5 enabled") {
		t.Fatalf("main menu view missing expected text: %s", view)
	}

	// Select 1. Agent (key "1")
	res, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("1")})
	mAgent := res.(tui.Model)
	if mAgent.State != tui.StateAgentSetup {
		t.Fatalf("expected StateAgentSetup after selecting 1, got %v", mAgent.State)
	}

	// Enter option 3 (Back -> key "3")
	res, _ = mAgent.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("3")})
	mStep1 := res.(tui.Model)
	if mStep1.WizardStep != 1 {
		t.Fatalf("expected wizard step 1, got %d", mStep1.WizardStep)
	}

	// Press Enter through steps 1 -> 2 -> 3 -> StateAgentWorkspace
	res, _ = mStep1.Update(tea.KeyMsg{Type: tea.KeyEnter})
	mStep2 := res.(tui.Model)
	if mStep2.WizardStep != 2 {
		t.Fatalf("expected wizard step 2, got %d", mStep2.WizardStep)
	}

	res, _ = mStep2.Update(tea.KeyMsg{Type: tea.KeyEnter})
	mStep3 := res.(tui.Model)
	if mStep3.WizardStep != 3 {
		t.Fatalf("expected wizard step 3, got %d", mStep3.WizardStep)
	}

	viewUnderstanding := mStep3.View()
	if !strings.Contains(viewUnderstanding, "PROJECT REVIEW") {
		t.Fatalf("missing project review card: %s", viewUnderstanding)
	}

	res, _ = mStep3.Update(tea.KeyMsg{Type: tea.KeyEnter})
	mWorkspace := res.(tui.Model)
	if mWorkspace.State != tui.StateAgentWorkspace {
		t.Fatalf("expected StateAgentWorkspace, got %v", mWorkspace.State)
	}
}

func TestTUI_ExtensionsMenu(t *testing.T) {
	m := tui.NewModel("llama3", "/test/workspace", "general")
	res, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("6")})
	mExt := res.(tui.Model)
	if mExt.State != tui.StateExtensionsMenu {
		t.Fatalf("expected StateExtensionsMenu, got %v", mExt.State)
	}

	view := mExt.View()
	if !strings.Contains(view, "EXTENSIONS") || !strings.Contains(view, "1. Installed") || !strings.Contains(view, "8. Extension Health") {
		t.Fatalf("missing expected text in extensions menu: %s", view)
	}

	res, _ = mExt.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("0")})
	mBack := res.(tui.Model)
	if mBack.State != tui.StateMainMenu {
		t.Fatalf("expected return to StateMainMenu, got %v", mBack.State)
	}
}

func TestTUI_Renderers(t *testing.T) {
	sb := tui.RenderStatusBar("READY", "llama3", "tdd", 120)
	if !strings.Contains(sb, "NOVA") || !strings.Contains(sb, "tdd") {
		t.Fatalf("unexpected status bar render: %s", sb)
	}

	modal := tui.RenderModal(&tui.ModalRequest{
		ToolName:  "write_file",
		RiskLevel: "HIGH",
		Arguments: `{"path":"main.go"}`,
	})
	if !strings.Contains(modal, "PERMISSION REQUIRED") || !strings.Contains(modal, "write_file") {
		t.Fatalf("unexpected modal render: %s", modal)
	}

	chat := tui.RenderChatHistory([]llm.Message{
		{Role: "user", Content: "Hello NOVA"},
		{Role: "assistant", Content: "Ready to code."},
	})
	if !strings.Contains(chat, "Hello NOVA") || !strings.Contains(chat, "Ready to code.") {
		t.Fatalf("unexpected chat render: %s", chat)
	}
}
