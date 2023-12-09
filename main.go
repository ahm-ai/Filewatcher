package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	"github.com/fsnotify/fsnotify"
)

func main() {
	pathFlag := flag.String("path", "", "Path to the folder to watch")
	regexFlag := flag.String("regex", "", "Regex pattern to match file names")
	commandFlag := flag.String("command", "", "Command to run on file change")

	flag.Parse()

	if *pathFlag == "" || *regexFlag == "" || *commandFlag == "" {
		log.Fatal("All flags 'path', 'regex', and 'command' are required")
	}

	re, err := regexp.Compile(string(*regexFlag))
	if err != nil {
		log.Fatalf("Error compiling regex: %v", err)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// Recursive function to watch directories
	watchDir := func(path string, fileInfo os.FileInfo, err error) error {
		// Skip node_modules directory
		if fileInfo.IsDir() {
			if fileInfo.Name() == "node_modules" {
				return filepath.SkipDir
			}
			return watcher.Add(path)
		}
		return nil
	}

	// Traverse the directory and watch each subdirectory
	if err := filepath.Walk(*pathFlag, watchDir); err != nil {
		log.Fatal(err)
	}

	done := make(chan bool)
	var cmd *exec.Cmd
	commandRunning := false

	// Function to execute the command
	executeCommand := func() {
		// If a command is already running, kill it
		if commandRunning && cmd != nil && cmd.Process != nil {
			_ = cmd.Process.Kill()
		}

		cmd = exec.Command("zsh", "-c", fmt.Sprintf("ZSH_DISABLE_COMPFIX=true; source ~/.zshrc; %s", *commandFlag))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		commandRunning = true
		err := cmd.Start()
		if err != nil {
			log.Printf("Error executing command: %v", err)
			commandRunning = false
		}
	}

	// Execute the command at the start
	executeCommand()

	// Start listening for events
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					if re.MatchString(event.Name) {
						fmt.Printf("Modified file: %s\n", event.Name)
						executeCommand()
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("Error:", err)
			}
		}
	}()

	<-done
}
