package scanner

import (
	"fmt"
	"sync/atomic"
	"time"
)

type Progress struct {
	ScannedFiles int64
}

func (s *Scanner) DisplayProgress(done chan struct{}) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			scanned := atomic.LoadInt64(&s.Progress.ScannedFiles)
			fmt.Printf("\rProgress: %d files scanned\n", scanned)
			fmt.Println("Done")
			return
		default:
			scanned := atomic.LoadInt64(&s.Progress.ScannedFiles)
			if scanned > 0 {
				fmt.Printf("\rProgress: %d files scanned", scanned)
			}
		}

		// Wait for next tick
		<-ticker.C
	}
}
