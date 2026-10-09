package cmd

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/muesli/cancelreader"
	"github.com/rpcarvs/faz/internal/model"
)

// TestReadMonitorKeys checks immediate quit, ignored keys, EOF, and input failure.
func TestReadMonitorKeys(t *testing.T) {
	inputErr := errors.New("input failed")
	tests := []struct {
		name   string
		reader io.Reader
		want   error
		quit   bool
	}{
		{name: "q without enter", reader: strings.NewReader("q"), quit: true},
		{name: "Ctrl+C", reader: strings.NewReader("\x03"), quit: true},
		{name: "other keys before quit", reader: strings.NewReader("abc\nq"), quit: true},
		{name: "EOF", reader: strings.NewReader("abc")},
		{name: "canceled read", reader: monitorErrorReader{cancelreader.ErrCanceled}},
		{name: "input error", reader: monitorErrorReader{inputErr}, quit: true, want: inputErr},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			quit := make(chan error, 1)
			readMonitorKeys(test.reader, quit)
			select {
			case err := <-quit:
				if !test.quit || !errors.Is(err, test.want) {
					t.Fatalf("quit result = %v, want quit=%v, error=%v", err, test.quit, test.want)
				}
			default:
				if test.quit {
					t.Fatal("expected quit notification")
				}
			}
		})
	}
}

// TestMonitorKeyboardSkipsNonterminalInput ensures pipes never enable raw mode.
func TestMonitorKeyboardSkipsNonterminalInput(t *testing.T) {
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = reader.Close()
		_ = writer.Close()
	})
	for _, input := range []io.Reader{reader, strings.NewReader("q")} {
		keyboard, err := newMonitorKeyboard(input)
		if err != nil || keyboard != nil {
			t.Fatalf("nonterminal keyboard = %v, error = %v, want disabled input", keyboard, err)
		}
	}
}

// TestRunMonitorStopsOnInputAndCancellation checks shutdown without database events.
func TestRunMonitorStopsOnInputAndCancellation(t *testing.T) {
	inputErr := errors.New("input failed")
	for _, name := range []string{"q", "context", "input error", "closed input"} {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			svc := newMonitorStubService([][]model.Issue{{}})
			quit := make(chan error)
			events := make(chan fsnotify.Event)
			done := make(chan error, 1)
			go func() {
				done <- runMonitor(ctx, svc, model.ListFilter{}, &bytes.Buffer{}, quit,
					events, make(chan error), make(chan time.Time), time.Millisecond)
			}()
			waitForMonitorListCall(t, svc)
			var want error
			switch name {
			case "q":
				quit <- nil
			case "context":
				cancel()
			case "input error":
				want = inputErr
				quit <- inputErr
			case "closed input":
				close(quit)
				events <- fsnotify.Event{}
				waitForMonitorListCall(t, svc)
				close(events)
			}
			waitForMonitorResult(t, done, want)
		})
	}
}

// TestMonitorTerminalWriter checks raw-mode line starts and output errors.
func TestMonitorTerminalWriter(t *testing.T) {
	data := []byte("first\nsecond\n")
	var output bytes.Buffer
	writer := monitorTerminalWriter{&output}
	written, err := writer.Write(data)
	if err != nil || written != len(data) || output.String() != "first\r\nsecond\r\n" {
		t.Fatalf("terminal write = %d, %v, %q", written, err, output.String())
	}
	outputErr := errors.New("output failed")
	for _, test := range []struct {
		writer io.Writer
		want   error
	}{
		{monitorErrorWriter{outputErr}, outputErr},
		{monitorErrorWriter{}, io.ErrShortWrite},
	} {
		if _, err := (monitorTerminalWriter{test.writer}).Write(data); !errors.Is(err, test.want) {
			t.Fatalf("write error = %v, want %v", err, test.want)
		}
	}
}

// monitorErrorReader injects a deterministic input failure.
type monitorErrorReader struct{ err error }

// Read returns the configured error without consuming input.
func (reader monitorErrorReader) Read([]byte) (int, error) { return 0, reader.err }

// monitorErrorWriter injects a deterministic write failure or short write.
type monitorErrorWriter struct{ err error }

// Write returns the configured error without accepting output.
func (writer monitorErrorWriter) Write([]byte) (int, error) { return 0, writer.err }
