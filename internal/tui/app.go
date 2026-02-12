package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mkrowiarz/mcp-symfony-stack/internal/core/commands"
)

type Model struct {
	focusedPane   int
	width         int
	height        int
	projectRoot   string
	projectName   string
	projectType   string
	projectStatus string
	worktrees     []worktreeInfo
	databases     []databaseInfo
	dumps         []dumpInfo
}

func NewModel() Model {
	return Model{
		focusedPane: 3,
		projectRoot: ".",
		projectName: "Loading...",
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.loadProject, m.loadWorktrees, m.loadDatabases, m.loadDumps)
}

func (m Model) loadProject() tea.Msg {
	info, err := commands.Info(m.projectRoot)
	if err != nil {
		return projectLoadedMsg{name: "Error", ptype: err.Error(), status: "✗"}
	}

	name := "Unknown"
	ptype := "generic"
	status := "✗"

	if info.ConfigSummary != nil {
		name = info.ConfigSummary.Name
		ptype = info.ConfigSummary.Type
	}
	if info.DockerComposeExists {
		status = "✓"
	}

	return projectLoadedMsg{name: name, ptype: ptype, status: status}
}

func (m Model) loadWorktrees() tea.Msg {
	result, err := commands.List(m.projectRoot)
	if err != nil {
		return worktreesLoadedMsg{worktrees: []worktreeInfo{}}
	}

	var wtis []worktreeInfo
	for _, wt := range result {
		wtis = append(wtis, worktreeInfo{
			branch: wt.Branch,
			path:   wt.Path,
			isMain: wt.IsMain,
		})
	}
	return worktreesLoadedMsg{worktrees: wtis}
}

func (m Model) loadDatabases() tea.Msg {
	result, err := commands.ListDBs(m.projectRoot)
	if err != nil {
		return databasesLoadedMsg{databases: []databaseInfo{}}
	}

	var dbis []databaseInfo
	for _, db := range result.Databases {
		dbis = append(dbis, databaseInfo{
			name:      db.Name,
			isDefault: db.IsDefault,
		})
	}
	return databasesLoadedMsg{databases: dbis}
}

func (m Model) loadDumps() tea.Msg {
	result, err := commands.ListDumps(m.projectRoot)
	if err != nil {
		return dumpsLoadedMsg{dumps: []dumpInfo{}}
	}

	var dis []dumpInfo
	for _, d := range result.Dumps {
		dis = append(dis, dumpInfo{
			name: d.Name,
			size: formatSize(d.Size),
			date: formatDate(d.Modified),
		})
	}
	return dumpsLoadedMsg{dumps: dis}
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func formatDate(modified string) string {
	t, err := time.Parse(time.RFC3339, modified)
	if err != nil {
		return modified
	}
	return t.Format("Jan 02 15:04")
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case projectLoadedMsg:
		m.projectName = msg.name
		m.projectType = msg.ptype
		m.projectStatus = msg.status
	case worktreesLoadedMsg:
		m.worktrees = msg.worktrees
	case databasesLoadedMsg:
		m.databases = msg.databases
	case dumpsLoadedMsg:
		m.dumps = msg.dumps
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

	infoContent := fmt.Sprintf("Project: %s\nType: %s\nCompose: %s", m.projectName, m.projectType, m.projectStatus)
	infoPane := m.renderPane("Info", infoContent, 1, leftWidth, paneHeight)

	wtContent := ""
	for _, wt := range m.worktrees {
		prefix := "  "
		if wt.isMain {
			prefix = "* "
		}
		wtContent += prefix + wt.branch + "\n"
	}
	if wtContent == "" {
		wtContent = "No worktrees"
	}
	worktreesPane := m.renderPane("Worktrees", wtContent, 2, leftWidth, paneHeight)

	dumpsContent := ""
	for _, d := range m.dumps {
		dumpsContent += fmt.Sprintf("%s (%s)\n", d.name, d.size)
	}
	if dumpsContent == "" {
		dumpsContent = "No dumps"
	}
	dumpsPane := m.renderPane("Dumps", dumpsContent, 4, leftWidth, paneHeight)

	dbContent := ""
	for _, db := range m.databases {
		prefix := "  "
		if db.isDefault {
			prefix = "* "
		}
		dbContent += prefix + db.name + "\n"
	}
	if dbContent == "" {
		dbContent = "No databases"
	}
	dbPane := m.renderPane("Databases", dbContent, 3, rightWidth, m.height-3)

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
