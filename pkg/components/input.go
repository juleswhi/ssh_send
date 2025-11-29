package components

import (
	"bytes"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
)

const correctPhrase = "open-sesame"
const script_path = ""

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

	// Password input
	passInput textinput.Model

	// Title + Body fields
	title textinput.Model
	body  textarea.Model

	// Focus management
	focus focusIndex

	// Script output
	output string
	err    error
}

func NewSecureInput() Model {
	pass := textinput.New()
	pass.Placeholder = "?"
	pass.Focus()
	pass.CharLimit = 128
	pass.Width = 32

	title := textinput.New()
	title.Placeholder = "title"
	title.Width = 40

	body := textarea.New()
	body.Placeholder = "body"
	body.SetWidth(80)
	body.SetHeight(20)

	return Model{
		stage:     stagePassword,
		passInput: pass,
		title:     title,
		body:      body,
		focus:     focusTitle,
	}
}

// INITIALIZE
func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				// Cycle focus
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
					// Submit pressed!
					out, err := runShellScript(m.title.Value(), m.body.Value())
					m.output = out
					m.err = err
					m.stage = stageDone
					return m, tea.Quit
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

// Manage focus behaviour
func (m *Model) updateFocus() (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch m.focus {
	case focusTitle:
		m.title.Focus()
		m.body.Blur()

	case focusBody:
		m.title.Blur()
		m.body.Focus()

	case focusSubmit:
		m.title.Blur()
		m.body.Blur()
	}

	return m, cmd
}

func (m Model) View() string {
	switch m.stage {

	case stagePassword:
		return m.passInput.View()

	case stageEditor:
		btn := "push"
		if m.focus == focusSubmit {
			btn = "push ?"
		}

		return (
				"Title:\n" +
				m.title.View() + "\n\n" +
				"Body:\n" +
				m.body.View() + "\n\n" +
				btn)

	case stageDone:
		msg := "\nScript executed.\n\nOutput:\n" + m.output
		if m.err != nil {
			msg += "\n\nError:\n" + m.err.Error()
		}
		return msg + "\n\nPress Ctrl+C to exit."
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

