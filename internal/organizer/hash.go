package organizer

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"giles/internal/database"
	"io"
	"log"
	"os"
	"path/filepath"
)

func ByHash(files []database.File, dest string) error {
	for i := range files {
		file := &files[i]

		prefixLen := 2
		if len(file.Hash) < prefixLen {
			prefixLen = len(file.Hash)
		}
		dirPrefix := file.Hash[:prefixLen]
		newPath := filepath.Join(dest, dirPrefix)

		if err := os.MkdirAll(newPath, 0755); err != nil {
			return fmt.Errorf("create directory %s: %w", newPath, err)
		}

		destPath := filepath.Join(newPath, file.Hash)

		err := os.Rename(file.Path, destPath)
		if err == nil {
			log.Printf("Moved %s to %s", file.Path, destPath)
			continue
		}

		if err := copyAndVerify(file.Path, destPath, file.Hash); err != nil {
			return fmt.Errorf("copy and verify %s: %w", destPath, err)
		}

		if err := os.Remove(file.Path); err != nil {
			return fmt.Errorf("remove original %s: %w", file.Path, err)
		}

		file.Name = file.Hash
		file.Path = newPath
	}
	return nil
}

func copyAndVerify(src, dest, expectedHash string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open source: %w", err)
	}
	defer srcFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("create destination: %w", err)
	}
	defer destFile.Close()

	h := sha256.New()
	if _, err := io.Copy(io.MultiWriter(destFile, h), srcFile); err != nil {
		os.Remove(dest) // Clean up on copy failure
		return fmt.Errorf("copy file: %w", err)
	}

	newHash := hex.EncodeToString(h.Sum(nil))
	if newHash != expectedHash {
		os.Remove(dest) // Clean up on hash mismatch
		return fmt.Errorf("hash mismatch: got %s, want %s", newHash, expectedHash)
	}

	return nil
}
