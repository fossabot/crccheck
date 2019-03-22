// +build mage

package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// Default target to run when none is specified
// If not set, running mage will list available targets
var Default = Build

// A build step that requires additional params, or platform specific steps for example
func Build() error {
	mg.Deps(VerifyDeps)
	fmt.Println("Building...")
	return sh.Run("go", "build", "-ldflags", "-s -w", ".")
}

// A custom install step if you need your bin someplace other than go/bin
func Install() error {
	mg.Deps(VerifyDeps, Lint, Test)
	fmt.Println("Installing...")
	return sh.Run("go", "install", "-ldflags", "-s -w", ".")
}

// Manage your deps, or running package managers.
func InstallDeps() error {
	fmt.Println("Installing Deps...")
	return sh.Run("go", "mod", "download")
}

// VerifyDeps Verify the downloaded dependencies.
func VerifyDeps() error {
	mg.Deps(InstallDeps)
	fmt.Println("Verifying Deps...")
	return sh.Run("go", "mod", "verify")
}

// Clean Clean up after yourself
func Clean() {
	fmt.Println("Cleaning...")
	os.RemoveAll("crccheck")
}

func Lint() error {
	fmt.Println("Linting Code...")
	cmd := exec.Command("golangci-lint", "run")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func Test() error {
	mg.Deps(VerifyDeps)
	fmt.Println("Running Tests...")
	return sh.Run("go", "test", "./...")
}
