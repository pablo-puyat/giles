//go:build mage

package main

import (
	"fmt"
	"github.com/magefile/mage/sh"
)

// runs go mod download and go build .
func Build() error {
	if err := sh.Run("go", "mod", "download"); err != nil {
		return err
	}
	sh.Run("go", "build", ".")
	completed()
	return nil
}

// formats the source files using go fmt
func Format() error {
	if err := sh.Run("go", "fmt"); err != nil {
		return err
	}
	completed()
	return nil
}

// BuildStatic creates a statically linked binary for ARM processors using musl
func BuildStaticArm() error {
	env := map[string]string{
		"CC":          "x86_64-linux-musl-gcc",
		"CXX":         "x86_64-linux-musl-g++",
		"GOARCH":      "amd64",
		"GOOS":        "linux",
		"CGO_ENABLED": "1",
	}

	if err := sh.RunWith(env, "go", "build",
		"-ldflags", "-linkmode external -extldflags -static"); err != nil {
		return err
	}
	completed()
	return nil
}

// Watch reruns the program when Go files change
func Watch() error {
	// Using sh.RunV to show the find and entr output
	if err := sh.RunV("sh", "-c", "find . -name \"*.go\" | entr -rc go run ."); err != nil {
		return err
	}
	completed()
	return nil
}

func completed() {
	fmt.Println("Completed.")
}
