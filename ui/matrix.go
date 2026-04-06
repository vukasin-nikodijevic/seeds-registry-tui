package ui

import (
	"fmt"
	"seeds-registry-tui/models"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// MatrixView shows a lot list on the left and the selected lot's grid on the right.
type MatrixView struct {
	store       *models.Store
	lotCursor   int
	searchInput textinput.Model
	searching   bool
	query       string
	scrollX     int
	scrollY     int
	width       int
	height      int
}

func NewMatrixView(store *models.Store) MatrixView {
	si := textinput.New()
	si.Placeholder = "Search plant..."
	si.Width = 30

	return MatrixView{
		store:       store,
		searchInput: si,
	}
}

func (m MatrixView) Init() tea.Cmd {
	return nil
}

func (m MatrixView) IsSearching() bool {
	return m.searching
}

func (m MatrixView) Update(msg tea.Msg) (MatrixView, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	if m.searching {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				m.searching = false
				m.searchInput.Blur()
				return m, nil
			case "enter":
				m.query = m.searchInput.Value()
				m.searching = false
				m.searchInput.Blur()
				return m, nil
			}
		}
		var cmd tea.Cmd
		m.searchInput, cmd = m.searchInput.Update(msg)
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		lots := m.store.GetLots()
		switch msg.String() {
		case "up", "k":
			if m.lotCursor > 0 {
				m.lotCursor--
				m.scrollX = 0
				m.scrollY = 0
			}
		case "down", "j":
			if m.lotCursor < len(lots)-1 {
				m.lotCursor++
				m.scrollX = 0
				m.scrollY = 0
			}
		case "left", "h":
			if m.scrollX > 0 {
				m.scrollX--
			}
		case "right", "l":
			m.scrollX++
		case "ctrl+u":
			if m.scrollY > 0 {
				m.scrollY--
			}
		case "ctrl+d":
			m.scrollY++
		case "/":
			m.searching = true
			m.searchInput.Focus()
			return m, textinput.Blink
		case "c":
			m.query = ""
			m.searchInput.SetValue("")
		}
	}
	return m, nil
}

func (m MatrixView) View() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render(" 🌱 Matrix View "))
	b.WriteString("\n\n")

	lots := m.store.GetLots()
	if len(lots) == 0 {
		b.WriteString(MutedStyle.Render("No lots created. Add lots in Admin panel."))
		b.WriteString("\n\n")
		b.WriteString(HelpStyle.Render("esc: back"))
		return b.String()
	}

	if m.lotCursor >= len(lots) {
		m.lotCursor = len(lots) - 1
	}

	// ── Left panel: lot list ──
	leftPanel := m.renderLotList(lots)

	// ── Right panel: selected lot grid ──
	lot := lots[m.lotCursor]
	rightPanel := m.renderGrid(lot)

	// Join panels side by side
	layout := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, "  ", rightPanel)
	b.WriteString(layout)

	// Search bar
	b.WriteString("\n\n")
	if m.searching {
		b.WriteString("Search: ")
		b.WriteString(m.searchInput.View())
	} else if m.query != "" {
		b.WriteString(fmt.Sprintf("Filter: %s  ", lipgloss.NewStyle().Foreground(Secondary).Render(m.query)))
		b.WriteString(MutedStyle.Render("(c to clear)"))
	}

	// Help
	b.WriteString("\n\n")
	if m.searching {
		b.WriteString(HelpStyle.Render("type to search • enter: apply • esc: cancel"))
	} else {
		b.WriteString(HelpStyle.Render("↑↓: select lot • ←→: scroll cols • ctrl+u/d: scroll rows • /: search • c: clear • esc: back"))
	}

	return b.String()
}

// renderLotList renders the left panel with the list of lots.
func (m MatrixView) renderLotList(lots []models.Lot) string {
	listWidth := 28

	var items []string
	items = append(items, SubtitleStyle.Render("Lots"))
	items = append(items, MutedStyle.Render(strings.Repeat("─", listWidth-2)))

	for i, lot := range lots {
		seeds := m.store.GetSeedsInLot(lot.Name)
		totalCells := lot.Rows * lot.Columns
		filledCells := m.countFilled(lot, seeds)
		pct := 0.0
		if totalCells > 0 {
			pct = float64(filledCells) / float64(totalCells) * 100
		}

		label := fmt.Sprintf("%s %d×%d %d%%", lot.Name, lot.Rows, lot.Columns, int(pct))

		if i == m.lotCursor {
			items = append(items, SelectedStyle.Render(fmt.Sprintf("▸ %-*s", listWidth-4, label)))
		} else {
			items = append(items, NormalStyle.Render(fmt.Sprintf("  %-*s", listWidth-4, label)))
		}
	}

	listStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Primary).
		Width(listWidth).
		Padding(1, 1)

	return listStyle.Render(strings.Join(items, "\n"))
}

// renderGrid renders the right panel with the selected lot's matrix grid.
func (m MatrixView) renderGrid(lot models.Lot) string {
	var b strings.Builder

	b.WriteString(SubtitleStyle.Render(fmt.Sprintf("%s (%d×%d)", lot.Name, lot.Rows, lot.Columns)))
	if lot.MatrixSize != "" {
		b.WriteString(MutedStyle.Render(fmt.Sprintf("  size: %s", lot.MatrixSize)))
	}
	b.WriteString("\n\n")

	seeds := m.store.GetSeedsInLot(lot.Name)

	// Create grid
	grid := make([][]string, lot.Rows)
	for r := 0; r < lot.Rows; r++ {
		grid[r] = make([]string, lot.Columns)
	}
	for _, seed := range seeds {
		startR := seed.Row
		endR := seed.Row
		if seed.RowEnd > 0 && seed.RowEnd != seed.Row {
			endR = seed.RowEnd
		}
		for r := startR; r <= endR; r++ {
			cS := seed.ColStart
			cE := seed.ColEnd
			if seed.RowEnd > 0 && seed.RowEnd != seed.Row {
				// Multi-row: full columns
				cS = 1
				cE = lot.Columns
			}
			for c := cS; c <= cE; c++ {
				if r-1 >= 0 && r-1 < lot.Rows && c-1 >= 0 && c-1 < lot.Columns {
					grid[r-1][c-1] = seed.Plant
				}
			}
		}
	}

	// Highlighted rows from search
	highlightRows := make(map[int]bool)
	if m.query != "" {
		q := strings.ToLower(m.query)
		for _, seed := range seeds {
			if strings.Contains(strings.ToLower(seed.Plant), q) ||
				strings.Contains(strings.ToLower(seed.Description), q) {
				startR := seed.Row
				endR := seed.Row
				if seed.RowEnd > 0 && seed.RowEnd != seed.Row {
					endR = seed.RowEnd
				}
				for r := startR; r <= endR; r++ {
					highlightRows[r] = true
				}
			}
		}
	}

	// Scrolling
	maxVisibleCols := 6
	maxVisibleRows := 12
	startCol := m.scrollX
	startRow := m.scrollY
	if startCol >= lot.Columns {
		startCol = lot.Columns - 1
	}
	if startCol < 0 {
		startCol = 0
	}
	if startRow >= lot.Rows {
		startRow = lot.Rows - 1
	}
	if startRow < 0 {
		startRow = 0
	}
	endCol := startCol + maxVisibleCols
	if endCol > lot.Columns {
		endCol = lot.Columns
	}
	endRow := startRow + maxVisibleRows
	if endRow > lot.Rows {
		endRow = lot.Rows
	}

	// Column header
	colHeader := HeaderCellStyle.Render("Row\\Col")
	for c := startCol; c < endCol; c++ {
		colHeader += HeaderCellStyle.Render(fmt.Sprintf("C%d", c+1))
	}
	b.WriteString(colHeader)
	b.WriteString("\n")

	sepLen := (endCol - startCol + 1) * 12
	b.WriteString(MutedStyle.Render(strings.Repeat("─", sepLen)))
	b.WriteString("\n")

	// Grid rows
	for r := startRow; r < endRow; r++ {
		isHighlighted := highlightRows[r+1]
		b.WriteString(HeaderCellStyle.Render(fmt.Sprintf("R%d", r+1)))

		for c := startCol; c < endCol; c++ {
			cellValue := grid[r][c]
			if cellValue == "" {
				if isHighlighted {
					b.WriteString(HighlightCellStyle.Render("·"))
				} else {
					b.WriteString(EmptyCellStyle.Render("·"))
				}
			} else {
				display := cellValue
				if len(display) > 10 {
					display = display[:10]
				}
				if isHighlighted {
					b.WriteString(HighlightCellStyle.Render(display))
				} else if m.query != "" && strings.Contains(strings.ToLower(cellValue), strings.ToLower(m.query)) {
					b.WriteString(HighlightCellStyle.Render(display))
				} else {
					b.WriteString(FilledCellStyle.Render(display))
				}
			}
		}
		b.WriteString("\n")
	}

	// Scroll indicator
	if lot.Columns > maxVisibleCols || lot.Rows > maxVisibleRows {
		b.WriteString("\n")
		b.WriteString(MutedStyle.Render(fmt.Sprintf("rows %d-%d/%d  cols %d-%d/%d",
			startRow+1, endRow, lot.Rows, startCol+1, endCol, lot.Columns)))
	}

	// Stats
	b.WriteString("\n\n")
	totalCells := lot.Rows * lot.Columns
	filledCells := m.countFilled(lot, seeds)
	freeCells := totalCells - filledCells
	occupancy := 0.0
	if totalCells > 0 {
		occupancy = float64(filledCells) / float64(totalCells) * 100
	}

	statsStyle := lipgloss.NewStyle().Foreground(Primary).Bold(true)
	b.WriteString(statsStyle.Render(fmt.Sprintf("Total: %d  Planted: %d  Free: %d  (%.0f%%)",
		totalCells, filledCells, freeCells, occupancy)))

	return b.String()
}

// countFilled counts cells occupied by seeds in a lot.
func (m MatrixView) countFilled(lot models.Lot, seeds []models.SeedItem) int {
	grid := make([][]bool, lot.Rows)
	for r := 0; r < lot.Rows; r++ {
		grid[r] = make([]bool, lot.Columns)
	}
	for _, seed := range seeds {
		startR := seed.Row
		endR := seed.Row
		if seed.RowEnd > 0 && seed.RowEnd != seed.Row {
			endR = seed.RowEnd
		}
		for r := startR; r <= endR; r++ {
			cS := seed.ColStart
			cE := seed.ColEnd
			if seed.RowEnd > 0 && seed.RowEnd != seed.Row {
				cS = 1
				cE = lot.Columns
			}
			for c := cS; c <= cE; c++ {
				if r-1 >= 0 && r-1 < lot.Rows && c-1 >= 0 && c-1 < lot.Columns {
					grid[r-1][c-1] = true
				}
			}
		}
	}
	count := 0
	for r := 0; r < lot.Rows; r++ {
		for c := 0; c < lot.Columns; c++ {
			if grid[r][c] {
				count++
			}
		}
	}
	return count
}
