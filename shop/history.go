package shop

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Purchase represents a single purchase record
type Purchase struct {
	Date    string `json:"date"`
	Item    string `json:"item"`
	Price   int    `json:"price"`
	Session string `json:"session"`
}

// PurchaseHistory represents a character's purchase log
type PurchaseHistory struct {
	Character string     `json:"character"`
	Purchases []Purchase `json:"purchases"`
}

var historyPath = "data/history"

// LoadHistory loads the purchase history for a character
func LoadHistory(characterFile string) (*PurchaseHistory, error) {
	filename := filepath.Join(historyPath, characterFile+".json")
	data, err := os.ReadFile(filename)
	if err != nil {
		// If file doesn't exist, return empty history
		if os.IsNotExist(err) {
			return &PurchaseHistory{
				Character: characterFile,
				Purchases: []Purchase{},
			}, nil
		}
		return nil, fmt.Errorf("failed to read history for '%s': %w", characterFile, err)
	}

	var history PurchaseHistory
	if err := json.Unmarshal(data, &history); err != nil {
		return nil, fmt.Errorf("failed to parse history for '%s': %w", characterFile, err)
	}

	return &history, nil
}

// SaveHistory saves the purchase history to file
func SaveHistory(history *PurchaseHistory) error {
	filename := filepath.Join(historyPath, history.Character+".json")
	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal history: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write history: %w", err)
	}

	return nil
}

// AppendPurchase adds a new purchase to a character's history
func AppendPurchase(characterFile string, item Item, session string) error {
	history, err := LoadHistory(characterFile)
	if err != nil {
		return err
	}

	purchase := Purchase{
		Date:    time.Now().Format("2006-01-02"),
		Item:    item.Name,
		Price:   item.Price,
		Session: session,
	}

	history.Purchases = append(history.Purchases, purchase)
	return SaveHistory(history)
}

// FormatHistory returns a formatted string of purchase history
func (h *PurchaseHistory) FormatHistory() string {
	if len(h.Purchases) == 0 {
		return "No purchase history found."
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("**Purchase History for %s**\n\n", h.Character))

	for _, p := range h.Purchases {
		sb.WriteString(fmt.Sprintf("• **%s** - %d gp (%s)\n", p.Item, p.Price, p.Date))
		if p.Session != "" {
			sb.WriteString(fmt.Sprintf("  *Session: %s*\n", p.Session))
		}
	}
	return sb.String()
}

// GetTotalSpent returns the total gold spent by this character
func (h *PurchaseHistory) GetTotalSpent() int {
	total := 0
	for _, p := range h.Purchases {
		total += p.Price
	}
	return total
}
