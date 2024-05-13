package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type Events struct {
}

func (Events) Event(context interface{}, event string, format string, data ...interface{}) {
	// fmt.Printf("%s: %s\n", event, fmt.Sprintf(format, data...))
}

func (Events) ErrorEvent(context interface{}, event string, err error, format string, data ...interface{}) {

}

func Execute() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
