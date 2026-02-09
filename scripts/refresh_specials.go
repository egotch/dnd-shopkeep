package main

import (
	"fmt"
	"os"

	"github.com/egotch/dnd-shopkeep/shop"
)

func main() {
	fmt.Println("Refreshing session specials via LLM curation...")
	fmt.Println("This may take a few minutes while Ollama generates recommendations.")
	fmt.Println()

	items, err := shop.RefreshSessionSpecials()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Session specials refreshed! %d items generated.\n\n", len(items))
	fmt.Println(shop.FormatItemList(items))
}
