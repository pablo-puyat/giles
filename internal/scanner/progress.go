package scanner

import (
	"giles/internal/database"
	"log"
)

func (s *Scanner) DisplayProgress(processed <-chan database.File) {
	for file := range processed {
		log.Printf("%s\tscanned\n", file.Name)
	}
}
