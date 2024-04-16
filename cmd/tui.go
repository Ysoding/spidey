package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fatih/color"
	"github.com/ysoding/spidey/config"
)

type stepInfo struct {
	Placeholder string
	Text        string
}

var steps = []stepInfo{{config.DEFALT_TARGET_URL, "What is the website url address you want to check?"},
	{"false", "Have you enabled checking external addresses?"}}

type model struct {
	curStep   int
	textInput textinput.Model
	done      bool
	started   bool
}

type startSpideyMsg struct{}

func startSpidey() tea.Cmd {
	return func() tea.Msg {
		return startSpideyMsg{}
	}
}

var cnf config.SpideyConfig

func init() {
	cnf = config.NewDefaultConfig()
}

func initialModel() model {
	ti := textinput.New()
	ti.Focus()
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

	case startSpideyMsg:

		m.done = true
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter", " ":
			if m.started || m.done {
				return m, tea.Quit
			}

			if m.curStep == 0 && strings.TrimSpace(m.textInput.Value()) != "" {
				cnf.TargetURL = m.textInput.Value()
			} else if m.curStep == 1 && strings.Contains(m.textInput.Value(), "y") {
				cnf.EnableCheckExternal = true
			}

			m.textInput.Reset()

			m.curStep += 1

			if m.curStep == len(steps) {
				m.started = true
				return m, startSpidey()
			}
			fallthrough
		default:
			m.textInput, cmd = m.textInput.Update(msg)
			m.textInput.Placeholder = steps[m.curStep].Placeholder
			return m, cmd
		}
	}
	return m, cmd
}

func (m model) View() string {
	quitMsg := "(esc or ctrl-c to quit)"

	if m.started || m.done {
		return fmt.Sprintf(
			"😭 %s\n\n%s\n",
			"All Done!",
			quitMsg,
		)
	}

	return fmt.Sprintf(
		"%s\n\n%s\n\n%s\n",
		steps[m.curStep].Text,
		m.textInput.View(),
		quitMsg,
	)
}
