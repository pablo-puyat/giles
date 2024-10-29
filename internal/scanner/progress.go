package scanner

import (
	"fmt"
	"giles/internal/database"
)

type Progress struct {
	ScannedFiles int64
}

func (s *Scanner) DisplayProgress(processed <-chan database.File) {
	for file := range processed {
		fmt.Printf("\rScanned: %s ", file.Name)
	}
}
