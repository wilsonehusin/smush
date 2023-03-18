package smush_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"go.husin.dev/smush"
)

func TestReadConfigValidData(t *testing.T) {
	rawYaml := `---
commands:
  - name: "one"
    runs: "echo one"
  - name: "two"
    runs: "echo two"`
	r := strings.NewReader(rawYaml)
	c, err := smush.ReadConfig(r)
	if err != nil {
		t.Fatalf("received error: %v", err)
	}

	compareValues(t, 2, len(c.Commands), c.Commands)

	compareValues(t, "one", c.Commands[0].Name, c.Commands)
	compareValues(t, "echo one", c.Commands[0].Runs, c.Commands)

	compareValues(t, "two", c.Commands[1].Name, c.Commands)
	compareValues(t, "echo two", c.Commands[1].Runs, c.Commands)
}

func TestCommandRunSuccessWithStdout(t *testing.T) {
	cmd := smush.Command{Name: "hello", Runs: "echo hello"}

	var stdoutByte []byte
	stdout := bytes.NewBuffer(stdoutByte)

	var stderrByte []byte
	stderr := bytes.NewBuffer(stderrByte)

	if err := cmd.Run(context.Background(), stdout, stderr); err != nil {
		t.Fatalf("received error: %v", err)
	}

	compareValues(t, "hello\n", stdout.String(), cmd)
	compareValues(t, ">>> exited 0\n", stderr.String(), cmd)
}
