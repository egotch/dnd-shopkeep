package shop

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Character represents a player character profile
type Character struct {
	Name             string   `json:"name"`
	ClassLevel       string   `json:"class_level"`
	DiscordHandle    string   `json:"discord_handle"`
	CurrentInventory []string `json:"current_inventory"`
	BackstorySummary string   `json:"backstory_summary"`
}

var charactersPath = "data/characters"

// userCharacterMap maps Discord usernames to character file names (built dynamically)
var userCharacterMap map[string]string
var characterCache map[string]*Character
var mapOnce sync.Once

// initCharacterMap scans the characters directory and builds the username mapping
func initCharacterMap() {
	userCharacterMap = make(map[string]string)
	characterCache = make(map[string]*Character)

	entries, err := os.ReadDir(charactersPath)
	if err != nil {
		fmt.Printf("Warning: could not read characters directory: %v\n", err)
		return
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		// Get filename without .json extension
		charFile := strings.TrimSuffix(entry.Name(), ".json")

		// Load the character to get their Discord handle
		char, err := loadCharacterFile(charFile)
		if err != nil {
			fmt.Printf("Warning: could not load character %s: %v\n", charFile, err)
			continue
		}

		// Map Discord handle to character file
		if char.DiscordHandle != "" {
			handle := strings.ToLower(char.DiscordHandle)
			userCharacterMap[handle] = charFile
			characterCache[charFile] = char
		}
	}
}

// ensureMapLoaded ensures the character map is initialized
func ensureMapLoaded() {
	mapOnce.Do(initCharacterMap)
}

// loadCharacterFile loads a character without using the cache (internal use)
func loadCharacterFile(name string) (*Character, error) {
	filename := filepath.Join(charactersPath, name+".json")
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read character '%s': %w", name, err)
	}

	var character Character
	if err := json.Unmarshal(data, &character); err != nil {
		return nil, fmt.Errorf("failed to parse character '%s': %w", name, err)
	}

	return &character, nil
}

// LoadCharacter loads a character profile by filename (without .json extension)
func LoadCharacter(name string) (*Character, error) {
	ensureMapLoaded()

	// Check cache first
	if char, exists := characterCache[name]; exists {
		return char, nil
	}

	// Load from file if not cached
	return loadCharacterFile(name)
}

// GetCharacterForUser returns the character filename for a Discord username
func GetCharacterForUser(discordUsername string) (string, error) {
	ensureMapLoaded()

	username := strings.ToLower(discordUsername)
	charName, exists := userCharacterMap[username]
	if !exists {
		return "", fmt.Errorf("no character found for user '%s'", discordUsername)
	}
	return charName, nil
}

// GetCharacterNameForUser returns the character's display name for a Discord username
func GetCharacterNameForUser(discordUsername string) (string, error) {
	charFile, err := GetCharacterForUser(discordUsername)
	if err != nil {
		return "", err
	}

	char, err := LoadCharacter(charFile)
	if err != nil {
		return "", err
	}

	return char.Name, nil
}

// GetAllCharacters returns all loaded characters
func GetAllCharacters() []*Character {
	ensureMapLoaded()

	chars := make([]*Character, 0, len(characterCache))
	for _, char := range characterCache {
		chars = append(chars, char)
	}
	return chars
}

// GetUserCharacterMap returns a copy of the Discord handle to character file mapping
func GetUserCharacterMap() map[string]string {
	ensureMapLoaded()

	// Return a copy to prevent external modification
	result := make(map[string]string, len(userCharacterMap))
	for k, v := range userCharacterMap {
		result[k] = v
	}
	return result
}

// ReloadCharacters forces a reload of the character map (useful after adding new characters)
func ReloadCharacters() {
	mapOnce = sync.Once{} // Reset the once
	ensureMapLoaded()
}

// FormatInventory returns a formatted string of the character's inventory
func (c *Character) FormatInventory() string {
	if len(c.CurrentInventory) == 0 {
		return "Your pack is empty."
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("**%s's Inventory** (%s)\n\n", c.Name, c.ClassLevel))
	for _, item := range c.CurrentInventory {
		sb.WriteString(fmt.Sprintf("• %s\n", item))
	}
	return sb.String()
}

// FormatCharacterSummary returns a brief summary of the character for LLM context
func (c *Character) FormatCharacterSummary() string {
	return fmt.Sprintf("%s (%s): %s", c.Name, c.ClassLevel, c.BackstorySummary)
}
