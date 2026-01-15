// +build windows

package main

import (
	"log"
	"os"
	"os/exec"
	"time"
)

func restartExecutable() {
	log.Println("Restarting process using os/exec (Windows fallback)...")
	
	exe, err := os.Executable()
	if err != nil {
		log.Printf("Failed to get executable path: %v", err)
		exe = os.Args[0] // Fallback to Args[0]
	}

	// On Windows, we can't replace the process. 
	// We'll start a new one and exit the current one.
	cmd := exec.Command(exe, os.Args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	err = cmd.Start()
	if err != nil {
		log.Printf("Failed to restart process: %v", err)
		os.Exit(1)
	}
	
	// Give a very short breath for the process to actually start
	time.Sleep(100 * time.Millisecond)
	os.Exit(0)
}
