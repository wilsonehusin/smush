package smush

import (
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
