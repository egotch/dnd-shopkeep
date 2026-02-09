package ai

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

type ChatMessage struct {
	Role	string	`json:"role"`
	Content	string	`json:"content"`
}

type OllamaRequest struct {
	Model	string	`json:"model"`
	Messages	[]ChatMessage	`json:"messages"`
	Stream	bool	`json:"stream"`
	Options	map[string]any	`json:"options,omitempty"`
}

type OllamaResponse struct {
	Message ChatMessage	`json:"message"`
}

// Conversation manages the entire conversation history
type Conversation struct {
	Messages []ChatMessage
	Model	string
}

// NewConversation creates a new conversation with optional system prompt
func NewConversation(model, systemPrompt string) *Conversation {
	conv := &Conversation{
		Messages: make([]ChatMessage, 0),
		Model: model,
	}

	// Add a system prompt if provided
	if systemPrompt != "" {
		conv.AddMessage("system", systemPrompt)
	}

	return conv
}

// AddMessage adds a message to the conversation history
func (c *Conversation) AddMessage(role, content string) {
	message := ChatMessage{
		Role: role,	
		Content: content,
	}
	c.Messages = append(c.Messages, message)
}

// GetStats returns conversation statistics for debugging
func (c *Conversation) GetStats() map[string]int {
	return map[string]int{
		"total_messages": len(c.Messages),
		"user_messages": c.countMessagesByRole("user"),
		"ai_messages": c.countMessagesByRole("assistant"),
	}
}

// countMessagesByRole counts the number of messages per given role
func (c *Conversation) countMessagesByRole(role string) int {
	count := 0

	for _, msg := range c.Messages {
		if msg.Role == role {
			count++
		}
	}
	return count
}

// SendtoOllama sends the ENTIRE conversation to Ollama and adds the response
func (c *Conversation) SendToOllama() (string, error) {
	request := OllamaRequest {
		Model: c.Model,
		Messages: c.Messages,
		Stream: false,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Post(
		"http://localhost:11434/api/chat",
		"application/json",
		bytes.NewBuffer(jsonData),
		)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var response OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", err
	}

	// IMPORTANT: Add the AI's rsponse to the converstion history
	c.AddMessage("assistant", response.Message.Content)

	return response.Message.Content, nil
}

// SendToOllamaWithTimeout sends the conversation with a configurable timeout and Ollama options
func (c *Conversation) SendToOllamaWithTimeout(timeout time.Duration, options map[string]any) (string, error) {
	request := OllamaRequest{
		Model:    c.Model,
		Messages: c.Messages,
		Stream:   false,
		Options:  options,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	client := &http.Client{Timeout: timeout}
	resp, err := client.Post(
		"http://localhost:11434/api/chat",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var response OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", err
	}

	c.AddMessage("assistant", response.Message.Content)

	return response.Message.Content, nil
}

func main() {

	// Init the conversation with a system prompt
	systemPrompt := "You are a helpful, but sassy assistant.  Remember the information the user tells you, and use it to inform future responses.  Have a helpful, but sarcastic attitude"
	conv :=  NewConversation("llama3.1:8b", systemPrompt)

	scanner := bufio.NewScanner(os.Stdin)

	// Demo sequence to show memory working
	fmt.Println("🎭 Demo Sequence:")
	fmt.Println("1. Tell the AI your name")
	fmt.Println("2. Ask what your name is")
	fmt.Println("3. Tell it your favorite color")
	fmt.Println("4. Ask it to remember both facts")
	fmt.Println("Type 'quit' to exit")

	// Main interaction loop
	for {
		fmt.Print("👤 You: ")
		scanner.Scan()
		userInput := scanner.Text()
		if strings.ToLower(userInput) == "quit" {
			break
		}
		if strings.ToLower(userInput) == "debug" {
			stats := conv.GetStats()
			fmt.Printf("🔍 Debug: %+v\n", stats)
			fmt.Printf("📝 Message history: \n")
			for i, msg := range conv.Messages {
				fmt.Printf("   %d. [%s]: %.50s...\n", i+1, msg.Role, msg.Content)
			}
			continue
		}
		// Add user message to the converstion
		conv.AddMessage("user", userInput)

		fmt.Print("🤖 AI: ")
		response, err := conv.SendToOllama()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}

		fmt.Println(response)

		// Show context info
		stats := conv.GetStats()
		fmt.Printf("💭 (Context: %d messages total)\n\n", stats["total_messages"])
	}

	fmt.Println("👋 Final conversation stats: ")
	stats := conv.GetStats()
	fmt.Printf("	User messages: %d\n", stats["user_messages"])
	fmt.Printf("	AI responses: %d\n", stats["ai_messages"])
	fmt.Printf("	total messages: %d\n", stats["total_messages"])
}

