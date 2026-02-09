package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type MagicItem struct {
	Name        string `json:"name"`
	Type        string `json:"type,omitempty"`
	Rarity      string `json:"rarity"`
	Attunement  string `json:"attunement,omitempty"`
	Description string `json:"description"`
}

type ItemFile struct {
	Items []MagicItem `json:"items"`
}

func main() {
	jsonFiles := []string{
		"data/magic_weapons.json",
		"data/magic_armor.json",
		"data/magic_potions.json",
		"data/wondrous_items.json",
	}

	// Step 1: Load all item names from all JSON files
	allItems := map[string]*MagicItem{}   // lowercase name -> item pointer
	fileData := map[string]*ItemFile{}     // file path -> parsed file
	nameSet := map[string]bool{}           // lowercase names for boundary detection

	for _, path := range jsonFiles {
		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Printf("Error reading %s: %v\n", path, err)
			return
		}
		var itemFile ItemFile
		if err := json.Unmarshal(data, &itemFile); err != nil {
			fmt.Printf("Error parsing %s: %v\n", path, err)
			return
		}
		fileData[path] = &itemFile
		for i := range itemFile.Items {
			lower := strings.ToLower(itemFile.Items[i].Name)
			allItems[lower] = &itemFile.Items[i]
			nameSet[lower] = true
		}
	}
	fmt.Printf("Loaded %d item names from JSON files\n", len(nameSet))

	// Step 2: Read the CSV into lines
	csvFile, err := os.Open("data/dmguide_magic_items.csv")
	if err != nil {
		fmt.Printf("Error opening CSV: %v\n", err)
		return
	}
	defer csvFile.Close()

	var lines []string
	scanner := bufio.NewScanner(csvFile)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024) // handle long lines
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	fmt.Printf("Read %d lines from CSV\n", len(lines))

	// Step 3: Scan CSV. When we hit a known item name, skip the type line,
	// then collect description until we hit another known item name.
	updated := 0
	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		item, found := allItems[strings.ToLower(line)]
		if !found {
			continue
		}

		// Found a known item. Skip forward past: blank line, type line, blank line
		j := i + 1

		// Skip blanks after item name
		for j < len(lines) && strings.TrimSpace(lines[j]) == "" {
			j++
		}
		// Skip the type/rarity line (e.g. "Wondrous Item, Rare (Requires Attunement)")
		if j < len(lines) {
			j++
		}
		// Skip blanks after type line
		for j < len(lines) && strings.TrimSpace(lines[j]) == "" {
			j++
		}

		// Collect description lines until we hit another known item name
		var descLines []string
		for j < len(lines) {
			descLine := strings.TrimSpace(lines[j])

			// Is this line another known item name? If so, stop.
			if nameSet[strings.ToLower(descLine)] {
				break
			}

			// Skip empty lines but keep collecting
			if descLine != "" {
				descLines = append(descLines, descLine)
			}
			j++
		}

		newDesc := strings.Join(descLines, " ")
		if len(newDesc) > len(item.Description) {
			old := item.Description
			item.Description = newDesc
			updated++
			if len(old) < len(newDesc) {
				fmt.Printf("  %-40s %4d -> %4d chars\n", item.Name, len(old), len(newDesc))
			}
		}
	}

	fmt.Printf("\nUpdated %d item descriptions\n\n", updated)

	// Step 4: Write updated JSON files
	for _, path := range jsonFiles {
		data, err := json.MarshalIndent(fileData[path], "", "  ")
		if err != nil {
			fmt.Printf("Error marshaling %s: %v\n", path, err)
			continue
		}
		if err := os.WriteFile(path, data, 0644); err != nil {
			fmt.Printf("Error writing %s: %v\n", path, err)
			continue
		}
		fmt.Printf("Wrote %s (%d items)\n", path, len(fileData[path].Items))
	}
}
