package smush

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"

	"golang.org/x/sync/semaphore"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Commands []*Command `yaml:"commands"`
}

type Command struct {
	Name string `yaml:"name"`
	Runs string `yaml:"runs"`

	cmd *exec.Cmd
}

type Failure struct {
	Command *Command
	Error   error
	Index   int
}

func ReadConfig(r io.Reader) (*Config, error) {
	data := &Config{}
	if err := yaml.NewDecoder(r).Decode(data); err != nil {
		return nil, fmt.Errorf("decoding yaml: %w", err)
	}
	return data, nil
}

func (c *Command) Label() string {
	if c.Name != "" {
		return c.Name
	}

	return strings.Split(c.Runs, " ")[0]
}

func (c *Command) Run(ctx context.Context, stdout, stderr io.Writer) error {
	cmdArgs := strings.Split(c.Runs, " ")
	c.cmd = exec.CommandContext(ctx, cmdArgs[0], cmdArgs[1:]...)
	c.cmd.Stdout = stdout
	c.cmd.Stderr = stderr

	err := c.cmd.Run()
	exitCode := 0
	if err != nil {
		exitErr, ok := err.(*exec.ExitError)
		if !ok {
			exitCode = -99
		} else {
			exitCode = exitErr.ExitCode()
		}
	}

	if stderr != nil {
		str := color.New(color.Bold)
		str.Fprintf(stderr, ">>> exited %d\n", exitCode)
	}

	if err != nil {
		return fmt.Errorf("running command '%s': %w", c.Runs, err)
	}
	return nil
}

func RunAll(ctx context.Context, maxProcs int64, commands []*Command) error {
	throttle := semaphore.NewWeighted(maxProcs)
	failures := make(chan *Failure, len(commands))
	// Set minimum to 3, matching the anchor used for non-program logs.
	leftpad := 3
	for _, command := range commands {
		label := command.Label()
		if len(label) > leftpad {
			leftpad = len(label)
		}
	}
	leftpad++

	for i, command := range commands {
		if err := throttle.Acquire(ctx, 1); err != nil {
			return fmt.Errorf("acquire semaphore: %w", err)
		}

		go func(cmd *Command, i int) {
			defer throttle.Release(1)

			label := fmt.Sprintf("%*s |> ", leftpad, cmd.Label())
			stdout := NewLogger(os.Stdout, label, i)
			stderr := NewLogger(os.Stderr, label, i)
			if err := cmd.Run(ctx, stdout, stderr); err != nil {
				writeNewlineAtEnd(stderr, err)
				failures <- &Failure{
					Command: cmd,
					Error:   err,
					Index:   i,
				}
			}
		}(command, i)
	}
	go func() {
		// Ensure all processes have exited by acquiring maximum weight of semaphore...
		_ = throttle.Acquire(context.Background(), maxProcs)
		// ...before finally closing the channel, so the channel receiver doesn't terminate early.
		close(failures)
	}()

	// Buffer the report so that all threads can exit first.
	var report bytes.Buffer
	report.WriteRune('\n')
	report.WriteRune('\n')

	rw := &Logger{
		w:      &report,
		prefix: []byte(fmt.Sprintf("%*s |> ", leftpad, "***")),
	}
	color.New(color.Bold).Fprint(rw, "RUN REPORT")
	fmt.Fprintf(rw, "\n")
	var f int
	for failure := range failures {
		if failure.Error == nil {
			continue
		}

		f++
		label := fmt.Sprintf("%*s |> ", leftpad, failure.Command.Label())
		if !errors.Is(failure.Error, context.Canceled) {
			writeNewlineAtEnd(
				NewLogger(&report, label, failure.Index),
				failure.Error)
		}
	}
	total := len(commands)
	pass := total - f
	fail := f
	if f == 0 {
		fmt.Fprintf(rw, "Ran %d total commands.\n", total)
		fmt.Fprint(rw, color.New(color.Bold, color.FgGreen).Sprint("PASS"))
	} else {
		fmt.Fprintf(rw, "Total: %d, Pass: %d, Fail: %d (%0.1f%%)\n",
			total,
			pass,
			fail,
			(float64(pass)/float64(total))*100.)
		fmt.Fprint(rw, color.New(color.Bold, color.FgRed).Sprint("FAIL"))
	}

	fmt.Fprint(&report, "\n")
	fmt.Fprint(os.Stderr, report.String())

	if f > 0 {
		return fmt.Errorf("at least one command exited with error")
	}
	return nil
}

func writeNewlineAtEnd(w io.Writer, err error) {
	if err == nil {
		return
	}

	msg := err.Error()
	if len(msg) == 0 {
		fmt.Fprint(w, "\n")
		return
	}

	switch msg[len(msg)-1] {
	case '\n', '\r':
		fmt.Fprint(w, msg)
		return
	}

	fmt.Fprintf(w, "%s\n", msg)
}
