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
	for i := 0; i < len(files); i++ {
		dirPrefix := files[i].Hash[:1]
		newPath := filepath.Join(dest, dirPrefix)

		if err := os.MkdirAll(newPath, 0755); err != nil {
			return fmt.Errorf("create directory %s: %w", newPath, err)
		}
		destPath := filepath.Join(newPath, files[i].Hash)
		err := os.Rename(files[i].Path, destPath)

		if err != nil {
			// Copy and verify since rename failed
			if err := copyAndVerify(files[i].Path, destPath, files[i].Hash); err != nil {
				return fmt.Errorf("copy and verify %s: %w", destPath, err)
			}
			// Remove original file after successful copy
			if err := os.Remove(files[i].Path); err != nil {
				return fmt.Errorf("remove original %s: %w", files[i].Path, err)
			}
		} else {
			log.Printf("Moved %s to %s", files[i].Path, destPath)
		}

		files[i].Name = files[i].Hash
		files[i].Path = newPath
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
		return fmt.Errorf("error copying file: %w", err)
	}

	newHash := hex.EncodeToString(h.Sum(nil))

	if newHash != expectedHash {
		os.Remove(dest) // Clean up on hash mismatch
		return fmt.Errorf("hash mismatch: got %s, want %s", newHash, expectedHash)
	}

	return nil
}
