package main

import (
	"os"

	bot "github.com/egotch/dnd-shopkeep/bot"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	bot.BotToken = os.Getenv("BOT_TOKEN")
	bot.GuildID = os.Getenv("GUILD_ID") // Optional: set for faster dev registration

	bot.Run()
}
