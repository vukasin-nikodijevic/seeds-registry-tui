# Seeds Registry TUI

Go TUI app for tracking planted seeds. Uses Bubble Tea framework.

1) it should support search for each item
2) items should be configurable to different kinds:
    * matrix nuber of rows and columns
    * single
3) each item / subitem should be searchable and indexable

Matrix item type should have index and description defined by `LOT`. 
    Example: Lot_001-1/5
    Explanation of index: It is from lot 1, row 1 and column 5
Single item type is simple and defined by numeric.
     Example:s_001
     Explanation: s means single, and 001 means numeric representation of it.

All items will have 2 fields:
    Plant: e.g. cucamber, paprika, tomato, ....
    Description: e.g. white cucamber, cherry tomato, ....

All items should be searchable by plant and description fields.

Matrix item should support of populating range of columns in the row as same plant type. Same seed can be planted in few rows or multiple columns of row e.g 1-4.

Plant should be uniq for entire app and should be avalable for pickup (like prepopulated manu) in item creation dialog.

For matrix item Lot is uniq as well and name consturcted by `lot_` + `001` like index.

I'll say that UI should be constructed by admin part where we add:
    Lots for matrix items
    Matix size for matrix items
    Plant types

User should be able to perform CRUD operations over items (planted seeds).
User should be able to easily search by type and description.
User should be able to see representation of matrix item in some visual so it can be understandad if there's free spots in matrix lot or general overview.

When searched item is in the matrix lot, highlite rows where it's planted.
åMatrix visual representation should have indexes from both axes so user can easily navigate.

create readme file so kid of 10 years can understand
save data in yaml file and make option to export to printable format and easy. to read for human. As well, readme should contains technical part how to compile source into reusable binary.

Yaml file should be saved into ~/.seeds-registry-tui in home folder of the user.

## Structure
- `models/` — Data models, YAML persistence (~/.seeds-registry-tui/)
- `ui/` — TUI views: admin, seeds CRUD, matrix visualization
- `main.go` — Entry point, menu navigation

## Domain
- PlantType: unique name (case-insensitive), e.g. "tomato"
- Lot: matrix grid, auto-named lot_001, lot_002; defined by rows × columns
- SeedItem: "matrix" (Lot_001-1/5) or "single" (s_001)
  - Fields: plant, description
  - Matrix supports column ranges (cols 1-4 same plant)

## Build
- `go run .` — run from source
- `go build -o seeds-registry .` — compile binary
- Dependencies: bubbletea, lipgloss, bubbles, yaml.v3

## Conventions
- Data stored as YAML, not JSON
- Search is case-insensitive across plant, description, index
- All IDs auto-increment
- Exports go to ~/.seeds-registry-tui/ as timestamped .txt
- UI uses 90% of terminal width; matrix view uses 90% width and height
- Seeds list is paginated to fit terminal height (PgUp/PgDn/Home/End supported)
- Terminal dimensions are forwarded to all subviews via SetHeight/SetSize on WindowSizeMsg