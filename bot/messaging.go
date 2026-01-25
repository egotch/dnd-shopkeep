package bot

import (
	"fmt"
	"log"
	"log/slog"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// postMsg is a simple function to post a message (string) to the discord channel
// takes in the discord session and message create methods
func postMsg(msgPost string, discord *discordgo.Session, message *discordgo.MessageCreate) {

	msg, err := discord.ChannelMessageSend(message.ChannelID, msgPost)
	if err != nil {
		log.Fatalf("failed to post message: %v", err)
	}
	log.Println(msg)
}

// checkMessageContents is a helper function that checks the contents of
// an incomming message for the string
//
// used to speed up switch/cases
func checkMessageContents(message *discordgo.MessageCreate, checkStr string) (found bool) {

	if strings.Contains(strings.ToLower(message.Content), checkStr) {
		return true
	} else {
		return false
	}
}

// Note: newMessage handler is now in bot.go with updated quartermaster logic

// augmentMessageWithUsername takes in a discord message and
// returns a new message with the username prepended to the content
// resulting format is "User [username]: message content ..."
func augmentMessageWithUsername(message *discordgo.Message) *discordgo.Message {

	var newContent strings.Builder
	var usernameMap = map[string]string{
		"timmehhey":   "Tim",
		"egotch":  "Eric",
		"dhrudolp": "Dieter",
	}

	// get the user from the message
	username := strings.ToLower(message.Author.Username)
	mappedUserName, exists := usernameMap[username]
	if ! exists {
		mappedUserName = username
	}

	slog.Info(fmt.Sprintf("augmenting message with username: %s", mappedUserName))

	newContent.WriteString(fmt.Sprintf("User [%s]: ", mappedUserName))
	newContent.WriteString(message.Content)

	message.Content = newContent.String()
	return message
}
