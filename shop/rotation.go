package shop

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/egotch/dnd-shopkeep/ai"
	"github.com/egotch/dnd-shopkeep/config"
)

// GetSessionSpecials returns the special items available for the current session
func GetSessionSpecials() []Item {
	items, err := loadItemsFromFile(config.DataPaths.SessionSpecials, "specials")
	if err != nil {
		return []Item{}
	}
	return items
}

// FormatSessionSpecials returns a formatted string of the session specials
func FormatSessionSpecials() string {
	items := GetSessionSpecials()
	return FormatItemList(items)
}

// CuratorItem represents a single item recommendation from the curator LLM
type CuratorItem struct {
	Name   string `json:"name"`
	Reason string `json:"reason"`
}

// CuratorSelection represents the curator's picks for one character
type CuratorSelection struct {
	Character string        `json:"character"`
	Items     []CuratorItem `json:"items"`
}

// CuratorResponse is the top-level response from the curator LLM
type CuratorResponse struct {
	Selections []CuratorSelection `json:"selections"`
}

// LoadAllMagicItems loads items from all magic item data files
func LoadAllMagicItems() ([]MagicItem, error) {
	paths := []string{
		config.DataPaths.MagicWeapons,
		config.DataPaths.MagicArmor,
		config.DataPaths.MagicPotions,
		config.DataPaths.WondrousItems,
	}

	var allItems []MagicItem
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", path, err)
		}

		var file struct {
			Items []MagicItem `json:"items"`
		}
		if err := json.Unmarshal(data, &file); err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", path, err)
		}

		allItems = append(allItems, file.Items...)
	}

	return allItems, nil
}

// FilterItemsByLevel filters magic items to only those whose rarity is allowed at the given level
func FilterItemsByLevel(items []MagicItem, level int) []MagicItem {
	var filtered []MagicItem
	for _, item := range items {
		if IsRarityAllowed(item.Rarity, level) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

// curatorSystemPrompt is the system prompt for the item curator (separate from Grash personality)
const curatorSystemPrompt = `You are a D&D 5e magic item curator. Your job is to select personalized magic items for each character in a party.

RULES:
- Select exactly 4 items per character
- Pick items that match the character's class, backstory, and playstyle
- Mix of 1-2 mechanically useful items + 1-2 thematically interesting items
- No duplicate items across characters (each item can only be recommended once)
- Return ONLY valid JSON, no other text
- CRITICAL: The "name" field must be copied EXACTLY from the item pool below. Do not rename, modify, abbreviate, or add parenthetical notes to item names. If the pool says "Instrument of the Bards", write exactly "Instrument of the Bards", NOT "Bard's Instrument (Instrument of the Bards)".
- Only select items that appear in the provided pool. Do not invent items or reference the character's existing equipment.

RESPONSE FORMAT (JSON only):
{
  "selections": [
    {
      "character": "Character Name",
      "items": [
        {"name": "Exact Item Name From Pool", "reason": "Brief reason why this suits them"},
        {"name": "Exact Item Name From Pool", "reason": "Brief reason why this suits them"},
        {"name": "Exact Item Name From Pool", "reason": "Brief reason why this suits them"},
        {"name": "Exact Item Name From Pool", "reason": "Brief reason why this suits them"}
      ]
    }
  ]
}`

// buildCuratorMessage constructs the user message with character profiles and item pool
func buildCuratorMessage(characters []*Character, items []MagicItem) string {
	var sb strings.Builder

	sb.WriteString("## Characters\n\n")
	for _, char := range characters {
		sb.WriteString(fmt.Sprintf("### %s (%s)\n", char.Name, char.ClassLevel))
		sb.WriteString(fmt.Sprintf("- Backstory: %s\n", char.BackstorySummary))
		if char.Playstyle != "" {
			sb.WriteString(fmt.Sprintf("- Playstyle: %s\n", char.Playstyle))
		}
		if len(char.CurrentInventory) > 0 {
			sb.WriteString(fmt.Sprintf("- Current inventory: %s\n", strings.Join(char.CurrentInventory, ", ")))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("## Available Item Pool\n\n")
	for _, item := range items {
		sb.WriteString(fmt.Sprintf("- %s (%s)\n", item.Name, normalizeRarity(item.Rarity)))
	}

	sb.WriteString(fmt.Sprintf("\nSelect exactly 4 items for each of the %d characters. Return ONLY JSON.\n", len(characters)))

	return sb.String()
}

// extractJSON attempts to extract valid JSON from a raw LLM response
func extractJSON(raw string) ([]byte, error) {
	// Try raw parse first
	raw = strings.TrimSpace(raw)
	if json.Valid([]byte(raw)) {
		return []byte(raw), nil
	}

	// Try extracting from markdown code fence
	if idx := strings.Index(raw, "```json"); idx >= 0 {
		start := idx + len("```json")
		if end := strings.Index(raw[start:], "```"); end >= 0 {
			candidate := strings.TrimSpace(raw[start : start+end])
			if json.Valid([]byte(candidate)) {
				return []byte(candidate), nil
			}
		}
	}
	if idx := strings.Index(raw, "```"); idx >= 0 {
		start := idx + len("```")
		if end := strings.Index(raw[start:], "```"); end >= 0 {
			candidate := strings.TrimSpace(raw[start : start+end])
			if json.Valid([]byte(candidate)) {
				return []byte(candidate), nil
			}
		}
	}

	// Try brace matching — find first { and last }
	first := strings.Index(raw, "{")
	last := strings.LastIndex(raw, "}")
	if first >= 0 && last > first {
		candidate := raw[first : last+1]
		if json.Valid([]byte(candidate)) {
			return []byte(candidate), nil
		}
	}

	return nil, fmt.Errorf("could not extract valid JSON from response")
}

// validateSelections checks curator selections against the actual item pool,
// dropping any hallucinated items and logging warnings
func validateSelections(response *CuratorResponse, pool []MagicItem) {
	// Build lookup structures
	poolNames := make(map[string]bool, len(pool))
	poolList := make([]string, 0, len(pool))
	for _, item := range pool {
		lower := strings.ToLower(item.Name)
		poolNames[lower] = true
		poolList = append(poolList, lower)
	}

	seen := make(map[string]bool)
	for i := range response.Selections {
		var valid []CuratorItem
		for _, ci := range response.Selections[i].Items {
			nameLower := strings.ToLower(ci.Name)

			// Try exact match first
			matched := nameLower
			if !poolNames[matched] {
				// Try stripping parenthetical suffixes: "Cape of the Mountebank (with modifications)" -> "cape of the mountebank"
				matched = fuzzyMatchPool(nameLower, poolNames, poolList)
			}

			if matched == "" {
				slog.Warn("curator hallucinated item, dropping",
					"item", ci.Name,
					"character", response.Selections[i].Character)
				continue
			}
			if seen[matched] {
				slog.Warn("curator duplicated item across characters, dropping",
					"item", ci.Name,
					"character", response.Selections[i].Character)
				continue
			}
			seen[matched] = true
			// Fix the name to the canonical pool name
			ci.Name = canonicalName(matched, pool)
			valid = append(valid, ci)
		}
		response.Selections[i].Items = valid
	}
}

// fuzzyMatchPool tries to match an LLM-returned name against the item pool.
// Strategies: strip parentheticals, check if a pool name is contained in the response or vice versa.
func fuzzyMatchPool(name string, poolNames map[string]bool, poolList []string) string {
	// Strip parenthetical suffix: "goggles of night (for enhanced magical sight)" -> "goggles of night"
	if idx := strings.Index(name, " ("); idx > 0 {
		stripped := strings.TrimSpace(name[:idx])
		if poolNames[stripped] {
			return stripped
		}
	}

	// Check if any pool item name is contained within the LLM name
	for _, poolName := range poolList {
		if strings.Contains(name, poolName) {
			return poolName
		}
	}

	// Check if the LLM name is contained within any pool item name
	for _, poolName := range poolList {
		if strings.Contains(poolName, name) {
			return poolName
		}
	}

	return ""
}

// canonicalName returns the original-cased item name from the pool
func canonicalName(lowerName string, pool []MagicItem) string {
	for _, item := range pool {
		if strings.ToLower(item.Name) == lowerName {
			return item.Name
		}
	}
	return lowerName
}

// selectionsToItems converts validated curator selections to shop Items with rolled prices
func selectionsToItems(response *CuratorResponse, pool []MagicItem) []Item {
	// Build lookup map (lowercase name -> MagicItem)
	poolMap := make(map[string]MagicItem, len(pool))
	for _, item := range pool {
		poolMap[strings.ToLower(item.Name)] = item
	}

	var items []Item
	for _, sel := range response.Selections {
		for _, ci := range sel.Items {
			magicItem, ok := poolMap[strings.ToLower(ci.Name)]
			if !ok {
				continue
			}
			shopItem := magicItem.ToShopItem()
			shopItem.Category = "specials"
			shopItem.Description = fmt.Sprintf("Recommended for %s: %s | %s",
				sel.Character, ci.Reason, magicItem.Description)
			if magicItem.Attunement != "" {
				shopItem.Rarity = fmt.Sprintf("%s, %s", normalizeRarity(magicItem.Rarity), magicItem.Attunement)
			}
			items = append(items, shopItem)
		}
	}

	return items
}

// RefreshSessionSpecials uses LLM curation to generate personalized magic item recommendations
func RefreshSessionSpecials() ([]Item, error) {
	slog.Info("refreshing session specials via LLM curation")

	// 1. Load all magic items
	allItems, err := LoadAllMagicItems()
	if err != nil {
		return nil, fmt.Errorf("failed to load magic items: %w", err)
	}
	slog.Info("loaded magic items", "count", len(allItems))

	// 2. Load all characters
	characters := GetAllCharacters()
	if len(characters) == 0 {
		return nil, fmt.Errorf("no characters found")
	}
	slog.Info("loaded characters", "count", len(characters))

	// 3. Find minimum party level and filter by rarity
	minLevel := 20
	for _, char := range characters {
		level := ParseLevel(char.ClassLevel)
		if level < minLevel {
			minLevel = level
		}
	}
	filtered := FilterItemsByLevel(allItems, minLevel)
	slog.Info("filtered items by level", "min_level", minLevel, "eligible", len(filtered))

	// 4. Build curator prompt and send to Ollama
	conv := ai.NewConversation("llama3.1:8b", curatorSystemPrompt)
	userMsg := buildCuratorMessage(characters, filtered)
	conv.AddMessage("user", userMsg)

	slog.Info("sending curator prompt to Ollama", "message_length", len(userMsg))
	start := time.Now()

	rawResponse, err := conv.SendToOllamaWithTimeout(5*time.Minute, map[string]any{
		"num_ctx": 8192,
	})
	if err != nil {
		return nil, fmt.Errorf("ollama curator call failed: %w", err)
	}
	slog.Info("ollama curator response received", "duration", time.Since(start))

	// 5. Parse and validate response
	jsonBytes, err := extractJSON(rawResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse curator response: %w\nRaw response:\n%s", err, rawResponse)
	}

	var curatorResp CuratorResponse
	if err := json.Unmarshal(jsonBytes, &curatorResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal curator response: %w", err)
	}

	validateSelections(&curatorResp, filtered)

	// 6. Convert to shop Items with rolled prices
	items := selectionsToItems(&curatorResp, filtered)
	slog.Info("generated session specials", "item_count", len(items))

	// 7. Write to session_specials.json
	output := itemFile{Items: items}
	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal specials: %w", err)
	}

	if err := os.WriteFile(config.DataPaths.SessionSpecials, data, 0644); err != nil {
		return nil, fmt.Errorf("failed to write session specials: %w", err)
	}

	slog.Info("session specials written", "path", config.DataPaths.SessionSpecials)
	return items, nil
}
