package components

import (
	"bytes"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

const correctPhrase = "open-sesame"
const script_path = "./run.sh"


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

	return Model {
		stage:     stagePassword,
		passInput: pass,
		title:     title,
		body:      body,
		focus:     focusTitle,
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

    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
		case "ctrl+c":
            return m, tea.Quit
        }
    }

    switch m.stage {

    case stagePassword:
        switch msg := msg.(type) {
        case tea.KeyMsg:
            if msg.String() == "enter" {
                if m.passInput.Value() == correctPhrase {
                    m.stage = stageEditor
                    m.title.Focus()
                    return m, textinput.Blink
                }
                m.passInput.SetValue("")
                m.passInput.Placeholder = "!?"
            }
        }
        var cmd tea.Cmd
        m.passInput, cmd = m.passInput.Update(msg)
        return m, cmd

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

    case stageDone:
        return m, tea.Quit
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

	case stagePassword:
		box := lipgloss.NewStyle().Width(40).Align(lipgloss.Center).Render(m.passInput.View())
		return lipgloss.Place(
			m.terminalWidth, m.terminalHeight,
			lipgloss.Center, lipgloss.Center,
			box,
		)

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

	case stageDone:
		return lipgloss.Place(
			m.terminalWidth, m.terminalHeight,
			lipgloss.Center, lipgloss.Center,
			"Press Ctrl+C to exit.",
		)
	}

	return ""
}

func runShellScript(title string, body string) (string, error) {
	cmd := exec.Command(script_path, title, body)

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	return out.String(), err
}

