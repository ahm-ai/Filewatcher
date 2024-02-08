package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
)

var (
	currentCmd   *exec.Cmd
	commandMutex sync.Mutex
)

func expandPath(path string) (string, error) {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, strings.TrimPrefix(path, "~")), nil
	}
	return path, nil
}

func killProcessOnPort(port string) {

	if port == "" {
		log.Printf("Port is empty")
		return
	}
	// Find process using the port
	cmd := exec.Command("lsof", "-t", "-i:"+port)
	out, err := cmd.Output()
	if err != nil {
		log.Printf("Error finding process for port %s: %v", port, err)
		return
	}
	pid := strings.TrimSpace(string(out))
	if pid == "" {
		log.Printf("No process found listening on port %s", port)
		return
	}
	// Kill the process
	killCmd := exec.Command("kill", pid)
	err = killCmd.Run()
	if err != nil {
		log.Printf("Error killing process %s on port %s: %v", pid, port, err)
	} else {
		log.Printf("Successfully killed process %s on port %s", pid, port)
	}
}

func main() {
	appPort := flag.String("port", "3000", "Port to kill the process on")
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

	fmt.Printf("Watching directory: %s\n", *pathFlag)
	// Recursive function to watch directories
	watchDir := func(path string, fileInfo os.FileInfo, err error) error {
		// Skip node_modules directory

		if fileInfo.IsDir() {
			if fileInfo.Name() == "node_modules" {
				return filepath.SkipDir
			}
			fmt.Printf("\033[32mWatching directory: %s\033[0m\n", path)
			return watcher.Add(path)
		}
		return nil
	}

	watchPath, err := expandPath(*pathFlag)
	if err != nil {
		log.Fatalf("Error expanding path: %v", err)
	}

	// Traverse the directory and watch each subdirectory
	if err := filepath.Walk(watchPath, watchDir); err != nil {
		log.Fatal(err)
	}

	done := make(chan bool)
	// var cmd *exec.Cmd
	// commandRunning := false

	// Function to execute the command
	executeCommand := func() {
		commandMutex.Lock()
		defer commandMutex.Unlock()

		killProcessOnPort(*appPort)

		// Kill the previous command if it's still running
		if currentCmd != nil && currentCmd.Process != nil {
			fmt.Println("Killing the previous running command.")
			_ = currentCmd.Process.Kill()
		}

		fmt.Println("Executing command:", *commandFlag)
		currentCmd = exec.Command("zsh", "-c", fmt.Sprintf("ZSH_DISABLE_COMPFIX=true; source ~/.zshrc; %s", *commandFlag))
		currentCmd.Dir = watchPath // Set the working directory

		stdoutPipe, err := currentCmd.StdoutPipe()
		if err != nil {
			log.Printf("Error creating stdout pipe: %v", err)
			return
		}
		stderrPipe, err := currentCmd.StderrPipe()
		if err != nil {
			log.Printf("Error creating stderr pipe: %v", err)
			return
		}

		if err := currentCmd.Start(); err != nil {
			log.Printf("Error executing command: %v", err)
			return
		} else {
			fmt.Println("Command started successfully.")
		}

		// Read and log stdout
		go func() {
			scanner := bufio.NewScanner(stdoutPipe)
			for scanner.Scan() {
				fmt.Printf("%s\n", scanner.Text())
			}
		}()

		// Read and log stderr
		go func() {
			scanner := bufio.NewScanner(stderrPipe)
			for scanner.Scan() {
				fmt.Printf("%s\n", scanner.Text())
			}
		}()

		// Wait for the command to finish
		if err := currentCmd.Wait(); err != nil {
			log.Printf("Command finished with error: %v", err)
		} else {
			fmt.Println("Command completed successfully.")
		}

		// Reset currentCmd after command completion
		currentCmd = nil
	}

	// Execute the command at the start
	go executeCommand()

	// Start listening for events
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
						fmt.Printf("Modified file detected: %s\n", event.Name)

						// commandMutex.Lock()
						if currentCmd != nil && currentCmd.Process != nil {
							fmt.Println("Killing the currently running command.")
							if err := currentCmd.Process.Kill(); err != nil {
								fmt.Println("Error killing the process:", err)
							} else {
								fmt.Println("Command killed successfully.")
							}
							currentCmd = nil
						}
						// commandMutex.Unlock()

						go executeCommand() // Restart the command in a new goroutine
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
