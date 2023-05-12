package smush

import (
	"io"
	"os"
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

func init() {
	if os.Getenv("GITHUB_ACTIONS") != "" {
		color.NoColor = false
	}
}

// ANSI text colors excluding:
// - grays (hard to read in terminal)
// - red (implies error when it's actually not)
// - green (implies success when it's actually not)
var colors = []color.Attribute{
	color.FgYellow,
	color.FgBlue,
	color.FgMagenta,
	color.FgCyan,
	color.FgHiYellow,
	color.FgHiBlue,
	color.FgHiMagenta,
	color.FgHiCyan,
}

func NewLogger(w io.Writer, prefix string, colorIndex int) *Logger {
	// Foreground colors are defined with Iota 30-38.
	// See github.com/fatih/color for details.
	colorizer := color.New(colors[colorIndex%len(colors)], color.Bold)
	colorized := []byte(colorizer.Sprint(prefix))

	return &Logger{
		w:      w,
		prefix: colorized,
	}
}

func (l *Logger) Write(p []byte) (n int, err error) {
	result := []byte{}
	l.firstLine.Do(func() {
		result = append(result, l.prefix...)
	})
	if l.endedWithNewLine {
		result = append(result, l.prefix...)
		l.endedWithNewLine = false
	}
	for i, b := range p {
		result = append(result, b)
		if b == NewLineByte {
			if i == len(p)-1 {
				l.endedWithNewLine = true
			} else {
				result = append(result, l.prefix...)
			}
		}
	}

	n, err = l.w.Write(result)
	if n == len(result) {
		// If everything was written, pretend we only wrote as much as asked to preserve cursor.
		return len(p), err
	}
	return n, err
}
