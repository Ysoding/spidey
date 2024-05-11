package cmd

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fatih/color"
	"github.com/ysoding/spidey/spidey"
)

const context = "Spidey"

var msg = "checking~~~"

type stepInfo struct {
	Placeholder string
	Text        string
}

var steps = []stepInfo{{spidey.DEFAULT_TARGET_URL, "What is the website url address you want to check?"},
	{"false", "Have you enabled checking external addresses?"}}

type model struct {
	curStep   int
	textInput textinput.Model
	done      bool
	started   bool
	showText  string
}

type doingSpideyMsg struct{}

func doSpidey() tea.Cmd {
	return func() tea.Msg {
		return doingSpideyMsg{}
	}
}

type tickMsg struct {
}

var cnf spidey.Config
var spideyResult string
var startOnce sync.Once

var events Events

type Events struct {
}

func (Events) Event(context interface{}, event string, format string, data ...interface{}) {}

func (Events) ErrorEvent(context interface{}, event string, err error, format string, data ...interface{}) {

}

func init() {
	cnf = spidey.NewDefaultConfig(events)
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
	case tickMsg:
		if len(m.showText) < len(spideyResult) {
			m.showText += string(spideyResult[len(m.showText)])
			return m, tea.Tick(time.Millisecond, func(time.Time) tea.Msg {
				return tickMsg{}
			})
		}
		return m, nil

	case doingSpideyMsg:
		startOnce.Do(func() {
			sr, err := spidey.Run(context, &cnf)
			if err != nil {
				//  TODO:
			}
			spideyResult = sr.ResultFormat()
		})

		if spideyResult != "" {
			m.started = false
			m.done = true
			return m, tea.Tick(time.Millisecond, func(time.Time) tea.Msg {
				return tickMsg{}
			})
		}

		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter", " ":
			if m.started {
				return m, doSpidey()
			}

			if m.done {
				return m, tea.Quit
			}

			if m.curStep == 0 && strings.TrimSpace(m.textInput.Value()) != "" {
				cnf.URL = m.textInput.Value()
			} else if m.curStep == 1 && strings.Contains(m.textInput.Value(), "y") {
				cnf.EnableCheckExternal = true
			}

			m.textInput.Reset()

			m.curStep += 1

			if m.curStep == len(steps) {
				m.started = true
				return m, doSpidey()
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

	if m.done {
		// TODO: print result
		return fmt.Sprintf("%s", m.showText)
		// return fmt.Sprintf(
		// 	"😭 %s\n\n%s\n\n%s\n",
		// 	"All Done!",
		// 	m.showText,
		// 	quitMsg,
		// )
	}

	if m.started {
		// TODO: print progress bar
		return fmt.Sprintf("%s\n\n%s\n", msg, quitMsg)
	}

	return fmt.Sprintf(
		"%s\n\n%s\n\n%s\n",
		steps[m.curStep].Text,
		m.textInput.View(),
		quitMsg,
	)
}
