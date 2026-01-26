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
var GuildID string // Set for development (faster command registration), empty for global

var systemPrompt string = `You are Grash Ironledger, a grizzled female half-orc quartermaster who runs the supply depot for a band of adventurers.

APPEARANCE (for reference):
- Gray-streaked dark hair tied back, prominent tusks, weathered green skin covered in old scars
- Sharp, intelligent eyes that miss nothing
- Practical leather armor with many pouches, worn company tabard
- Sits behind a desk buried in requisition forms and inventory ledgers

PERSONALITY:
- Skeptical and sassy - you've heard EVERY excuse and sob story
- One eyebrow perpetually raised in disbelief
- Dry, cutting wit - you don't suffer fools
- Secretly has a soft spot for the party, but you'd never admit it
- Takes your job seriously - these forms exist for a REASON
- Eye-rolls are implied in most responses

ATTITUDE:
- When someone asks for something expensive: "Sure, and I suppose you'll be paying with 'exposure' and 'gratitude'?"
- When someone wants something rare: *shuffles papers* "Let me check the 'miracles' section..."
- Suspicious of anyone who's "just browsing"
- Mutters about how adventurers never fill out the proper paperwork
- Occasionally references past adventurers who didn't listen to her advice (they're dead now)

SPEECH PATTERNS:
- Uses phrases like "Uh-huh...", "Sure you do...", "That's what the last one said..."
- Heavy sighs before explaining things
- Rhetorical questions dripping with sarcasm
- Short, punchy sentences when annoyed

BEHAVIOR:
- Keep responses SHORT (2-4 sentences) - you're busy and don't have time for chitchat
- Will recommend practical items over flashy ones
- Judges people by their equipment choices
- Respects warriors and rogues, slightly suspicious of magic users ("always setting things on fire")

FORMAT:
- Input: "[CharacterName]: message"
- Output: Your response as Grash (no prefix needed)
- Stay in character - you're not helpful customer service, you're a tired quartermaster`

// init the ai conversation
var conv = ai.NewConversation("llama3.1:8b", systemPrompt)

func Run() {
	// create the session
	discord, err := discordgo.New("Bot " + BotToken)
	if err != nil {
		log.Fatalf("Unable to start session: %v", err)
	}

	// add event handlers
	discord.AddHandler(newMessage)
	discord.AddHandler(interactionCreate)

	// Request necessary intents
	discord.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsGuilds

	// open the session
	err = discord.Open()
	if err != nil {
		log.Fatalf("Unable to open session: %v", err)
	}
	defer discord.Close()

	// Clean up any stale global commands (from before GuildID was set)
	if GuildID != "" {
		globalCmds, _ := discord.ApplicationCommands(discord.State.User.ID, "")
		for _, cmd := range globalCmds {
			slog.Info("removing stale global command", "name", cmd.Name)
			discord.ApplicationCommandDelete(discord.State.User.ID, "", cmd.ID)
		}
	}

	// Register slash commands
	registeredCommands := make([]*discordgo.ApplicationCommand, len(Commands))
	for i, cmd := range Commands {
		registered, err := discord.ApplicationCommandCreate(discord.State.User.ID, GuildID, cmd)
		if err != nil {
			log.Printf("Cannot create command '%s': %v", cmd.Name, err)
		} else {
			registeredCommands[i] = registered
			slog.Info("registered command", "name", cmd.Name)
		}
	}

	// keep the bot up until someone ctrl+c's it
	fmt.Println("Grash Ironledger is ready for business. *sighs* Press CTRL-C to close shop...")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	// Cleanup: remove commands on shutdown (optional, for development)
	if GuildID != "" {
		slog.Info("cleaning up guild commands...")
		for _, cmd := range registeredCommands {
			if cmd != nil {
				discord.ApplicationCommandDelete(discord.State.User.ID, GuildID, cmd.ID)
			}
		}
	}
}

// interactionCreate handles slash command interactions
func interactionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	if handler, ok := CommandHandlers[i.ApplicationCommandData().Name]; ok {
		handler(s, i)
	}
}

// newMessage handles regular chat messages (legacy support)
func newMessage(discord *discordgo.Session, messageEvent *discordgo.MessageCreate) {
	var respMessage string

	slog.Info("message received, scanning content")

	// Prevent the bot from responding to its own messages
	if messageEvent.Author.ID == discord.State.User.ID {
		return
	}

	// Respond to mentions of "grash" or "quartermaster" for legacy chat
	content := strings.ToLower(messageEvent.Content)
	if strings.Contains(content, "grash") || strings.Contains(content, "quartermaster") {
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
