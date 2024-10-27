//go:build mage

package main

import (
	"fmt"
	"github.com/magefile/mage/sh"
)

func Build() error {
	if err := sh.Run("go", "mod", "download"); err != nil {
		return err
	}
	sh.Run("go", "build", ".")
	completed()
	return nil
}

func Format() error {
	if err := sh.Run("go", "fmt"); err != nil {
		return err
	}
	completed()
	return nil
}

func completed() {
	fmt.Println("Completed.")
}
