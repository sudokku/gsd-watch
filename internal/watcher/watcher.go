package watcher

import (
	"fmt"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fsnotify/fsnotify"
	"github.com/radu/gsd-watch/internal/tui"
)

// Run monitors root recursively using fsnotify. All subdirectories are added
// on startup via filepath.WalkDir (WATCH-01). Newly created directories are
// dynamically added to the watcher (WATCH-02). Rapid file writes are
// debounced at 300ms per path using a per-path timer map (WATCH-03).
//
// Run runs forever; call it as a goroutine: go watcher.Run(root, events).
// Only Write and Create fsnotify ops produce FileChangedMsg values.
func Run(root string, events chan<- tea.Msg) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Fprintf(os.Stderr, "watcher: failed to create fsnotify watcher: %v\n", err)
		return
	}
	defer w.Close()

	// WATCH-01: Add root and all subdirectories on startup.
	if walkErr := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable entries, don't abort walk
		}
		if d.IsDir() {
			return w.Add(path)
		}
		return nil
	}); walkErr != nil {
		fmt.Fprintf(os.Stderr, "watcher: WalkDir error: %v\n", walkErr)
	}

	// Debounce state: one timer per watched path.
	var mu sync.Mutex
	timers := make(map[string]*time.Timer)

	for {
		select {
		case e, ok := <-w.Events:
			if !ok {
				return
			}

			// Filter: only Write and Create ops produce messages.
			if !e.Has(fsnotify.Write) && !e.Has(fsnotify.Create) {
				continue
			}

			// WATCH-02: Dynamically add newly created directories.
			// If the created path is a directory, add it to the watcher and skip
			// sending a FileChangedMsg — directory creation is not a file change.
			if e.Has(fsnotify.Create) {
				if info, statErr := os.Stat(e.Name); statErr == nil && info.IsDir() {
					_ = w.Add(e.Name) // idempotent — safe to call on already-watched path
					continue          // directory create: don't debounce/send a FileChangedMsg
				}
			}

			// Skip known editor temp/swap files (e.g. micro's .tmp.PID.ts, vim's .swp).
			if isTempFile(e.Name) {
				continue
			}

			// WATCH-03: Debounce — reset per-path timer to 300ms.
			// Capture path locally to avoid closure-capture issues.
			path := e.Name

			mu.Lock()
			t, exists := timers[path]
			if !exists {
				// Create timer in stopped state; it will be reset below.
				t = time.AfterFunc(time.Duration(math.MaxInt64), func() {
					events <- tui.FileChangedMsg{Path: path}
					mu.Lock()
					delete(timers, path)
					mu.Unlock()
				})
				t.Stop()
				timers[path] = t
			}
			mu.Unlock()

			t.Reset(300 * time.Millisecond)

		case err, ok := <-w.Errors:
			if !ok {
				return
			}
			fmt.Fprintf(os.Stderr, "watcher error: %v\n", err)
		}
	}
}

// isTempFile reports whether path is a known editor temp or swap file that
// should be ignored by the watcher. Patterns covered:
//   - micro: file.ext.tmp.PID.timestamp  (contains ".tmp.")
//   - vim:   file.swp / file.swx / file.swo
//   - generic: files ending in "~"
func isTempFile(path string) bool {
	base := filepath.Base(path)
	if strings.Contains(base, ".tmp.") {
		return true
	}
	ext := filepath.Ext(base)
	if ext == ".swp" || ext == ".swx" || ext == ".swo" {
		return true
	}
	return strings.HasSuffix(base, "~")
}
