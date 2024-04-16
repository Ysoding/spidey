package cmd

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fatih/color"
)

type stepInfo struct {
	Placeholder string
	Text        string
}

var steps = []stepInfo{{"https:://google.com", "What is the website url address you want to check?"},
	{"false", "Have you enabled checking external addresses?"}}

type model struct {
	curStep   int
	textInput textinput.Model
}

var vals [2]string

func initialModel() model {
	ti := textinput.New()
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20
	ti.Placeholder = steps[0].Placeholder
	ti.TextStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA"))

	return model{
		textInput: ti,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	color.Set(color.FgCyan)

	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter", " ":

			m.curStep += 1

			if m.curStep == len(steps) {
				m.curStep = 0
				return m, cmd
			}
		}
	}

	m.textInput, cmd = m.textInput.Update(msg)
	m.textInput.Placeholder = steps[m.curStep].Placeholder
	vals[m.curStep] = m.textInput.View()

	return m, cmd
}

func (m model) View() string {

	if m.curStep == len(steps) {
		return "All done!"
	}

	return fmt.Sprintf(
		"%s\n\n%s",
		steps[m.curStep].Text,
		m.textInput.View(),
	) + "\n"
}
