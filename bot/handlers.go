package bot

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/egotch/dnd-shopkeep/shop"
)

// handleShop processes the /shop command
func handleShop(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options
	category := "all"
	if len(options) > 0 {
		category = options[0].StringValue()
	}

	slog.Info("shop command received", "category", category, "user", i.Member.User.Username)

	catalog, err := shop.LoadCatalog()
	if err != nil {
		respondWithError(s, i, "Failed to load catalog: "+err.Error())
		return
	}

	var items []shop.Item
	var title string

	if category == "monthly" {
		month := shop.GetCurrentMonth()
		items = shop.GetMonthlyRotation(month)
		title = fmt.Sprintf("Monthly Specials (%s)", month)
	} else {
		items = catalog.GetItemsByCategory(category)
		if category == "all" {
			title = "All Shop Items"
		} else {
			title = strings.Title(category)
		}
	}

	if len(items) == 0 {
		respondWithMessage(s, i, fmt.Sprintf("No items found in category: %s", category))
		return
	}

	// Get character context for personalized recommendations
	charFile, _ := shop.GetCharacterForUser(i.Member.User.Username)
	char, _ := shop.LoadCharacter(charFile)

	// Build response with AI flavor if available
	var response string
	if char != nil {
		// Send to Ollama for quartermaster flavor
		prompt := fmt.Sprintf("[%s]: Show me %s items", char.Name, category)
		conv.AddMessage("user", prompt)
		aiResponse, err := conv.SendToOllama()
		if err == nil {
			response = aiResponse + "\n\n"
		}
	}

	response += fmt.Sprintf("**%s**\n\n%s", title, shop.FormatItemList(items))

	respondWithMessage(s, i, response)
}

// handleBuy processes the /buy command
func handleBuy(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options
	itemName := ""
	quantity := 1

	for _, opt := range options {
		switch opt.Name {
		case "item":
			itemName = opt.StringValue()
		case "quantity":
			quantity = int(opt.IntValue())
		}
	}

	slog.Info("buy command received", "item", itemName, "quantity", quantity, "user", i.Member.User.Username)

	// Get character for this user
	charFile, err := shop.GetCharacterForUser(i.Member.User.Username)
	if err != nil {
		respondWithError(s, i, "You don't have a character registered. Contact the GM.")
		return
	}

	char, err := shop.LoadCharacter(charFile)
	if err != nil {
		respondWithError(s, i, "Failed to load character: "+err.Error())
		return
	}

	// Find the item in catalog
	catalog, err := shop.LoadCatalog()
	if err != nil {
		respondWithError(s, i, "Failed to load catalog: "+err.Error())
		return
	}

	item, err := catalog.FindItem(itemName)
	if err != nil {
		respondWithError(s, i, fmt.Sprintf("Item '%s' not found. Try /shop to see available items.", itemName))
		return
	}

	// Log purchase for each quantity
	totalCost := item.Price * quantity
	for j := 0; j < quantity; j++ {
		if err := shop.AppendPurchase(charFile, *item, "Between sessions"); err != nil {
			respondWithError(s, i, "Failed to record purchase: "+err.Error())
			return
		}
	}

	// Generate AI response for flavor
	prompt := fmt.Sprintf("[%s]: I want to buy %d %s", char.Name, quantity, item.Name)
	conv.AddMessage("user", prompt)
	aiResponse, err := conv.SendToOllama()

	var response string
	if err == nil && aiResponse != "" {
		response = aiResponse + "\n\n"
	}

	response += fmt.Sprintf("**Purchase Recorded!**\n• Item: %s (x%d)\n• Total: %d gp\n• Character: %s\n\n*Remember to deduct gold from your character sheet!*",
		item.Name, quantity, totalCost, char.Name)

	respondWithMessage(s, i, response)
}

// handleInventory processes the /inventory command
func handleInventory(s *discordgo.Session, i *discordgo.InteractionCreate) {
	slog.Info("inventory command received", "user", i.Member.User.Username)

	charFile, err := shop.GetCharacterForUser(i.Member.User.Username)
	if err != nil {
		respondWithError(s, i, "You don't have a character registered. Contact the GM.")
		return
	}

	char, err := shop.LoadCharacter(charFile)
	if err != nil {
		respondWithError(s, i, "Failed to load character: "+err.Error())
		return
	}

	// Also load recent purchases
	history, _ := shop.LoadHistory(charFile)

	response := char.FormatInventory()

	if history != nil && len(history.Purchases) > 0 {
		response += "\n**Recent Purchases (pending session confirmation):**\n"
		for _, p := range history.Purchases {
			response += fmt.Sprintf("• %s (%d gp) - %s\n", p.Item, p.Price, p.Date)
		}
	}

	respondWithMessage(s, i, response)
}

// handleHistory processes the /history command
func handleHistory(s *discordgo.Session, i *discordgo.InteractionCreate) {
	slog.Info("history command received", "user", i.Member.User.Username)

	charFile, err := shop.GetCharacterForUser(i.Member.User.Username)
	if err != nil {
		respondWithError(s, i, "You don't have a character registered. Contact the GM.")
		return
	}

	history, err := shop.LoadHistory(charFile)
	if err != nil {
		respondWithError(s, i, "Failed to load history: "+err.Error())
		return
	}

	char, _ := shop.LoadCharacter(charFile)
	charName := charFile
	if char != nil {
		charName = char.Name
	}

	response := fmt.Sprintf("**Purchase History for %s**\n\n", charName)

	if len(history.Purchases) == 0 {
		response += "No purchases yet. Use /shop to browse available items!"
	} else {
		for _, p := range history.Purchases {
			response += fmt.Sprintf("• **%s** - %d gp (%s)\n", p.Item, p.Price, p.Date)
		}
		response += fmt.Sprintf("\n**Total Spent:** %d gp", history.GetTotalSpent())
	}

	respondWithMessage(s, i, response)
}

// respondWithMessage sends an interaction response
func respondWithMessage(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	// Discord has a 2000 character limit, truncate if needed
	if len(message) > 1900 {
		message = message[:1900] + "\n\n*...response truncated*"
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
		},
	})
	if err != nil {
		slog.Error("failed to respond to interaction", "error", err)
	}
}

// respondWithError sends an error response
func respondWithError(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	slog.Error("command error", "error", message)
	respondWithMessage(s, i, "Error: "+message)
}
