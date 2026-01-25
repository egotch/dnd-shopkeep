package shop

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Character represents a player character profile
type Character struct {
	Name             string   `json:"name"`
	ClassLevel       string   `json:"class_level"`
	CurrentInventory []string `json:"current_inventory"`
	BackstorySummary string   `json:"backstory_summary"`
}

var charactersPath = "data/characters"

// UserCharacterMap maps Discord usernames to character file names
var UserCharacterMap = map[string]string{
	"timmehhey": "tim_paladin",
	"egotch":    "eric_wizard",
	"dhrudolp":  "dieter_rogue",
}

// LoadCharacter loads a character profile by filename (without .json extension)
func LoadCharacter(name string) (*Character, error) {
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

// GetCharacterForUser returns the character filename for a Discord username
func GetCharacterForUser(discordUsername string) (string, error) {
	username := strings.ToLower(discordUsername)
	charName, exists := UserCharacterMap[username]
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
