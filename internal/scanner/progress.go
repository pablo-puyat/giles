package scanner

import (
	"fmt"
	"sync/atomic"
	"time"
)

type Progress struct {
	ScannedFiles int64
}

func (s *Scanner) DisplayProgress(done chan bool) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			scanned := atomic.LoadInt64(&s.Progress.ScannedFiles)
			if scanned > 0 {
				fmt.Printf("\rProgress: %d files scanned", scanned)
			}

		case <-done:
			scanned := atomic.LoadInt64(&s.Progress.ScannedFiles)
			fmt.Printf("\rProgress: %d files scanned\n", scanned)
			fmt.Println("Done")
			return
		}
	}
}
