package models

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// PlantType represents a unique plant type in the registry.
type PlantType struct {
	Name string `yaml:"name"`
}

// MatrixSize represents a reusable matrix size template (rows × columns).
type MatrixSize struct {
	Name    string `yaml:"name"` // e.g. "3x4"
	Rows    int    `yaml:"rows"`
	Columns int    `yaml:"columns"`
}

// Label returns a human-readable label for the matrix size.
func (m MatrixSize) Label() string {
	return fmt.Sprintf("%s (%d×%d)", m.Name, m.Rows, m.Columns)
}

// Lot represents a matrix lot with a specific size and index.
type Lot struct {
	Name       string `yaml:"name"`        // e.g. lot_001
	MatrixSize string `yaml:"matrix_size"` // reference to MatrixSize.Name
	Rows       int    `yaml:"rows"`
	Columns    int    `yaml:"columns"`
}

// LotIndex returns the numeric index from the lot name.
func (l Lot) Index() int {
	var idx int
	fmt.Sscanf(l.Name, "lot_%d", &idx)
	return idx
}

// SeedItem represents a planted seed.
type SeedItem struct {
	ID          string `yaml:"id"`
	Type        string `yaml:"type"` // "matrix" or "single"
	Plant       string `yaml:"plant"`
	Description string `yaml:"description"`

	// Matrix-specific fields
	LotName  string `yaml:"lot_name,omitempty"`
	Row      int    `yaml:"row,omitempty"`
	RowEnd   int    `yaml:"row_end,omitempty"` // 0 means single row
	ColStart int    `yaml:"col_start,omitempty"`
	ColEnd   int    `yaml:"col_end,omitempty"`
}

// Index returns the display index for the item.
func (s SeedItem) Index() string {
	if s.Type == "matrix" {
		lotIdx := 0
		fmt.Sscanf(s.LotName, "lot_%d", &lotIdx)
		if s.RowEnd > 0 && s.RowEnd != s.Row {
			// Multi-row: full columns assumed
			return fmt.Sprintf("Lot_%03d-R%d-%d", lotIdx, s.Row, s.RowEnd)
		}
		if s.ColStart == s.ColEnd {
			return fmt.Sprintf("Lot_%03d-R%d/%d", lotIdx, s.Row, s.ColStart)
		}
		return fmt.Sprintf("Lot_%03d-R%d/%d-%d", lotIdx, s.Row, s.ColStart, s.ColEnd)
	}
	singleIdx := 0
	fmt.Sscanf(s.ID, "s_%d", &singleIdx)
	return fmt.Sprintf("s_%03d", singleIdx)
}

// MatchesSearch checks if the item matches a search query (case-insensitive).
func (s SeedItem) MatchesSearch(query string) bool {
	q := strings.ToLower(query)
	return strings.Contains(strings.ToLower(s.Plant), q) ||
		strings.Contains(strings.ToLower(s.Description), q) ||
		strings.Contains(strings.ToLower(s.Index()), q) ||
		strings.Contains(strings.ToLower(s.Type), q)
}

// Store holds all application data and persists to a YAML file.
type Store struct {
	mu          sync.RWMutex
	dataFile    string
	dataDir     string
	Plants      []PlantType  `yaml:"plants"`
	MatrixSizes []MatrixSize `yaml:"matrix_sizes"`
	Lots        []Lot        `yaml:"lots"`
	Seeds       []SeedItem   `yaml:"seeds"`
	NextSingle  int          `yaml:"next_single"`
}

func NewStore(dataDir string) (*Store, error) {
	if err := os.MkdirAll(dataDir, 0750); err != nil {
		return nil, err
	}
	s := &Store{
		dataFile:    filepath.Join(dataDir, "seeds_registry.yaml"),
		dataDir:     dataDir,
		Plants:      []PlantType{},
		MatrixSizes: []MatrixSize{},
		Lots:        []Lot{},
		Seeds:       []SeedItem{},
		NextSingle:  1,
	}
	// Try loading YAML first, fall back to legacy JSON
	if err := s.Load(); err != nil && !os.IsNotExist(err) {
		jsonFile := filepath.Join(dataDir, "seeds_registry.json")
		if _, jerr := os.Stat(jsonFile); jerr == nil {
			if err := s.loadLegacyJSON(jsonFile); err != nil {
				return nil, err
			}
			_ = s.save()
		}
	}
	return s, nil
}

func (s *Store) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.dataFile)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, s)
}

func (s *Store) loadLegacyJSON(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	// yaml.v3 is a superset of JSON, so it can parse JSON fine
	return yaml.Unmarshal(data, s)
}

// save persists data to disk. Caller must hold s.mu.
func (s *Store) save() error {
	data, err := yaml.Marshal(s)
	if err != nil {
		return err
	}
	return os.WriteFile(s.dataFile, data, 0600)
}

// Plant operations

func (s *Store) AddPlant(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := strings.TrimSpace(name)
	if n == "" {
		return fmt.Errorf("plant name cannot be empty")
	}
	for _, p := range s.Plants {
		if strings.EqualFold(p.Name, n) {
			return fmt.Errorf("plant '%s' already exists", n)
		}
	}
	s.Plants = append(s.Plants, PlantType{Name: n})
	return s.save()
}

func (s *Store) RemovePlant(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, p := range s.Plants {
		if strings.EqualFold(p.Name, name) {
			s.Plants = append(s.Plants[:i], s.Plants[i+1:]...)
			return s.save()
		}
	}
	return fmt.Errorf("plant '%s' not found", name)
}

func (s *Store) RenamePlant(oldName, newName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := strings.TrimSpace(newName)
	if n == "" {
		return fmt.Errorf("plant name cannot be empty")
	}
	for _, p := range s.Plants {
		if strings.EqualFold(p.Name, n) && !strings.EqualFold(p.Name, oldName) {
			return fmt.Errorf("plant '%s' already exists", n)
		}
	}
	found := false
	for i, p := range s.Plants {
		if strings.EqualFold(p.Name, oldName) {
			s.Plants[i].Name = n
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("plant '%s' not found", oldName)
	}
	// Update all seeds referencing the old name
	for i, seed := range s.Seeds {
		if strings.EqualFold(seed.Plant, oldName) {
			s.Seeds[i].Plant = n
		}
	}
	return s.save()
}

func (s *Store) GetPlantNames() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	names := make([]string, len(s.Plants))
	for i, p := range s.Plants {
		names[i] = p.Name
	}
	return names
}

// MatrixSize operations

func (s *Store) AddMatrixSize(name string, rows, cols int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := strings.TrimSpace(name)
	if n == "" {
		n = fmt.Sprintf("%dx%d", rows, cols)
	}
	if rows <= 0 || cols <= 0 {
		return fmt.Errorf("rows and columns must be positive")
	}
	for _, m := range s.MatrixSizes {
		if strings.EqualFold(m.Name, n) {
			return fmt.Errorf("matrix size '%s' already exists", n)
		}
	}
	s.MatrixSizes = append(s.MatrixSizes, MatrixSize{Name: n, Rows: rows, Columns: cols})
	return s.save()
}

func (s *Store) RemoveMatrixSize(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Check if any lot references this size
	for _, l := range s.Lots {
		if l.MatrixSize == name {
			return fmt.Errorf("cannot remove: lot '%s' uses this matrix size", l.Name)
		}
	}
	for i, m := range s.MatrixSizes {
		if m.Name == name {
			s.MatrixSizes = append(s.MatrixSizes[:i], s.MatrixSizes[i+1:]...)
			return s.save()
		}
	}
	return fmt.Errorf("matrix size '%s' not found", name)
}

func (s *Store) UpdateMatrixSize(name string, newName string, rows, cols int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if rows <= 0 || cols <= 0 {
		return fmt.Errorf("rows and columns must be positive")
	}
	newName = strings.TrimSpace(newName)
	if newName == "" {
		return fmt.Errorf("name cannot be empty")
	}
	// Check for duplicate name if renaming
	if !strings.EqualFold(newName, name) {
		for _, m := range s.MatrixSizes {
			if strings.EqualFold(m.Name, newName) {
				return fmt.Errorf("matrix size '%s' already exists", newName)
			}
		}
	}
	found := false
	for i, m := range s.MatrixSizes {
		if m.Name == name {
			s.MatrixSizes[i].Name = newName
			s.MatrixSizes[i].Rows = rows
			s.MatrixSizes[i].Columns = cols
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("matrix size '%s' not found", name)
	}
	// Also update all lots that use this matrix size
	for i, l := range s.Lots {
		if l.MatrixSize == name {
			s.Lots[i].MatrixSize = newName
			s.Lots[i].Rows = rows
			s.Lots[i].Columns = cols
		}
	}
	return s.save()
}

func (s *Store) GetMatrixSizes() []MatrixSize {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]MatrixSize, len(s.MatrixSizes))
	copy(result, s.MatrixSizes)
	return result
}

// Lot operations

func (s *Store) NextLotIndex() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	maxIdx := 0
	for _, l := range s.Lots {
		idx := l.Index()
		if idx > maxIdx {
			maxIdx = idx
		}
	}
	return maxIdx + 1
}

func (s *Store) NextLotName() string {
	return fmt.Sprintf("lot_%03d", s.NextLotIndex())
}

func (s *Store) AddLotWithIndex(matrixSizeName string, index int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Find matrix size
	var ms *MatrixSize
	for _, m := range s.MatrixSizes {
		if m.Name == matrixSizeName {
			ms = &m
			break
		}
	}
	if ms == nil {
		return fmt.Errorf("matrix size '%s' not found", matrixSizeName)
	}
	if index <= 0 {
		return fmt.Errorf("lot index must be positive")
	}
	name := fmt.Sprintf("lot_%03d", index)
	// Check uniqueness
	for _, l := range s.Lots {
		if l.Name == name {
			return fmt.Errorf("lot '%s' already exists", name)
		}
	}
	s.Lots = append(s.Lots, Lot{
		Name:       name,
		MatrixSize: matrixSizeName,
		Rows:       ms.Rows,
		Columns:    ms.Columns,
	})
	return s.save()
}

func (s *Store) RemoveLot(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, l := range s.Lots {
		if l.Name == name {
			filtered := make([]SeedItem, 0)
			for _, seed := range s.Seeds {
				if seed.LotName != name {
					filtered = append(filtered, seed)
				}
			}
			s.Seeds = filtered
			s.Lots = append(s.Lots[:i], s.Lots[i+1:]...)
			return s.save()
		}
	}
	return fmt.Errorf("lot '%s' not found", name)
}

func (s *Store) GetLot(name string) *Lot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, l := range s.Lots {
		if l.Name == name {
			return &l
		}
	}
	return nil
}

func (s *Store) GetLots() []Lot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	lots := make([]Lot, len(s.Lots))
	copy(lots, s.Lots)
	return lots
}

// Seed operations

func (s *Store) AddMatrixSeed(lotName, plant, desc string, row, rowEnd, colStart, colEnd int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	lotIdx := 0
	fmt.Sscanf(lotName, "lot_%d", &lotIdx)
	var id string
	if rowEnd > 0 && rowEnd != row {
		id = fmt.Sprintf("Lot_%03d-R%d-%d", lotIdx, row, rowEnd)
	} else {
		id = fmt.Sprintf("Lot_%03d-R%d/%d-%d", lotIdx, row, colStart, colEnd)
	}

	seed := SeedItem{
		ID:          id,
		Type:        "matrix",
		Plant:       plant,
		Description: desc,
		LotName:     lotName,
		Row:         row,
		RowEnd:      rowEnd,
		ColStart:    colStart,
		ColEnd:      colEnd,
	}
	s.Seeds = append(s.Seeds, seed)
	return s.save()
}

func (s *Store) AddSingleSeed(plant, desc string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := fmt.Sprintf("s_%03d", s.NextSingle)
	s.NextSingle++
	seed := SeedItem{
		ID:          id,
		Type:        "single",
		Plant:       plant,
		Description: desc,
	}
	s.Seeds = append(s.Seeds, seed)
	return s.save()
}

func (s *Store) RemoveSeed(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, seed := range s.Seeds {
		if seed.ID == id {
			s.Seeds = append(s.Seeds[:i], s.Seeds[i+1:]...)
			return s.save()
		}
	}
	return fmt.Errorf("seed '%s' not found", id)
}

func (s *Store) UpdateSeed(id, plant, desc string, row, rowEnd, colStart, colEnd int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, seed := range s.Seeds {
		if seed.ID == id {
			s.Seeds[i].Plant = plant
			s.Seeds[i].Description = desc
			if seed.Type == "matrix" {
				s.Seeds[i].Row = row
				s.Seeds[i].RowEnd = rowEnd
				s.Seeds[i].ColStart = colStart
				s.Seeds[i].ColEnd = colEnd
			}
			return s.save()
		}
	}
	return fmt.Errorf("seed '%s' not found", id)
}

func (s *Store) SearchSeeds(query string) []SeedItem {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if query == "" {
		result := make([]SeedItem, len(s.Seeds))
		copy(result, s.Seeds)
		return result
	}
	var results []SeedItem
	for _, seed := range s.Seeds {
		if seed.MatchesSearch(query) {
			results = append(results, seed)
		}
	}
	return results
}

func (s *Store) GetSeedsInLot(lotName string) []SeedItem {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var results []SeedItem
	for _, seed := range s.Seeds {
		if seed.LotName == lotName {
			results = append(results, seed)
		}
	}
	return results
}

func (s *Store) GetAllSeeds() []SeedItem {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]SeedItem, len(s.Seeds))
	copy(result, s.Seeds)
	return result
}

// ExportDir returns the data directory for placing exports.
func (s *Store) ExportDir() string {
	return s.dataDir
}

// ExportPrintable generates a human-readable text report and saves it to a file.
func (s *Store) ExportPrintable() (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var b strings.Builder

	b.WriteString("╔══════════════════════════════════════════════════════════════╗\n")
	b.WriteString("║                    SEEDS REGISTRY REPORT                    ║\n")
	b.WriteString("╚══════════════════════════════════════════════════════════════╝\n")
	b.WriteString(fmt.Sprintf("  Generated: %s\n", time.Now().Format("2006-01-02 15:04")))
	b.WriteString("\n")

	// Summary
	b.WriteString("┌──────────────────────────────────────────────────────────────┐\n")
	b.WriteString("│  SUMMARY                                                    │\n")
	b.WriteString("├──────────────────────────────────────────────────────────────┤\n")
	b.WriteString(fmt.Sprintf("│  Plant types:   %-43d│\n", len(s.Plants)))
	b.WriteString(fmt.Sprintf("│  Matrix sizes:  %-43d│\n", len(s.MatrixSizes)))
	b.WriteString(fmt.Sprintf("│  Matrix lots:   %-43d│\n", len(s.Lots)))
	b.WriteString(fmt.Sprintf("│  Total seeds:   %-43d│\n", len(s.Seeds)))
	b.WriteString("└──────────────────────────────────────────────────────────────┘\n")
	b.WriteString("\n")

	// Plant types
	b.WriteString("── PLANT TYPES ──────────────────────────────────────────────\n\n")
	if len(s.Plants) == 0 {
		b.WriteString("  (none)\n")
	} else {
		for i, p := range s.Plants {
			b.WriteString(fmt.Sprintf("  %d. %s\n", i+1, p.Name))
		}
	}
	b.WriteString("\n")

	// Single seeds
	b.WriteString("── SINGLE SEEDS ─────────────────────────────────────────────\n\n")
	singles := 0
	b.WriteString(fmt.Sprintf("  %-12s  %-20s  %s\n", "Index", "Plant", "Description"))
	b.WriteString(fmt.Sprintf("  %-12s  %-20s  %s\n", "────────────", "────────────────────", "──────────────────────────"))
	for _, seed := range s.Seeds {
		if seed.Type == "single" {
			b.WriteString(fmt.Sprintf("  %-12s  %-20s  %s\n", seed.Index(), seed.Plant, seed.Description))
			singles++
		}
	}
	if singles == 0 {
		b.WriteString("  (none)\n")
	}
	b.WriteString("\n")

	// Matrix lots
	b.WriteString("── MATRIX LOTS ──────────────────────────────────────────────\n\n")
	for _, lot := range s.Lots {
		seeds := make([]SeedItem, 0)
		for _, seed := range s.Seeds {
			if seed.LotName == lot.Name {
				seeds = append(seeds, seed)
			}
		}

		// Build grid
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

		totalCells := lot.Rows * lot.Columns
		filled := 0
		for r := 0; r < lot.Rows; r++ {
			for c := 0; c < lot.Columns; c++ {
				if grid[r][c] != "" {
					filled++
				}
			}
		}

		b.WriteString(fmt.Sprintf("  %s  (%d rows × %d columns)  —  %d/%d planted (%.0f%%)\n",
			strings.ToUpper(lot.Name), lot.Rows, lot.Columns, filled, totalCells,
			float64(filled)/float64(totalCells)*100))
		b.WriteString("\n")

		// Column header
		colWidth := 12
		b.WriteString(fmt.Sprintf("  %*s", colWidth, ""))
		for c := 0; c < lot.Columns; c++ {
			b.WriteString(fmt.Sprintf("%-*s", colWidth, fmt.Sprintf("Col %d", c+1)))
		}
		b.WriteString("\n")
		b.WriteString("  " + strings.Repeat("─", colWidth+lot.Columns*colWidth) + "\n")

		// Rows
		for r := 0; r < lot.Rows; r++ {
			b.WriteString(fmt.Sprintf("  %-*s", colWidth, fmt.Sprintf("Row %d", r+1)))
			for c := 0; c < lot.Columns; c++ {
				cell := grid[r][c]
				if cell == "" {
					cell = "·"
				} else if len(cell) > colWidth-2 {
					cell = cell[:colWidth-2]
				}
				b.WriteString(fmt.Sprintf("%-*s", colWidth, cell))
			}
			b.WriteString("\n")
		}

		// Seed list for this lot
		if len(seeds) > 0 {
			b.WriteString("\n  Planted seeds:\n")
			b.WriteString(fmt.Sprintf("    %-22s  %-15s  %s\n", "Index", "Plant", "Description"))
			b.WriteString(fmt.Sprintf("    %-22s  %-15s  %s\n", "──────────────────────", "───────────────", "──────────────────────"))
			for _, seed := range seeds {
				b.WriteString(fmt.Sprintf("    %-22s  %-15s  %s\n", seed.Index(), seed.Plant, seed.Description))
			}
		}
		b.WriteString("\n\n")
	}

	if len(s.Lots) == 0 {
		b.WriteString("  (none)\n\n")
	}

	b.WriteString("══════════════════════════════════════════════════════════════\n")
	b.WriteString("  End of report\n")

	report := b.String()

	// Save to file
	filename := fmt.Sprintf("seeds_report_%s.txt", time.Now().Format("2006-01-02_1504"))
	outPath := filepath.Join(s.dataDir, filename)
	if err := os.WriteFile(outPath, []byte(report), 0600); err != nil {
		return "", err
	}

	return outPath, nil
}
