package smush

import (
	"bytes"
	"fmt"
	"io"
	"sync"

	"github.com/fatih/color"
)

type Logger struct {
	firstLine sync.Once
	w         io.Writer
	prefix    []byte

	endedWithNewLine bool
}

const NewLineByte = byte('\n')

// ANSI text colors excluding:
// - grays (hard to read in terminal)
// - red (implies error when it's actually not)
var colors = []color.Attribute{
	color.FgGreen,
	color.FgYellow,
	color.FgBlue,
	color.FgMagenta,
	color.FgCyan,
}

func NewLogger(w io.Writer, prefix string, colorIndex int) (*Logger, error) {
	// Foreground colors are defined with Iota 30-38.
	// See github.com/fatih/color for details.
	colorizer := color.New(colors[colorIndex%len(colors)], color.Bold)
	colorized := []byte(colorizer.Sprint(prefix))

	return &Logger{
		w:      w,
		prefix: colorized,
	}, nil
}

func (l *Logger) Write(p []byte) (n int, err error) {
	clone := make([]byte, len(p))
	copied := copy(clone, p)
	if copied != len(p) {
		return 0, fmt.Errorf("copy: %d != %d", copied, len(p))
	}

	parts := bytes.Split(clone, []byte{NewLineByte})

	b := []byte{}
	for i, part := range parts {
		if (i == len(parts)-1) && len(part) == 0 {
			// Skip last part as it seems to always be empty slice.
			continue
		}
		b = append(b, NewLineByte)
		b = append(b, l.prefix...)
		b = append(b, part...)
	}

	n, err = l.w.Write(b)
	if n == len(b) {
		return len(p), err
	}
	return n, err
}
