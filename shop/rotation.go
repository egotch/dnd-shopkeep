package shop

import (
	"hash/fnv"
	"math/rand"
	"time"
)

// UncommonItems is the master list of uncommon items for rotation
var UncommonItems = []Item{
	{Name: "Bag of Holding", Category: "wondrous", Price: 500, Description: "This bag has an interior space considerably larger than its outside dimensions.", Rarity: "uncommon"},
	{Name: "Boots of Elvenkind", Category: "wondrous", Price: 500, Description: "Your steps make no sound.", Rarity: "uncommon"},
	{Name: "Cloak of Elvenkind", Category: "wondrous", Price: 500, Description: "Advantage on Stealth checks.", Rarity: "uncommon"},
	{Name: "Goggles of Night", Category: "wondrous", Price: 500, Description: "Grants 60 ft. darkvision.", Rarity: "uncommon"},
	{Name: "Immovable Rod", Category: "wondrous", Price: 500, Description: "This rod stays in place when activated.", Rarity: "uncommon"},
	{Name: "Ring of Jumping", Category: "ring", Price: 500, Description: "Cast Jump on yourself at will.", Rarity: "uncommon"},
	{Name: "Ring of Water Walking", Category: "ring", Price: 500, Description: "Stand and walk on liquid surfaces.", Rarity: "uncommon"},
	{Name: "Wand of Magic Detection", Category: "wand", Price: 500, Description: "3 charges, cast Detect Magic.", Rarity: "uncommon"},
	{Name: "Wand of Secrets", Category: "wand", Price: 500, Description: "3 charges, detect secret doors and traps.", Rarity: "uncommon"},
	{Name: "Circlet of Blasting", Category: "wondrous", Price: 500, Description: "Cast Scorching Ray once per day.", Rarity: "uncommon"},
	{Name: "Gloves of Thievery", Category: "wondrous", Price: 500, Description: "+5 bonus to Sleight of Hand and lockpicking.", Rarity: "uncommon"},
	{Name: "Hat of Disguise", Category: "wondrous", Price: 500, Description: "Cast Disguise Self at will.", Rarity: "uncommon"},
	{Name: "Lantern of Revealing", Category: "wondrous", Price: 500, Description: "Reveals invisible creatures in 30 ft.", Rarity: "uncommon"},
	{Name: "Slippers of Spider Climbing", Category: "wondrous", Price: 500, Description: "Walk on walls and ceilings.", Rarity: "uncommon"},
	{Name: "Sending Stones", Category: "wondrous", Price: 500, Description: "Cast Sending once per day between paired stones.", Rarity: "uncommon"},
}

// GetMonthlyRotation returns the rotating uncommon items for a given month
// Uses seed-based selection so the same month always produces the same rotation
func GetMonthlyRotation(month string) []Item {
	if month == "" {
		month = time.Now().Format("2006-01")
	}

	// Create deterministic seed from month string
	h := fnv.New64a()
	h.Write([]byte(month))
	seed := int64(h.Sum64())

	// Create seeded RNG
	rng := rand.New(rand.NewSource(seed))

	// Shuffle a copy of the uncommon items
	items := make([]Item, len(UncommonItems))
	copy(items, UncommonItems)

	rng.Shuffle(len(items), func(i, j int) {
		items[i], items[j] = items[j], items[i]
	})

	// Return first 5 items
	rotationSize := 5
	if len(items) < rotationSize {
		rotationSize = len(items)
	}

	return items[:rotationSize]
}

// GetCurrentMonth returns the current month in YYYY-MM format
func GetCurrentMonth() string {
	return time.Now().Format("2006-01")
}

// FormatRotationItems returns a formatted string of the monthly rotation
func FormatRotationItems(month string) string {
	items := GetMonthlyRotation(month)
	return FormatItemList(items)
}
