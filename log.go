package smush

import (
	"bytes"
	"fmt"
	"io"
	"sync"
)

type Logger struct {
	firstLine sync.Once
	w         io.Writer
	prefix    []byte

	endedWithNewLine bool
}

const NewLineByte = byte('\n')

func NewLogger(w io.Writer, prefix []byte) (*Logger, error) {
	dupPrefix := make([]byte, len(prefix))
	if n := copy(dupPrefix, prefix); n != len(prefix) {
		return nil, fmt.Errorf("cloning prefix: %d < %d", n, len(prefix))
	}

	return &Logger{
		w:      w,
		prefix: dupPrefix,
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
