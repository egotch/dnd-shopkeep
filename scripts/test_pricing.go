package main

import (
	"fmt"

	"github.com/egotch/dnd-shopkeep/shop"
)

func main() {
	fmt.Println(shop.FormatPricingTable())

	fmt.Println("\n**Example Price Rolls:**")

	// Test rolling prices for different rarities
	rarities := []string{"Common", "Uncommon", "Rare", "Very Rare", "Legendary"}
	for _, r := range rarities {
		tier := shop.GetPriceTier(r)
		fmt.Printf("%s: %d gp (range: %d-%d)\n", r, tier.RollPrice(), tier.Min, tier.Max)
	}

	fmt.Println("\n**Consumable Price Rolls:**")
	for _, r := range rarities[:4] {
		tier := shop.GetConsumablePriceTier(r)
		fmt.Printf("%s Potion: %d gp (range: %d-%d)\n", r, tier.RollPrice(), tier.Min, tier.Max)
	}

	fmt.Println("\n**MagicItem.ToShopItem() Example:**")
	item := shop.MagicItem{
		Name:        "Cloak of Elvenkind",
		Rarity:      "Uncommon",
		Description: "Advantage on Stealth checks.",
	}
	shopItem := item.ToShopItem()
	fmt.Printf("Name: %s, Cost: %.0f gp, Desc: %s\n", shopItem.Name, shopItem.Cost, shopItem.Description)
}
