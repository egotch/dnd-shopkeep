package main

import (
	bot "github.com/egotch/dnd-shopkeep/bot"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	bot.BotToken = godotenv.Get("BOT_TOKEN")
	bot.Run()
}
