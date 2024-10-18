package organizer

import (
	"fmt"
	"giles/models"
)

func NewOrganizer(sourceDir string) *Organizer {
	return &Organizer{sourceDir: sourceDir}
}

type Organizer struct {
	sourceDir string
}

func (ds *Organizer) OrganizeFiles(files []models.FileData, destinationDir string) {
	for _, file := range files {
		fmt.Printf("%s", file.Name)
	}
	fmt.Printf("%s", destinationDir)
}
