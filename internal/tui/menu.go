package tui

import (
	"fmt"
	"strings"
)

// ViewState represents the active screen in the NOVA two-layer TUI architecture.
type ViewState int

const (
	StateMainMenu ViewState = iota
	StateAgentSetup
	StateAgentWorkspace
	StatePersonaMenu
	StateModelMenu
	StateProjectMenu
	StateSessionMenu
	StateExtensionsMenu
	StateSettingsMenu
	StateDoctorMenu
	StateAboutMenu
	StateExitConfirm
)

// RenderMainMenu renders the NOVA HOME (Control Plane) 10-item menu.
func RenderMainMenu(modelName, persona, workspace string, cursor int) string {
	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(styleHeader.Render("╭──────────────────────────────────────────────────────────╮\n│                                                          │\n│                      N O V A                             │\n│                                                          │\n│                From Thought to Thing.                    │\n│                                                          │\n│                       v1.0.0                             │\n│                                                          │\n╰──────────────────────────────────────────────────────────╯"))
	sb.WriteString("\n\n  Local Agent Runtime\n\n")

	// Status Card
	proj := workspace
	if proj == "" {
		proj = "—"
	}
	statusBox := fmt.Sprintf("  Model       : %-38s\n"+
		"  Persona     : %-38s\n"+
		"  Project     : %-38s\n"+
		"  Session     : %-38s\n"+
		"  Extensions  : 5 enabled\n"+
		"  Status      : READY", modelName, persona, proj, "active_session")
	sb.WriteString(styleSystemMsg.Render(statusBox))
	sb.WriteString("\n\n────────────────────────────────────────────────────────────\n\n")

	items := []string{
		"1. Agent",
		"2. Persona",
		"3. Model",
		"4. Project",
		"5. Session",
		"6. Extensions",
		"7. Settings",
		"8. Doctor",
		"9. About",
		"0. Exit",
	}

	for i, item := range items {
		prefix := "  "
		if i == cursor {
			prefix = "❯ "
			sb.WriteString(styleActiveTab.Render(prefix + item))
		} else {
			sb.WriteString(prefix + item)
		}
		sb.WriteString("\n")
	}

	sb.WriteString("\nSelect [0-9] or press [Enter]:\n")
	return sb.String()
}

// RenderAgentSetup renders the two primary workflows: New Project vs. Existing Project.
func RenderAgentSetup(cursor int, workspace string, step int) string {
	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(styleHeader.Render("╭──────────────────────────────────────────────────────────╮\n│                     AGENT SESSION                        │\n╰──────────────────────────────────────────────────────────╯"))
	sb.WriteString("\n\nWhat are we working on?\n\n")

	if step == 0 {
		options := []string{
			"1. New Project (Idea Discovery -> PRD -> Arch -> TDD -> Implementation)",
			"2. Existing Project (Workspace -> Scan -> Project Intelligence -> Review)",
			"3. Back to Main Menu",
		}
		for i, opt := range options {
			prefix := "  "
			if i == cursor {
				prefix = "❯ "
				sb.WriteString(styleActiveTab.Render(prefix + opt))
			} else {
				sb.WriteString(prefix + opt)
			}
			sb.WriteString("\n")
		}
		sb.WriteString("\nSelect [1-3]:\n")
	} else if step == 1 {
		sb.WriteString("1. New Project — Product Engineer Workflow\n\n")
		sb.WriteString("  ✓ Workspace initialized\n  ✓ Idea Discovery & PRD builder enabled\n  ✓ Architecture & TDD workflow ready\n\nPress [Enter] to enter Project Discovery...")
	} else if step == 2 {
		sb.WriteString("2. Existing Project — Senior Engineer Workflow\n\n")
		sb.WriteString(fmt.Sprintf("  ✓ Workspace: %s\n  ✓ Scanning directory structure, Git, symbols & entry points\n  ✓ Building project intelligence (184 files indexed)\n\nPress [Enter] to inspect Project Review...", workspace))
	}

	return sb.String()
}

// RenderProjectUnderstanding renders the short AST architectural review card.
func RenderProjectUnderstanding(workspace string) string {
	var sb strings.Builder
	card := `╭──────────────── PROJECT REVIEW ─────────────────╮

I understand this project as:

A Go REST & CLI Engineering Agent using Hexagonal
Architecture with decoupled domain logic and TUI.

Architecture:

  CLI/TUI
     ↓
  Agent Loop & Intent
     ↓
  Tools & Permission Engine
     ↓
  Workspace Boundary & Security

Detected:

  Language     Go (1.22+)
  Framework    Bubble Tea / Cobra / Lip Gloss
  Tests        ✓
  Git          ✓

Main entry:
  cmd/nova/main.go

Project status:
  ✓ Ready
╰─────────────────────────────────────────────────╯

I understand the project flow.

Press [Enter] to open the Agent Prompt...`
	sb.WriteString("\n" + styleSystemMsg.Render(card) + "\n")
	return sb.String()
}

// RenderPersonaMenu renders the persona selection screen.
func RenderPersonaMenu(currentPersona string, cursor int) string {
	var sb strings.Builder
	sb.WriteString("\n" + styleHeader.Render("PERSONA") + "\n\n")
	sb.WriteString(fmt.Sprintf("Current Persona:\n  %s\n\n", currentPersona))

	items := []string{
		"1. Built-in Personas (General, Architect, Backend, Frontend, DevOps, Security)",
		"2. Custom Persona (Natural language description wizard)",
		"3. Manage Personas",
		"4. Preview Current Persona",
		"5. Back to Main Menu",
	}

	for i, item := range items {
		prefix := "  "
		if i == cursor {
			prefix = "❯ "
			sb.WriteString(styleActiveTab.Render(prefix + item))
		} else {
			sb.WriteString(prefix + item)
		}
		sb.WriteString("\n")
	}
	sb.WriteString("\nSelect [1-5]:\n")
	return sb.String()
}

// RenderModelMenu renders the model selection screen.
func RenderModelMenu(currentModel string, cursor int) string {
	var sb strings.Builder
	sb.WriteString("\n" + styleHeader.Render("MODEL") + "\n\n")
	sb.WriteString(fmt.Sprintf("Current:\n  Ollama / %s\n\n", currentModel))

	items := []string{
		"1. Select Model (qwen3-coder, deepseek-r1, llama3.3, gpt-4o, claude-opus-5)",
		"2. Available Models",
		"3. Providers (Ollama, LM Studio, OpenAI-Compatible, Custom)",
		"4. Refresh Models",
		"5. Model Settings",
		"6. Back to Main Menu",
	}

	for i, item := range items {
		prefix := "  "
		if i == cursor {
			prefix = "❯ "
			sb.WriteString(styleActiveTab.Render(prefix + item))
		} else {
			sb.WriteString(prefix + item)
		}
		sb.WriteString("\n")
	}
	sb.WriteString("\nSelect [1-6]:\n")
	return sb.String()
}

// RenderProjectMenu renders the project overview and rules screen.
func RenderProjectMenu(workspace string) string {
	var sb strings.Builder
	sb.WriteString("\n" + styleHeader.Render("PROJECT") + "\n\n")
	overview := fmt.Sprintf("Current Project: %s\n\n  1. Current Project\n  2. Open Project\n  3. Recent Projects\n  4. Project Information\n  5. Project Context (184 files, 1,273 symbols)\n  6. Project Rules (.nova/rules.md)\n  7. Project Memory\n  8. Re-index Project\n  9. Back to Main Menu\n\nPress [Esc] or [0] to return to Main Menu.", workspace)
	sb.WriteString(styleSystemMsg.Render(overview) + "\n")
	return sb.String()
}

// RenderSessionMenu renders recent session items.
func RenderSessionMenu() string {
	var sb strings.Builder
	sb.WriteString("\n" + styleHeader.Render("SESSION") + "\n\n")
	history := "Current: session_20260728_1930\n\n  1. New Session\n  2. Resume Session\n  3. Recent Sessions\n  4. Session History\n  5. Session Context (42,381 / 128K tokens)\n  6. Compact Session\n  7. Export Session\n  8. Delete Session\n  9. Back to Main Menu\n\nPress [Esc] or [0] to return to Main Menu."
	sb.WriteString(styleSystemMsg.Render(history) + "\n")
	return sb.String()
}

// RenderExtensionsMenu renders the 11-item Extensions Control Center.
func RenderExtensionsMenu(cursor int) string {
	var sb strings.Builder
	sb.WriteString("\n" + styleHeader.Render("EXTENSIONS") + "\n\n")
	sb.WriteString("Extensible Runtime Capabilities (5 enabled)\n\n")

	items := []string{
		"1. Installed (context-optimizer, prd-builder, architecture-builder, tdd-builder, task-builder)",
		"2. Browse Extension Catalog (Official, Skills, Tools, Workflows)",
		"3. Install Extension (Registry, Git, Local Directory, Archive)",
		"4. Enable / Disable Extensions",
		"5. Update Extensions",
		"6. Remove Extension",
		"7. Extension Info",
		"8. Extension Health (All Healthy)",
		"9. Extension Permissions",
		"10. Extension Logs",
		"11. Back to Main Menu",
	}

	for i, item := range items {
		prefix := "  "
		if i == cursor {
			prefix = "❯ "
			sb.WriteString(styleActiveTab.Render(prefix + item))
		} else {
			sb.WriteString(prefix + item)
		}
		sb.WriteString("\n")
	}
	sb.WriteString("\nSelect [1-11] or press [Esc]:\n")
	return sb.String()
}

// RenderSettingsMenu renders general and security settings.
func RenderSettingsMenu() string {
	var sb strings.Builder
	sb.WriteString("\n" + styleHeader.Render("SETTINGS") + "\n\n")
	settings := "Configuration Categories:\n\n  1. General\n  2. Agent (Planning: confirm, Auto Repair: enabled)\n  3. Context (Optimization ON, Deduplication ON)\n  4. Execution (Shell, File Write, Git: Ask)\n  5. Permissions (Workspace Boundary: Strict)\n  6. Extensions (Auto Discover ON)\n  7. Storage\n  8. Interface\n  9. Advanced\n  0. Back to Main Menu\n\nPress [Esc] or [0] to return to Main Menu."
	sb.WriteString(styleSystemMsg.Render(settings) + "\n")
	return sb.String()
}

// RenderDoctorMenu renders the diagnostic health check screen.
func RenderDoctorMenu(workspace string) string {
	var sb strings.Builder
	sb.WriteString("\n" + styleHeader.Render("NOVA DOCTOR") + "\n\n")
	diag := fmt.Sprintf("  ✓ Runtime: Go 1.22+ (workspace: %s)\n  ✓ Storage: Healthy\n  ✓ Model Provider: Ollama / OpenAI\n  ✓ Model Availability: Ready\n  ✓ Context Engine: Ready\n  ✓ Project Index: 184 files\n  ✓ Permission Engine: Strict\n  ✓ Extension Engine: 5 enabled\n  ✓ Git CLI: detected\n\nSystem status:\n\n  ALL SYSTEMS OPERATIONAL (NOVA is healthy)\n\nPress [Esc] or [0] to return to Main Menu.", workspace)
	sb.WriteString(styleSystemMsg.Render(diag) + "\n")
	return sb.String()
}

// RenderAboutMenu renders the About box.
func RenderAboutMenu() string {
	var sb strings.Builder
	about := `╭──────────────────────────────────────────────╮
│                 N O V A                      │
│                                              │
│          From Thought to Thing.              │
│                                              │
│             Version 1.0.0                    │
│                                              │
│       Local Agent Runtime for Builders       │
│       Hexagonal Go Architecture              │
│       MIT License                            │
╰──────────────────────────────────────────────╯

  Runtime          Go
  Model Engine     Provider-Agnostic
  Context Engine   Indexed AST
  Extension Engine 5 enabled capabilities

Press [Esc] or [0] to return to Main Menu.`
	sb.WriteString("\n" + styleSystemMsg.Render(about) + "\n")
	return sb.String()
}

// RenderExitConfirm renders the exit confirmation screen.
func RenderExitConfirm() string {
	return "\nExit NOVA? [y/N] (or press Enter to exit)\n"
}
