package main

import (
	"fmt"
	"os"
	"path/filepath"
	"seeds-registry-tui/models"
	"seeds-registry-tui/ui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type page int

const (
	pageMenu page = iota
	pageAdmin
	pageSeeds
	pageMatrix
)

type model struct {
	store        *models.Store
	page         page
	menuCursor   int
	admin        ui.AdminView
	seeds        ui.SeedsView
	matrix       ui.MatrixView
	width        int
	height       int
	quitting     bool
	exportMsg    string
	exportMsgErr bool
}

func initialModel(store *models.Store) model {
	return model{
		store:  store,
		page:   pageMenu,
		admin:  ui.NewAdminView(store),
		seeds:  ui.NewSeedsView(store),
		matrix: ui.NewMatrixView(store),
	}
}

func (m model) Init() tea.Cmd {
	return tea.SetWindowTitle("Seeds Registry TUI")
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.seeds.SetHeight(msg.Height)
		m.matrix.SetSize(msg.Width, msg.Height)
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || (msg.String() == "q" && m.page == pageMenu) {
			m.quitting = true
			return m, tea.Quit
		}
	}

	switch m.page {
	case pageMenu:
		return m.updateMenu(msg)
	case pageAdmin:
		return m.updateAdmin(msg)
	case pageSeeds:
		return m.updateSeeds(msg)
	case pageMatrix:
		return m.updateMatrix(msg)
	}
	return m, nil
}

func (m model) updateMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.menuCursor > 0 {
				m.menuCursor--
			}
		case "down", "j":
			if m.menuCursor < 4 {
				m.menuCursor++
			}
		case "enter":
			switch m.menuCursor {
			case 0:
				m.page = pageAdmin
				m.exportMsg = ""
			case 1:
				m.seeds.RefreshResults()
				m.page = pageSeeds
				m.exportMsg = ""
			case 2:
				m.page = pageMatrix
				m.exportMsg = ""
			case 3:
				path, err := m.store.ExportPrintable()
				if err != nil {
					m.exportMsg = "Export failed: " + err.Error()
					m.exportMsgErr = true
				} else {
					m.exportMsg = "Report saved to: " + path
					m.exportMsgErr = false
				}
			case 4:
				m.quitting = true
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m model) updateAdmin(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "esc" {
		a := m.admin
		if !a.IsAddMode() {
			m.page = pageMenu
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.admin, cmd = m.admin.Update(msg)
	return m, cmd
}

func (m model) updateSeeds(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "esc" {
		if !m.seeds.IsFormMode() && !m.seeds.IsSearching() && !m.seeds.IsPlantPicking() && !m.seeds.IsLotPicking() {
			m.page = pageMenu
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.seeds, cmd = m.seeds.Update(msg)
	return m, cmd
}

func (m model) updateMatrix(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "esc" {
		if !m.matrix.IsSearching() {
			m.page = pageMenu
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.matrix, cmd = m.matrix.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.quitting {
		return "Goodbye! 🌱\n"
	}

	var content string

	switch m.page {
	case pageMenu:
		content = m.menuView()
	case pageAdmin:
		content = m.admin.View()
	case pageSeeds:
		content = m.seeds.View()
	case pageMatrix:
		content = m.matrix.View()
	}

	minWidth := m.width * 90 / 100
	frame := lipgloss.NewStyle().
		Padding(1, 2).
		Width(minWidth).
		MaxWidth(m.width).
		MaxHeight(m.height)

	return frame.Render(content)
}

func (m model) menuView() string {
	title := ui.TitleStyle.Render(" 🌱 Seeds Registry ")

	items := []string{
		"Admin Panel     - Manage plants, lots & matrix sizes",
		"Seeds Registry  - Add, edit, search planted seeds",
		"Matrix View     - Visual lot representation",
		"Export Report   - Save printable report to file",
		"Quit",
	}

	var menu string
	for i, item := range items {
		if i == m.menuCursor {
			menu += ui.SelectedStyle.Render("▸ "+item) + "\n"
		} else {
			menu += ui.NormalStyle.Render("  "+item) + "\n"
		}
	}

	stats := m.statsView()

	var exportLine string
	if m.exportMsg != "" {
		if m.exportMsgErr {
			exportLine = "\n" + ui.ErrorStyle.Render("✗ "+m.exportMsg) + "\n"
		} else {
			exportLine = "\n" + ui.SuccessStyle.Render("✓ "+m.exportMsg) + "\n"
		}
	}

	help := ui.HelpStyle.Render("↑/↓: navigate • enter: select • q: quit")

	return fmt.Sprintf("%s\n\n%s\n%s%s\n%s", title, menu, stats, exportLine, help)
}

func (m model) statsView() string {
	plants := len(m.store.GetPlantNames())
	sizes := len(m.store.GetMatrixSizes())
	lots := len(m.store.GetLots())
	seeds := len(m.store.GetAllSeeds())

	statsStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ui.Muted).
		Padding(0, 2).
		Foreground(ui.Muted)

	return statsStyle.Render(fmt.Sprintf("Plants: %d  │  Sizes: %d  │  Lots: %d  │  Seeds: %d", plants, sizes, lots, seeds))
}

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	dataDir := filepath.Join(home, ".seeds-registry-tui")
	store, err := models.NewStore(dataDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing store: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(initialModel(store), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
