package scanner

import (
	"fmt"
	"log"
	"sync/atomic"
	"time"
)

type Progress struct {
	ScannedFiles int64
}

func (s *Scanner) DisplayProgress(done chan struct{}) {
	log.Println("Progress display started")

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		// Check for done signal first
		select {
		case <-done:
			scanned := atomic.LoadInt64(&s.Progress.ScannedFiles)
			fmt.Printf("\rProgress: %d files scanned\n", scanned)
			fmt.Println("Done")
			return
		default:
			// Update progress if not done
			scanned := atomic.LoadInt64(&s.Progress.ScannedFiles)
			if scanned > 0 {
				fmt.Printf("\rProgress: %d files scanned", scanned)
			}
		}

		// Wait for next tick
		<-ticker.C
	}
}
