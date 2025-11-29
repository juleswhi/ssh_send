package components

import (
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbletea/input"
	"github.com/charmbracelet/bubbletea/viewport"
)

type Model struct {
	input     input.Model
	viewport  viewport.Model
	success   bool
}

const secretString = "secret123" // Replace this with your desired string

func CreateHome(name, location, jobTitle, description string) Model {
	inputModel := input.NewModel()
	inputModel.Placeholder = "Enter secret string"
	inputModel.Focus()

	return Model{
		input:    inputModel,
		viewport: viewport.New(80, 24),
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if m.input.Value() == secretString {
				// Execute the shell script when the correct string is entered
				cmd = executeShellScript()
				m.success = true
			}
		}
	}

	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	var status string
	if m.success {
		status = "Success! Shell script executed."
	} else {
		status = "Type the secret string and press Enter"
	}

	// Position the input box in the middle of the screen
	width, height := m.viewport.Dimensions()
	inputBox := m.input.View()

	return inputBox + "\n" + status
}

func executeShellScript() tea.Cmd {
	// Replace with your script path
	cmd := exec.Command("/path/to/your/script.sh")
	cmd.Stdout = tea.NewOutput()

	// Execute the script and return a command to run it
	err := cmd.Run()
	if err != nil {
		return tea.Cmd(func() tea.Msg {
			return "Error executing shell script: " + err.Error()
		})
	}

	return nil
}