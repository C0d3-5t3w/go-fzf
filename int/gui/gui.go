package gui

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"

	"fyne.io/fyne/v2/widget"
	"github.com/C0d3-5t3w/go-fzf/int/config"
	"github.com/C0d3-5t3w/go-fzf/int/ripgrep"
	"github.com/C0d3-5t3w/go-fzf/int/storage"
)

// GUI represents the application's graphical user interface.
type GUI struct {
	App     fyne.App
	Window  fyne.Window
	Config  *config.Config
	Storage *storage.Storage

	searchInput *widget.Entry
	resultsList *widget.List
	statusBar   *widget.Label

	// Data binding for results
	resultsData binding.StringList

	// Debouncing search
	searchTimer *time.Timer
	searchMutex sync.Mutex
	debounceDur time.Duration
}

// NewGUI creates and initializes a new GUI instance.
func NewGUI(cfg *config.Config, store *storage.Storage) *GUI {
	a := app.New()
	w := a.NewWindow("Go-FZF Ripgrep Interface")

	gui := &GUI{
		App:         a,
		Window:      w,
		Config:      cfg,
		Storage:     store,
		resultsData: binding.NewStringList(),
		debounceDur: 300 * time.Millisecond,
		statusBar:   widget.NewLabel("Ready."),
	}

	// Load history before setting up widgets that might use it
	err := gui.Storage.LoadHistory()
	if err != nil {
		log.Printf("Warning: Failed to load search history: %v", err)
	}

	gui.setupWidgets()
	w.SetContent(gui.createContent())
	w.Resize(fyne.NewSize(800, 600))

	// Save history on close
	w.SetOnClosed(func() {
		err := gui.Storage.SaveHistory()
		if err != nil {
			log.Printf("Error saving history on close: %v", err)
		}
	})

	return gui
}

func (g *GUI) setupWidgets() {
	g.searchInput = widget.NewEntry()
	g.searchInput.SetPlaceHolder("Enter search pattern...")

	g.resultsList = widget.NewListWithData(g.resultsData,
		func() fyne.CanvasObject {
			return widget.NewLabel("template")
		},
		func(item binding.DataItem, obj fyne.CanvasObject) {
			label := obj.(*widget.Label)
			strItem := item.(binding.String)
			s, _ := strItem.Get()
			label.SetText(s)
		})

	// --- Event Handling ---
	g.searchInput.OnChanged = func(text string) {
		g.searchMutex.Lock()
		// Reset the timer if it exists
		if g.searchTimer != nil {
			g.searchTimer.Stop()
		}
		// Start a new timer
		g.searchTimer = time.AfterFunc(g.debounceDur, func() {
			g.performSearch(text)
		})
		g.searchMutex.Unlock()
	}

	// Handle Enter key press for immediate search
	g.searchInput.OnSubmitted = func(text string) {
		g.searchMutex.Lock()
		// Cancel any pending debounced search
		if g.searchTimer != nil {
			g.searchTimer.Stop()
			g.searchTimer = nil
		}
		g.searchMutex.Unlock()
		// Perform search immediately
		g.performSearch(text)
	}

	g.resultsList.OnSelected = func(id widget.ListItemID) {
		selectedItem, err := g.resultsData.GetValue(id)
		if err != nil {
			log.Printf("Error getting selected item: %v", err)
			g.statusBar.SetText(fmt.Sprintf("Error getting selection: %v", err))
			return
		}

		// Attempt to copy to clipboard using Window's clipboard
		clipboard := g.Window.Clipboard() // Corrected: Use g.Window
		if clipboard != nil {
			clipboard.SetContent(selectedItem)
			g.statusBar.SetText(fmt.Sprintf("Copied: %s", selectedItem))
		} else {
			// This case should be rare unless running in a very restricted environment
			log.Printf("Selected (clipboard not available): %s", selectedItem)
			g.statusBar.SetText(fmt.Sprintf("Selected: %s (clipboard N/A)", selectedItem))
		}

		// Clear selection and focus input for next search
		g.resultsList.Unselect(id)
		g.Window.Canvas().Focus(g.searchInput)
	}
}

func (g *GUI) createContent() fyne.CanvasObject {
	// Add padding around the list and include the status bar
	paddedList := container.NewPadded(g.resultsList)
	return container.NewBorder(g.searchInput, g.statusBar, nil, nil, paddedList)
}

// performSearch runs ripgrep and updates the results list.
func (g *GUI) performSearch(pattern string) {
	trimmedPattern := strings.TrimSpace(pattern)

	// Update status immediately
	if trimmedPattern == "" {
		g.resultsData.Set([]string{})
		g.statusBar.SetText("Ready. Enter a search pattern.")
		return
	}
	g.statusBar.SetText(fmt.Sprintf("Searching for '%s'...", trimmedPattern))

	// Add to history and save *before* starting the async search
	g.Storage.AddSearchTerm(trimmedPattern)
	go func() {
		err := g.Storage.SaveHistory()
		if err != nil {
			log.Printf("Error saving search history: %v", err)
		}
	}()

	// Run ripgrep in a goroutine to avoid blocking the UI thread
	go func(p string) {
		results, err := ripgrep.RunRipgrep(g.Config.RipgrepPath, p, g.Config.SearchDirs)
		if err != nil {
			log.Printf("Error running ripgrep: %v", err)
			fyne.CurrentApp().SendNotification(&fyne.Notification{
				Title:   "Ripgrep Error",
				Content: err.Error(),
			})
			g.resultsData.Set([]string{"Error running search..."})
			g.statusBar.SetText(fmt.Sprintf("Error: %v", err))
			return
		}

		var statusText string
		if len(results) == 0 {
			results = []string{"No matches found."}
			statusText = fmt.Sprintf("No matches found for '%s'", p)
		} else {
			statusText = fmt.Sprintf("Found %d match(es) for '%s'", len(results), p)
		}

		err = g.resultsData.Set(results)
		if err != nil {
			log.Printf("Error setting results data: %v", err)
			g.statusBar.SetText(fmt.Sprintf("Error updating results: %v", err))
		} else {
			g.statusBar.SetText(statusText)
		}
	}(trimmedPattern)
}

// Run shows the main window and starts the Fyne application event loop.
func (g *GUI) Run() {
	g.Window.ShowAndRun()
}
