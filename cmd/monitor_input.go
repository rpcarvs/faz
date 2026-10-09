package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/x/term"
	"github.com/muesli/cancelreader"
)

// monitorKeyboard owns terminal input and its restoration after monitor exits.
type monitorKeyboard struct {
	file   *os.File
	state  *term.State
	reader cancelreader.CancelReader
	quit   <-chan error
	done   <-chan struct{}
}

// newMonitorKeyboard enables immediate key input only when stdin is a terminal.
func newMonitorKeyboard(input io.Reader) (*monitorKeyboard, error) {
	file, ok := input.(*os.File)
	if !ok || !term.IsTerminal(file.Fd()) {
		return nil, nil
	}
	reader, err := cancelreader.NewReader(file)
	if err != nil {
		return nil, fmt.Errorf("open monitor input: %w", err)
	}
	state, err := term.MakeRaw(file.Fd())
	if err != nil {
		return nil, errors.Join(fmt.Errorf("configure monitor terminal: %w", err), reader.Close())
	}
	quit := make(chan error, 1)
	done := make(chan struct{})
	keyboard := &monitorKeyboard{file: file, state: state, reader: reader, quit: quit, done: done}
	go func() {
		defer close(done)
		defer close(quit)
		readMonitorKeys(reader, quit)
	}()
	return keyboard, nil
}

// Close cancels and joins the input reader, then restores the terminal state.
func (keyboard *monitorKeyboard) Close() error {
	keyboard.reader.Cancel()
	<-keyboard.done
	closeErr := keyboard.reader.Close()
	restoreErr := term.Restore(keyboard.file.Fd(), keyboard.state)
	if closeErr != nil {
		closeErr = fmt.Errorf("close monitor input: %w", closeErr)
	}
	if restoreErr != nil {
		restoreErr = fmt.Errorf("restore monitor terminal: %w", restoreErr)
	}
	return errors.Join(closeErr, restoreErr)
}

// readMonitorKeys reports q, Ctrl+C, or input errors; other keys are ignored.
func readMonitorKeys(reader io.Reader, quit chan<- error) {
	var key [1]byte
	for {
		if _, err := io.ReadFull(reader, key[:]); err != nil {
			if !errors.Is(err, io.EOF) && !errors.Is(err, cancelreader.ErrCanceled) {
				quit <- fmt.Errorf("read monitor input: %w", err)
			}
			return
		}
		if key[0] == 'q' || key[0] == '\x03' {
			quit <- nil
			return
		}
	}
}

// monitorTerminalWriter preserves line starts when raw mode disables newline conversion.
type monitorTerminalWriter struct {
	io.Writer
}

// Write converts newlines for raw terminal output and reports consumed source bytes.
func (writer monitorTerminalWriter) Write(data []byte) (int, error) {
	expanded := bytes.ReplaceAll(data, []byte("\n"), []byte("\r\n"))
	written, err := writer.Writer.Write(expanded)
	if err != nil {
		return 0, err
	}
	if written != len(expanded) {
		return 0, io.ErrShortWrite
	}
	return len(data), nil
}
