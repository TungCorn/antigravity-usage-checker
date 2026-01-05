// Package display handles terminal output formatting with colors and progress bars.
package display

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/TungCorn/antigravity-usage-checker/internal/api"
)

const (
	// ANSI color codes
	Reset   = "\033[0m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"
	Bold    = "\033[1m"
	Dim     = "\033[2m"
)

// ShowUsage displays the usage data in the terminal.
func ShowUsage(data *api.UsageData, asJSON bool, isCached bool) {
	if asJSON {
		showJSON(data)
		return
	}
	
	showTable(data, isCached)
}

// showJSON outputs the usage data as formatted JSON.
func showJSON(data *api.UsageData) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Printf("Error formatting JSON: %v\n", err)
		return
	}
	fmt.Println(string(jsonData))
}

// showTable displays the usage data as a formatted table.
func showTable(data *api.UsageData, isCached bool) {
	// Header
	fmt.Println()
	fmt.Printf("%s%s🚀 Antigravity Usage Monitor%s\n", Bold, Cyan, Reset)
	fmt.Println(strings.Repeat("─", 70))
	
	// Cache indicator
	if isCached || data.IsCached {
		fmt.Printf("%s⚠️  Cached data from %s%s\n", Yellow, formatTime(data.FetchedAt), Reset)
		fmt.Println(strings.Repeat("─", 70))
	}
	
	// Table header
	fmt.Printf("%-22s %4s %5s %6s %-14s %s\n",
		"Model", "Used", "Limit", "Left", "Progress", "Reset")
	fmt.Println(strings.Repeat("─", 70))
	
	// Calculate totals with deduplication for shared quota pools
	// Use a map to track unique quota pools by their Used/Limit/Remaining signature
	type quotaKey struct {
		Used      int
		Limit     int
		Remaining int
	}
	uniqueQuotas := make(map[quotaKey]bool)
	
	totalUsed := 0
	totalLimit := 0
	totalRemaining := 0
	
	// Model rows
	for _, model := range data.Models {
		// Create a unique key for this quota pool
		key := quotaKey{
			Used:      model.Used,
			Limit:     model.Limit,
			Remaining: model.Remaining,
		}
		
		// Only count this quota pool if we haven't seen it before
		if !uniqueQuotas[key] {
			uniqueQuotas[key] = true
			totalUsed += model.Used
			totalLimit += model.Limit
			totalRemaining += model.Remaining
		}
		
		// Color based on remaining percentage (not usage)
		remainingPercent := 100 - model.UsagePercent
		color := getRemainingColor(remainingPercent)
		progressBar := createProgressBar(model.UsagePercent, 10)
		resetStr := formatResetTime(model.ResetTime)
		
		fmt.Printf("%-22s %s%4d%s %5d %s%6d%s %-14s %s\n",
			truncateString(model.ModelName, 22),
			Cyan, model.Used, Reset,
			model.Limit,
			color, model.Remaining, Reset,
			progressBar,
			resetStr,
		)
	}
	
	// Footer separator
	fmt.Println(strings.Repeat("─", 70))
	
	// Total usage summary
	totalUsagePercent := 0
	if totalLimit > 0 {
		totalUsagePercent = (totalUsed * 100) / totalLimit
	}
	totalRemainingPercent := 100 - totalUsagePercent
	summaryColor := getRemainingColor(totalRemainingPercent)
	fmt.Printf("%s📊 Total: %d/%d used (%d%% remaining)%s\n",
		summaryColor, totalUsed, totalLimit, totalRemainingPercent, Reset)
	fmt.Println(strings.Repeat("─", 70))
	
	// Tier and credits
	footer := []string{}
	if data.Tier != "" {
		footer = append(footer, fmt.Sprintf("Tier: %s%s%s", Cyan, data.Tier, Reset))
	}
	if data.PromptCredit > 0 {
		footer = append(footer, fmt.Sprintf("Credits: %s%d%s", Green, data.PromptCredit, Reset))
	}
	
	if len(footer) > 0 {
		fmt.Printf("%s%s%s\n", Dim, strings.Join(footer, " | "), Reset)
	}
	
	fmt.Println()
}

// createProgressBar generates a visual progress bar.
func createProgressBar(percent, width int) string {
	if percent > 100 {
		percent = 100
	}
	if percent < 0 {
		percent = 0
	}
	
	filled := width * percent / 100
	empty := width - filled
	
	// Color based on remaining percentage
	remainingPercent := 100 - percent
	color := getRemainingColor(remainingPercent)
	
	bar := fmt.Sprintf("%s%s%s%s %2d%%",
		color,
		strings.Repeat("█", filled),
		strings.Repeat("░", empty),
		Reset,
		percent,
	)
	
	return bar
}

// getStatusColor returns the appropriate color based on usage percentage.
func getStatusColor(percent int) string {
	switch {
	case percent < 50:
		return Green
	case percent < 80:
		return Yellow
	default:
		return Red
	}
}

// getRemainingColor returns the appropriate color based on remaining percentage.
func getRemainingColor(remainingPercent int) string {
	switch {
	case remainingPercent > 50:
		return Green
	case remainingPercent > 20:
		return Yellow
	default:
		return Red
	}
}

// truncateString limits string length with ellipsis.
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// formatTime formats a time.Time for display.
func formatTime(t time.Time) string {
	return t.Format("15:04:05 02/01/2006")
}

// formatResetTime converts an ISO timestamp to relative time (e.g., "in 2h 30m").
func formatResetTime(resetTimeStr string) string {
	if resetTimeStr == "" {
		return "-"
	}
	
	// Parse ISO 8601 timestamp
	resetTime, err := time.Parse(time.RFC3339, resetTimeStr)
	if err != nil {
		return "-"
	}
	
	now := time.Now()
	diff := resetTime.Sub(now)
	
	// If already reset
	if diff <= 0 {
		return Dim + "reset" + Reset
	}
	
	// Format as relative time
	hours := int(diff.Hours())
	minutes := int(diff.Minutes()) % 60
	
	if hours > 0 {
		return fmt.Sprintf("%s%dh %dm%s", Dim, hours, minutes, Reset)
	}
	return fmt.Sprintf("%s%dm%s", Dim, minutes, Reset)
}
