package organizer

import (
	"giles/models"
)

func NewOrganizer(sourceDir string) *Organizer {
	return &Organizer{sourceDir: sourceDir}
}

type Organizer struct {
	sourceDir string
}

func (ds *Organizer) OrganizeFiles(files []models.FileData, destination string) []models.FileData {
	for i, file := range files {
		n := 2
		if len(file.Hash) < n {
			n = len(file.Hash)
		}
		directory := file.Hash[:n]
		files[i].Name = file.Hash
		files[i].Path = destination + "/" + directory
	}
	return files
}
