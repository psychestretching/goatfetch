package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	reset      = "\033[0m"
	bold       = "\033[1m"
	green      = "\033[32m"
	cyan       = "\033[36m"
	colorBlock = "███"
)

func main() {
	user := getUsername()
	host := getHostname()
	shell := getShell()
	kernel := getKernel()
	osName := getOSName()
	arch := runtime.GOARCH

	// Calculate underline length
	userHost := user + "@" + host
	userHostLen := len(userHost)

	// Header with centered formatting
	fmt.Println()
	fmt.Printf("%s%s%s%s@%s%s%s\n", 
		strings.Repeat(" ", 19), bold+green, user, reset,
		bold+green, host, reset)
	fmt.Printf("%s%s%s\n",
		strings.Repeat(" ", 19),
		strings.Repeat("-", userHostLen),
		strings.Repeat(" ", 2))

	// System information with aligned formatting
	fmt.Printf("  %s(_(%s             %sOS:%s %s\n", bold, reset, cyan, reset, osName)
	fmt.Printf("  %s/_/'______/)%s    %sKernel:%s %s\n", bold, reset, cyan, reset, kernel)
	fmt.Printf("  %s\"  |      |%s     %sShell:%s %s\n", bold, reset, cyan, reset, shell)
	fmt.Printf("  %s   |\"\"\"\"\"\"|%s     %sArch:%s %s\n", bold, reset, cyan, reset, arch)

	// Color blocks
	printColorBlocks()
}

func getUsername() string {
	if user := os.Getenv("USER"); user != "" {
		return user
	}
	if user := os.Getenv("LOGNAME"); user != "" {
		return user
	}
	if runtime.GOOS == "windows" {
		return os.Getenv("USERNAME")
	}
	return "unknown"
}

func getHostname() string {
	host, err := os.Hostname()
	if err == nil {
		// Remove domain if present
		return strings.Split(host, ".")[0]
	}
	
	// Fallback for Windows
	if runtime.GOOS == "windows" {
		cmd := exec.Command("hostname")
		output, err := cmd.Output()
		if err == nil {
			return strings.TrimSpace(string(output))
		}
	}
	return "unknown"
}

func getShell() string {
	shellPath := os.Getenv("SHELL")
	if shellPath == "" {
		if runtime.GOOS == "windows" {
			return filepath.Base(os.Getenv("ComSpec"))
		}
		return "unknown"
	}
	return filepath.Base(shellPath)
}

func getKernel() string {
	switch runtime.GOOS {
	case "linux", "darwin", "freebsd", "openbsd":
		cmd := exec.Command("uname", "-rs")
		output, err := cmd.Output()
		if err == nil {
			return strings.TrimSpace(string(output))
		}
	case "windows":
		cmd := exec.Command("cmd", "/c", "ver")
		output, err := cmd.Output()
		if err == nil {
			ver := strings.TrimSpace(string(output))
			// Clean up Windows version output
			ver = strings.TrimPrefix(ver, "Microsoft Windows [Version ")
			ver = strings.TrimSuffix(ver, "]")
			return ver
		}
	}
	return runtime.GOOS
}

func getOSName() string {
	// Try Linux/BSD release files
	osFiles := []string{
		"/etc/os-release",
		"/usr/lib/os-release",
		"/etc/openwrt_release",
		"/etc/lsb-release",
	}
	
	for _, file := range osFiles {
		if name := parseOSRelease(file); name != "" {
			return name
		}
	}

	// macOS detection
	if runtime.GOOS == "darwin" {
		cmd := exec.Command("sw_vers", "-productName")
		output, err := cmd.Output()
		if err == nil {
			name := strings.TrimSpace(string(output))
			cmd = exec.Command("sw_vers", "-productVersion")
			output, err = cmd.Output()
			if err == nil {
				return name + " " + strings.TrimSpace(string(output))
			}
			return name
		}
	}

	// Windows WSL detection
	if wsl := os.Getenv("WSL_DISTRO_NAME"); wsl != "" {
		return "WSL: " + wsl
	}

	// Android detection
	if _, err := os.Stat("/system/bin/adb"); err == nil {
		return "Android"
	}

	return runtime.GOOS
}

func parseOSRelease(filename string) string {
	file, err := os.Open(filename)
	if err != nil {
		return ""
	}
	defer file.Close()

	prettyName := ""
	name := ""

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "PRETTY_NAME="):
			value := strings.SplitN(line, "=", 2)[1]
			prettyName = strings.Trim(value, `"`)
		case strings.HasPrefix(line, "NAME="):
			value := strings.SplitN(line, "=", 2)[1]
			name = strings.Trim(value, `"`)
		}
	}
	
	if prettyName != "" {
		return prettyName
	}
	return name
}

func printColorBlocks() {
	// Regular colors
	fmt.Print(strings.Repeat(" ", 19))
	for i := 30; i <= 37; i++ {
		fmt.Printf("\033[%dm%s", i, colorBlock)
	}
	fmt.Println()

	// Bright colors
	fmt.Print(strings.Repeat(" ", 19))
	for i := 90; i <= 97; i++ {
		fmt.Printf("\033[%dm%s", i, colorBlock)
	}
	fmt.Printf("%s\n\n", reset)
}
