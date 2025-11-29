package components

import (
	"bytes"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	submitStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FFCC"))

	submitFocusedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF00AA")).
				Bold(true)
)

type stage int

const (
	stagePassword stage = iota
	stageMenu
	stageEditor
	stageDone
)

type focusIndex int

const (
	focusTitle focusIndex = iota
	focusBody
	focusSubmit
	focusMax
)

type Model struct {
	stage stage

	passInput textinput.Model

	// Menu
	menuItems  []string
	menuCursor int

	// Form fields
	title textinput.Model
	body  textinput.Model

	focus focusIndex

	output string
	err    error

	terminalWidth  int
	terminalHeight int
}

func NewInput() Model {
	pass := textinput.New()
	pass.Placeholder = "?"
	pass.Focus()
	pass.Width = 32

	title := textinput.New()
	title.Placeholder = "title"
	title.Width = 50

	body := textinput.New()
	body.Placeholder = "body"
	body.Width = 50

	return Model{
		stage:     stagePassword,
		passInput: pass,
		title:     title,
		body:      body,
		focus:     focusTitle,

		menuItems: []string{
			"notify",
			"ssh key",
			"typing", // new option
		},
		menuCursor: 0,
	}
}

func (m Model) Init() tea.Cmd { return textinput.Blink }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.terminalWidth = msg.Width
		m.terminalHeight = msg.Height
		return m, nil
	}

	// Global quit (Ctrl+C)
	if km, ok := msg.(tea.KeyMsg); ok {
		if km.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}

	switch m.stage {

	// -----------------------------------
	// PASSWORD SCREEN
	// -----------------------------------
	case stagePassword:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.String() == "enter" {
				phrase := os.Getenv("CORRECT_PHRASE")
				if m.passInput.Value() == phrase {
					m.stage = stageMenu
					return m, nil
				}
				m.passInput.SetValue("")
				m.passInput.Placeholder = "!?"
			}
		}
		var cmd tea.Cmd
		m.passInput, cmd = m.passInput.Update(msg)
		return m, cmd

	// -----------------------------------
	// MENU SCREEN
	// -----------------------------------
	case stageMenu:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "up":
				if m.menuCursor > 0 {
					m.menuCursor--
				}
			case "down":
				if m.menuCursor < len(m.menuItems)-1 {
					m.menuCursor++
				}

			case "enter":
				switch m.menuCursor {

				case 0: // notify
					m.stage = stageEditor
					m.title.Focus()

				case 1: // ssh key
					key, err := readSSHKey()
					m.output = key
					m.err = err
					m.stage = stageDone

				case 2: // typing (launch thokr)
					err := runThokr()
					if err != nil {
						m.output = "Error running thokr: " + err.Error()
					} else {
						m.output = "Exited typing."
					}
					m.stage = stageDone
				}
				return m, nil
			}
		}
		return m, nil

	// -----------------------------------
	// EDITOR (notify form)
	// -----------------------------------
	case stageEditor:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {

			case "tab", "shift+tab":
				if msg.String() == "tab" {
					m.focus++
				} else {
					m.focus--
				}

				if m.focus < 0 {
					m.focus = focusMax - 1
				}
				if m.focus >= focusMax {
					m.focus = 0
				}

				return m.updateFocus()

			case "enter":
				if m.focus == focusSubmit {
					out, err := runShellScript(m.title.Value(), m.body.Value())
					m.output = out
					m.err = err
					m.stage = stageDone
					return m, nil
				}
			}
		}

		switch m.focus {
		case focusTitle:
			var cmd tea.Cmd
			m.title, cmd = m.title.Update(msg)
			return m, cmd

		case focusBody:
			var cmd tea.Cmd
			m.body, cmd = m.body.Update(msg)
			return m, cmd
		}

		return m, nil

	// -----------------------------------
	// DONE (output shown)
	// -----------------------------------
	case stageDone:
		return m, nil
	}

	return m, nil
}

func (m *Model) updateFocus() (tea.Model, tea.Cmd) {
	m.title.Blur()
	m.body.Blur()

	switch m.focus {
	case focusTitle:
		m.title.Focus()
	case focusBody:
		m.body.Focus()
	}

	return m, nil
}

func (m Model) View() string {
	switch m.stage {

	// -----------------------
	// PASSWORD
	// -----------------------
	case stagePassword:
		box := lipgloss.NewStyle().Width(40).Align(lipgloss.Center).Render(m.passInput.View())
		return lipgloss.Place(
			m.terminalWidth, m.terminalHeight,
			lipgloss.Center, lipgloss.Center,
			box,
		)

	// -----------------------
	// MENU
	// -----------------------
	case stageMenu:
		var items string
		for i, item := range m.menuItems {
			cursor := " "
			if i == m.menuCursor {
				cursor = ">"
			}
			items += cursor + " " + item + "\n"
		}

		return lipgloss.Place(
			m.terminalWidth, m.terminalHeight,
			lipgloss.Center, lipgloss.Center,
			items,
		)

	// -----------------------
	// EDITOR (notify screen)
	// -----------------------
	case stageEditor:
		submit := submitStyle.Render("->")
		if m.focus == focusSubmit {
			submit = submitFocusedStyle.Render("!")
		}

		ui := lipgloss.JoinVertical(lipgloss.Center,
			m.title.View(),
			"",
			m.body.View(),
			"",
			submit,
		)

		return lipgloss.Place(
			m.terminalWidth, m.terminalHeight,
			lipgloss.Center, lipgloss.Center,
			ui,
		)

	// -----------------------
	// DONE SCREEN
	// -----------------------
	case stageDone:
		out := m.output
		if m.err != nil {
			out = "Error:\n" + m.err.Error()
		}

		return lipgloss.Place(
			m.terminalWidth, m.terminalHeight,
			lipgloss.Center, lipgloss.Center,
			out+"\n\nPress Ctrl+C to exit.",
		)
	}

	return ""
}

// ---------------------------------------------------------
// SCRIPT EXECUTION
// ---------------------------------------------------------
func runShellScript(title string, body string) (string, error) {
	script_path := os.Getenv("SCRIPT_PATH")
	cmd := exec.Command(script_path, title, body)

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	return out.String(), err
}

// ---------------------------------------------------------
// SSH KEY READING
// ---------------------------------------------------------
func readSSHKey() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}

	keyPath := filepath.Join(usr.HomeDir, ".ssh", "id_ed25519.pub")

	data, err := os.ReadFile(keyPath)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// ---------------------------------------------------------
// RUN THOKR (typing TUI app)
// ---------------------------------------------------------
func runThokr() error {
	cmd := exec.Command("thokr") // assumes thokr is installed and in PATH
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

