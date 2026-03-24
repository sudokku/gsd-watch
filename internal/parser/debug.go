package parser

import (
	"fmt"
	"io"
	"time"
)

// DebugOut is the writer for debug output. nil means silent (default).
// Set to os.Stderr from main.go when --debug is passed.
var DebugOut io.Writer

// debugf writes "[debug HH:MM:SS] event: detail\n" to DebugOut if non-nil.
// All errors from Fprintf are silently discarded — debug output is best-effort.
func debugf(event, format string, args ...any) {
	if DebugOut == nil {
		return
	}
	ts := time.Now().Format("15:04:05")
	detail := fmt.Sprintf(format, args...)
	fmt.Fprintf(DebugOut, "[debug %s] %s: %s\n", ts, event, detail)
}
