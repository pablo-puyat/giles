package hash

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"os"
)

func Calculate(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		log.Printf("Error opening file: %v\n", err)
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Printf("Error hashing file: %v\n", err)
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
