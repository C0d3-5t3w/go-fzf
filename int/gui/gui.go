package gui

import (
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
	"github.com/C0d3-5t3w/go-fzf/int/storage" // Import storage package
)

// GUI represents the application's graphical user interface.
type GUI struct {
	App     fyne.App
	Window  fyne.Window
	Config  *config.Config
	Storage *storage.Storage // Add storage field

	searchInput *widget.Entry
	resultsList *widget.List

	// Data binding for results
	resultsData binding.StringList

	// Debouncing search
	searchTimer *time.Timer
	searchMutex sync.Mutex
	debounceDur time.Duration
}

// NewGUI creates and initializes a new GUI instance.
func NewGUI(cfg *config.Config, store *storage.Storage) *GUI { // Add storage parameter
	a := app.New()
	w := a.NewWindow("Go-FZF Ripgrep Interface")

	gui := &GUI{
		App:         a,
		Window:      w,
		Config:      cfg,
		Storage:     store, // Assign storage
		resultsData: binding.NewStringList(),
		debounceDur: 300 * time.Millisecond, // Debounce duration
	}

	// Load history before setting up widgets that might use it
	err := gui.Storage.LoadHistory()
	if err != nil {
		log.Printf("Warning: Failed to load search history: %v", err)
		// Non-fatal, continue with empty history
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

	g.resultsList.OnSelected = func(id widget.ListItemID) {
		selectedItem, err := g.resultsData.GetValue(id)
		if err != nil {
			log.Printf("Error getting selected item: %v", err)
			return
		}
		// Example action: Log the selected item
		log.Printf("Selected: %s", selectedItem)
		// Potentially open the file/line in an editor here
		// For now, just clear selection and focus input
		g.resultsList.Unselect(id)
		g.Window.Canvas().Focus(g.searchInput)
	}
}

func (g *GUI) createContent() fyne.CanvasObject {
	return container.NewBorder(g.searchInput, nil, nil, nil, g.resultsList)
}

// performSearch runs ripgrep and updates the results list.
func (g *GUI) performSearch(pattern string) {
	trimmedPattern := strings.TrimSpace(pattern)
	if trimmedPattern == "" {
		g.resultsData.Set([]string{}) // Clear results if input is empty
		return
	}

	// Add to history and save *before* starting the async search
	// to capture the intent immediately.
	g.Storage.AddSearchTerm(trimmedPattern)
	go func() { // Save history asynchronously to avoid blocking UI slightly
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
			// Optionally display error to the user via a status bar or dialog
			fyne.CurrentApp().SendNotification(&fyne.Notification{
				Title:   "Ripgrep Error",
				Content: err.Error(),
			})
			g.resultsData.Set([]string{"Error running search..."}) // Show error in list
			return
		}

		if len(results) == 0 {
			results = []string{"No matches found."}
		}

		// Update the results data binding (must be done on main thread implicitly by Fyne)
		err = g.resultsData.Set(results)
		if err != nil {
			log.Printf("Error setting results data: %v", err)
		}
	}(trimmedPattern)
}

// Run shows the main window and starts the Fyne application event loop.
func (g *GUI) Run() {
	g.Window.ShowAndRun()
}
