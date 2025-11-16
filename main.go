package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Config holds the application configuration
type Config struct {
	NotionToken string
	DatabaseID  string
}

// NotionPage represents a page in the Notion database
type NotionPage struct {
	ID         string                 `json:"id"`
	Properties map[string]interface{} `json:"properties"`
	URL        string                 `json:"url"`
}

// NotionQueryResponse represents the response from Notion API
type NotionQueryResponse struct {
	Results []NotionPage `json:"results"`
}

// Reminder holds formatted reminder information
type Reminder struct {
	Title    string
	Created  string
	Priority string
	Category string
	URL      string
}

func main() {
	// Load configuration
	config, err := loadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Validate configuration
	if config.NotionToken == "" || config.DatabaseID == "" {
		log.Fatal("ERROR: NOTION_TOKEN and DATABASE_ID must be set in config file")
	}

	log.Println("Checking Notion for pending reminders...")

	// Fetch reminders from Notion with retries
	var reminders []Reminder
	maxRetries := 5

	for attempt := 0; attempt <= maxRetries; attempt++ {
		reminders, err = getPendingReminders(config)
		if err == nil {
			break
		}

		if attempt < maxRetries {
			// Exponential backoff: 2s, 4s, 8s, 16s, 32s
			waitTime := time.Duration(1<<uint(attempt+1)) * time.Second
			log.Printf("Failed to fetch reminders (attempt %d/%d): %v", attempt+1, maxRetries+1, err)
			log.Printf("Retrying in %v...", waitTime)
			time.Sleep(waitTime)
		}
	}

	if err != nil {
		log.Fatalf("Error fetching reminders after %d attempts: %v", maxRetries+1, err)
	}

	if len(reminders) == 0 {
		log.Println("No pending reminders found!")
		showNotificationSimple("All caught up! No pending reminders. 🎉")
		return
	}

	log.Printf("Found %d pending reminder(s)\n", len(reminders))

	// Log individual reminders to console
	for i, reminder := range reminders {
		log.Printf("  %d. %s (from %s)\n", i+1, reminder.Title, reminder.Created)
	}

	// Build simple summary message
	var message string
	urgency := "normal"

	if len(reminders) == 1 {
		message = "You have 1 pending reminder.\n\nRight-click to open in Notion."
	} else {
		message = fmt.Sprintf("You have %d pending reminders.\n\nRight-click to open in Notion.", len(reminders))
		if len(reminders) > 3 {
			urgency = "critical"
		}
	}

	// Get database URL
	databaseURL := fmt.Sprintf("https://www.notion.so/%s", config.DatabaseID)

	// Show summary notification
	showNotification("Notion Reminders", message, urgency, databaseURL)
}

// loadConfig reads configuration from the config file
func loadConfig() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	configPath := filepath.Join(homeDir, ".config", "notion-reminder", "config.conf")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("config file not found at %s. Please run setup script first", configPath)
	}

	config := &Config{}
	lines := strings.Split(string(data), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "NOTION_TOKEN":
			config.NotionToken = value
		case "DATABASE_ID":
			config.DatabaseID = value
		}
	}

	return config, nil
}

// getPendingReminders fetches pending reminders from Notion
func getPendingReminders(config *Config) ([]Reminder, error) {
	url := fmt.Sprintf("https://api.notion.com/v1/databases/%s/query", config.DatabaseID)

	// Build request body
	requestBody := map[string]interface{}{
		"filter": map[string]interface{}{
			"property": "Status",
			"checkbox": map[string]bool{
				"equals": false,
			},
		},
		"sorts": []map[string]string{
			{
				"property":  "Created At",
				"direction": "descending",
			},
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+config.NotionToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Notion-Version", "2022-06-28")

	// Send request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("notion API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var queryResponse NotionQueryResponse
	if err := json.Unmarshal(body, &queryResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Format reminders
	var reminders []Reminder
	for _, page := range queryResponse.Results {
		reminder := formatReminder(page)
		if reminder != nil {
			reminders = append(reminders, *reminder)
		}
	}

	return reminders, nil
}

// formatReminder extracts and formats reminder information from a Notion page
func formatReminder(page NotionPage) *Reminder {
	reminder := &Reminder{
		URL: page.URL,
	}

	// Extract title
	if nameProperty, ok := page.Properties["Name"].(map[string]interface{}); ok {
		if titleArray, ok := nameProperty["title"].([]interface{}); ok && len(titleArray) > 0 {
			if titleObj, ok := titleArray[0].(map[string]interface{}); ok {
				if plainText, ok := titleObj["plain_text"].(string); ok {
					reminder.Title = plainText
				}
			}
		}
	}

	if reminder.Title == "" {
		reminder.Title = "Untitled"
	}

	// Extract created time
	if createdProperty, ok := page.Properties["Created At"].(map[string]interface{}); ok {
		if createdTime, ok := createdProperty["created_time"].(string); ok {
			// Parse ISO 8601 time
			t, err := time.Parse(time.RFC3339, createdTime)
			if err == nil {
				reminder.Created = t.Format("Jan 02, 2006 at 15:04")
			} else {
				reminder.Created = "Unknown date"
			}
		}
	}

	// Extract priority if it exists
	if priorityProperty, ok := page.Properties["Priority"].(map[string]interface{}); ok {
		if selectObj, ok := priorityProperty["select"].(map[string]interface{}); ok {
			if name, ok := selectObj["name"].(string); ok {
				reminder.Priority = fmt.Sprintf(" [Priority: %s]", name)
			}
		}
	}

	// Extract category if it exists
	if categoryProperty, ok := page.Properties["Category"].(map[string]interface{}); ok {
		if selectObj, ok := categoryProperty["select"].(map[string]interface{}); ok {
			if name, ok := selectObj["name"].(string); ok {
				reminder.Category = fmt.Sprintf(" [Category: %s]", name)
			}
		}
	}

	return reminder
}

// showNotification displays a desktop notification using notify-send with retry logic
func showNotification(title, message, urgency, url string) {
	// Get home directory for icon path
	homeDir, homeErr := os.UserHomeDir()
	iconPath := "dialog-information" // fallback
	if homeErr == nil {
		customIcon := filepath.Join(homeDir, ".local", "share", "notion-reminder", "logo.svg")
		if _, statErr := os.Stat(customIcon); statErr == nil {
			iconPath = customIcon
		}
	}

	args := []string{
		"-u", urgency,
		"-i", iconPath,
		"-a", "Notion Reminders",
		"-t", "0", // Stay on screen until dismissed
	}

	// Add click action if URL is provided
	if url != "" {
		args = append(args, "-A", "default=Open in Notion")
	}

	args = append(args, title, message)

	// Retry logic for notification - D-Bus/notification daemon may not be ready at boot
	maxRetries := 5
	var err error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		cmd := exec.Command("notify-send", args...)
		err = cmd.Run()

		if err == nil {
			// Success!
			if url != "" {
				log.Printf("Click notification to open: %s\n", url)
			}
			return
		}

		if attempt < maxRetries {
			// Exponential backoff: 1s, 2s, 4s, 8s, 16s
			waitTime := time.Duration(1<<uint(attempt)) * time.Second
			log.Printf("Failed to show notification (attempt %d/%d): %v", attempt+1, maxRetries+1, err)
			log.Printf("Retrying in %v... (notification daemon may not be ready yet)", waitTime)
			time.Sleep(waitTime)
		}
	}

	// All retries failed
	log.Printf("Warning: Failed to show notification after %d attempts: %v", maxRetries+1, err)
	log.Printf("Make sure libnotify is installed: sudo pacman -S libnotify")
	log.Printf("And that the notification daemon is running")

	if url != "" {
		log.Printf("Click notification to open: %s\n", url)
	}
}

// showNotificationSimple displays a simple auto-dismissing notification with retry logic
func showNotificationSimple(message string) {
	// Get home directory for icon path
	homeDir, homeErr := os.UserHomeDir()
	iconPath := "dialog-information" // fallback
	if homeErr == nil {
		customIcon := filepath.Join(homeDir, ".local", "share", "notion-reminder", "logo.svg")
		if _, statErr := os.Stat(customIcon); statErr == nil {
			iconPath = customIcon
		}
	}

	args := []string{
		"-u", "low",
		"-i", iconPath,
		"-a", "Notion Reminders - Complete",
		"Notion Reminders",
		message,
	}

	// Retry logic for notification - D-Bus/notification daemon may not be ready at boot
	maxRetries := 5
	var err error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		cmd := exec.Command("notify-send", args...)
		err = cmd.Run()

		if err == nil {
			// Success!
			return
		}

		if attempt < maxRetries {
			// Exponential backoff: 1s, 2s, 4s, 8s, 16s
			waitTime := time.Duration(1<<uint(attempt)) * time.Second
			log.Printf("Failed to show notification (attempt %d/%d): %v", attempt+1, maxRetries+1, err)
			log.Printf("Retrying in %v... (notification daemon may not be ready yet)", waitTime)
			time.Sleep(waitTime)
		}
	}

	// All retries failed
	log.Printf("Warning: Failed to show notification after %d attempts: %v", maxRetries+1, err)
	log.Printf("Make sure libnotify is installed: sudo pacman -S libnotify")
	log.Printf("And that the notification daemon is running")
}
