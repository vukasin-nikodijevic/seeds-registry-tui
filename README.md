# 🌱 Seeds Registry

A terminal app to keep track of all your planted seeds — like a notebook for your garden, but on a computer!

## What Does It Do?

Imagine you have a big garden. You plant tomatoes, cucumbers, peppers, and more. After a while, you forget what you planted and where. This app helps you remember everything!

There are **two ways** to plant seeds in this app:

### 1. Matrix Lots (like a grid or a chessboard)

Think of a big tray with rows and columns — like a waffle or an ice cube tray.

```
        C1        C2        C3        C4
  ┌─────────┬─────────┬─────────┬─────────┐
R1│ tomato  │ tomato  │  empty  │  empty  │
  ├─────────┼─────────┼─────────┼─────────┤
R2│ pepper  │ pepper  │ pepper  │ pepper  │
  ├─────────┼─────────┼─────────┼─────────┤
R3│  empty  │  empty  │  empty  │ cucumber│
  └─────────┴─────────┴─────────┴─────────┘
```

Each tray is called a **Lot** and gets a name like `lot_001`, `lot_002`, etc. Every little square in the grid has an address like `Lot_001-2/3` which means "Lot 1, Row 2, Column 3". You can fill a whole row or part of it with the same plant at once!

### 2. Single Seeds

These are simple — just one seed, one entry. They get names like `s_001`, `s_002`, etc.

---

## The Main Menu

When you open the app, you see 4 choices:

| Choice | What It Does |
|---|---|
| **Admin Panel** | Set things up — add plant types and create lot trays |
| **Seeds Registry** | Add, edit, search, or delete your planted seeds |
| **Matrix View** | See a visual picture of your lot trays (like the grid above!) |
| **Export Report** | Save a printable text report of everything to a file |
| **Quit** | Close the app |

---

## How to Use It (Step by Step)

### Step 1: Set Up (Admin Panel)

Before planting anything, you need to tell the app what kinds of plants you have and create your trays.

1. Go to **Admin Panel**
2. On the **Plants** tab, press `a` and type a plant name (like "tomato"). You can add as many as you want!
3. Switch to the **Lots** tab with `Tab`, press `a`, and type how many rows and columns your tray should have

### Step 2: Plant Seeds (Seeds Registry)

1. Go to **Seeds Registry**
2. Press `a` to add a new seed
3. Pick the type:
   - Press `Ctrl+T` to switch between **Single** and **Matrix**
4. Press `Ctrl+P` to pick which plant it is
5. If it's a matrix seed, press `Ctrl+L` to pick which lot/tray, then type the row and column numbers
6. Type a description (like "cherry tomato" or "big red pepper")
7. Press `Enter` to save!

### Step 3: Find Your Seeds

- Press `/` to search — type any part of the plant name, description, or index
- Press `c` to clear your search and see everything again

### Step 4: See the Big Picture (Matrix View)

1. Go to **Matrix View**
2. You'll see your lot tray drawn as a grid
3. Use `Tab` to switch between different lots
4. Use arrow keys to scroll around big trays
5. Press `/` to search — rows with matching plants get **highlighted in yellow**!
6. At the bottom, you'll see how full your tray is (like "5 out of 20 spots used")

---

## Keyboard Shortcuts

| Key | What It Does |
|---|---|
| `↑` `↓` or `k` `j` | Move up and down |
| `Enter` | Select or confirm |
| `Esc` | Go back |
| `/` | Start searching |
| `c` | Clear search |
| `a` | Add something new |
| `e` | Edit something |
| `d` | Delete something |
| `Tab` | Switch tabs |
| `Ctrl+T` | Switch seed type (single/matrix) |
| `Ctrl+P` | Pick a plant |
| `Ctrl+L` | Pick a lot |
| `q` | Quit (from main menu) |

---

## How to Run

Make sure you have [Go](https://go.dev) installed (version 1.25 or newer), then:

```bash
cd seeds-registry-tui
go run .
```

Your data is saved automatically in a YAML file at `~/.seeds-registry-tui/seeds_registry.yaml`, so everything is still there next time you open the app!

---

## Building a Binary (Technical)

Instead of running from source every time, you can **compile** the app into a single executable file (a "binary") that works without Go installed.

### Build for your current machine

```bash
cd seeds-registry-tui
go build -o seeds-registry .
```

This creates a file called `seeds-registry` in the current folder. Run it with:

```bash
./seeds-registry
```

### Build for other operating systems

You can build for a different OS or CPU even from your own machine. Just set two variables before the build command:

**Linux (64-bit):**
```bash
GOOS=linux GOARCH=amd64 go build -o seeds-registry-linux .
```

**Windows:**
```bash
GOOS=windows GOARCH=amd64 go build -o seeds-registry.exe .
```

**Mac (Apple Silicon / M1, M2, ...):**
```bash
GOOS=darwin GOARCH=arm64 go build -o seeds-registry-mac .
```

**Mac (Intel):**
```bash
GOOS=darwin GOARCH=amd64 go build -o seeds-registry-mac-intel .
```

### Make a smaller binary

Add `-ldflags="-s -w"` to strip debug info and make the file smaller:

```bash
go build -ldflags="-s -w" -o seeds-registry .
```

### Install globally

To install it so you can run `seeds-registry` from anywhere in your terminal:

```bash
go install .
```

This puts the binary in your `$GOPATH/bin` (usually `~/go/bin`). Make sure that folder is in your `PATH`.

---

## Data & Export

- **Data file:** `~/.seeds-registry-tui/seeds_registry.yaml` — human-readable YAML, easy to edit by hand or back up
- **Export reports:** Select "Export Report" from the main menu. A printable `.txt` file is saved to `~/.seeds-registry-tui/` with a timestamped name like `seeds_report_2026-04-06_1430.txt`
