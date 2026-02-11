# TUI Mode - Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build an interactive terminal interface for managing Docker Compose projects with keyboard-driven navigation.

**Architecture:** Bubble Tea (Elm architecture) with Lip Gloss styling. Master-detail layout with 4 panes: Info, Worktrees, Databases, Dumps. All state in single Model struct, updates via messages.

**Tech Stack:** Go, github.com/charmbracelet/bubbletea, github.com/charmbracelet/lipgloss, github.com/charmbracelet/bubbles

---

## Task 1: Add TUI Dependencies

**Files:**
- Modify: `go.mod`

**Step 1: Add bubbletea, lipgloss, bubbles dependencies**

```bash
cd .worktrees/tui-mode
go get github.com/charmbracelet/bubbletea
go get github.com/charmbracelet/lipgloss
go get github.com/charmbracelet/bubbles
```

**Step 2: Verify dependencies**

```bash
go mod tidy
go build ./...
```

Expected: No errors

**Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "chore: add Bubble Tea TUI dependencies"
```

---

## Task 2: Create TUI Package Structure

**Files:**
- Create: `internal/tui/app.go`
- Create: `internal/tui/messages.go`

**Step 1: Create app.go with basic Bubble Tea setup**

```go
package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	focusedPane int // 1-4: Info, Worktrees, Databases, Dumps
}

func NewModel() Model {
	return Model{
		focusedPane: 3, // Start with Databases focused
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
```

**Step 2: Create messages.go (placeholder for now)**

```go
package tui

// Message types will be added as needed
```

**Step 3: Verify build**

```bash
go build ./internal/tui
```

Expected: No errors

**Step 4: Commit**

```bash
git add internal/tui/
git commit -m "feat(tui): add basic Bubble Tea setup with navigation"
```

---

## Task 3: Update main.go to Route to TUI

**Files:**
- Modify: `cmd/pm/main.go`

**Step 1: Read current main.go**

```bash
cat cmd/pm/main.go
```

**Step 2: Update main.go to route to TUI by default**

```go
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/mkrowiarz/mcp-symfony-stack/internal/mcp"
	"github.com/mkrowiarz/mcp-symfony-stack/internal/tui"
)

func main() {
	mcpFlag := flag.Bool("mcp", false, "Run as MCP server")
	flag.Parse()

	if *mcpFlag {
		if err := mcp.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "MCP server error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Default: run TUI
	if err := tui.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "TUI error: %v\n", err)
		os.Exit(1)
	}
}
```

**Step 3: Build and test**

```bash
go build -o pm ./cmd/pm
./pm
```

Expected: TUI shows "TUI Mode - Press q to quit", q exits

**Step 4: Commit**

```bash
git add cmd/pm/main.go
git commit -m "feat: route to TUI by default, MCP with --mcp flag"
```

---

## Task 4: Add Lip Gloss Styles

**Files:**
- Create: `internal/tui/styles.go`

**Step 1: Create styles.go with base styling**

```go
package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	primaryColor   = lipgloss.Color("#7D56F4")
	secondaryColor = lipgloss.Color("#6C737D")
	successColor   = lipgloss.Color("#2ECC71")
	errorColor     = lipgloss.Color("#E74C3C")
	
	// Base styles
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			Padding(0, 1)
	
	paneStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(secondaryColor).
			Padding(0, 1)
	
	focusedPaneStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(primaryColor).
				Padding(0, 1)
	
	selectedItemStyle = lipgloss.NewStyle().
				Foreground(primaryColor).
				Bold(true)
	
	statusBarStyle = lipgloss.NewStyle().
			Background(secondaryColor).
			Foreground(lipgloss.Color("#FFFFFF")).
			Padding(0, 1)
)
```

**Step 2: Verify build**

```bash
go build ./internal/tui
```

**Step 3: Commit**

```bash
git add internal/tui/styles.go
git commit -m "feat(tui): add Lip Gloss styles for panes and UI elements"
```

---

## Task 5: Implement Master-Detail Layout

**Files:**
- Modify: `internal/tui/app.go`

**Step 1: Update Model to hold layout dimensions**

```go
package tui

import (
	"fmt"
	"strings"
	
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
	
	// Layout: left column (30%) | right column (70%)
	leftWidth := m.width * 30 / 100
	rightWidth := m.width - leftWidth - 3
	paneHeight := (m.height - 4) / 2 // Account for status bar and borders
	
	// Left column panes
	infoPane := m.renderPane("Info", "Project: phoenix\nType: symfony", 1, leftWidth, paneHeight)
	worktreesPane := m.renderPane("Worktrees", "* main\n  feature/abc", 2, leftWidth, paneHeight)
	dumpsPane := m.renderPane("Dumps", "dump_01.sql\ndump_02.sql", 4, leftWidth, paneHeight)
	
	// Right column (databases)
	dbPane := m.renderPane("Databases", "mytower_eu (default)\nmytower_eu_test", 3, rightWidth, m.height-3)
	
	// Left column
	leftCol := lipgloss.JoinVertical(lipgloss.Top, infoPane, worktreesPane, dumpsPane)
	
	// Main layout
	mainLayout := lipgloss.JoinHorizontal(lipgloss.Top, leftCol, " ", dbPane)
	
	// Status bar
	statusBar := statusBarStyle.Width(m.width).Render(
		"[1-4]pane [Tab]switch [r]efresh [q]uit",
	)
	
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
```

**Step 2: Build and test**

```bash
go build -o pm ./cmd/pm
./pm
```

Expected: Master-detail layout visible, Tab/1-4 switches focus, q quits

**Step 3: Commit**

```bash
git add internal/tui/app.go
git commit -m "feat(tui): implement master-detail layout with 4 panes"
```

---

## Task 6: Load Real Data - Project Info

**Files:**
- Modify: `internal/tui/app.go`
- Create: `internal/tui/messages.go`

**Step 1: Update messages.go with data loading messages**

```go
package tui

type projectLoadedMsg struct {
	name   string
	ptype  string
	status string
}
```

**Step 2: Update Model to store project info**

```go
// Add to Model struct in app.go
type Model struct {
	focusedPane int
	width       int
	height      int
	projectRoot string
	
	// Project info
	projectName   string
	projectType   string
	projectStatus string
}

// Update NewModel
func NewModel() Model {
	return Model{
		focusedPane:   3,
		projectRoot:   ".",
		projectName:   "Loading...",
		projectType:   "",
		projectStatus: "",
	}
}
```

**Step 3: Add Init command to load project**

```go
import (
	"github.com/mkrowiarz/mcp-symfony-stack/internal/core/commands"
)

func (m Model) Init() tea.Cmd {
	return m.loadProject
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

// Add to Update:
case projectLoadedMsg:
	m.projectName = msg.name
	m.projectType = msg.ptype
	m.projectStatus = msg.status
```

**Step 4: Update View to use real data**

```go
// In View(), update infoPane:
infoContent := fmt.Sprintf("Project: %s\nType: %s\nCompose: %s", 
	m.projectName, m.projectType, m.projectStatus)
infoPane := m.renderPane("Info", infoContent, 1, leftWidth, paneHeight)
```

**Step 5: Build and test**

```bash
go build -o pm ./cmd/pm
./pm
```

Expected: Project info pane shows real project data

**Step 6: Commit**

```bash
git add internal/tui/
git commit -m "feat(tui): load and display real project info"
```

---

## Task 7: Load Real Data - Worktrees

**Files:**
- Modify: `internal/tui/app.go`
- Modify: `internal/tui/messages.go`

**Step 1: Add worktree message and state**

```go
// In messages.go
type worktreesLoadedMsg struct {
	worktrees []worktreeInfo
}

type worktreeInfo struct {
	branch string
	path   string
	isMain bool
}
```

**Step 2: Add to Model**

```go
type Model struct {
	// ...existing fields...
	worktrees []worktreeInfo
}
```

**Step 3: Add worktree loading**

```go
// In Init, return tea.Batch to run multiple commands
func (m Model) Init() tea.Cmd {
	return tea.Batch(m.loadProject, m.loadWorktrees)
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

// In Update
case worktreesLoadedMsg:
	m.worktrees = msg.worktrees
```

**Step 4: Update View**

```go
// In View()
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
```

**Step 5: Build and test**

```bash
go build -o pm ./cmd/pm
./pm
```

**Step 6: Commit**

```bash
git add internal/tui/
git commit -m "feat(tui): load and display worktrees list"
```

---

## Task 8: Load Real Data - Databases

**Files:**
- Modify: `internal/tui/app.go`
- Modify: `internal/tui/messages.go`

**Step 1: Add database message and state**

```go
// In messages.go
type databasesLoadedMsg struct {
	databases []databaseInfo
}

type databaseInfo struct {
	name      string
	isDefault bool
}
```

**Step 2: Add to Model and load function (similar to worktrees)**

```go
databases []databaseInfo

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
```

**Step 3: Update View to show databases**

**Step 4: Build, test, commit**

---

## Task 9: Load Real Data - Dumps

**Files:**
- Modify: `internal/tui/app.go`
- Modify: `internal/tui/messages.go`

**Step 1: Add dump message and state (similar pattern)**

**Step 2: Load dumps using commands.ListDumps**

**Step 3: Update View**

**Step 4: Build, test, commit**

---

## Task 10: Add Refresh Commands

**Files:**
- Modify: `internal/tui/app.go`

**Step 1: Handle 'r' and 'R' keys**

```go
case "r":
	return m, m.refreshCurrentPane()
case "R":
	return m, tea.Batch(m.loadProject, m.loadWorktrees, m.loadDatabases, m.loadDumps)
```

**Step 2: Add refreshCurrentPane function**

```go
func (m Model) refreshCurrentPane() tea.Cmd {
	switch m.focusedPane {
	case 1:
		return m.loadProject
	case 2:
		return m.loadWorktrees
	case 3:
		return m.loadDatabases
	case 4:
		return m.loadDumps
	}
	return nil
}
```

**Step 3: Build, test, commit**

---

## Task 11: Add Pane Selection with Arrows

**Files:**
- Modify: `internal/tui/app.go`

**Step 1: Add selectedIndex to Model**

```go
selectedIndex map[int]int // pane -> selected item index
```

**Step 2: Handle up/down arrows**

```go
case "up", "k":
	if idx, ok := m.selectedIndex[m.focusedPane]; ok {
		m.selectedIndex[m.focusedPane] = max(0, idx-1)
	}
case "down", "j":
	// Increment with bounds check based on list length
```

**Step 3: Highlight selected item in View**

**Step 4: Build, test, commit**

---

## Task 12: Add Status Bar with Contextual Shortcuts

**Files:**
- Modify: `internal/tui/app.go`

**Step 1: Create contextual status bar**

```go
func (m Model) statusBarText() string {
	switch m.focusedPane {
	case 2: // Worktrees
		return "[n]ew [r]emove [o]pen [Tab]switch [q]uit"
	case 3: // Databases
		return "[d]ump [c]lone [x]drop [Tab]switch [q]uit"
	case 4: // Dumps
		return "[i]mport [x]delete [Tab]switch [q]uit"
	default:
		return "[Tab]switch [r]efresh [q]uit"
	}
}
```

**Step 2: Update View to use contextual status**

**Step 3: Build, test, commit**

---

## Task 13: Add Progress Modal for Operations

**Files:**
- Create: `internal/tui/components/modal.go`
- Modify: `internal/tui/app.go`

**Step 1: Create modal component**

```go
package components

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ProgressModal struct {
	title   string
	message string
}

func NewProgressModal(title, message string) ProgressModal {
	return ProgressModal{title: title, message: message}
}

func (m ProgressModal) View(width, height int) string {
	// Center modal
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(1, 2)
	
	content := style.Render(m.title + "\n\n" + m.message)
	
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
}
```

**Step 2: Add modal state to Model**

```go
type Model struct {
	// ...existing...
	showModal   bool
	modalTitle  string
	modalMessage string
}
```

**Step 3: Show modal in View when showModal is true**

**Step 4: Build, test, commit**

---

## Task 14: Implement Dump Operation

**Files:**
- Modify: `internal/tui/app.go`
- Modify: `internal/tui/messages.go`

**Step 1: Add dump command and messages**

```go
type dumpStartedMsg struct {
	dbName string
}
type dumpFinishedMsg struct {
	result *types.DumpResult
	err    error
}
```

**Step 2: Handle 'd' key in Update**

```go
case "d":
	if m.focusedPane == 3 && len(m.databases) > 0 {
		db := m.databases[m.selectedIndex[3]]
		m.showModal = true
		m.modalTitle = "Dumping database..."
		m.modalMessage = db.name
		return m, m.dumpDatabase(db.name)
	}
```

**Step 3: Implement dumpDatabase command**

```go
func (m Model) dumpDatabase(dbName string) tea.Cmd {
	return func() tea.Msg {
		result, err := commands.Dump(m.projectRoot, dbName, nil)
		return dumpFinishedMsg{result: result, err: err}
	}
}
```

**Step 4: Handle completion**

**Step 5: Build, test, commit**

---

## Task 15: Implement Clone Operation

**Files:**
- Modify: `internal/tui/app.go`

**Step 1: Add input mode for clone target name**

**Step 2: Handle 'c' key to enter input mode**

**Step 3: Execute clone on Enter**

**Step 4: Build, test, commit**

---

## Task 16: Implement Drop Operation with Confirmation

**Files:**
- Modify: `internal/tui/app.go`

**Step 1: Add confirmation state**

```go
confirmingDrop bool
pendingDropDB  string
```

**Step 2: Handle 'x' key**

```go
case "x":
	if m.focusedPane == 3 {
		// Show confirmation prompt
		m.confirmingDrop = true
		m.pendingDropDB = m.databases[m.selectedIndex[3]].name
	}
```

**Step 3: Handle y/n for confirmation**

**Step 4: Execute drop on 'y'**

**Step 5: Build, test, commit**

---

## Task 17: Implement Worktree Operations

**Files:**
- Modify: `internal/tui/app.go`

**Step 1: Implement 'n' for new worktree (with input)**

**Step 2: Implement 'r' for remove worktree (with confirmation)**

**Step 3: Implement 'o' for open in terminal**

**Step 4: Build, test, commit**

---

## Task 18: Implement Dump Import Operation

**Files:**
- Modify: `internal/tui/app.go`

**Step 1: Handle 'i' from dumps pane**

**Step 2: Show database selector**

**Step 3: Execute import**

**Step 4: Build, test, commit**

---

## Task 19: Add Error Modal

**Files:**
- Modify: `internal/tui/app.go`
- Modify: `internal/tui/components/modal.go`

**Step 1: Add error modal state**

**Step 2: Show error modal on operation failure**

**Step 3: Dismiss on any key**

**Step 4: Build, test, commit**

---

## Task 20: Add Help Overlay

**Files:**
- Create: `internal/tui/components/help.go`
- Modify: `internal/tui/app.go`

**Step 1: Create help overlay component**

**Step 2: Show on '?' key**

**Step 3: Dismiss on any key**

**Step 4: Build, test, commit**

---

## Task 21: Final Testing & Polish

**Files:**
- All TUI files

**Step 1: Run full test suite**

```bash
go test ./...
```

**Step 2: Manual testing checklist:**
- [ ] All panes load data correctly
- [ ] Tab/1-4 navigation works
- [ ] Arrow navigation within panes
- [ ] Dump operation with progress
- [ ] Clone operation with input
- [ ] Drop with confirmation
- [ ] Worktree create/remove
- [ ] Error handling
- [ ] Help overlay

**Step 3: Fix any issues**

**Step 4: Final commit**

```bash
git add -A
git commit -m "feat(tui): complete TUI mode implementation"
```

---

## Success Criteria

- ✅ `./pm` launches TUI
- ✅ All 4 panes display real data
- ✅ Keyboard navigation (Tab, 1-4, arrows)
- ✅ Operations (dump, clone, drop) work with progress feedback
- ✅ Confirmation for destructive actions
- ✅ Error modal shows on failure
- ✅ Help overlay on '?'
- ✅ All tests passing
