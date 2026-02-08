package shop

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/egotch/dnd-shopkeep/config"
)

// Item represents a single shop item
// Fields are optional depending on item type (weapon, armor, gear, potion)
type Item struct {
	Name        string  `json:"name"`
	Category    string  `json:"category,omitempty"`    // Set during load based on file
	Cost        float64 `json:"cost"`                  // Price in GP
	Description string  `json:"description,omitempty"` // For potions/gear
	Rarity      string  `json:"rarity,omitempty"`      // common, uncommon, rare
	// Weapon-specific fields
	Damage     string `json:"damage,omitempty"`
	Properties string `json:"properties,omitempty"`
	Mastery    string `json:"mastery,omitempty"`
	Weight     string `json:"weight,omitempty"`
	// Armor-specific fields
	AC          string `json:"ac,omitempty"`
	Strength    string `json:"strength,omitempty"`
	Stealth     string `json:"stealth,omitempty"`
}

// Catalog represents the full shop inventory
type Catalog struct {
	Items           []Item          `json:"items"`
	SessionSpecials SessionSpecials `json:"session_specials"`
}

// SessionSpecials holds special items available for the current session
type SessionSpecials struct {
	Items []Item `json:"items"`
}

// itemFile represents the structure of each category JSON file
type itemFile struct {
	Items []Item `json:"items"`
}

// categoryMap maps file paths to category names
var categoryMap = map[string]string{
	config.DataPaths.Weapons:         "weapons",
	config.DataPaths.Armor:           "armor",
	config.DataPaths.Potions:         "potions",
	config.DataPaths.AdventuringGear: "gear",
	config.DataPaths.SessionSpecials: "specials",
}

// LoadCatalog loads the catalog from multiple category JSON files
func LoadCatalog() (*Catalog, error) {
	catalog := &Catalog{}

	// Load each category file
	categoryFiles := []string{
		config.DataPaths.Weapons,
		config.DataPaths.Armor,
		config.DataPaths.Potions,
		config.DataPaths.AdventuringGear,
	}

	for _, path := range categoryFiles {
		items, err := loadItemsFromFile(path, categoryMap[path])
		if err != nil {
			return nil, fmt.Errorf("failed to load %s: %w", path, err)
		}
		catalog.Items = append(catalog.Items, items...)
	}

	// Load session specials
	specials, err := loadItemsFromFile(config.DataPaths.SessionSpecials, "specials")
	if err != nil {
		return nil, fmt.Errorf("failed to load session specials: %w", err)
	}
	catalog.SessionSpecials.Items = specials

	return catalog, nil
}

// loadItemsFromFile loads items from a single JSON file and sets their category
func loadItemsFromFile(path string, category string) ([]Item, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var file itemFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	// Set category for each item
	for i := range file.Items {
		file.Items[i].Category = category
	}

	return file.Items, nil
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

	// Check session specials too
	for i, item := range c.SessionSpecials.Items {
		if strings.ToLower(item.Name) == name {
			return &c.SessionSpecials.Items[i], nil
		}
		if strings.Contains(strings.ToLower(item.Name), name) {
			return &c.SessionSpecials.Items[i], nil
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
		// Format cost - show as int if whole number, otherwise with decimal
		costStr := fmt.Sprintf("%.0f", item.Cost)
		if item.Cost != float64(int(item.Cost)) {
			costStr = fmt.Sprintf("%.1f", item.Cost)
		}
		sb.WriteString(fmt.Sprintf("• **%s** - %s gp\n", item.Name, costStr))

		// Show details based on item type
		switch item.Category {
		case "weapons":
			sb.WriteString(fmt.Sprintf("  *%s, %s*\n", item.Damage, item.Properties))
		case "armor":
			if item.AC != "" {
				details := []string{fmt.Sprintf("AC %s", item.AC)}
				if item.Strength != "" && item.Strength != "—" {
					details = append(details, fmt.Sprintf("%s required", item.Strength))
				}
				if item.Stealth != "" && item.Stealth != "—" {
					details = append(details, fmt.Sprintf("Stealth %s", item.Stealth))
				}
				if item.Weight != "" && item.Weight != "—" {
					details = append(details, item.Weight)
				}
				sb.WriteString(fmt.Sprintf("  *%s*\n", strings.Join(details, ", ")))
			}
		case "gear":
			if item.Weight != "" && item.Weight != "—" {
				sb.WriteString(fmt.Sprintf("  *%s*\n", item.Weight))
			}
		default:
			if item.Description != "" {
				sb.WriteString(fmt.Sprintf("  *%s*\n", item.Description))
			}
		}
	}
	return sb.String()
}
