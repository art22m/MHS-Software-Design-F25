package main

import (
	"syscall"

	"github.com/art22m/MHS-Software-Design-F25/gocli/internal/shell"
)

func main() {
	shell := shell.NewShell()
	exitCode := shell.Run()
	syscall.Exit(exitCode)
}
