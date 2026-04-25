package ui

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/mattn/go-isatty"
)

// Interactive reports whether stdout is a terminal.
var Interactive = isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())

var frames = [...]string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// Spinner displays an animated progress indicator.
type Spinner struct {
	mu     sync.Mutex
	msg    string
	stopCh chan struct{}
	doneCh chan struct{}
}

// NewSpinner starts a new animated spinner with the given message.
func NewSpinner(msg string) *Spinner {
	s := &Spinner{
		msg:    msg,
		stopCh: make(chan struct{}),
		doneCh: make(chan struct{}),
	}

	if Interactive {
		go s.run()
	}

	return s
}

func (s *Spinner) run() {
	defer close(s.doneCh)

	idx := 0
	ticker := time.NewTicker(80 * time.Millisecond)

	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			fmt.Print("\r\033[K")

			return
		case <-ticker.C:
			s.mu.Lock()
			msg := s.msg
			s.mu.Unlock()

			fmt.Printf("\r  %s %s\033[K", YellowBold(frames[idx]), msg)

			idx = (idx + 1) % len(frames)
		}
	}
}

// Update changes the spinner message.
func (s *Spinner) Update(msg string) {
	s.mu.Lock()
	s.msg = msg
	s.mu.Unlock()
}

// Stop halts the spinner animation.
func (s *Spinner) Stop() {
	if !Interactive {
		return
	}

	close(s.stopCh)
	<-s.doneCh
}
