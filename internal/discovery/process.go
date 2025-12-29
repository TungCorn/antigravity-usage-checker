// Package discovery provides functions to detect the running Antigravity
// language server process and extract connection information.
package discovery

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// ProcessInfo contains connection information extracted from the Antigravity process.
type ProcessInfo struct {
	PID         int
	ConnectPort int    // HTTPS port for API calls
	HTTPPort    int    // HTTP fallback port
	CSRFToken   string // X-Codeium-Csrf-Token value
}

// FindAntigravityProcess scans running processes to find the Antigravity
// language server and extracts connection information from its command line.
func FindAntigravityProcess() (*ProcessInfo, error) {
	if runtime.GOOS == "windows" {
		return findProcessWindows()
	}
	return findProcessUnix()
}

// findProcessWindows uses PowerShell to find the language server process on Windows.
func findProcessWindows() (*ProcessInfo, error) {
	// PowerShell command to get language server process with command line as JSON
	cmd := exec.Command("powershell", "-NoProfile", "-Command",
		`Get-CimInstance Win32_Process | Where-Object { $_.CommandLine -like "*extension_server_port*" -and $_.Name -like "*language_server*" } | Select-Object -First 1 ProcessId, CommandLine | ConvertTo-Json -Compress`)
	
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute PowerShell: %w", err)
	}
	
	info, err := parseProcessInfoJSON(string(output))
	if err != nil {
		return nil, err
	}
	
	// Now find all listening ports for this PID and test to find the API port
	ports, err := getListeningPortsForPID(info.PID)
	if err == nil && len(ports) > 0 {
		// Test each port to find the one that responds to API
		for _, port := range ports {
			if testAPIPort(port, info.CSRFToken) {
				info.ConnectPort = port
				break
			}
		}
	}
	
	// If no working API port found, use HTTP port
	if info.ConnectPort == 0 {
		info.ConnectPort = info.HTTPPort
	}
	
	return info, nil
}

// findProcessUnix uses ps command to find the language server process on Unix systems.
func findProcessUnix() (*ProcessInfo, error) {
	cmd := exec.Command("ps", "aux")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute ps: %w", err)
	}
	
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "language_server") {
			return parseCommandLine(line)
		}
	}
	
	return nil, fmt.Errorf("Antigravity language server not found")
}

// PowerShell JSON output structure
type psProcessInfo struct {
	ProcessId   int    `json:"ProcessId"`
	CommandLine string `json:"CommandLine"`
}

// parseProcessInfoJSON parses JSON output from PowerShell.
func parseProcessInfoJSON(output string) (*ProcessInfo, error) {
	output = strings.TrimSpace(output)
	if output == "" || output == "null" {
		return nil, fmt.Errorf("Antigravity language server not found")
	}
	
	var psInfo psProcessInfo
	if err := json.Unmarshal([]byte(output), &psInfo); err != nil {
		return nil, fmt.Errorf("failed to parse process info: %w", err)
	}
	
	if psInfo.ProcessId == 0 {
		return nil, fmt.Errorf("Antigravity language server not found")
	}
	
	// Parse command line to extract port and token
	info := &ProcessInfo{PID: psInfo.ProcessId}
	
	// Extract extension_server_port
	portRe := regexp.MustCompile(`--extension_server_port\s+(\d+)`)
	if match := portRe.FindStringSubmatch(psInfo.CommandLine); len(match) > 1 {
		info.HTTPPort, _ = strconv.Atoi(match[1])
	}
	
	// Extract csrf_token
	tokenRe := regexp.MustCompile(`--csrf_token\s+([a-zA-Z0-9-]+)`)
	if match := tokenRe.FindStringSubmatch(psInfo.CommandLine); len(match) > 1 {
		info.CSRFToken = match[1]
	}
	
	if info.HTTPPort == 0 || info.CSRFToken == "" {
		return nil, fmt.Errorf("could not extract connection info from process")
	}
	
	info.ConnectPort = info.HTTPPort
	return info, nil
}

// parseProcessInfo parses PowerShell output to extract process information.
// PowerShell Format-List wraps long command lines across multiple lines.
func parseProcessInfo(output string) (*ProcessInfo, error) {
	lines := strings.Split(output, "\n")
	
	var pid int
	var cmdLine strings.Builder
	inCommandLine := false
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		// Check for ProcessId field
		if strings.HasPrefix(trimmed, "ProcessId") {
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				pid, _ = strconv.Atoi(strings.TrimSpace(parts[1]))
			}
			inCommandLine = false
			continue
		}
		
		// Check for CommandLine field (may span multiple lines)
		if strings.HasPrefix(trimmed, "CommandLine") {
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				cmdLine.WriteString(strings.TrimSpace(parts[1]))
				cmdLine.WriteString(" ")
			}
			inCommandLine = true
			continue
		}
		
		// Continuation of CommandLine (indented lines after CommandLine:)
		if inCommandLine && len(trimmed) > 0 && !strings.Contains(trimmed, ":") {
			cmdLine.WriteString(trimmed)
			cmdLine.WriteString(" ")
		} else if strings.Contains(trimmed, ":") {
			// New field found, stop continuing CommandLine
			inCommandLine = false
		}
	}
	
	cmdLineStr := cmdLine.String()
	if cmdLineStr == "" {
		return nil, fmt.Errorf("Antigravity language server not found")
	}
	
	return parseCommandLine(cmdLineStr + fmt.Sprintf(" PID=%d", pid))
}

// parseCommandLine extracts ports and CSRF token from command line arguments.
func parseCommandLine(cmdLine string) (*ProcessInfo, error) {
	info := &ProcessInfo{}
	
	// Extract PID if present
	pidRe := regexp.MustCompile(`PID[=:]?\s*(\d+)`)
	if match := pidRe.FindStringSubmatch(cmdLine); len(match) > 1 {
		info.PID, _ = strconv.Atoi(match[1])
	}
	
	// Extract extension_server_port (HTTP port)
	extPortRe := regexp.MustCompile(`--extension_server_port[=\s]+(\d+)`)
	if match := extPortRe.FindStringSubmatch(cmdLine); len(match) > 1 {
		info.HTTPPort, _ = strconv.Atoi(match[1])
	}
	
	// Extract CSRF token (looks for a long alphanumeric string after csrf-related flags)
	csrfRe := regexp.MustCompile(`--[a-z_]*csrf[a-z_]*[=\s]+([a-zA-Z0-9_-]{20,})`)
	if match := csrfRe.FindStringSubmatch(cmdLine); len(match) > 1 {
		info.CSRFToken = match[1]
	}
	
	// If no CSRF token found, try alternative patterns
	if info.CSRFToken == "" {
		altRe := regexp.MustCompile(`"?([a-zA-Z0-9_-]{32,64})"?`)
		if match := altRe.FindStringSubmatch(cmdLine); len(match) > 1 {
			// Use as potential token if it looks like one
			info.CSRFToken = match[1]
		}
	}
	
	// Validate we have minimum required info
	if info.HTTPPort == 0 && info.CSRFToken == "" {
		return nil, fmt.Errorf("could not extract connection info from process")
	}
	
	// If no connect port found, use HTTP port for both
	if info.ConnectPort == 0 {
		info.ConnectPort = info.HTTPPort
	}
	
	return info, nil
}

// GetListeningPorts returns all TCP ports that a process is listening on.
func GetListeningPorts(pid int) ([]int, error) {
	var cmd *exec.Cmd
	
	if runtime.GOOS == "windows" {
		cmd = exec.Command("netstat", "-ano")
	} else {
		cmd = exec.Command("lsof", "-iTCP", "-sTCP:LISTEN", "-n", "-P", "-p", strconv.Itoa(pid))
	}
	
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get listening ports: %w", err)
	}
	
	return parseListeningPorts(string(output), pid)
}

// parseListeningPorts extracts ports from netstat/lsof output.
func parseListeningPorts(output string, pid int) ([]int, error) {
	var ports []int
	pidStr := strconv.Itoa(pid)
	
	lines := strings.Split(output, "\n")
	portRe := regexp.MustCompile(`:(\d+)\s`)
	
	for _, line := range lines {
		if strings.Contains(line, "LISTENING") && strings.Contains(line, pidStr) {
			if match := portRe.FindStringSubmatch(line); len(match) > 1 {
				port, _ := strconv.Atoi(match[1])
				if port > 0 {
					ports = append(ports, port)
				}
			}
		}
	}
	
	return ports, nil
}

// getListeningPortsForPID returns all TCP ports that a specific process is listening on.
func getListeningPortsForPID(pid int) ([]int, error) {
	if runtime.GOOS != "windows" {
		return GetListeningPorts(pid)
	}
	
	// Windows: use netstat and filter by PID
	cmd := exec.Command("netstat", "-ano")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run netstat: %w", err)
	}
	
	var ports []int
	pidStr := strconv.Itoa(pid)
	portRe := regexp.MustCompile(`127\.0\.0\.1:(\d+)`)
	
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "LISTENING") && strings.HasSuffix(strings.TrimSpace(line), pidStr) {
			if match := portRe.FindStringSubmatch(line); len(match) > 1 {
				port, _ := strconv.Atoi(match[1])
				if port > 0 {
					ports = append(ports, port)
				}
			}
		}
	}
	
	return ports, nil
}

// testAPIPort tests if a port responds to the Antigravity API.
// Uses GetUnleashData endpoint which doesn't require authentication.
func testAPIPort(port int, csrfToken string) bool {
	client := &http.Client{
		Timeout: 3 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	
	// Test with GetUnleashData endpoint (works without full auth)
	url := fmt.Sprintf("https://127.0.0.1:%d/exa.language_server_pb.LanguageServerService/GetUnleashData", port)
	
	body := []byte(`{"context":{"properties":{"ide":"antigravity"}}}`)
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return false
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Connect-Protocol-Version", "1")
	req.Header.Set("X-Codeium-Csrf-Token", csrfToken)
	
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	
	// If we get 200, this is the right port
	return resp.StatusCode == http.StatusOK
}
