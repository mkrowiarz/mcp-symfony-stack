package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	focusedPane int
	width       int
	height      int
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
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
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
	if m.width == 0 {
		return "Loading..."
	}

	leftWidth := m.width * 30 / 100
	rightWidth := m.width - leftWidth - 3
	paneHeight := (m.height - 4) / 2

	infoPane := m.renderPane("Info", "Project: phoenix\nType: symfony", 1, leftWidth, paneHeight)
	worktreesPane := m.renderPane("Worktrees", "* main\n  feature/abc", 2, leftWidth, paneHeight)
	dumpsPane := m.renderPane("Dumps", "dump_01.sql\ndump_02.sql", 4, leftWidth, paneHeight)

	dbPane := m.renderPane("Databases", "mytower_eu (default)\nmytower_eu_test", 3, rightWidth, m.height-3)

	leftCol := lipgloss.JoinVertical(lipgloss.Top, infoPane, worktreesPane, dumpsPane)
	mainLayout := lipgloss.JoinHorizontal(lipgloss.Top, leftCol, " ", dbPane)

	statusBar := statusBarStyle.Width(m.width).Render("[1-4]pane [Tab]switch [r]efresh [q]uit")

	return lipgloss.JoinVertical(lipgloss.Top, mainLayout, statusBar)
}

func (m Model) renderPane(title, content string, paneNum, width, height int) string {
	style := paneStyle
	if m.focusedPane == paneNum {
		style = focusedPaneStyle
	}

	header := titleStyle.Render(title)
	body := lipgloss.NewStyle().Padding(1, 0).Render(content)

	pane := lipgloss.JoinVertical(lipgloss.Left, header, body)
	return style.Width(width).Height(height).Render(pane)
}

func Run() error {
	p := tea.NewProgram(NewModel(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
