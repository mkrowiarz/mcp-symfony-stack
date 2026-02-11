package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	focusedPane int
}

func NewModel() Model {
	return Model{
		focusedPane: 3,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab":
			m.focusedPane = m.focusedPane%4 + 1
		case "1", "2", "3", "4":
			m.focusedPane = int(msg.String()[0] - '0')
		}
	}
	return m, nil
}

func (m Model) View() string {
	return "TUI Mode - Press q to quit"
}

func Run() error {
	p := tea.NewProgram(NewModel())
	_, err := p.Run()
	return err
}
