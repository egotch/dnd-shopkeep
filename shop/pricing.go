package shop

import (
	"fmt"
	"math/rand/v2"
	"strings"
)

// PriceTier defines the price range for a rarity level
type PriceTier struct {
	Rarity string
	Min    int
	Max    int
}

// PricingTable defines the roll table for magic item prices by rarity
// Based on DMG guidelines with some shop markup
var PricingTable = []PriceTier{
	{Rarity: "Common", Min: 50, Max: 100},
	{Rarity: "Uncommon", Min: 100, Max: 500},
	{Rarity: "Rare", Min: 500, Max: 5000},
	{Rarity: "Very Rare", Min: 5000, Max: 50000},
	{Rarity: "Legendary", Min: 50000, Max: 200000},
	{Rarity: "Artifact", Min: 200000, Max: 500000},
}

// ConsumablePricingTable has reduced prices for one-use items (potions, scrolls, ammo)
var ConsumablePricingTable = []PriceTier{
	{Rarity: "Common", Min: 25, Max: 75},
	{Rarity: "Uncommon", Min: 75, Max: 300},
	{Rarity: "Rare", Min: 300, Max: 3000},
	{Rarity: "Very Rare", Min: 3000, Max: 30000},
	{Rarity: "Legendary", Min: 30000, Max: 100000},
}

// GetPriceTier returns the price tier for a given rarity
func GetPriceTier(rarity string) *PriceTier {
	rarity = normalizeRarity(rarity)
	for _, tier := range PricingTable {
		if strings.EqualFold(tier.Rarity, rarity) {
			return &tier
		}
	}
	// Default to Uncommon if rarity not found
	return &PricingTable[1]
}

// GetConsumablePriceTier returns the consumable price tier for a given rarity
func GetConsumablePriceTier(rarity string) *PriceTier {
	rarity = normalizeRarity(rarity)
	for _, tier := range ConsumablePricingTable {
		if strings.EqualFold(tier.Rarity, rarity) {
			return &tier
		}
	}
	return &ConsumablePricingTable[1]
}

// RollPrice generates a random price within the tier's range
// Uses a weighted distribution favoring the middle of the range
func (t *PriceTier) RollPrice() int {
	if t.Min == t.Max {
		return t.Min
	}

	// Roll 2d100-style for bell curve distribution toward middle
	roll1 := rand.IntN(t.Max-t.Min+1) + t.Min
	roll2 := rand.IntN(t.Max-t.Min+1) + t.Min
	avg := (roll1 + roll2) / 2

	// Round to nice numbers
	return roundToNiceNumber(avg)
}

// RollPriceLinear generates a uniformly random price within the tier's range
func (t *PriceTier) RollPriceLinear() int {
	if t.Min == t.Max {
		return t.Min
	}
	price := rand.IntN(t.Max-t.Min+1) + t.Min
	return roundToNiceNumber(price)
}

// RollPriceForItem determines price based on item rarity, using consumable table for potions/scrolls/ammo
func RollPriceForItem(item MagicItem) int {
	name := strings.ToLower(item.Name)
	isConsumable := strings.Contains(name, "potion") ||
		strings.Contains(name, "scroll") ||
		strings.Contains(name, "ammunition") ||
		strings.Contains(name, "oil") ||
		strings.Contains(name, "elixir") ||
		strings.Contains(name, "philter")

	var tier *PriceTier
	if isConsumable {
		tier = GetConsumablePriceTier(item.Rarity)
	} else {
		tier = GetPriceTier(item.Rarity)
	}

	return tier.RollPrice()
}

// MagicItem represents an item from the magic item reference files
type MagicItem struct {
	Name        string `json:"name"`
	Type        string `json:"type,omitempty"`
	Rarity      string `json:"rarity"`
	Attunement  string `json:"attunement,omitempty"`
	Description string `json:"description"`
}

// ToShopItem converts a MagicItem to a shop Item with a rolled price
func (m *MagicItem) ToShopItem() Item {
	return Item{
		Name:        m.Name,
		Description: m.Description,
		Rarity:      m.Rarity,
		Cost:        float64(RollPriceForItem(*m)),
	}
}

// normalizeRarity handles variations like "Rarity Varies", "Very Rare", etc.
func normalizeRarity(rarity string) string {
	rarity = strings.TrimSpace(rarity)

	// Handle "Rarity Varies" - default to Rare
	if strings.Contains(strings.ToLower(rarity), "varies") {
		return "Rare"
	}

	// Handle compound rarities like "Uncommon (+1), Rare (+2)"
	if strings.Contains(rarity, ",") {
		// Take the first rarity mentioned
		parts := strings.Split(rarity, ",")
		rarity = strings.TrimSpace(parts[0])
		// Remove any parenthetical like "(+1)"
		if idx := strings.Index(rarity, "("); idx > 0 {
			rarity = strings.TrimSpace(rarity[:idx])
		}
	}

	return rarity
}

// roundToNiceNumber rounds a price to a "shop friendly" number
func roundToNiceNumber(price int) int {
	switch {
	case price < 100:
		// Round to nearest 5
		return ((price + 2) / 5) * 5
	case price < 1000:
		// Round to nearest 25
		return ((price + 12) / 25) * 25
	case price < 10000:
		// Round to nearest 100
		return ((price + 50) / 100) * 100
	default:
		// Round to nearest 500
		return ((price + 250) / 500) * 500
	}
}

// FormatPricingTable returns a formatted string of the pricing table for reference
func FormatPricingTable() string {
	var sb strings.Builder
	sb.WriteString("**Magic Item Pricing Table**\n\n")
	sb.WriteString("| Rarity | Min GP | Max GP |\n")
	sb.WriteString("|--------|--------|--------|\n")
	for _, tier := range PricingTable {
		sb.WriteString(formatTableRow(tier))
	}
	sb.WriteString("\n**Consumable Pricing (Potions, Scrolls, Oils)**\n\n")
	sb.WriteString("| Rarity | Min GP | Max GP |\n")
	sb.WriteString("|--------|--------|--------|\n")
	for _, tier := range ConsumablePricingTable {
		sb.WriteString(formatTableRow(tier))
	}
	return sb.String()
}

func formatTableRow(tier PriceTier) string {
	return "| " + tier.Rarity + " | " + formatGold(tier.Min) + " | " + formatGold(tier.Max) + " |\n"
}

func formatGold(amount int) string {
	if amount >= 1000 {
		return fmt.Sprintf("%d,000", amount/1000)
	}
	return fmt.Sprintf("%d", amount)
}
