package bot

import "github.com/bwmarrin/discordgo"

// Commands defines all slash commands for the shop bot
var Commands = []*discordgo.ApplicationCommand{
	{
		Name:        "shop",
		Description: "Browse the quartermaster's wares",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "category",
				Description: "Filter by category (weapons, armor, potions, gear, wondrous)",
				Required:    false,
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{Name: "All Items", Value: "all"},
					{Name: "Weapons", Value: "weapons"},
					{Name: "Armor", Value: "armor"},
					{Name: "Potions", Value: "potions"},
					{Name: "Gear", Value: "gear"},
					{Name: "Wondrous Items", Value: "wondrous"},
					{Name: "Monthly Specials", Value: "monthly"},
				},
			},
		},
	},
	{
		Name:        "buy",
		Description: "Purchase an item from the shop",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "item",
				Description: "Name of the item to purchase",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "quantity",
				Description: "How many to purchase (default: 1)",
				Required:    false,
				MinValue:    floatPtr(1),
				MaxValue:    10,
			},
		},
	},
	{
		Name:        "inventory",
		Description: "View your character's current inventory",
	},
	{
		Name:        "history",
		Description: "View your purchase history",
	},
}

// CommandHandlers maps command names to their handler functions
var CommandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"shop":      handleShop,
	"buy":       handleBuy,
	"inventory": handleInventory,
	"history":   handleHistory,
}

// floatPtr is a helper to create a *float64 for MinValue
func floatPtr(f float64) *float64 {
	return &f
}
