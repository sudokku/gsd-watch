package watcher_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/radu/gsd-watch/internal/tui"
	"github.com/radu/gsd-watch/internal/watcher"
)

// TestRunAddsSubdirs verifies that Run() watches all subdirectories on startup.
// WATCH-01: WalkDir adds all subdirs to watcher so events fire for nested files.
func TestRunAddsSubdirs(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	sub1 := filepath.Join(root, "sub1")
	sub2 := filepath.Join(sub1, "sub2")
	if err := os.MkdirAll(sub2, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	events := make(chan tea.Msg, 10)
	go watcher.Run(root, events)

	// Give watcher time to start and add directories.
	time.Sleep(100 * time.Millisecond)

	// Write a file deep in the tree.
	testFile := filepath.Join(sub2, "test.txt")
	if err := os.WriteFile(testFile, []byte("hello"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	// Expect a FileChangedMsg within 2 seconds.
	select {
	case msg := <-events:
		fm, ok := msg.(tui.FileChangedMsg)
		if !ok {
			t.Fatalf("expected FileChangedMsg, got %T", msg)
		}
		if !strings.Contains(fm.Path, "test.txt") {
			t.Errorf("expected path containing test.txt, got %q", fm.Path)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout: no FileChangedMsg received for subdirectory file write")
	}
}

// TestDynamicDirAdd verifies that a newly created directory is dynamically added to the watcher.
// WATCH-02: fsnotify.Create events for directories trigger watcher.Add().
func TestDynamicDirAdd(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	events := make(chan tea.Msg, 10)
	go watcher.Run(root, events)

	// Give watcher time to start.
	time.Sleep(100 * time.Millisecond)

	// Create a new directory inside root.
	newDir := filepath.Join(root, "newdir")
	if err := os.Mkdir(newDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// Give watcher time to detect the new directory and add it.
	time.Sleep(150 * time.Millisecond)

	// Drain any messages from directory creation event.
	drain(events)

	// Write a file inside the new directory.
	testFile := filepath.Join(newDir, "file.txt")
	if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	// Expect a FileChangedMsg for the file in the dynamically-added directory.
	select {
	case msg := <-events:
		fm, ok := msg.(tui.FileChangedMsg)
		if !ok {
			t.Fatalf("expected FileChangedMsg, got %T", msg)
		}
		if !strings.Contains(fm.Path, "file.txt") {
			t.Errorf("expected path containing file.txt, got %q", fm.Path)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout: no FileChangedMsg received for file in dynamically-added directory")
	}
}

// TestDebounce verifies that rapid writes to the same file produce exactly one FileChangedMsg.
// WATCH-03: 300ms debounce collapses rapid writes into a single event.
func TestDebounce(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	events := make(chan tea.Msg, 20)
	go watcher.Run(root, events)

	// Give watcher time to start.
	time.Sleep(100 * time.Millisecond)

	testFile := filepath.Join(root, "test.txt")

	// Write to same file 5 times with 20ms gaps (well within 300ms window).
	for i := 0; i < 5; i++ {
		if err := os.WriteFile(testFile, []byte("write"), 0644); err != nil {
			t.Fatalf("write %d: %v", i, err)
		}
		time.Sleep(20 * time.Millisecond)
	}

	// Collect all messages for 1 second after last write.
	var received []tea.Msg
	timer := time.NewTimer(1 * time.Second)
	defer timer.Stop()

	// The debounce fires 300ms after the last write, so we wait up to 1 second.
collecting:
	for {
		select {
		case msg := <-events:
			received = append(received, msg)
		case <-timer.C:
			break collecting
		}
	}

	// Exactly 1 message should have been received.
	count := 0
	for _, msg := range received {
		if _, ok := msg.(tui.FileChangedMsg); ok {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected exactly 1 FileChangedMsg after 5 rapid writes, got %d", count)
	}
}

// TestFilterOps verifies that only Write and Create events produce FileChangedMsg.
// Remove events should not produce messages.
func TestFilterOps(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	events := make(chan tea.Msg, 10)
	go watcher.Run(root, events)

	// Give watcher time to start.
	time.Sleep(100 * time.Millisecond)

	// Create a file (this will trigger a Create event).
	testFile := filepath.Join(root, "toremove.txt")
	if err := os.WriteFile(testFile, []byte("data"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	// Wait for and drain the create/write message.
	select {
	case <-events:
		// Expected — drain the create event.
	case <-time.After(2 * time.Second):
		t.Fatal("timeout: expected FileChangedMsg for file creation")
	}
	// Drain any extra messages from the write flush.
	time.Sleep(350 * time.Millisecond)
	drain(events)

	// Now remove the file — this should NOT produce a FileChangedMsg.
	if err := os.Remove(testFile); err != nil {
		t.Fatalf("remove: %v", err)
	}

	// Wait 500ms to confirm no message is sent for Remove.
	select {
	case msg := <-events:
		if _, ok := msg.(tui.FileChangedMsg); ok {
			t.Error("got unexpected FileChangedMsg for file removal (Remove op should be filtered)")
		}
	case <-time.After(500 * time.Millisecond):
		// Correct — no message received for Remove event.
	}
}

// drain reads and discards all messages currently in the channel.
func drain(ch chan tea.Msg) {
	for {
		select {
		case <-ch:
		default:
			return
		}
	}
}
