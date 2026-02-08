package shop

import (
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
