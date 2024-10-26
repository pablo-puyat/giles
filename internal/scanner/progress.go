package scanner

import (
	"fmt"
	"sync/atomic"
	"time"
)

type Progress struct {
	TotalFiles   int64
	ScannedFiles int64
}

func (s *Scanner) DisplayProgress(done chan bool) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			scanned := atomic.LoadInt64(&s.Progress.ScannedFiles)
			total := atomic.LoadInt64(&s.Progress.TotalFiles)
			if total > 0 {
				fmt.Printf("\rProgress: %d/%d files scanned (%.1f%%)",
					scanned, total,
					float64(scanned)/float64(total)*100)
			}
		case <-done:
			fmt.Println()
			return
		}
	}
}
