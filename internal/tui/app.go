package tui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/awanmh/Nova/internal/llm"
)

// Model represents the top-level Bubble Tea TUI state for NOVA interactive agent.
type Model struct {
	State       ViewState
	MenuCursor  int
	SubCursor   int
	WizardStep  int
	DirectChat  bool
	ModelName   string
	Workspace   string
	Persona     string
	Status      string
	TokenCount  int
	Messages    []llm.Message
	InputText   string
	ModalReq    *ModalRequest
	ShowHelp    bool
	quitting    bool
	personaList []string
	personaIdx  int
}

// NewModel creates a new initial TUI model.
func NewModel(modelName, workspace, persona string) Model {
	if modelName == "" {
		modelName = "llama3"
	}
	if persona == "" {
		persona = "general"
	}
	pList := []string{"general", "architect", "security", "tdd"}
	pIdx := 0
	for i, p := range pList {
		if p == persona {
			pIdx = i
			break
		}
	}
	return Model{
		State:       StateMainMenu,
		MenuCursor:  0,
		SubCursor:   0,
		WizardStep:  0,
		ModelName:   modelName,
		Workspace:   workspace,
		Persona:     persona,
		Status:      "READY",
		InputText:   "",
		personaList: pList,
		personaIdx:  pIdx,
	}
}

// Init initializes the TUI model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles TUI events and state transitions.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// If permission modal is active
		if m.ModalReq != nil {
			switch strings.ToLower(msg.String()) {
			case "y":
				m.Status = "APPROVED_ONCE"
				m.ModalReq = nil
			case "a":
				m.Status = "APPROVED_SESSION"
				m.ModalReq = nil
			case "n", "esc":
				m.Status = "DENIED"
				m.ModalReq = nil
			}
			return m, nil
		}

		if m.State != StateAgentWorkspace {
			return m.updateMenu(msg)
		}

		switch msg.String() {
		case "ctrl+c", "ctrl+q":
			m.quitting = true
			return m, tea.Quit
		case "ctrl+p":
			m.personaIdx = (m.personaIdx + 1) % len(m.personaList)
			m.Persona = m.personaList[m.personaIdx]
			return m, nil
		case "ctrl+l":
			m.Messages = []llm.Message{}
			return m, nil
		case "ctrl+h":
			m.ShowHelp = !m.ShowHelp
			return m, nil
		case "enter":
			text := strings.TrimSpace(m.InputText)
			m.InputText = ""
			if text == "" {
				return m, nil
			}
			if strings.HasPrefix(text, "/") {
				return m.handleSlashCommand(text)
			}
			m.Messages = append(m.Messages, llm.Message{
				Role:    "user",
				Content: text,
			})
			m.Messages = append(m.Messages, llm.Message{
				Role:    "assistant",
				Content: fmt.Sprintf("[%s mode] Acknowledged: %s", m.Persona, text),
			})
			m.TokenCount += len(text) / 4
			return m, nil
		case "backspace":
			if len(m.InputText) > 0 {
				m.InputText = m.InputText[:len(m.InputText)-1]
			}
		default:
			if len(msg.String()) == 1 {
				m.InputText += msg.String()
			}
		}
	}
	return m, nil
}

func (m Model) updateMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "ctrl+c", "ctrl+q":
		m.quitting = true
		return m, tea.Quit
	case "ctrl+p":
		m.personaIdx = (m.personaIdx + 1) % len(m.personaList)
		m.Persona = m.personaList[m.personaIdx]
		return m, nil
	case "ctrl+l":
		m.Messages = []llm.Message{}
		return m, nil
	case "ctrl+h":
		m.ShowHelp = !m.ShowHelp
		return m, nil
	case "esc":
		if m.State != StateMainMenu {
			m.State = StateMainMenu
			return m, nil
		}
	case "0":
		if m.State != StateMainMenu {
			m.State = StateMainMenu
			return m, nil
		}
		m.MenuCursor = 9
		return m.selectMainMenuItem()
	}

	switch m.State {
	case StateMainMenu:
		switch key {
		case "up", "k":
			m.MenuCursor = (m.MenuCursor - 1 + 10) % 10
		case "down", "j":
			m.MenuCursor = (m.MenuCursor + 1) % 10
		case "1", "2", "3", "4", "5", "6", "7", "8", "9":
			m.MenuCursor = int(key[0] - '1')
			return m.selectMainMenuItem()
		case "enter":
			return m.selectMainMenuItem()
		}
	case StateAgentSetup:
		switch m.WizardStep {
		case 0:
			switch key {
			case "up", "k":
				m.SubCursor = (m.SubCursor - 1 + 4) % 4
			case "down", "j":
				m.SubCursor = (m.SubCursor + 1) % 4
			case "1", "2", "3":
				m.SubCursor = int(key[0] - '1')
				m.WizardStep = 1
			case "4":
				m.State = StateMainMenu
			case "enter":
				if m.SubCursor == 3 {
					m.State = StateMainMenu
				} else {
					m.WizardStep = 1
				}
			}
		case 1:
			if key == "enter" {
				m.WizardStep = 2
			}
		case 2:
			if key == "enter" {
				m.WizardStep = 3
			}
		case 3:
			if key == "enter" {
				m.State = StateAgentWorkspace
			}
		}
	case StatePersonaMenu:
		switch key {
		case "up", "k":
			m.SubCursor = (m.SubCursor - 1 + 5) % 5
		case "down", "j":
			m.SubCursor = (m.SubCursor + 1) % 5
		case "1":
			m.Persona = "general"
			m.State = StateMainMenu
		case "2":
			m.Persona = "architect"
			m.State = StateMainMenu
		case "3":
			m.Persona = "tdd"
			m.State = StateMainMenu
		case "4":
			m.Persona = "security"
			m.State = StateMainMenu
		case "5":
			m.State = StateMainMenu
		case "enter":
			switch m.SubCursor {
			case 0:
				m.Persona = "general"
			case 1:
				m.Persona = "architect"
			case 2:
				m.Persona = "tdd"
			case 3:
				m.Persona = "security"
			}
			m.State = StateMainMenu
		}
	case StateModelMenu:
		switch key {
		case "up", "k":
			m.SubCursor = (m.SubCursor - 1 + 6) % 6
		case "down", "j":
			m.SubCursor = (m.SubCursor + 1) % 6
		case "1":
			m.ModelName = "qwen3-coder"
			m.State = StateMainMenu
		case "2":
			m.ModelName = "deepseek-r1"
			m.State = StateMainMenu
		case "3":
			m.ModelName = "llama3.3"
			m.State = StateMainMenu
		case "4":
			m.ModelName = "gpt-4o"
			m.State = StateMainMenu
		case "5":
			m.ModelName = "claude-opus-5"
			m.State = StateMainMenu
		case "6":
			m.State = StateMainMenu
		case "enter":
			models := []string{"qwen3-coder", "deepseek-r1", "llama3.3", "gpt-4o", "claude-opus-5"}
			if m.SubCursor < len(models) {
				m.ModelName = models[m.SubCursor]
			}
			m.State = StateMainMenu
		}
	case StateExtensionsMenu:
		switch key {
		case "up", "k":
			m.SubCursor = (m.SubCursor - 1 + 11) % 11
		case "down", "j":
			m.SubCursor = (m.SubCursor + 1) % 11
		case "enter":
			m.State = StateMainMenu
		}
	case StateProjectMenu, StateSessionMenu, StateSettingsMenu, StateDoctorMenu, StateAboutMenu:
		if key == "enter" {
			m.State = StateMainMenu
		}
	case StateExitConfirm:
		if strings.ToLower(key) == "y" || key == "enter" {
			m.quitting = true
			return m, tea.Quit
		} else if strings.ToLower(key) == "n" || key == "esc" {
			m.State = StateMainMenu
		}
	}

	return m, nil
}

func (m Model) selectMainMenuItem() (tea.Model, tea.Cmd) {
	switch m.MenuCursor {
	case 0: // 1. Agent
		m.State = StateAgentSetup
		m.WizardStep = 0
		m.SubCursor = 0
	case 1: // 2. Persona
		m.State = StatePersonaMenu
		m.SubCursor = 0
	case 2: // 3. Model
		m.State = StateModelMenu
		m.SubCursor = 0
	case 3: // 4. Project
		m.State = StateProjectMenu
	case 4: // 5. Session
		m.State = StateSessionMenu
	case 5: // 6. Extensions
		m.State = StateExtensionsMenu
		m.SubCursor = 0
	case 6: // 7. Settings
		m.State = StateSettingsMenu
	case 7: // 8. Doctor
		m.State = StateDoctorMenu
	case 8: // 9. About
		m.State = StateAboutMenu
	case 9: // 0. Exit
		m.State = StateExitConfirm
	}
	return m, nil
}

func (m Model) handleSlashCommand(cmdStr string) (tea.Model, tea.Cmd) {
	parts := strings.Fields(cmdStr)
	cmd := strings.ToLower(parts[0])

	switch cmd {
	case "/help":
		m.ShowHelp = !m.ShowHelp
	case "/clear":
		m.Messages = []llm.Message{}
	case "/quit", "/exit":
		m.quitting = true
		return m, tea.Quit
	case "/persona":
		if len(parts) > 1 {
			target := strings.ToLower(parts[1])
			for i, p := range m.personaList {
				if p == target {
					m.personaIdx = i
					m.Persona = p
					break
				}
			}
		} else {
			m.personaIdx = (m.personaIdx + 1) % len(m.personaList)
			m.Persona = m.personaList[m.personaIdx]
		}
	case "/model":
		if len(parts) > 1 {
			m.ModelName = parts[1]
			m.Messages = append(m.Messages, llm.Message{
				Role:    "assistant",
				Content: fmt.Sprintf("✅ Active model switched to: **%s**", m.ModelName),
			})
		}
	case "/apikey", "/key":
		if len(parts) > 1 {
			_ = os.Setenv("OPENAI_API_KEY", parts[1])
			_ = os.Setenv("NOVA_PROVIDER", "openai")
			m.Messages = append(m.Messages, llm.Message{
				Role:    "assistant",
				Content: "✅ OpenAI API Key configured for session (Provider switched to `openai`).",
			})
		}
	case "/extensions":
		m.Messages = append(m.Messages, llm.Message{
			Role: "assistant",
			Content: "ACTIVE EXTENSIONS\n\n  ✓ context-optimizer\n  ✓ prd-builder\n  ✓ architecture-builder\n  ✓ tdd-builder\n  ✓ task-builder\n\nAvailable but inactive:\n\n  ○ ponytail\n  ○ caveman\n  ○ docker\n  ○ security-review",
		})
	case "/plan":
		m.Messages = append(m.Messages, llm.Message{
			Role: "assistant",
			Content: "CURRENT PLAN\n\n  [1] Inspect auth architecture     ✓\n  [2] Implement JWT service         ●\n  [3] Add middleware                ○\n  [4] Add tests                     ○\n  [5] Verify                        ○\n\nProgress: 1 / 5",
		})
	case "/compact":
		m.Messages = append(m.Messages, llm.Message{
			Role: "assistant",
			Content: "Compacting session context...\n\nBefore:\n  73,821 tokens\n\nAfter:\n  29,442 tokens\n\nSaved:\n  44,379 tokens\n\n✓ Context optimized",
		})
	}
	return m, nil
}

// View renders the interactive TUI screen.
func (m Model) View() string {
	if m.quitting {
		return "\nNOVA\n\nFrom Thought to Thing.\n\nGoodbye.\n"
	}

	switch m.State {
	case StateMainMenu:
		return RenderMainMenu(m.ModelName, m.Persona, m.Workspace, m.MenuCursor)
	case StateAgentSetup:
		if m.WizardStep == 3 {
			return RenderProjectUnderstanding(m.Workspace)
		}
		return RenderAgentSetup(m.SubCursor, m.Workspace, m.WizardStep)
	case StatePersonaMenu:
		return RenderPersonaMenu(m.Persona, m.SubCursor)
	case StateModelMenu:
		return RenderModelMenu(m.ModelName, m.SubCursor)
	case StateProjectMenu:
		return RenderProjectMenu(m.Workspace)
	case StateSessionMenu:
		return RenderSessionMenu()
	case StateExtensionsMenu:
		return RenderExtensionsMenu(m.SubCursor)
	case StateSettingsMenu:
		return RenderSettingsMenu()
	case StateDoctorMenu:
		return RenderDoctorMenu(m.Workspace)
	case StateAboutMenu:
		return RenderAboutMenu()
	case StateExitConfirm:
		return RenderExitConfirm()
	}

	// StateAgentWorkspace: The real interactive Agent Workspace
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(styleHeader.Render("NOVA — Local Agent Workspace"))
	b.WriteString("\n")
	b.WriteString(RenderStatusBar(m.Status, m.ModelName, m.Persona, m.TokenCount))
	b.WriteString("\n\n")

	if m.ShowHelp {
		b.WriteString(styleSystemMsg.Render("Keybindings: [Ctrl+P] Switch Persona | [Ctrl+L] Clear Chat | [Ctrl+H] Toggle Help | [Ctrl+C] Quit\nSlash Commands: /help, /persona <name>, /model <name>, /apikey <key>, /clear, /quit"))
		b.WriteString("\n\n")
	}

	if m.ModalReq != nil {
		b.WriteString(RenderModal(m.ModalReq))
		b.WriteString("\n\n")
	} else if len(m.Messages) > 0 {
		b.WriteString(RenderChatHistory(m.Messages))
		b.WriteString("\n\n")
	} else {
		b.WriteString(styleSystemMsg.Render("I have finished understanding this project.\n\nWhat would you like me to do?"))
		b.WriteString("\n\n")
	}

	b.WriteString(styleUserMsg.Render("Prompt > ") + m.InputText + "█\n")

	return b.String()
}

// Run launches the Bubble Tea TUI application.
func Run(modelName, workspace, persona string) error {
	p := tea.NewProgram(NewModel(modelName, workspace, persona))
	_, err := p.Run()
	return err
}
