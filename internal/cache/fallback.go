// Package cache provides graceful degradation by caching the last known
// usage data for display when Antigravity is not running.
package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/tungcorn/antigravity-usage-checker/internal/api"
)

const cacheFileName = "usage_cache.json"

// getCachePath returns the path to the cache file.
func getCachePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".gemini", cacheFileName)
}

// Save stores the usage data to the cache file.
func Save(data *api.UsageData) error {
	cachePath := getCachePath()
	
	// Ensure directory exists
	dir := filepath.Dir(cachePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}
	
	// Mark as cached and update timestamp
	data.IsCached = false // Will be set to true when loaded
	data.FetchedAt = time.Now()
	
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache data: %w", err)
	}
	
	if err := os.WriteFile(cachePath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}
	
	return nil
}

// LoadLastKnown loads the last known usage data from cache.
// Returns error if cache doesn't exist or data is too old (>24 hours).
func LoadLastKnown() (*api.UsageData, error) {
	cachePath := getCachePath()
	
	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, fmt.Errorf("no cached data available: %w", err)
	}
	
	var usage api.UsageData
	if err := json.Unmarshal(data, &usage); err != nil {
		return nil, fmt.Errorf("failed to parse cached data: %w", err)
	}
	
	// Check if cache is too old (24 hours)
	age := time.Since(usage.FetchedAt)
	if age > 24*time.Hour {
		return nil, fmt.Errorf("cached data is too old (%v)", age)
	}
	
	usage.IsCached = true
	return &usage, nil
}

// Clear removes the cache file.
func Clear() error {
	cachePath := getCachePath()
	if err := os.Remove(cachePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to clear cache: %w", err)
	}
	return nil
}
