package main

import (
	"fmt"

	"github.com/egotch/dnd-shopkeep/shop"
)

func main() {
	fmt.Println("=== Dynamic Character Loading Test ===\n")

	// Get all characters
	chars := shop.GetAllCharacters()
	fmt.Printf("Found %d characters:\n\n", len(chars))

	for _, char := range chars {
		fmt.Printf("  %s (%s)\n", char.Name, char.ClassLevel)
		fmt.Printf("    Discord: %s\n", char.DiscordHandle)
		fmt.Printf("    Items: %d\n\n", len(char.CurrentInventory))
	}

	// Show the username map
	fmt.Println("=== Discord Handle Mapping ===\n")
	userMap := shop.GetUserCharacterMap()
	for handle, charFile := range userMap {
		fmt.Printf("  @%s -> %s\n", handle, charFile)
	}

	// Test lookup
	fmt.Println("\n=== Lookup Test ===\n")
	testHandles := []string{"egotch", "dhrudolp", "CalciferMashi", "unknown_user"}
	for _, handle := range testHandles {
		charFile, err := shop.GetCharacterForUser(handle)
		if err != nil {
			fmt.Printf("  @%s -> ERROR: %v\n", handle, err)
		} else {
			char, _ := shop.LoadCharacter(charFile)
			fmt.Printf("  @%s -> %s (%s)\n", handle, char.Name, char.ClassLevel)
		}
	}
}
