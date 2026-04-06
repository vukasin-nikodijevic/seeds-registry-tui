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

// AdminView handles plant types, matrix sizes, and lots management.
type AdminView struct {
	store       *models.Store
	tab         int // 0=plants, 1=matrix sizes, 2=lots
	cursor      int
	addMode     bool
	editMode    bool
	editTarget  string
	inputs      []textinput.Model
	focusIdx    int
	sizeIdx     int // for lot creation: which matrix size is selected
	message     string
	messageType string
}

func NewAdminView(store *models.Store) AdminView {
	return AdminView{store: store}
}

func (a AdminView) Init() tea.Cmd { return nil }

func (a AdminView) Update(msg tea.Msg) (AdminView, tea.Cmd) {
	if a.addMode || a.editMode {
		return a.updateFormMode(msg)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			a.tab = (a.tab + 1) % 3
			a.cursor = 0
			a.message = ""
		case "shift+tab":
			a.tab--
			if a.tab < 0 {
				a.tab = 2
			}
			a.cursor = 0
			a.message = ""
		case "up", "k":
			if a.cursor > 0 {
				a.cursor--
			}
		case "down", "j":
			max := a.listLen() - 1
			if a.cursor < max {
				a.cursor++
			}
		case "a":
			if a.tab == 2 {
				sizes := a.store.GetMatrixSizes()
				if len(sizes) == 0 {
					a.message = "Add a matrix size first before creating lots."
					a.messageType = "error"
					return a, nil
				}
			}
			a.addMode = true
			a.editMode = false
			a.message = ""
			a.sizeIdx = 0
			a.inputs = a.createInputs()
			a.focusIdx = 0
			if len(a.inputs) > 0 {
				a.inputs[0].Focus()
			}
			return a, textinput.Blink
		case "e":
			if a.tab == 2 {
				a.message = "Lots cannot be edited. Delete and recreate instead."
				a.messageType = "error"
				return a, nil
			}
			a.startEdit()
			if a.editMode {
				return a, textinput.Blink
			}
		case "d":
			a.deleteSelected()
		}
	}
	return a, nil
}

func (a *AdminView) startEdit() {
	switch a.tab {
	case 0:
		plants := a.store.GetPlantNames()
		if a.cursor >= len(plants) {
			return
		}
		a.editTarget = plants[a.cursor]
		ti := textinput.New()
		ti.SetValue(a.editTarget)
		ti.CharLimit = 50
		ti.Width = 30
		ti.Focus()
		a.inputs = []textinput.Model{ti}
	case 1:
		sizes := a.store.GetMatrixSizes()
		if a.cursor >= len(sizes) {
			return
		}
		ms := sizes[a.cursor]
		a.editTarget = ms.Name

		name := textinput.New()
		name.SetValue(ms.Name)
		name.CharLimit = 30
		name.Width = 30
		name.Focus()

		rows := textinput.New()
		rows.SetValue(fmt.Sprintf("%d", ms.Rows))
		rows.CharLimit = 5
		rows.Width = 20

		cols := textinput.New()
		cols.SetValue(fmt.Sprintf("%d", ms.Columns))
		cols.CharLimit = 5
		cols.Width = 20

		a.inputs = []textinput.Model{name, rows, cols}
	default:
		return
	}
	a.editMode = true
	a.addMode = false
	a.focusIdx = 0
	a.message = ""
}

func (a AdminView) updateFormMode(msg tea.Msg) (AdminView, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			a.addMode = false
			a.editMode = false
			a.inputs = nil
			return a, nil
		case "ctrl+s":
			if a.tab == 2 && a.addMode {
				sizes := a.store.GetMatrixSizes()
				if len(sizes) > 0 {
					a.sizeIdx = (a.sizeIdx + 1) % len(sizes)
				}
				return a, nil
			}
		case "tab", "shift+tab":
			if msg.String() == "tab" {
				a.focusIdx++
			} else {
				a.focusIdx--
			}
			if a.focusIdx >= len(a.inputs) {
				a.focusIdx = 0
			}
			if a.focusIdx < 0 {
				a.focusIdx = len(a.inputs) - 1
			}
			for i := range a.inputs {
				if i == a.focusIdx {
					a.inputs[i].Focus()
				} else {
					a.inputs[i].Blur()
				}
			}
			return a, textinput.Blink
		case "enter":
			if a.editMode {
				a.submitEdit()
			} else {
				a.submitAdd()
			}
			if a.messageType == "success" {
				a.addMode = false
				a.editMode = false
				a.inputs = nil
			}
			return a, nil
		}
	}

	var cmd tea.Cmd
	if a.focusIdx < len(a.inputs) {
		a.inputs[a.focusIdx], cmd = a.inputs[a.focusIdx].Update(msg)
	}
	return a, cmd
}

func (a *AdminView) createInputs() []textinput.Model {
	switch a.tab {
	case 0:
		ti := textinput.New()
		ti.Placeholder = "Plant name (e.g. tomato)"
		ti.CharLimit = 50
		ti.Width = 30
		return []textinput.Model{ti}
	case 1:
		name := textinput.New()
		name.Placeholder = "Name (e.g. small, large) — optional"
		name.CharLimit = 30
		name.Width = 30

		rows := textinput.New()
		rows.Placeholder = "Number of rows"
		rows.CharLimit = 5
		rows.Width = 20

		cols := textinput.New()
		cols.Placeholder = "Number of columns"
		cols.CharLimit = 5
		cols.Width = 20

		return []textinput.Model{name, rows, cols}
	case 2:
		idx := textinput.New()
		idx.SetValue(fmt.Sprintf("%d", a.store.NextLotIndex()))
		idx.Placeholder = "Lot index number"
		idx.CharLimit = 5
		idx.Width = 20
		return []textinput.Model{idx}
	}
	return nil
}

func (a *AdminView) submitAdd() {
	switch a.tab {
	case 0:
		name := strings.TrimSpace(a.inputs[0].Value())
		if err := a.store.AddPlant(name); err != nil {
			a.message = err.Error()
			a.messageType = "error"
		} else {
			a.message = fmt.Sprintf("Plant '%s' added", name)
			a.messageType = "success"
		}
	case 1:
		name := strings.TrimSpace(a.inputs[0].Value())
		rows, err1 := strconv.Atoi(strings.TrimSpace(a.inputs[1].Value()))
		cols, err2 := strconv.Atoi(strings.TrimSpace(a.inputs[2].Value()))
		if err1 != nil || err2 != nil || rows <= 0 || cols <= 0 {
			a.message = "Invalid dimensions — enter positive numbers"
			a.messageType = "error"
			return
		}
		if name == "" {
			name = fmt.Sprintf("%dx%d", rows, cols)
		}
		if err := a.store.AddMatrixSize(name, rows, cols); err != nil {
			a.message = err.Error()
			a.messageType = "error"
		} else {
			a.message = fmt.Sprintf("Matrix size '%s' (%d×%d) added", name, rows, cols)
			a.messageType = "success"
		}
	case 2:
		sizes := a.store.GetMatrixSizes()
		if len(sizes) == 0 {
			a.message = "No matrix sizes defined. Add one first."
			a.messageType = "error"
			return
		}
		ms := sizes[a.sizeIdx]
		idxStr := strings.TrimSpace(a.inputs[0].Value())
		idx, err := strconv.Atoi(idxStr)
		if err != nil || idx <= 0 {
			a.message = "Invalid index — enter a positive number"
			a.messageType = "error"
			return
		}
		if err := a.store.AddLotWithIndex(ms.Name, idx); err != nil {
			a.message = err.Error()
			a.messageType = "error"
		} else {
			a.message = fmt.Sprintf("Lot 'lot_%03d' created with size '%s' (%d×%d)", idx, ms.Name, ms.Rows, ms.Columns)
			a.messageType = "success"
		}
	}
}

func (a *AdminView) submitEdit() {
	switch a.tab {
	case 0:
		newName := strings.TrimSpace(a.inputs[0].Value())
		if err := a.store.RenamePlant(a.editTarget, newName); err != nil {
			a.message = err.Error()
			a.messageType = "error"
		} else {
			a.message = fmt.Sprintf("Plant renamed '%s' → '%s'", a.editTarget, newName)
			a.messageType = "success"
		}
	case 1:
		newName := strings.TrimSpace(a.inputs[0].Value())
		rows, err1 := strconv.Atoi(strings.TrimSpace(a.inputs[1].Value()))
		cols, err2 := strconv.Atoi(strings.TrimSpace(a.inputs[2].Value()))
		if newName == "" {
			a.message = "Name cannot be empty"
			a.messageType = "error"
			return
		}
		if err1 != nil || err2 != nil || rows <= 0 || cols <= 0 {
			a.message = "Invalid dimensions — enter positive numbers"
			a.messageType = "error"
			return
		}
		if err := a.store.UpdateMatrixSize(a.editTarget, newName, rows, cols); err != nil {
			a.message = err.Error()
			a.messageType = "error"
		} else {
			a.message = fmt.Sprintf("Matrix size '%s' updated → '%s' %d×%d", a.editTarget, newName, rows, cols)
			a.messageType = "success"
		}
	}
}

func (a *AdminView) deleteSelected() {
	switch a.tab {
	case 0:
		plants := a.store.GetPlantNames()
		if a.cursor < len(plants) {
			name := plants[a.cursor]
			if err := a.store.RemovePlant(name); err != nil {
				a.message = err.Error()
				a.messageType = "error"
			} else {
				a.message = fmt.Sprintf("Plant '%s' removed", name)
				a.messageType = "success"
				if a.cursor > 0 {
					a.cursor--
				}
			}
		}
	case 1:
		sizes := a.store.GetMatrixSizes()
		if a.cursor < len(sizes) {
			name := sizes[a.cursor].Name
			if err := a.store.RemoveMatrixSize(name); err != nil {
				a.message = err.Error()
				a.messageType = "error"
			} else {
				a.message = fmt.Sprintf("Matrix size '%s' removed", name)
				a.messageType = "success"
				if a.cursor > 0 {
					a.cursor--
				}
			}
		}
	case 2:
		lots := a.store.GetLots()
		if a.cursor < len(lots) {
			name := lots[a.cursor].Name
			if err := a.store.RemoveLot(name); err != nil {
				a.message = err.Error()
				a.messageType = "error"
			} else {
				a.message = fmt.Sprintf("Lot '%s' removed", name)
				a.messageType = "success"
				if a.cursor > 0 {
					a.cursor--
				}
			}
		}
	}
}

func (a AdminView) listLen() int {
	switch a.tab {
	case 0:
		return len(a.store.GetPlantNames())
	case 1:
		return len(a.store.GetMatrixSizes())
	case 2:
		return len(a.store.GetLots())
	}
	return 0
}

func (a AdminView) View() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render(" 🌱 Admin Panel "))
	b.WriteString("\n\n")

	tabs := []string{"Plants", "Matrix Sizes", "Lots"}
	var tabRow []string
	for i, t := range tabs {
		if i == a.tab {
			tabRow = append(tabRow, TabActiveStyle.Render(t))
		} else {
			tabRow = append(tabRow, TabInactiveStyle.Render(t))
		}
	}
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, tabRow...))
	b.WriteString("\n\n")

	if a.addMode || a.editMode {
		b.WriteString(a.renderForm())
	} else {
		b.WriteString(a.renderList())
	}

	if a.message != "" {
		b.WriteString("\n")
		if a.messageType == "error" {
			b.WriteString(ErrorStyle.Render("✗ " + a.message))
		} else {
			b.WriteString(SuccessStyle.Render("✓ " + a.message))
		}
	}

	b.WriteString("\n\n")
	if a.addMode || a.editMode {
		help := "tab: next field • enter: submit • esc: cancel"
		if a.tab == 2 && a.addMode {
			help = "ctrl+s: cycle size • tab: next field • enter: submit • esc: cancel"
		}
		b.WriteString(HelpStyle.Render(help))
	} else {
		help := "↑/↓: navigate • a: add • e: edit • d: delete • tab: switch tab • esc: back"
		if a.tab == 2 {
			help = "↑/↓: navigate • a: add • d: delete • tab: switch tab • esc: back"
		}
		b.WriteString(HelpStyle.Render(help))
	}

	return b.String()
}

func (a AdminView) renderList() string {
	var b strings.Builder

	switch a.tab {
	case 0:
		b.WriteString(SubtitleStyle.Render("Plant Types"))
		b.WriteString("\n\n")
		plants := a.store.GetPlantNames()
		if len(plants) == 0 {
			b.WriteString(MutedStyle.Render("  No plants added yet. Press 'a' to add."))
		}
		for i, p := range plants {
			if i == a.cursor {
				b.WriteString(SelectedStyle.Render("▸ " + p))
			} else {
				b.WriteString(NormalStyle.Render("  " + p))
			}
			b.WriteString("\n")
		}
	case 1:
		b.WriteString(SubtitleStyle.Render("Matrix Sizes"))
		b.WriteString("\n\n")
		sizes := a.store.GetMatrixSizes()
		if len(sizes) == 0 {
			b.WriteString(MutedStyle.Render("  No matrix sizes defined yet. Press 'a' to add."))
		}
		for i, ms := range sizes {
			info := fmt.Sprintf("%-15s  %d×%d", ms.Name, ms.Rows, ms.Columns)
			if i == a.cursor {
				b.WriteString(SelectedStyle.Render("▸ " + info))
			} else {
				b.WriteString(NormalStyle.Render("  " + info))
			}
			b.WriteString("\n")
		}
	case 2:
		b.WriteString(SubtitleStyle.Render("Matrix Lots"))
		b.WriteString("\n\n")
		lots := a.store.GetLots()
		if len(lots) == 0 {
			b.WriteString(MutedStyle.Render("  No lots created yet. Press 'a' to add."))
		}
		for i, l := range lots {
			info := fmt.Sprintf("%s  [%s]  %d×%d", l.Name, l.MatrixSize, l.Rows, l.Columns)
			if i == a.cursor {
				b.WriteString(SelectedStyle.Render("▸ " + info))
			} else {
				b.WriteString(NormalStyle.Render("  " + info))
			}
			b.WriteString("\n")
		}
	}
	return b.String()
}

func (a AdminView) IsAddMode() bool {
	return a.addMode || a.editMode
}

func (a AdminView) renderForm() string {
	var b strings.Builder

	switch a.tab {
	case 0:
		if a.editMode {
			b.WriteString(SubtitleStyle.Render(fmt.Sprintf("Edit Plant: %s", a.editTarget)))
		} else {
			b.WriteString(SubtitleStyle.Render("Add Plant Type"))
		}
		b.WriteString("\n\n")
		b.WriteString("  Name: ")
		b.WriteString(a.inputs[0].View())
	case 1:
		if a.editMode {
			b.WriteString(SubtitleStyle.Render(fmt.Sprintf("Edit Matrix Size: %s", a.editTarget)))
			b.WriteString("\n\n")
			b.WriteString("  Name:    ")
			b.WriteString(a.inputs[0].View())
			b.WriteString("\n")
			b.WriteString("  Rows:    ")
			b.WriteString(a.inputs[1].View())
			b.WriteString("\n")
			b.WriteString("  Columns: ")
			b.WriteString(a.inputs[2].View())
		} else {
			b.WriteString(SubtitleStyle.Render("Add Matrix Size"))
			b.WriteString("\n\n")
			b.WriteString("  Name:    ")
			b.WriteString(a.inputs[0].View())
			b.WriteString(MutedStyle.Render("  (leave empty for auto e.g. 3x4)"))
			b.WriteString("\n")
			b.WriteString("  Rows:    ")
			b.WriteString(a.inputs[1].View())
			b.WriteString("\n")
			b.WriteString("  Columns: ")
			b.WriteString(a.inputs[2].View())
		}
	case 2:
		b.WriteString(SubtitleStyle.Render("Add Lot"))
		b.WriteString("\n\n")
		sizes := a.store.GetMatrixSizes()
		if len(sizes) > 0 && a.sizeIdx < len(sizes) {
			ms := sizes[a.sizeIdx]
			b.WriteString(fmt.Sprintf("  Size:  %s",
				lipgloss.NewStyle().Foreground(Secondary).Bold(true).Render(
					fmt.Sprintf("%s (%d×%d)", ms.Name, ms.Rows, ms.Columns))))
			b.WriteString(MutedStyle.Render("  (ctrl+s to cycle)"))
			b.WriteString("\n")
		}
		b.WriteString("  Index: ")
		b.WriteString(a.inputs[0].View())
		idxVal := strings.TrimSpace(a.inputs[0].Value())
		if idx, err := strconv.Atoi(idxVal); err == nil && idx > 0 {
			b.WriteString(MutedStyle.Render(fmt.Sprintf("  → lot_%03d", idx)))
		}
	}
	b.WriteString("\n")
	return b.String()
}
