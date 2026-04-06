package ui

import (
	"fmt"
	"seeds-registry-tui/models"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type seedMode int

const (
	seedModeList seedMode = iota
	seedModeAdd
	seedModeEdit
	seedModeView
)

// SeedsView handles CRUD for seeds with search.
type SeedsView struct {
	store       *models.Store
	mode        seedMode
	cursor      int
	searchInput textinput.Model
	searching   bool
	query       string
	results     []models.SeedItem

	// Form fields for add/edit
	formInputs   []textinput.Model
	formFocusIdx int
	formType     string // "matrix" or "single"
	editID       string

	// Plant picker state
	plantSearch    textinput.Model
	plantFiltered  []string
	plantCursor    int
	plantPicking   bool
	plantSelected  string

	// Lot picker state
	lotSearch    textinput.Model
	lotFiltered  []models.Lot
	lotCursor    int
	lotPicking   bool
	lotSelected  string // lot name

	message     string
	messageType string
}

func NewSeedsView(store *models.Store) SeedsView {
	si := textinput.New()
	si.Placeholder = "Search by plant, description, index..."
	si.Width = 40

	sv := SeedsView{
		store:       store,
		searchInput: si,
		results:     store.GetAllSeeds(),
	}
	return sv
}

func (s SeedsView) Init() tea.Cmd {
	return nil
}

func (s SeedsView) Update(msg tea.Msg) (SeedsView, tea.Cmd) {
	switch s.mode {
	case seedModeAdd, seedModeEdit:
		return s.updateForm(msg)
	default:
		return s.updateList(msg)
	}
}

func (s SeedsView) updateList(msg tea.Msg) (SeedsView, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if s.searching {
			switch msg.String() {
			case "esc":
				s.searching = false
				s.searchInput.Blur()
				return s, nil
			case "enter":
				s.query = s.searchInput.Value()
				s.results = s.store.SearchSeeds(s.query)
				s.cursor = 0
				s.searching = false
				s.searchInput.Blur()
				return s, nil
			}
			var cmd tea.Cmd
			s.searchInput, cmd = s.searchInput.Update(msg)
			return s, cmd
		}

		switch msg.String() {
		case "up", "k":
			if s.cursor > 0 {
				s.cursor--
			}
		case "down", "j":
			if s.cursor < len(s.results)-1 {
				s.cursor++
			}
		case "/":
			s.searching = true
			s.searchInput.Focus()
			return s, textinput.Blink
		case "a":
			plants := s.store.GetPlantNames()
			if len(plants) == 0 {
				s.message = "No plants defined. Add plants in Admin first."
				s.messageType = "error"
				return s, nil
			}
			s.mode = seedModeAdd
			s.formType = "single"
			s.message = ""
			s.initForm()
			return s, textinput.Blink
		case "e":
			if s.cursor < len(s.results) {
				seed := s.results[s.cursor]
				s.mode = seedModeEdit
				s.editID = seed.ID
				s.message = ""
				s.initEditForm(seed)
				return s, textinput.Blink
			}
		case "d":
			if s.cursor < len(s.results) {
				seed := s.results[s.cursor]
				if err := s.store.RemoveSeed(seed.ID); err != nil {
					s.message = err.Error()
					s.messageType = "error"
				} else {
					s.message = fmt.Sprintf("Seed '%s' removed", seed.Index())
					s.messageType = "success"
					s.results = s.store.SearchSeeds(s.query)
					if s.cursor >= len(s.results) && s.cursor > 0 {
						s.cursor--
					}
				}
			}
		case "c":
			// Clear search
			s.query = ""
			s.searchInput.SetValue("")
			s.results = s.store.GetAllSeeds()
			s.cursor = 0
		}
	}
	return s, nil
}

func (s *SeedsView) initPlantPicker(preselect string) {
	ps := textinput.New()
	ps.Placeholder = "Type to search plants..."
	ps.CharLimit = 50
	ps.Width = 30
	s.plantSearch = ps
	s.plantSelected = preselect
	s.plantPicking = false
	s.plantCursor = 0
	s.plantFiltered = s.store.GetPlantNames()
}

func (s *SeedsView) initLotPicker(preselect string) {
	ls := textinput.New()
	ls.Placeholder = "Type to search lots..."
	ls.CharLimit = 50
	ls.Width = 30
	s.lotSearch = ls
	s.lotSelected = preselect
	s.lotPicking = false
	s.lotCursor = 0
	s.lotFiltered = s.store.GetLots()
}

func (s *SeedsView) initForm() {
	desc := textinput.New()
	desc.Placeholder = "Description (e.g. cherry tomato)"
	desc.CharLimit = 100
	desc.Width = 40
	desc.Focus()

	s.formInputs = []textinput.Model{desc}
	s.formFocusIdx = 0
	s.initPlantPicker("")
	s.initLotPicker("")

	if s.formType == "matrix" {
		row := textinput.New()
		row.Placeholder = "Row start"
		row.CharLimit = 5
		row.Width = 20

		rowEnd := textinput.New()
		rowEnd.Placeholder = "Row end (optional, fills full cols)"
		rowEnd.CharLimit = 5
		rowEnd.Width = 30

		colStart := textinput.New()
		colStart.Placeholder = "Column start"
		colStart.CharLimit = 5
		colStart.Width = 20

		colEnd := textinput.New()
		colEnd.Placeholder = "Column end"
		colEnd.CharLimit = 5
		colEnd.Width = 20

		s.formInputs = append(s.formInputs, row, rowEnd, colStart, colEnd)
	}
}

func (s *SeedsView) initEditForm(seed models.SeedItem) {
	desc := textinput.New()
	desc.SetValue(seed.Description)
	desc.CharLimit = 100
	desc.Width = 40
	desc.Focus()

	s.formInputs = []textinput.Model{desc}
	s.formFocusIdx = 0
	s.formType = seed.Type
	s.initPlantPicker(seed.Plant)
	s.initLotPicker(seed.LotName)

	if seed.Type == "matrix" {
		row := textinput.New()
		row.SetValue(fmt.Sprintf("%d", seed.Row))
		row.CharLimit = 5
		row.Width = 20

		rowEnd := textinput.New()
		if seed.RowEnd > 0 && seed.RowEnd != seed.Row {
			rowEnd.SetValue(fmt.Sprintf("%d", seed.RowEnd))
		}
		rowEnd.Placeholder = "Row end (optional)"
		rowEnd.CharLimit = 5
		rowEnd.Width = 30

		colStart := textinput.New()
		colStart.SetValue(fmt.Sprintf("%d", seed.ColStart))
		colStart.CharLimit = 5
		colStart.Width = 20

		colEnd := textinput.New()
		colEnd.SetValue(fmt.Sprintf("%d", seed.ColEnd))
		colEnd.CharLimit = 5
		colEnd.Width = 20

		s.formInputs = append(s.formInputs, row, rowEnd, colStart, colEnd)
	}
}

func (s SeedsView) updateForm(msg tea.Msg) (SeedsView, tea.Cmd) {
	// Plant picker is active — handle it first
	if s.plantPicking {
		return s.updatePlantPicker(msg)
	}
	// Lot picker is active — handle it first
	if s.lotPicking {
		return s.updateLotPicker(msg)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			s.mode = seedModeList
			s.formInputs = nil
			return s, nil
		case "ctrl+t":
			// Toggle type in add mode
			if s.mode == seedModeAdd {
				if s.formType == "single" {
					lots := s.store.GetLots()
					if len(lots) == 0 {
						s.message = "Cannot create matrix seed: no lots defined. Add lots in Admin first."
						s.messageType = "error"
						return s, nil
					}
					s.formType = "matrix"
				} else {
					s.formType = "single"
				}
				s.initForm()
				// Preserve selected plant
				return s, textinput.Blink
			}
		case "ctrl+p":
			// Open plant picker
			s.plantPicking = true
			s.plantSearch.SetValue("")
			s.plantSearch.Focus()
			s.plantFiltered = s.store.GetPlantNames()
			s.plantCursor = 0
			if s.formFocusIdx < len(s.formInputs) {
				s.formInputs[s.formFocusIdx].Blur()
			}
			return s, textinput.Blink
		case "ctrl+l":
			// Open lot picker (matrix only)
			if s.formType == "matrix" && s.mode == seedModeAdd {
				s.lotPicking = true
				s.lotSearch.SetValue("")
				s.lotSearch.Focus()
				s.lotFiltered = s.store.GetLots()
				s.lotCursor = 0
				if s.formFocusIdx < len(s.formInputs) {
					s.formInputs[s.formFocusIdx].Blur()
				}
				return s, textinput.Blink
			}
		case "tab", "shift+tab":
			dir := 1
			if msg.String() == "shift+tab" {
				dir = -1
			}
			// Advance, skipping disabled col fields when row end is filled
			for {
				s.formFocusIdx += dir
				if s.formFocusIdx >= len(s.formInputs) {
					s.formFocusIdx = 0
				}
				if s.formFocusIdx < 0 {
					s.formFocusIdx = len(s.formInputs) - 1
				}
				// Skip col start (idx 3) and col end (idx 4) when row end is filled
				if s.formType == "matrix" && len(s.formInputs) > 4 &&
					strings.TrimSpace(s.formInputs[2].Value()) != "" &&
					(s.formFocusIdx == 3 || s.formFocusIdx == 4) {
					continue
				}
				break
			}
			for i := range s.formInputs {
				if i == s.formFocusIdx {
					s.formInputs[i].Focus()
				} else {
					s.formInputs[i].Blur()
				}
			}
			return s, textinput.Blink
		case "enter":
			s.submitForm()
			if s.messageType == "success" {
				s.mode = seedModeList
				s.formInputs = nil
				s.results = s.store.SearchSeeds(s.query)
			}
			return s, nil
		}
	}

	var cmd tea.Cmd
	if s.formFocusIdx < len(s.formInputs) {
		s.formInputs[s.formFocusIdx], cmd = s.formInputs[s.formFocusIdx].Update(msg)
	}
	return s, cmd
}

func (s SeedsView) updatePlantPicker(msg tea.Msg) (SeedsView, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			s.plantPicking = false
			s.plantSearch.Blur()
			// Re-focus form input
			if s.formFocusIdx < len(s.formInputs) {
				s.formInputs[s.formFocusIdx].Focus()
			}
			return s, textinput.Blink
		case "up", "ctrl+k":
			if s.plantCursor > 0 {
				s.plantCursor--
			}
			return s, nil
		case "down", "ctrl+j":
			if s.plantCursor < len(s.plantFiltered)-1 {
				s.plantCursor++
			}
			return s, nil
		case "enter":
			if len(s.plantFiltered) > 0 && s.plantCursor < len(s.plantFiltered) {
				s.plantSelected = s.plantFiltered[s.plantCursor]
			}
			s.plantPicking = false
			s.plantSearch.Blur()
			if s.formFocusIdx < len(s.formInputs) {
				s.formInputs[s.formFocusIdx].Focus()
			}
			return s, textinput.Blink
		}
	}

	var cmd tea.Cmd
	s.plantSearch, cmd = s.plantSearch.Update(msg)
	// Filter plants based on search text
	query := strings.ToLower(s.plantSearch.Value())
	all := s.store.GetPlantNames()
	if query == "" {
		s.plantFiltered = all
	} else {
		s.plantFiltered = nil
		for _, p := range all {
			if strings.Contains(strings.ToLower(p), query) {
				s.plantFiltered = append(s.plantFiltered, p)
			}
		}
	}
	if s.plantCursor >= len(s.plantFiltered) {
		s.plantCursor = 0
	}
	return s, cmd
}

func (s SeedsView) updateLotPicker(msg tea.Msg) (SeedsView, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			s.lotPicking = false
			s.lotSearch.Blur()
			if s.formFocusIdx < len(s.formInputs) {
				s.formInputs[s.formFocusIdx].Focus()
			}
			return s, textinput.Blink
		case "up", "ctrl+k":
			if s.lotCursor > 0 {
				s.lotCursor--
			}
			return s, nil
		case "down", "ctrl+j":
			if s.lotCursor < len(s.lotFiltered)-1 {
				s.lotCursor++
			}
			return s, nil
		case "enter":
			if len(s.lotFiltered) > 0 && s.lotCursor < len(s.lotFiltered) {
				s.lotSelected = s.lotFiltered[s.lotCursor].Name
			}
			s.lotPicking = false
			s.lotSearch.Blur()
			if s.formFocusIdx < len(s.formInputs) {
				s.formInputs[s.formFocusIdx].Focus()
			}
			return s, textinput.Blink
		}
	}

	var cmd tea.Cmd
	s.lotSearch, cmd = s.lotSearch.Update(msg)
	query := strings.ToLower(s.lotSearch.Value())
	all := s.store.GetLots()
	if query == "" {
		s.lotFiltered = all
	} else {
		s.lotFiltered = nil
		for _, l := range all {
			label := fmt.Sprintf("%s %s %dx%d", l.Name, l.MatrixSize, l.Rows, l.Columns)
			if strings.Contains(strings.ToLower(label), query) {
				s.lotFiltered = append(s.lotFiltered, l)
			}
		}
	}
	if s.lotCursor >= len(s.lotFiltered) {
		s.lotCursor = 0
	}
	return s, cmd
}

func (s *SeedsView) submitForm() {
	if s.plantSelected == "" {
		s.message = "No plant selected. Press ctrl+p to pick a plant."
		s.messageType = "error"
		return
	}

	plant := s.plantSelected
	desc := strings.TrimSpace(s.formInputs[0].Value())

	if s.mode == seedModeEdit {
		row := 0
		rowEnd := 0
		colStart := 0
		colEnd := 0
		if s.formType == "matrix" && len(s.formInputs) > 1 {
			var err1, err2, err3 error
			row, err1 = strconv.Atoi(strings.TrimSpace(s.formInputs[1].Value()))
			rowEndStr := strings.TrimSpace(s.formInputs[2].Value())
			if rowEndStr != "" {
				rowEnd, err2 = strconv.Atoi(rowEndStr)
			}
			colStart, err3 = strconv.Atoi(strings.TrimSpace(s.formInputs[3].Value()))
			colEndVal, err4 := strconv.Atoi(strings.TrimSpace(s.formInputs[4].Value()))
			if err1 != nil || err3 != nil || err4 != nil {
				s.message = "Invalid row/column numbers"
				s.messageType = "error"
				return
			}
			if err2 != nil && rowEndStr != "" {
				s.message = "Invalid row end number"
				s.messageType = "error"
				return
			}
			colEnd = colEndVal
		}
		if err := s.store.UpdateSeed(s.editID, plant, desc, row, rowEnd, colStart, colEnd); err != nil {
			s.message = err.Error()
			s.messageType = "error"
		} else {
			s.message = "Seed updated"
			s.messageType = "success"
		}
		return
	}

	if s.formType == "single" {
		if err := s.store.AddSingleSeed(plant, desc); err != nil {
			s.message = err.Error()
			s.messageType = "error"
		} else {
			s.message = "Single seed added"
			s.messageType = "success"
		}
	} else {
		if s.lotSelected == "" {
			s.message = "No lot selected. Press ctrl+l to pick a lot."
			s.messageType = "error"
			return
		}
		// Find the selected lot
		var lot models.Lot
		found := false
		for _, l := range s.store.GetLots() {
			if l.Name == s.lotSelected {
				lot = l
				found = true
				break
			}
		}
		if !found {
			s.message = "Selected lot no longer exists."
			s.messageType = "error"
			return
		}

		row, err1 := strconv.Atoi(strings.TrimSpace(s.formInputs[1].Value()))
		rowEndStr := strings.TrimSpace(s.formInputs[2].Value())
		rowEnd := 0
		var err2 error
		if rowEndStr != "" {
			rowEnd, err2 = strconv.Atoi(rowEndStr)
		}
		colStart, err3 := strconv.Atoi(strings.TrimSpace(s.formInputs[3].Value()))
		colEnd, err4 := strconv.Atoi(strings.TrimSpace(s.formInputs[4].Value()))

		if err1 != nil || err3 != nil || err4 != nil {
			s.message = "Invalid row/column numbers"
			s.messageType = "error"
			return
		}
		if err2 != nil && rowEndStr != "" {
			s.message = "Invalid row end number"
			s.messageType = "error"
			return
		}
		if row < 1 || row > lot.Rows {
			s.message = fmt.Sprintf("Row must be 1-%d", lot.Rows)
			s.messageType = "error"
			return
		}
		if rowEnd > 0 {
			if rowEnd < row || rowEnd > lot.Rows {
				s.message = fmt.Sprintf("Row end must be %d-%d", row, lot.Rows)
				s.messageType = "error"
				return
			}
			// Multi-row: full columns, ignore col inputs
			colStart = 1
			colEnd = lot.Columns
		} else {
			if colStart < 1 || colEnd < colStart || colEnd > lot.Columns {
				s.message = fmt.Sprintf("Columns must be 1-%d and start ≤ end", lot.Columns)
				s.messageType = "error"
				return
			}
		}

		if err := s.store.AddMatrixSeed(lot.Name, plant, desc, row, rowEnd, colStart, colEnd); err != nil {
			s.message = err.Error()
			s.messageType = "error"
		} else {
			s.message = "Matrix seed added"
			s.messageType = "success"
		}
	}
}

func (s SeedsView) View() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render(" 🌱 Seeds Registry "))
	b.WriteString("\n\n")

	switch s.mode {
	case seedModeAdd, seedModeEdit:
		b.WriteString(s.renderForm())
	default:
		b.WriteString(s.renderListView())
	}

	if s.message != "" {
		b.WriteString("\n")
		if s.messageType == "error" {
			b.WriteString(ErrorStyle.Render("✗ " + s.message))
		} else {
			b.WriteString(SuccessStyle.Render("✓ " + s.message))
		}
	}

	b.WriteString("\n\n")
	if s.mode == seedModeAdd || s.mode == seedModeEdit {
		if s.plantPicking || s.lotPicking {
			b.WriteString(HelpStyle.Render("↑/↓: navigate • type to filter • enter: select • esc: cancel"))
		} else {
			b.WriteString(HelpStyle.Render("tab: next • ctrl+p: pick plant • ctrl+t: toggle type • ctrl+l: pick lot • enter: save • esc: cancel"))
		}
	} else if s.searching {
		b.WriteString(HelpStyle.Render("type to search • enter: apply • esc: cancel"))
	} else {
		b.WriteString(HelpStyle.Render("↑/↓: navigate • /: search • c: clear search • a: add • e: edit • d: delete • esc: back"))
	}

	return b.String()
}

func (s SeedsView) renderListView() string {
	var b strings.Builder

	// Search bar
	searchLabel := "Search: "
	if s.searching {
		b.WriteString(searchLabel)
		b.WriteString(s.searchInput.View())
	} else if s.query != "" {
		b.WriteString(searchLabel)
		b.WriteString(lipgloss.NewStyle().Foreground(Secondary).Render(s.query))
		b.WriteString(MutedStyle.Render(fmt.Sprintf("  (%d results)", len(s.results))))
	} else {
		b.WriteString(MutedStyle.Render("Press / to search"))
	}
	b.WriteString("\n\n")

	if len(s.results) == 0 {
		b.WriteString(MutedStyle.Render("  No seeds found."))
		return b.String()
	}

	// Table header
	header := fmt.Sprintf("  %-20s %-10s %-15s %-30s", "Index", "Type", "Plant", "Description")
	b.WriteString(SubtitleStyle.Render(header))
	b.WriteString("\n")
	b.WriteString(MutedStyle.Render(strings.Repeat("─", 78)))
	b.WriteString("\n")

	for i, seed := range s.results {
		line := fmt.Sprintf("%-20s %-10s %-15s %-30s", seed.Index(), seed.Type, seed.Plant, seed.Description)
		if i == s.cursor {
			b.WriteString(SelectedStyle.Render("▸ " + line))
		} else {
			b.WriteString(NormalStyle.Render("  " + line))
		}
		b.WriteString("\n")
	}
	return b.String()
}

func (s SeedsView) renderForm() string {
	var b strings.Builder

	title := "Add Seed"
	if s.mode == seedModeEdit {
		title = "Edit Seed"
	}
	b.WriteString(SubtitleStyle.Render(title))
	b.WriteString("\n\n")

	// Type selector
	if s.mode == seedModeAdd {
		typeLabel := "Type: "
		if s.formType == "single" {
			typeLabel += lipgloss.NewStyle().Foreground(Secondary).Bold(true).Render("[Single]") + MutedStyle.Render(" Matrix")
		} else {
			typeLabel += MutedStyle.Render("Single ") + lipgloss.NewStyle().Foreground(Secondary).Bold(true).Render("[Matrix]")
		}
		b.WriteString("  " + typeLabel)
		b.WriteString(MutedStyle.Render("  (ctrl+t to toggle)"))
		b.WriteString("\n\n")
	}

	// Plant picker
	if s.plantPicking {
		b.WriteString("  Plant: ")
		b.WriteString(s.plantSearch.View())
		b.WriteString("\n")
		if len(s.plantFiltered) == 0 {
			b.WriteString(MutedStyle.Render("    No matching plants found"))
			b.WriteString("\n")
		} else {
			maxShow := 8
			if len(s.plantFiltered) < maxShow {
				maxShow = len(s.plantFiltered)
			}
			for i := 0; i < maxShow; i++ {
				p := s.plantFiltered[i]
				if i == s.plantCursor {
					b.WriteString(SelectedStyle.Render("    ▸ " + p))
				} else {
					b.WriteString(NormalStyle.Render("      " + p))
				}
				b.WriteString("\n")
			}
			if len(s.plantFiltered) > maxShow {
				b.WriteString(MutedStyle.Render(fmt.Sprintf("    ... and %d more", len(s.plantFiltered)-maxShow)))
				b.WriteString("\n")
			}
		}
	} else if s.plantSelected != "" {
		b.WriteString(fmt.Sprintf("  Plant: %s", lipgloss.NewStyle().Foreground(Secondary).Bold(true).Render(s.plantSelected)))
		b.WriteString(MutedStyle.Render("  (ctrl+p to change)"))
		b.WriteString("\n")
	} else {
		plants := s.store.GetPlantNames()
		if len(plants) == 0 {
			b.WriteString(ErrorStyle.Render("  Plant: No plants defined! Add plants in Admin first."))
		} else {
			b.WriteString(MutedStyle.Render("  Plant: (none selected — press ctrl+p to pick)"))
		}
		b.WriteString("\n")
	}

	// Lot picker (matrix only)
	if s.formType == "matrix" && s.mode == seedModeAdd {
		if s.lotPicking {
			b.WriteString("  Lot:   ")
			b.WriteString(s.lotSearch.View())
			b.WriteString("\n")
			if len(s.lotFiltered) == 0 {
				b.WriteString(MutedStyle.Render("    No matching lots found"))
				b.WriteString("\n")
			} else {
				maxShow := 8
				if len(s.lotFiltered) < maxShow {
					maxShow = len(s.lotFiltered)
				}
				for i := 0; i < maxShow; i++ {
					l := s.lotFiltered[i]
					label := fmt.Sprintf("%s (%d×%d)", l.Name, l.Rows, l.Columns)
					if i == s.lotCursor {
						b.WriteString(SelectedStyle.Render("    ▸ " + label))
					} else {
						b.WriteString(NormalStyle.Render("      " + label))
					}
					b.WriteString("\n")
				}
				if len(s.lotFiltered) > maxShow {
					b.WriteString(MutedStyle.Render(fmt.Sprintf("    ... and %d more", len(s.lotFiltered)-maxShow)))
					b.WriteString("\n")
				}
			}
		} else if s.lotSelected != "" {
			// Find lot details for display
			lotLabel := s.lotSelected
			for _, l := range s.store.GetLots() {
				if l.Name == s.lotSelected {
					lotLabel = fmt.Sprintf("%s (%d×%d)", l.Name, l.Rows, l.Columns)
					break
				}
			}
			b.WriteString(fmt.Sprintf("  Lot:   %s", lipgloss.NewStyle().Foreground(Secondary).Bold(true).Render(lotLabel)))
			b.WriteString(MutedStyle.Render("  (ctrl+l to change)"))
			b.WriteString("\n")
		} else {
			lots := s.store.GetLots()
			if len(lots) == 0 {
				b.WriteString(ErrorStyle.Render("  Lot:   No lots defined! Add lots in Admin first."))
			} else {
				b.WriteString(MutedStyle.Render("  Lot:   (none selected — press ctrl+l to pick)"))
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")

	// Description
	b.WriteString("  Description: ")
	b.WriteString(s.formInputs[0].View())
	b.WriteString("\n")

	// Matrix-specific fields
	if s.formType == "matrix" && len(s.formInputs) > 1 {
		b.WriteString("  Row:         ")
		b.WriteString(s.formInputs[1].View())
		b.WriteString("\n")
		b.WriteString("  Row End:     ")
		b.WriteString(s.formInputs[2].View())
		b.WriteString(MutedStyle.Render("  (blank = single row)"))
		b.WriteString("\n")
		rowEndFilled := strings.TrimSpace(s.formInputs[2].Value()) != ""
		if rowEndFilled {
			b.WriteString(MutedStyle.Render("  Col Start:   (full columns)"))
			b.WriteString("\n")
			b.WriteString(MutedStyle.Render("  Col End:     (full columns)"))
			b.WriteString("\n")
		} else {
			b.WriteString("  Col Start:   ")
			b.WriteString(s.formInputs[3].View())
			b.WriteString("\n")
			b.WriteString("  Col End:     ")
			b.WriteString(s.formInputs[4].View())
			b.WriteString("\n")
		}
	}

	return b.String()
}

func (s *SeedsView) RefreshResults() {
	s.results = s.store.SearchSeeds(s.query)
}

func (s SeedsView) IsFormMode() bool {
	return s.mode == seedModeAdd || s.mode == seedModeEdit
}

func (s SeedsView) IsSearching() bool {
	return s.searching
}

func (s SeedsView) IsPlantPicking() bool {
	return s.plantPicking
}

func (s SeedsView) IsLotPicking() bool {
	return s.lotPicking
}
