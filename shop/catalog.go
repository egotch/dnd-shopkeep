package shop

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Item represents a single shop item
type Item struct {
	Name        string `json:"name"`
	Category    string `json:"category"`
	Price       int    `json:"price"`
	Description string `json:"description"`
	Rarity      string `json:"rarity"` // common, uncommon, rare
}

// Catalog represents the full shop inventory
type Catalog struct {
	Items           []Item         `json:"items"`
	MonthlyRotation MonthlySection `json:"monthly_rotation"`
}

// MonthlySection holds the rotating uncommon items
type MonthlySection struct {
	Month string `json:"month"`
	Items []Item `json:"items"`
}

var catalogPath = "data/catalog.json"

// LoadCatalog loads the catalog from the JSON file
func LoadCatalog() (*Catalog, error) {
	data, err := os.ReadFile(catalogPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read catalog: %w", err)
	}

	var catalog Catalog
	if err := json.Unmarshal(data, &catalog); err != nil {
		return nil, fmt.Errorf("failed to parse catalog: %w", err)
	}

	return &catalog, nil
}

// GetItemsByCategory returns items matching the given category
func (c *Catalog) GetItemsByCategory(category string) []Item {
	if category == "" || category == "all" {
		return c.Items
	}

	category = strings.ToLower(category)
	var filtered []Item
	for _, item := range c.Items {
		if strings.ToLower(item.Category) == category {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

// FindItem performs a fuzzy search for an item by name
func (c *Catalog) FindItem(name string) (*Item, error) {
	name = strings.ToLower(name)

	// First try exact match
	for i, item := range c.Items {
		if strings.ToLower(item.Name) == name {
			return &c.Items[i], nil
		}
	}

	// Then try contains match
	for i, item := range c.Items {
		if strings.Contains(strings.ToLower(item.Name), name) {
			return &c.Items[i], nil
		}
	}

	// Check monthly rotation items too
	for i, item := range c.MonthlyRotation.Items {
		if strings.ToLower(item.Name) == name {
			return &c.MonthlyRotation.Items[i], nil
		}
		if strings.Contains(strings.ToLower(item.Name), name) {
			return &c.MonthlyRotation.Items[i], nil
		}
	}

	return nil, fmt.Errorf("item '%s' not found in catalog", name)
}

// GetCategories returns all unique categories in the catalog
func (c *Catalog) GetCategories() []string {
	seen := make(map[string]bool)
	var categories []string
	for _, item := range c.Items {
		cat := strings.ToLower(item.Category)
		if !seen[cat] {
			seen[cat] = true
			categories = append(categories, item.Category)
		}
	}
	return categories
}

// FormatItemList returns a formatted string of items for display
func FormatItemList(items []Item) string {
	if len(items) == 0 {
		return "No items found."
	}

	var sb strings.Builder
	for _, item := range items {
		sb.WriteString(fmt.Sprintf("• **%s** - %d gp\n", item.Name, item.Price))
		if item.Description != "" {
			sb.WriteString(fmt.Sprintf("  *%s*\n", item.Description))
		}
	}
	return sb.String()
}
