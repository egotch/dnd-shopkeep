package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
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
	file, err := os.Open("data/dmguide_magic_items.csv")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	var weapons, armor, potions, wondrous []MagicItem

	scanner := bufio.NewScanner(file)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	// Patterns to identify item type lines
	weaponPattern := regexp.MustCompile(`^Weapon \(([^)]+)\),\s*(.+)`)
	armorPattern := regexp.MustCompile(`^Armor \(([^)]+)\),\s*(.+)`)
	potionPattern := regexp.MustCompile(`^Potion,\s*(.+)`)
	wondrousPattern := regexp.MustCompile(`^Wondrous Item,\s*(.+)`)

	// Skip patterns (image credits, table headers, etc.)
	skipPatterns := []string{"Conceptopolis", "Paul Scott Canavan", "1d100", "1d10", "1d20", "1d6", "1d8"}

	i := 0
	for i < len(lines) {
		line := strings.TrimSpace(lines[i])

		// Skip empty lines and noise
		if line == "" || shouldSkip(line, skipPatterns) {
			i++
			continue
		}

		// Check if this could be an item name (followed by blank line then type line)
		if i+2 < len(lines) && lines[i+1] == "" {
			typeLine := strings.TrimSpace(lines[i+2])

			var item MagicItem
			var category string

			if matches := weaponPattern.FindStringSubmatch(typeLine); matches != nil {
				item.Name = line
				item.Type = matches[1]
				item.Rarity, item.Attunement = parseRarityAttunement(matches[2])
				category = "weapon"
			} else if matches := armorPattern.FindStringSubmatch(typeLine); matches != nil {
				item.Name = line
				item.Type = matches[1]
				item.Rarity, item.Attunement = parseRarityAttunement(matches[2])
				category = "armor"
			} else if matches := potionPattern.FindStringSubmatch(typeLine); matches != nil {
				item.Name = line
				item.Rarity, item.Attunement = parseRarityAttunement(matches[1])
				category = "potion"
			} else if matches := wondrousPattern.FindStringSubmatch(typeLine); matches != nil {
				item.Name = line
				item.Rarity, item.Attunement = parseRarityAttunement(matches[1])
				category = "wondrous"
			}

			if category != "" {
				// Collect description (lines after type line until next item or section)
				i += 3 // Move past name, blank, type line
				if i < len(lines) && lines[i] == "" {
					i++ // Skip blank after type line
				}

				var descLines []string
				for i < len(lines) {
					descLine := strings.TrimSpace(lines[i])

					// Stop if we hit a new item pattern
					if i+2 < len(lines) && lines[i+1] == "" {
						nextType := strings.TrimSpace(lines[i+2])
						if weaponPattern.MatchString(nextType) ||
						   armorPattern.MatchString(nextType) ||
						   potionPattern.MatchString(nextType) ||
						   wondrousPattern.MatchString(nextType) {
							break
						}
					}

					// Stop on section headers
					if strings.HasPrefix(descLine, "Magic Items (") {
						break
					}

					// Skip noise
					if shouldSkip(descLine, skipPatterns) || descLine == "" {
						i++
						continue
					}

					// Skip obvious table data
					if strings.Contains(descLine, "\t") && !strings.HasPrefix(descLine, "You") && !strings.HasPrefix(descLine, "This") && !strings.HasPrefix(descLine, "When") {
						i++
						continue
					}

					descLines = append(descLines, descLine)
					i++

					// Limit description length
					if len(descLines) >= 3 {
						break
					}
				}

				item.Description = strings.Join(descLines, " ")

				// Truncate long descriptions
				if len(item.Description) > 300 {
					item.Description = item.Description[:297] + "..."
				}

				switch category {
				case "weapon":
					weapons = append(weapons, item)
				case "armor":
					armor = append(armor, item)
				case "potion":
					potions = append(potions, item)
				case "wondrous":
					wondrous = append(wondrous, item)
				}
				continue
			}
		}
		i++
	}

	// Write output files
	writeJSON("data/magic_weapons.json", weapons)
	writeJSON("data/magic_armor.json", armor)
	writeJSON("data/potions.json", potions)
	writeJSON("data/wondrous_items.json", wondrous)

	fmt.Printf("Extracted: %d weapons, %d armor, %d potions, %d wondrous items\n",
		len(weapons), len(armor), len(potions), len(wondrous))
}

func parseRarityAttunement(s string) (rarity, attunement string) {
	s = strings.TrimSpace(s)

	// Check for attunement
	if idx := strings.Index(s, "(Requires Attunement"); idx != -1 {
		rarity = strings.TrimSpace(s[:idx])
		attunement = strings.TrimPrefix(s[idx:], "(")
		attunement = strings.TrimSuffix(attunement, ")")
	} else {
		rarity = s
	}

	// Clean up rarity (remove +1/+2/+3 variants for simpler output)
	rarity = strings.TrimSpace(rarity)

	return
}

func shouldSkip(line string, patterns []string) bool {
	for _, p := range patterns {
		if strings.HasPrefix(line, p) {
			return true
		}
	}
	// Skip lines that look like table rows (number followed by tab)
	if len(line) > 0 && line[0] >= '0' && line[0] <= '9' && strings.Contains(line, "\t") {
		return true
	}
	return false
}

func writeJSON(filename string, items []MagicItem) {
	file := ItemFile{Items: items}
	data, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		fmt.Println("Error marshaling:", err)
		return
	}
	if err := os.WriteFile(filename, data, 0644); err != nil {
		fmt.Println("Error writing file:", err)
		return
	}
	fmt.Println("Wrote", filename)
}
