// +build !windows

package main

import (
	"log"
	"os"
	"syscall"
)

func restartExecutable() {
	log.Println("Restarting process using syscall.Exec (process replacement)...")
	
	exe, err := os.Executable()
	if err != nil {
		log.Printf("Failed to get executable path: %v", err)
		exe = os.Args[0] // Fallback to Args[0]
	}

	err = syscall.Exec(exe, os.Args, os.Environ())
	if err != nil {
		log.Printf("Failed to exec new process: %v", err)
		os.Exit(1)
	}
}
