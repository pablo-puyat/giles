package organizer

import (
	"fmt"
	"giles/internal/database"
	"io"
	"os"
)

func NewOrganizer(sourceDir string) *Organizer {
	return &Organizer{sourceDir: sourceDir}
}

type Organizer struct {
	sourceDir string
}

func (ds *Organizer) OrganizeFiles(files []database.File, destination string) []database.File {
	for i, file := range files {
		n := 2
		if len(file.Hash) < n {
			n = len(file.Hash)
		}
		directory := file.Hash[:n]
		files[i].Name = file.Hash
		files[i].Path = destination + "/" + directory
		newName := files[i].Path + files[i].Name
		// verify that source files exists
		println(newName)
		// check if directory exists and create if it doesn't
		//if _, err := os.Stat("data/csv_data"); os.IsNotExist(err) {
		//	err := os.MkdirAll("data/json_data", 0755)
		//	if err != nil {
		//		fmt.Println(err)
		//	}
		//	fmt.Println("Directory Created Successfully")
		//} else {
		//	fmt.Println("Directory exists")
		//}
		// copy file
		//copyFile(files[i].Path + files[i].Name)
		// verify that file exists in new location

		// remove previous file

	}
	return files
}

func copyFile(source string, destination string) error {
	srcFile, err := os.Open(source)
	if err != nil {
		fmt.Println(err)
	}
	defer srcFile.Close()

	destFile, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return err
	}
	fmt.Println("Copy done!")
	return nil
}
