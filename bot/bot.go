package bot

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/egotch/dnd-shopkeep/ai"
)

var BotToken string
var systemPrompt string = `You are a helpful, but sassy discord bot.
		- Remember the information the user tells you, and use it to inform future responses.
		- Your persona on the discord server is the Yellow Valkyrie, but we like to call you Val for short.
		- Have a playful, sarcastic, and slightly flirty attitude.
		- Be flirtatios and don't be afraid to cross the line into being a bit inappropriate or sexual.
		- Be playful with your responses, and use emojis when apppropriate.
		- The discord channel you are connected to is one for a small group of middle aged gamers that are Millenials (born in late 80s).
		- There are 3 members to the discord chat: Tim, Eric, and Dieter.
		- We play lots of dota (defense of the ancients).

	Here is the format at the input and response messages:
		- Input messages from the users will be prefixed with their username
			- example: "[username]: message content ..."

		- Output message should not include the original message from the user.
		- Output message should just be your response to the user.
			- When responding, address the user by their username when appropriate.
			- example: "Hey! I think you should try ..."
		`
// init the ai conversation
var conv =  ai.NewConversation("llama3.1:8b", systemPrompt)

func Run() {

	// create the sessoin
	discord, err := discordgo.New("Bot " + BotToken)
	if err != nil {
		log.Fatalf("Unable to start session: %v", err)
	}


	// add event handlers
	discord.AddHandler(newMessage)
	// discord.AddHandler(pingPong)
	// discord.AddHandler(userJoin)

	// open the session
	discord.Open()
	defer discord.Close()

	// keep the bot up until someone ctrl+c's it
	fmt.Println("Bot is up.  Press CTRL-C to exit...")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}

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

func newMessage(discord *discordgo.Session, messageEvent *discordgo.MessageCreate) {

	var respMessage string

	// Prevent the bot from responding to its own messages
	// by checking the author ID
	if messageEvent.Author.ID == discord.State.User.ID {
		return
	}

	if checkMessageContents(messageEvent, "val"){
		slog.Info("message received, sending to ollama")
		augmentMessageWithUsername(messageEvent.Message)

		conv.AddMessage("user", messageEvent.Content)
		response, err := conv.SendToOllama()
		if err != nil {
			respMessage = fmt.Sprintf("Error: %v\n", err)
		} else {
			respMessage = response
		}

		postMsg(respMessage, discord, messageEvent)
	}
}

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

func pingPong(discord *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == discord.State.User.ID {
		return
	}

	switch m.Content {
	case "ping":
		discord.ChannelMessageSend(m.ChannelID, "Pong!")
	case "pong":
		discord.ChannelMessageSend(m.ChannelID, "Ping!")
	}
}

func userJoin(discord *discordgo.Session, pu *discordgo.GuildMemberUpdate) {

	if pu.User.ID == discord.State.User.ID {
		return
	}

	log.Println(pu.User.Username)
}

