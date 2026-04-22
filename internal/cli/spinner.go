package cli

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/mattn/go-isatty"
)

var interactive = isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())

var frames = [...]string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

type spinner struct {
	mu     sync.Mutex
	msg    string
	stopCh chan struct{}
	doneCh chan struct{}
}

func newSpinner(msg string) *spinner {
	s := &spinner{
		msg:    msg,
		stopCh: make(chan struct{}),
		doneCh: make(chan struct{}),
	}

	if interactive {
		go s.run()
	}

	return s
}

func (s *spinner) run() {
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

			fmt.Printf("\r  %s %s\033[K", yellowBold(frames[idx]), msg)

			idx = (idx + 1) % len(frames)
		}
	}
}

func (s *spinner) update(msg string) {
	s.mu.Lock()
	s.msg = msg
	s.mu.Unlock()
}

func (s *spinner) stop() {
	if !interactive {
		return
	}

	close(s.stopCh)
	<-s.doneCh
}
