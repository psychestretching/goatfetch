package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
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
	uptime := getUptime()
	processes := getProcessCount()
	terminal := getTerminal()
	mem := getMemInfo()
	cpuCores := getCPUCores()

	info := []string{
		fmt.Sprintf("%sOS:%s %s", cyan, reset, osName),
		fmt.Sprintf("%sHost:%s %s", cyan, reset, host),
		fmt.Sprintf("%sUser:%s %s", cyan, reset, user),
		fmt.Sprintf("%sKernel:%s %s", cyan, reset, kernel),
		fmt.Sprintf("%sArch:%s %s", cyan, reset, arch),
		fmt.Sprintf("%sShell:%s %s", cyan, reset, shell),
		fmt.Sprintf("%sUptime:%s %s", cyan, reset, uptime),
		fmt.Sprintf("%sProcs:%s %s", cyan, reset, processes),
		fmt.Sprintf("%sCores:%s %s", cyan, reset, cpuCores),
		fmt.Sprintf("%sTerminal:%s %s", cyan, reset, terminal),
		fmt.Sprintf("%sMemory:%s %s\n", cyan, reset, mem),
	}

	goatLines := []string{
		"  (_(",
		"  /_/'______/)",
		"  \"  |      |",
		"     |\"\"\"\"\"\"|",
		"",
	}

	// Compute max visual width (not counting ANSI)
	asciiWidth := 0
	for _, l := range goatLines {
		if w := visibleLength(l); w > asciiWidth {
			asciiWidth = w
		}
	}

	// Center header above both columns
	userHost := fmt.Sprintf("%s@%s", user, host)
	infoWidth := maxInfoLineLen(info)
	totalWidth := asciiWidth + 2 + infoWidth
	centerPad := (totalWidth - len(userHost)) / 2
	if centerPad < 0 {
		centerPad = 0
	}
	fmt.Printf("%s%s%s%s%s%s%s\n",
		strings.Repeat(" ", centerPad),
		bold+green, user, reset,
		bold+green, "@"+host, reset)
	fmt.Printf("%s%s\n", strings.Repeat(" ", centerPad), strings.Repeat("-", len(userHost)+1))

	// Main output
	lines := max(len(goatLines), len(info))
	for i := 0; i < lines; i++ {
		ascii := ""
		if i < len(goatLines) {
			ascii = bold + goatLines[i] + reset
		}
		infoStr := ""
		if i < len(info) {
			infoStr = info[i]
		}
		fmt.Printf("%s%s  %s\n", ascii, strings.Repeat(" ", asciiWidth-visibleLength(goatLinesAt(i, goatLines))), infoStr)
	}

	printColorBlocks(asciiWidth)
}

func goatLinesAt(i int, goatLines []string) string {
	if i < len(goatLines) {
		return goatLines[i]
	}
	return ""
}

func visibleLength(str string) int {
	// Strips ANSI codes and returns rune length
	out := ""
	inEsc := false
	for _, r := range str {
		if r == 0x1b {
			inEsc = true
			continue
		}
		if inEsc {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEsc = false // End of ANSI code
			}
			continue
		}
		out += string(r)
	}
	return utf8.RuneCountInString(out)
}

func maxInfoLineLen(info []string) int {
	maxLen := 0
	for _, s := range info {
		l := visibleLength(s)
		if l > maxLen {
			maxLen = l
		}
	}
	return maxLen
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
		return strings.Split(host, ".")[0]
	}
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
			ver = strings.TrimPrefix(ver, "Microsoft Windows [Version ")
			ver = strings.TrimSuffix(ver, "]")
			return ver
		}
	}
	return runtime.GOOS
}
func getOSName() string {
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
	if wsl := os.Getenv("WSL_DISTRO_NAME"); wsl != "" {
		return "WSL: " + wsl
	}
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

func getUptime() string {
	switch runtime.GOOS {
	case "linux":
		if b, err := os.ReadFile("/proc/uptime"); err == nil {
			parts := strings.Fields(string(b))
			if len(parts) > 0 {
				if sec, err := strconv.ParseFloat(parts[0], 64); err == nil {
					d := time.Duration(sec) * time.Second
					return formatDuration(d)
				}
			}
		}
	case "darwin":
		cmd := exec.Command("sysctl", "-n", "kern.boottime")
		output, err := cmd.Output()
		if err == nil {
			parts := strings.Split(string(output), " ")
			for _, part := range parts {
				if strings.HasPrefix(part, "sec=") {
					sec, _ := strconv.ParseInt(strings.TrimPrefix(part, "sec="), 10, 64)
					uptime := time.Since(time.Unix(sec, 0))
					return formatDuration(uptime)
				}
			}
		}
	}
	return "unknown"
}
func formatDuration(d time.Duration) string {
	days := d / (24 * time.Hour)
	d -= days * 24 * time.Hour
	hours := d / time.Hour
	d -= hours * time.Hour
	mins := d / time.Minute
	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, mins)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, mins)
	}
	return fmt.Sprintf("%dm", mins)
}

func getProcessCount() string {
	switch runtime.GOOS {
	case "linux":
		files, err := os.ReadDir("/proc")
		if err == nil {
			count := 0
			for _, f := range files {
				if f.IsDir() && isAllDigits(f.Name()) {
					count++
				}
			}
			return strconv.Itoa(count)
		}
	case "darwin":
		out, err := exec.Command("ps", "-e").Output()
		if err == nil {
			lines := strings.Split(string(out), "\n")
			return strconv.Itoa(len(lines) - 2)
		}
	}
	return "unknown"
}
func isAllDigits(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func getTerminal() string {
	term := os.Getenv("TERM")
	if term != "" {
		return term
	}
	return "unknown"
}

func getCPUCores() string {
	return strconv.Itoa(runtime.NumCPU())
}

func getMemInfo() string {
	switch runtime.GOOS {
	case "linux":
		f, err := os.Open("/proc/meminfo")
		if err == nil {
			defer f.Close()
			var total, free int
			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				line := scanner.Text()
				if strings.HasPrefix(line, "MemTotal:") {
					fmt.Sscanf(line, "MemTotal: %d kB", &total)
				} else if strings.HasPrefix(line, "MemAvailable:") {
					fmt.Sscanf(line, "MemAvailable: %d kB", &free)
				}
			}
			if total > 0 {
				used := total - free
				return fmt.Sprintf("%.1fMiB / %.1fMiB", float64(used)/1024, float64(total)/1024)
			}
		}
	case "darwin":
		out, err := exec.Command("vm_stat").Output()
		if err == nil {
			lines := strings.Split(string(out), "\n")
			var pageSize, freePages, totalPages int64
			pageSize = 4096
			for _, line := range lines {
				if strings.Contains(line, "page size of") {
					fmt.Sscanf(line, "Mach Virtual Memory Statistics: (page size of %d bytes)", &pageSize)
				}
				if strings.HasPrefix(line, "Pages free:") {
					fmt.Sscanf(line, "Pages free: %d.", &freePages)
				}
				if strings.HasPrefix(line, "Pages active:") ||
					strings.HasPrefix(line, "Pages inactive:") ||
					strings.HasPrefix(line, "Pages speculative:") ||
					strings.HasPrefix(line, "Pages wired down:") ||
					strings.HasPrefix(line, "Pages throttled:") ||
					strings.HasPrefix(line, "Pages purgeable:") {
					var v int64
					fmt.Sscanf(line, "%*s %d.", &v)
					totalPages += v
				}
			}
			if totalPages > 0 {
				used := totalPages - freePages
				return fmt.Sprintf("%.1fMiB / %.1fMiB", float64(used*pageSize)/1024/1024, float64(totalPages*pageSize)/1024/1024)
			}
		}
	}
	return "unknown"
}

func printColorBlocks(asciiWidth int) {
	// Print color blocks left-aligned with ASCII art
	fmt.Printf("%s%s", strings.Repeat(" ", asciiWidth), "  ")
	for i := 30; i <= 37; i++ {
		fmt.Printf("\033[%dm%s", i, colorBlock)
	}
	fmt.Println()
	fmt.Printf("%s%s", strings.Repeat(" ", asciiWidth), "  ")
	for i := 90; i <= 97; i++ {
		fmt.Printf("\033[%dm%s", i, colorBlock)
	}
	fmt.Printf("%s\n\n", reset)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
